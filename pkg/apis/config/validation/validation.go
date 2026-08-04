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

			// Each filter rule may only use the fields valid for this signal.
			for j, rule := range target.Filters {
				allErrs = append(allErrs, validateFilterRule(rule, sig, targetPath.Child("filters").Index(j))...)
			}

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

// validateFilterRule ensures a filter rule only uses fields valid for the given
// signal. The metrics-only fields are valid on the metrics signal, the
// logs-only fields on the logs and events signals, and the traces and profiles
// signals accept only the signal-agnostic Conditions form. The Resource field
// is valid on metrics, logs and events. ErrorMode is valid everywhere.
func validateFilterRule(rule config.FilterRule, sig config.SignalType, path *field.Path) field.ErrorList {
	allErrs := make(field.ErrorList, 0)

	metricUsed := len(rule.Metric) > 0 || len(rule.DataPoint) > 0 ||
		rule.MetricInclude != nil || rule.MetricExclude != nil || len(rule.MetricConditions) > 0
	logUsed := len(rule.LogRecord) > 0 ||
		rule.LogInclude != nil || rule.LogExclude != nil || len(rule.LogConditions) > 0
	resourceUsed := len(rule.Resource) > 0
	conditionsUsed := len(rule.Conditions) > 0

	reject := func(used bool, child, detail string) {
		if used {
			allErrs = append(allErrs, field.Invalid(path.Child(child), "", detail))
		}
	}

	switch sig {
	case config.SignalMetrics:
		reject(logUsed, "log_record", "log filter fields are not valid on the metrics signal")
		reject(conditionsUsed, "conditions", "the signal-agnostic conditions form is not valid on the metrics signal; use metric/datapoint/metric_conditions")
	case config.SignalLogs, config.SignalEvents:
		reject(metricUsed, "metric", "metric filter fields are not valid on the logs signal")
		reject(conditionsUsed, "conditions", "the signal-agnostic conditions form is not valid on the logs signal; use log_record/log_conditions")
	case config.SignalTraces, config.SignalProfiles:
		reject(metricUsed, "metric", "metric filter fields are not valid on this signal; use conditions")
		reject(logUsed, "log_record", "log filter fields are not valid on this signal; use conditions")
		reject(resourceUsed, "resource", "the resource field is not valid on this signal; use conditions")
	default:
		// Unknown signal; no per-signal field restrictions to apply.
	}

	return allErrs
}
