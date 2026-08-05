// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"fmt"
	"net/url"

	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/gardener/gardener-extension-otelcol/pkg/apis/config"
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

		// Each filter rule may only set a block for a signal this target serves.
		for j, rule := range target.Filters {
			allErrs = append(allErrs, validateFilterRule(rule, enabledSignals, targetPath.Child("filters").Index(j))...)
		}

		// A debug target writes to the collector's own logs; it does not use an
		// endpoint, TLS or a token, so those checks are skipped.
		if target.Exporter.Protocol == config.ExporterProtocolDebug {
			continue
		}

		allErrs = append(allErrs, validateExporter(target.Exporter, targetPath.Child("exporter"))...)

		// A non-debug target must specify an endpoint.
		if target.Exporter.Endpoint == "" {
			allErrs = append(
				allErrs,
				field.Required(targetPath.Child("exporter", "endpoint"), "no endpoint specified for the target"),
			)
		}
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

// validateExporter validates the fields of a single exporter configuration.
func validateExporter(exp config.ExporterConfig, path *field.Path) field.ErrorList {
	allErrs := make(field.ErrorList, 0)

	if exp.Endpoint != "" {
		if _, err := url.Parse(exp.Endpoint); err != nil {
			allErrs = append(allErrs, field.Invalid(path.Child("endpoint"), exp.Endpoint, "invalid URL specified"))
		}
	}

	if exp.ReadBufferSize < 0 {
		allErrs = append(allErrs, field.Invalid(path.Child("read_buffer_size"), exp.ReadBufferSize, "value cannot be negative"))
	}
	if exp.WriteBufferSize < 0 {
		allErrs = append(allErrs, field.Invalid(path.Child("write_buffer_size"), exp.WriteBufferSize, "value cannot be negative"))
	}

	allErrs = append(allErrs, validateResourceReference(exp.Token, path.Child("token"))...)
	if exp.TLS != nil {
		allErrs = append(allErrs, validateResourceReference(exp.TLS.CA, path.Child("tls", "ca"))...)
		allErrs = append(allErrs, validateResourceReference(exp.TLS.Cert, path.Child("tls", "cert"))...)
		allErrs = append(allErrs, validateResourceReference(exp.TLS.Key, path.Child("tls", "key"))...)
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

// validateFilterRule ensures a filter rule only sets a block for a signal the
// enclosing target serves. The Metrics block requires the metrics signal; the
// Logs block requires the logs or events signal.
func validateFilterRule(rule config.FilterRule, enabledSignals map[config.SignalType]bool, path *field.Path) field.ErrorList {
	allErrs := make(field.ErrorList, 0)

	if rule.Metrics != nil && !enabledSignals[config.SignalMetrics] {
		allErrs = append(allErrs, field.Invalid(
			path.Child("metrics"), "",
			"the metrics filter block requires the target to serve the metrics signal",
		))
	}

	if rule.Logs != nil && !enabledSignals[config.SignalLogs] && !enabledSignals[config.SignalEvents] {
		allErrs = append(allErrs, field.Invalid(
			path.Child("logs"), "",
			"the logs filter block requires the target to serve the logs or events signal",
		))
	}

	return allErrs
}
