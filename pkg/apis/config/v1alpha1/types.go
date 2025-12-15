// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Encoding specifies the encoding used by the collector exporters.
//
// +k8s:enum
type Encoding string

const (
	// EncodingProto specifies that proto encoding is used for messages.
	EncodingProto Encoding = "proto"
	// EncodingJSON specifies that JSON is used for encoding messages.
	EncodingJSON Encoding = "json"
)

// Compression specifies the compression used by the collector.
//
// +k8s:enum
type Compression string

const (
	// CompressionGzip specifies that gzip compression is used.
	CompressionGzip Compression = "gzip"
	// CompressionZstd specifies that zstd compression is used.
	CompressionZstd Compression = "zstd"
	// CompressionSnappy specifies that snappy compression is used.
	CompressionSnappy Compression = "snappy"
	// CompressionNone specifies that no compression is used.
	CompressionNone Compression = "none"
)

const (
	// DefaultRetryInitialInterval specifies the default initial interval to
	// wait after the first failure, before attempting a retry.
	DefaultRetryInitialInterval = 5 * time.Second
	// DefaultRetryMaxInterval specifies the default upper bound on backoff.
	DefaultRetryMaxInterval = 30 * time.Second
	// DefaultRetryMaxElapsedTime specifies the default maximum amount of
	// time spent trying to send a batch.
	DefaultRetryMaxElapsedTime = 300 * time.Second
	// DefaultRetryMultiplier specifies the default factor by which the
	// retry interval is multiplied on each attempt.
	DefaultRetryMultiplier = 1.5

	// DefaultExporterClientTimeout specifies the default client timeout for
	// HTTP requests made by exporters.
	DefaultExporterClientTimeout = 30 * time.Second
	// DefaultExporterClientReadBufferSize specifies the default
	// ReadBufferSize for the HTTP client used by exporters.
	DefaultExporterClientReadBufferSize = 0
	// DefaultExporterClientWriteBufferSize specifies the default
	// WriteBufferSize for the HTTP client used by the exporters.
	DefaultExporterClientWriteBufferSize = 512 * 1024
)

// RetryOnFailureConfig provides the retry policy for an exporter.
type RetryOnFailureConfig struct {
	// Enabled specifies whether retry on failure is enabled or not. Default
	// is true.
	//
	// +k8s:optional
	// +default=true
	Enabled *bool `json:"enabled,omitzero"`

	// InitialInterval specifies the time to wait after the first failure
	// before retrying. The default value is [DefaultRetryInitialInterval].
	//
	// +k8s:optional
	// +default=ref(DefaultRetryInitialInterval)
	InitialInterval time.Duration `json:"initial_interval,omitzero"`

	// MaxInterval specifies the upper bound on backoff. Default value is
	// [DefaultRetryMaxInterval].
	//
	// +k8s:optional
	// +default=ref(DefaultRetryMaxInterval)
	MaxInterval time.Duration `json:"max_interval,omitzero"`

	// MaxElapsedTime specifies the maximum amount of time spent trying to
	// send a batch. If set to 0, the retries are never stopped. The default
	// value is [DefaultRetryMaxElapsedTime].
	//
	// +k8s:optional
	// +default=ref(DefaultRetryMaxElapsedTime)
	MaxElapsedTime time.Duration `json:"max_elapsed_time,omitzero"`

	// Multiplier specifies the factor by which the retry interval is
	// multiplied on each attempt. The default value is
	// [DefaultRetryMultiplier].
	//
	// +k8s:optional
	// +default=ref(DefaultRetryMultiplier)
	Multiplier float64 `json:"multiplier,omitzero"`
}

// OTLPHTTPExporterConfig provides the OTLP HTTP Exporter configuration settings.
//
// See [OTLP HTTP Exporter] for more details.
//
// [OTLP HTTP Exporter]: https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlphttpexporter
type OTLPHTTPExporterConfig struct {
	// Enabled specifies whether the OTLP HTTP exporter is enabled or not.
	//
	// +k8s:optional
	// +default=false
	Enabled *bool `json:"enabled,omitzero"`

	// Endpoint specifies the target base URL to send data to, e.g. https://example.com:4318
	//
	// To send each signal a corresponding path will be added to this base
	// URL, i.e. for traces "/v1/traces" will appended, for metrics
	// "/v1/metrics" will be appended, for logs "/v1/logs" will be appended.
	//
	// +k8s:optional
	Endpoint string `json:"endpoint,omitzero"`

	// TracesEndpoint specifies the target URL to send trace data to, e.g. https://example.com:4318/v1/traces.
	//
	// When this setting is present the base endpoint setting is ignored for
	// traces.
	//
	// +k8s:optional
	TracesEndpoint string `json:"traces_endpoint,omitzero"`

	// MetricsEndpoint specifies the target URL to send metric data to, e.g. https://example.com:4318/v1/metrics.
	//
	// When this setting is present the base endpoint setting is ignored for
	// metrics.
	//
	// +k8s:optional
	MetricsEndpoint string `json:"metrics_endpoint,omitzero"`

	// LogsEndpoint specifies the target URL to send log data to, e.g. https://example.com:4318/v1/logs
	//
	// When this setting is present the base endpoint setting is ignored for
	// logs.
	//
	// +k8s:optional
	LogsEndpoint string `json:"logs_endpoint,omitzero"`

	// ProfilesEndpoint specifies the target URL to send profile data to, e.g. https://example.com:4318/v1development/profiles.
	//
	// When this setting is present the endpoint setting is ignored for
	// profile data.
	//
	// +k8s:optional
	ProfilesEndpoint string `json:"profiles_endpoint,omitzero"`

	// TLS specifies the TLS configuration settings for the exporter.
	//
	// +k8s:optional
	TLS *TLSConfig `json:"tls,omitempty"`
	// Token references a bearer token for authentication.
	//
	// +k8s:optional
	Token *ResourceReference `json:"token,omitempty"`

	// Timeout specifies the HTTP request time limit. Default value is
	// [DefaultExporterClientTimeout].
	//
	// +k8s:optional
	// +default=ref(DefaultExporterClientTimeout)
	Timeout time.Duration `json:"timeout,omitzero"`

	// ReadBufferSize specifies the ReadBufferSize for the HTTP
	// client. Default value is [DefaultExporterClientReadBufferSize].
	//
	// +k8s:optional
	// +default=ref(DefaultExporterClientReadBufferSize)
	ReadBufferSize int `json:"read_buffer_size,omitzero"`

	// WriteBufferSize specifies the WriteBufferSize for the HTTP
	// client. Default value is [DefaultExporterClientWriteBufferSize].
	//
	// +k8s:optional
	// +default=ref(DefaultExporterClientWriteBufferSize)
	WriteBufferSize int `json:"write_buffer_size,omitzero"`

	// Encoding specifies the encoding to use for the messages. The default
	// value is [EncodingProto].
	//
	// +k8s:optional
	// +default=ref(EncodingProto)
	Encoding Encoding `json:"encoding,omitzero"`

	// RetryOnFailure specifies the retry policy of the exporter.
	//
	// +k8s:optional
	RetryOnFailure RetryOnFailureConfig `json:"retry_on_failure,omitzero"`

	// Compression specifies the compression to use. The default value is
	// [CompressionGzip].
	//
	// +k8s:optional
	// +default=ref(CompressionGzip)
	Compression Compression `json:"compression,omitzero"`
}

