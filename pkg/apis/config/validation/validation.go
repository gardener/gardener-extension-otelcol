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

	// At least one signal must be enabled.
	anyEnabled := false
	for _, sig := range config.AllSignals() {
		signal := cfg.Spec.Signals.Signal(sig)
		if !signal.IsEnabled() {
			continue
		}
		anyEnabled = true

		signalPath := specPath.Child("signals", string(sig))

		// An enabled signal must define at least one target.
		if len(signal.Targets) == 0 {
			allErrs = append(
				allErrs,
				field.Required(signalPath.Child("targets"), "an enabled signal must define at least one target"),
			)

			continue
		}

		for i, target := range signal.Targets {
			targetPath := signalPath.Child("targets").Index(i)

			// A debug target writes to the collector's own logs; it does not
			// use an endpoint, TLS or a token, so those checks are skipped.
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
	}

	if !anyEnabled {
		allErrs = append(allErrs, field.Required(specPath.Child("signals"), "no signal enabled"))
	}

	return allErrs.ToAggregate()
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
