// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"fmt"
	"net/url"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor"
	"go.opentelemetry.io/collector/confmap"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/gardener/gardener-extension-otelcol/pkg/apis/config"
	"github.com/gardener/gardener-extension-otelcol/pkg/filterrender"
)

// Validate validates the given [config.CollectorConfig]
func Validate(cfg config.CollectorConfig) error {
	allErrs := make(field.ErrorList, 0)

	specPath := field.NewPath("spec")
	targetsPath := specPath.Child("targets")

	// At least one target must be defined.
	if len(cfg.Spec.Targets) == 0 {
		allErrs = append(allErrs, field.Required(targetsPath, "at least one target must be defined"))

		return allErrs.ToAggregate()
	}

	// anyEnabled tracks whether the targets collectively enable at least one
	// signal.
	anyEnabled := false

	for i, target := range cfg.Spec.Targets {
		targetPath := targetsPath.Index(i)

		enabledSignals := validateSignals(target.Signals, targetPath.Child("signals"), &allErrs)
		if len(enabledSignals) > 0 {
			anyEnabled = true
		}

		// Validate the target's opaque filter configuration, if any.
		allErrs = append(allErrs, validateTargetFilter(target, targetPath.Child("filters"))...)

		// Validate the target's exporters. A target must enable at least one
		// transport; each enabled OTLP transport is validated (endpoint, buffers,
		// TLS/token). The debug exporter needs no endpoint/TLS/token.
		allErrs = append(allErrs, validateExporters(target.Exporter, targetPath.Child("exporter"))...)
	}

	if !anyEnabled {
		allErrs = append(allErrs, field.Required(targetsPath, "no signal enabled by any target"))
	}

	return allErrs.ToAggregate()
}

// validateSignals validates a target's signal list: every entry must be a known
// signal type and must not be duplicated. It returns the set of valid signals
// found. The map does double duty: an entry only exists for known signals, and
// its value counts how many times we have already seen that signal.
func validateSignals(signals []config.SignalType, path *field.Path, allErrs *field.ErrorList) map[config.SignalType]bool {
	if len(signals) == 0 {
		*allErrs = append(*allErrs, field.Required(path, "a target must serve at least one signal"))

		return nil
	}

	seenSignals := map[config.SignalType]int{
		config.SignalLogs:    0,
		config.SignalEvents:  0,
		config.SignalMetrics: 0,
	}
	enabled := map[config.SignalType]bool{}

	for i, s := range signals {
		cnt, ok := seenSignals[s]
		if !ok {
			*allErrs = append(
				*allErrs,
				field.NotSupported(path.Index(i), string(s), []string{
					string(config.SignalLogs),
					string(config.SignalEvents),
					string(config.SignalMetrics),
				}),
			)

			continue
		}
		if cnt >= 1 {
			*allErrs = append(*allErrs, field.Duplicate(path.Index(i), string(s)))
		}
		seenSignals[s]++
		enabled[s] = true
	}

	return enabled
}

// validateExporters validates a target's per-transport exporters. At least one
// transport must be enabled. Each enabled OTLP transport must specify an
// endpoint and have valid buffer sizes and resource references; the debug
// exporter needs none of those.
func validateExporters(exp config.CollectorExportersConfig, path *field.Path) field.ErrorList {
	allErrs := make(field.ErrorList, 0)

	if exp.OTLPHTTPExporter == nil && exp.OTLPGRPCExporter == nil && exp.DebugExporter == nil {
		allErrs = append(allErrs, field.Required(path, "at least one exporter transport must be enabled"))

		return allErrs
	}

	if e := exp.OTLPHTTPExporter; e != nil {
		allErrs = append(allErrs, validateOTLPExporter(e.Endpoint, e.ReadBufferSize, e.WriteBufferSize, e.Token, e.TLS, path.Child("otlp_http"))...)
	}
	if e := exp.OTLPGRPCExporter; e != nil {
		allErrs = append(allErrs, validateOTLPExporter(e.Endpoint, e.ReadBufferSize, e.WriteBufferSize, e.Token, e.TLS, path.Child("otlp_grpc"))...)
	}

	return allErrs
}

// validateOTLPExporter validates the common fields of an OTLP (HTTP or gRPC)
// exporter: a required endpoint, non-negative buffer sizes, and valid token/TLS
// resource references.
func validateOTLPExporter(endpoint string, readBufferSize, writeBufferSize int, token *config.ResourceReference, tls *config.TLSConfig, path *field.Path) field.ErrorList {
	allErrs := make(field.ErrorList, 0)

	if endpoint == "" {
		allErrs = append(allErrs, field.Required(path.Child("endpoint"), "no endpoint specified for the exporter"))
	} else if _, err := url.Parse(endpoint); err != nil {
		allErrs = append(allErrs, field.Invalid(path.Child("endpoint"), endpoint, "invalid URL specified"))
	}

	if readBufferSize < 0 {
		allErrs = append(allErrs, field.Invalid(path.Child("read_buffer_size"), readBufferSize, "value cannot be negative"))
	}
	if writeBufferSize < 0 {
		allErrs = append(allErrs, field.Invalid(path.Child("write_buffer_size"), writeBufferSize, "value cannot be negative"))
	}

	allErrs = append(allErrs, validateResourceReference(token, path.Child("token"))...)
	if tls != nil {
		allErrs = append(allErrs, validateResourceReference(tls.CA, path.Child("tls", "ca"))...)
		allErrs = append(allErrs, validateResourceReference(tls.Cert, path.Child("tls", "cert"))...)
		allErrs = append(allErrs, validateResourceReference(tls.Key, path.Child("tls", "key"))...)
	}

	return allErrs
}

// validateResourceReference ensures a resource reference has a non-empty name
// and data key. A nil reference is valid (the field is optional).
func validateResourceReference(ref *config.ResourceReference, path *field.Path) field.ErrorList {
	if ref == nil {
		return nil
	}
	if ref.ResourceRef.Name == "" || ref.ResourceRef.DataKey == "" {
		return field.ErrorList{
			field.Invalid(path, fmt.Sprintf("%+v", ref.ResourceRef), "name or dataKey is empty"),
		}
	}

	return nil
}

// validateTargetFilter validates a target's opaque filter configuration. The
// body is decoded and round-tripped through a factory-created
// [filterprocessor.Config] (parse + Validate), which checks OTTL condition
// syntax, match properties and severities, and rejects unknown keys. This
// delegates the filter semantics to the upstream processor, so new upstream
// validation is inherited on a dependency version bump.
//
// The factory base is required: it populates the OTTL function maps the config
// needs to parse and validate conditions. A zero-value Config would not validate
// conditions correctly.
func validateTargetFilter(target config.Target, path *field.Path) field.ErrorList {
	rendered, err := filterrender.FilterProcessorConfig(target)
	if err != nil {
		return field.ErrorList{field.Invalid(path, "", err.Error())}
	}
	if len(rendered) == 0 {
		return nil
	}

	base := filterprocessor.NewFactory().CreateDefaultConfig()
	conf := confmap.NewFromStringMap(rendered)
	if err := conf.Unmarshal(base); err != nil {
		return field.ErrorList{field.Invalid(path, "", err.Error())}
	}
	if err := base.(*filterprocessor.Config).Validate(); err != nil {
		return field.ErrorList{field.Invalid(path, "", err.Error())}
	}

	return nil
}
