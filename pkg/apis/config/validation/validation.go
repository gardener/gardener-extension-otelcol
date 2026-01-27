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
		cfg.Spec.Exporters.OTLPGRPCExporter.IsEnabled(),
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
			path:  "spec.exporters.otlp_http.endpoint",
			value: cfg.Spec.Exporters.OTLPHTTPExporter.Endpoint,
		},
		{
			path:  "spec.exporters.otlp_http.traces_endpoint",
			value: cfg.Spec.Exporters.OTLPHTTPExporter.TracesEndpoint,
		},
		{
			path:  "spec.exporters.otlp_http.metrics_endpoint",
			value: cfg.Spec.Exporters.OTLPHTTPExporter.MetricsEndpoint,
		},
		{
			path:  "spec.exporters.otlp_http.logs_endpoint",
			value: cfg.Spec.Exporters.OTLPHTTPExporter.LogsEndpoint,
		},
		{
			path:  "spec.exporters.otlp_http.profiles_endpoint",
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
	type nonNegativeField struct {
		path  string
		value int
	}

	nonNegativeFields := []nonNegativeField{
		{
			path:  "spec.exporters.otlp_http.read_buffer_size",
			value: cfg.Spec.Exporters.OTLPHTTPExporter.ReadBufferSize,
		},
		{
			path:  "spec.exporters.otlp_http.write_buffer_size",
			value: cfg.Spec.Exporters.OTLPHTTPExporter.WriteBufferSize,
		},
		{
			path:  "spec.exporters.otlp_grpc.read_buffer_size",
			value: cfg.Spec.Exporters.OTLPGRPCExporter.ReadBufferSize,
		},
		{
			path:  "spec.exporters.otlp_grpc.write_buffer_size",
			value: cfg.Spec.Exporters.OTLPGRPCExporter.WriteBufferSize,
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
	type resourceRef struct {
		path string
		ref  *config.ResourceReference
	}

	resourceRefs := []resourceRef{
		{
			path: "spec.exporters.otlp_http.token",
			ref:  cfg.Spec.Exporters.OTLPHTTPExporter.Token,
		},
		{
			path: "spec.exporters.otlp_grpc.token",
			ref:  cfg.Spec.Exporters.OTLPGRPCExporter.Token,
		},
	}

	// Referenced resources from the OTLP HTTP exporter
	if cfg.Spec.Exporters.OTLPHTTPExporter.TLS != nil {
		resourceRefs = append(
			resourceRefs,
			resourceRef{
				path: "spec.exporters.otlp_http.tls.ca",
				ref:  cfg.Spec.Exporters.OTLPHTTPExporter.TLS.CA,
			},
			resourceRef{
				path: "spec.exporters.otlp_http.tls.cert",
				ref:  cfg.Spec.Exporters.OTLPHTTPExporter.TLS.Cert,
			},
			resourceRef{
				path: "spec.exporters.otlp_http.tls.key",
				ref:  cfg.Spec.Exporters.OTLPHTTPExporter.TLS.Key,
			},
		)
	}

	// Referenced resources from the OTLP gRPC exporter
	if cfg.Spec.Exporters.OTLPGRPCExporter.TLS != nil {
		resourceRefs = append(
			resourceRefs,
			resourceRef{
				path: "spec.exporters.otlp_grpc.tls.ca",
				ref:  cfg.Spec.Exporters.OTLPGRPCExporter.TLS.CA,
			},
			resourceRef{
				path: "spec.exporters.otlp_grpc.tls.cert",
				ref:  cfg.Spec.Exporters.OTLPGRPCExporter.TLS.Cert,
			},
			resourceRef{
				path: "spec.exporters.otlp_grpc.tls.key",
				ref:  cfg.Spec.Exporters.OTLPGRPCExporter.TLS.Key,
			},
		)
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

	// Validate expected string values are not empty
	type nonEmptyString struct {
		path  string
		value string
	}

	nonEmptyStrings := make([]nonEmptyString, 0)
	if cfg.Spec.Exporters.OTLPGRPCExporter.IsEnabled() {
		nonEmptyStrings = append(
			nonEmptyStrings,
			nonEmptyString{
				path:  "spec.exporters.otlp_grpc.endpoint",
				value: cfg.Spec.Exporters.OTLPGRPCExporter.Endpoint,
			},
		)
	}

	for _, f := range nonEmptyStrings {
		if f.value == "" {
			allErrs = append(
				allErrs,
				field.Invalid(field.NewPath(f.path), f.path, "empty value specified"),
			)
		}
	}

	return allErrs.ToAggregate()
}
