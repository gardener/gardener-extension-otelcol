// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"cmp"
	"net/url"

	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/gardener/gardener-extension-otelcol/pkg/apis/config"
)

// Validate validates the given [config.CollectorConfig]
func Validate(cfg config.CollectorConfig) error {
	allErrs := make(field.ErrorList, 0)

	// We require at least one exporter to be enabled
	anyExporterEnabled := []bool{
		cfg.Spec.Exporters.DebugExporter.IsEnabled(),
		cfg.Spec.Exporters.OTLPHTTPExporter.IsEnabled(),
	}

	if !cmp.Or(anyExporterEnabled...) {
		allErrs = append(
			allErrs,
			field.Required(field.NewPath("spec.exporters"), "no exporter enabled"),
		)
	}

	// Validate URL fields
	urlFields := []struct {
		path  string
		value string
	}{
		{
			path:  "spec.exporters.otlphttp.endpoint",
			value: cfg.Spec.Exporters.OTLPHTTPExporter.Endpoint,
		},
		{
			path:  "spec.exporters.otlphttp.traces_endpoint",
			value: cfg.Spec.Exporters.OTLPHTTPExporter.TracesEndpoint,
		},
		{
			path:  "spec.exporters.otlphttp.metrics_endpoint",
			value: cfg.Spec.Exporters.OTLPHTTPExporter.MetricsEndpoint,
		},
		{
			path:  "spec.exporters.otlphttp.logs_endpoint",
			value: cfg.Spec.Exporters.OTLPHTTPExporter.LogsEndpoint,
		},
		{
			path:  "spec.exporters.otlphttp.profiles_endpoint",
			value: cfg.Spec.Exporters.OTLPHTTPExporter.ProfilesEndpoint,
		},
	}

	for _, f := range urlFields {
		if f.value != "" {
			if _, err := url.Parse(f.value); err != nil {
				allErrs = append(
					allErrs,
					field.Invalid(field.NewPath(f.path), f.value, "invalid URL specified"),
				)
			}
		}
	}

	// Make sure that the HTTP client read/write buffers are good
	nonNegativeFields := []struct {
		path  string
		value int
	}{
		{
			path:  "spec.exporters.otlphttp.read_buffer_size",
			value: cfg.Spec.Exporters.OTLPHTTPExporter.ReadBufferSize,
		},
		{
			path:  "spec.exporters.otlphttp.write_buffer_size",
			value: cfg.Spec.Exporters.OTLPHTTPExporter.WriteBufferSize,
		},
	}

	for _, f := range nonNegativeFields {
		if f.value < 0 {
			allErrs = append(
				allErrs,
				field.Invalid(field.NewPath(f.path), f.value, "value cannot be negative"),
			)
		}
	}

	// Validate resource references
	resourceRefs := []struct {
		path string
		ref  *config.ResourceReference
	}{
		{
			path: "spec.exporters.otlphttp.token",
			ref:  cfg.Spec.Exporters.OTLPHTTPExporter.Token,
		},
		{
			path: "spec.exporters.otlphttp.tls.ca",
			ref:  cfg.Spec.Exporters.OTLPHTTPExporter.TLS.CA,
		},
		{
			path: "spec.exporters.otlphttp.tls.cert",
			ref:  cfg.Spec.Exporters.OTLPHTTPExporter.TLS.Cert,
		},
		{
			path: "spec.exporters.otlphttp.tls.key",
			ref:  cfg.Spec.Exporters.OTLPHTTPExporter.TLS.Key,
		},
	}

	for _, f := range resourceRefs {
		if f.ref != nil {
			if f.ref.ResourceRef.Name == "" || f.ref.ResourceRef.DataKey == "" {
				allErrs = append(
					allErrs,
					field.Invalid(field.NewPath(f.path), f.path, "name or dataKey is empty"),
				)
			}
		}
	}

	return allErrs.ToAggregate()
}