// DebugExporterVerbosity specifies the verbosity level for the debug exporter.
//
// +k8s:enum
type DebugExporterVerbosity string

const (
	// DebugExporterVerbosityBasic specifies basic level of verbosity.
	DebugExporterVerbosityBasic DebugExporterVerbosity = "basic"
	// DebugExporterVerbosityNormal specifies normal level of verbosity.
	DebugExporterVerbosityNormal DebugExporterVerbosity = "normal"
	// DebugExporterVerbosityDetailed specifies detailed level of verbosity.
	DebugExporterVerbosityDetailed DebugExporterVerbosity = "detailed"
)

// DebugExporterConfig provides the settings for the debug exporter
type DebugExporterConfig struct {
	// Enabled specifies whether the debug exporter is enabled or not.
	//
	// +k8s:optional
	// +default=false
	Enabled *bool `json:"enabled,omitzero"`

	// Verbosity specifies the verbosity level for the debug exporter.
	//
	// +k8s:optional
	// +default=ref(DebugExporterVerbosityBasic)
	Verbosity DebugExporterVerbosity `json:"verbosity,omitzero"`
}

// CollectorExportersConfig provides the OTLP exporter settings.
type CollectorExportersConfig struct {
	// HTTPExporter provides the OTLP HTTP Exporter settings.
	//
	// +k8s:optional
	OTLPHTTPExporter OTLPHTTPExporterConfig `json:"otlphttp,omitzero"`

	// DebugExporter provides the settings for the debug exporter.
	//
	// +k8s:optional
	DebugExporter DebugExporterConfig `json:"debug,omitzero"`
}

// CollectorConfigSpec specifies the desired state of [CollectorConfig]
type CollectorConfigSpec struct {
	// Exporters specify exporters configuration of the collector.
	//
	// +k8s:required
	Exporters CollectorExportersConfig `json:"exporters,omitzero"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CollectorConfig provides the OpenTelemetry Collector API configuration.
type CollectorConfig struct {
	metav1.TypeMeta `json:",inline"`

	// Spec provides the extension configuration spec.
	Spec CollectorConfigSpec `json:"spec,omitzero"`
}

// TLSConfig provides the TLS settings used by exporters.
type TLSConfig struct {
	// InsecureSkipVerify specifies whether to skip verifying the
	// certificate or not.
	// +k8s:optional
	// +default=false
	InsecureSkipVerify *bool `json:"insecureSkipVerify,omitempty"`
	// CA references the CA certificate to use for verifying the server certificate.
	// For a client this verifies the server certificate.
	// For a server this verifies client certificates.
	// If empty uses system root CA.
	//
	// +k8s:optional
	CA *ResourceReference `json:"ca,omitempty"`
	// Cert references the client certificate to use for TLS required connections.
	//
	// +k8s:optional
	Cert *ResourceReference `json:"cert,omitempty"`
	// Key references the client key to use for TLS required connections.
	//
	// +k8s:optional
	Key *ResourceReference `json:"key,omitempty"`
}

// ResourceReference references data from a Secret.
type ResourceReference struct {
	// ResourceRef references a resource in the shoot.
	//
	// +k8s:required
	ResourceRef ResourceReferenceDetails `json:"resourceRef"`
}

// ResourceReferenceDetails references a resource (e.g., a Secret) in the garden cluster.
type ResourceReferenceDetails struct {
	// Name is the name of thresource e reference in `.spec.resources` in the Shoot resource.
	//
	// +k8s:required
	Name string `json:"name"`
	// DataKey is the key in the resource data map.
	//
	// +k8s:required
	DataKey string `json:"dataKey"`
}
