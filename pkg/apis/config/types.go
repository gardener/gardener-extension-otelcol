// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MetricsVerbosityLevel specifies the verbosity of the internal collector
// metrics.
//
// See the link below for more details.
//
// https://opentelemetry.io/docs/collector/internal-telemetry/#metric-verbosity
type MetricsVerbosityLevel string

const (
	// MetricsVerbosityLevelNone disables the internal collector metrics.
	MetricsVerbosityLevelNone MetricsVerbosityLevel = "none"
	// MetricsVerbosityLevelBasic configures the collector to emit basic
	// metrics only.
	MetricsVerbosityLevelBasic MetricsVerbosityLevel = "basic"
	// MetricsVerbosityLevelNormal configures the collector with standard
	// indicators on top of the basic ones.
	MetricsVerbosityLevelNormal MetricsVerbosityLevel = "normal"
	// MetricsVerbosityDetailed configures the collector with the most
	// verbose level, which includes dimensions and views.
	MetricsVerbosityLevelDetailed MetricsVerbosityLevel = "detailed"
)

// LogLevel specifies the minimum enabled logging level for the collector.
//
// See the link below for more details.
//
// https://opentelemetry.io/docs/collector/internal-telemetry/#configure-internal-logs
type LogLevel string

const (
	// LogLevelInfo sets the collector's internal logger to INFO level.
	LogLevelInfo LogLevel = "INFO"
	// LogLevelWarn sets the collector's internal logger to WARN level.
	LogLevelWarn LogLevel = "WARN"
	// LogLevelError sets the collector's internal logger to ERROR level.
	LogLevelError LogLevel = "ERROR"
	// LogLevelDebug sets the collector's internal logger to DEBUG level.
	LogLevelDebug LogLevel = "DEBUG"
)

// LogEncoding specifies the encoding for the internal collector logger.
//
// See the link below for more details.
//
// https://opentelemetry.io/docs/collector/internal-telemetry/#configure-internal-logs
type LogEncoding string

const (
	// LogEncodingConsole sets the collector's internal logger with console
	// encoding.
	LogEncodingConsole LogEncoding = "console"
	// LogEncodingJSON sets the collector's internal logger with JSON
	// encoding.
	LogEncodingJSON LogEncoding = "json"
)

// MessageEncoding specifies the encoding used by the collector exporters.
type MessageEncoding string

const (
	// MessageEncodingProto specifies that proto encoding is used for
	// messages.
	MessageEncodingProto MessageEncoding = "proto"
	// MessageEncodingJSON specifies that JSON is used for encoding
	// messages.
	MessageEncodingJSON MessageEncoding = "json"
)

// Compression specifies the compression used by the collector.
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

// RetryOnFailureConfig provides the retry policy for an exporter.
type RetryOnFailureConfig struct {
	// Enabled specifies whether retry on failure is enabled or not.
	Enabled *bool

	// InitialInterval specifies the time to wait after the first failure
	// before retrying.
	InitialInterval time.Duration

	// MaxInterval specifies the upper bound on backoff.
	MaxInterval time.Duration

	// MaxElapsedTime specifies the maximum amount of time spent trying to
	// send a batch. If set to 0, the retries are never stopped.
	MaxElapsedTime time.Duration

	// Multiplier specifies the factor by which the retry interval is
	// multiplied on each attempt.
	Multiplier float64
}

// OTLPHTTPExporterConfig provides the OTLP HTTP Exporter configuration settings.
//
// See [OTLP HTTP Exporter] for more details.
//
// [OTLP HTTP Exporter]: https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlphttpexporter
type OTLPHTTPExporterConfig struct {
	// Enabled specifies whether the OTLP HTTP exporter is enabled or not.
	Enabled *bool

	// Endpoint specifies the target base URL to send data to, e.g. https://example.com:4318
	//
	// To send each signal a corresponding path will be added to this base
	// URL, i.e. for traces "/v1/traces" will appended, for metrics
	// "/v1/metrics" will be appended, for logs "/v1/logs" will be appended.
	Endpoint string

	// TracesEndpoint specifies the target URL to send trace data to, e.g. https://example.com:4318/v1/traces.
	//
	// When this setting is present the base endpoint setting is ignored for
	// traces.
	TracesEndpoint string

	// MetricsEndpoint specifies the target URL to send metric data to, e.g. https://example.com:4318/v1/metrics.
	//
	// When this setting is present the base endpoint setting is ignored for
	// metrics.
	MetricsEndpoint string

	// LogsEndpoint specifies the target URL to send log data to, e.g. https://example.com:4318/v1/logs
	//
	// When this setting is present the base endpoint setting is ignored for
	// logs.
	LogsEndpoint string

	// ProfilesEndpoint specifies the target URL to send profile data to, e.g. https://example.com:4318/v1development/profiles.
	//
	// When this setting is present the endpoint setting is ignored for
	// profile data.
	ProfilesEndpoint string

	// TLS specifies the TLS configuration settings for the exporter.
	TLS *TLSConfig

	// Token references a bearer token for authentication.
	Token *ResourceReference

	// Timeout specifies the HTTP request time limit.
	Timeout time.Duration

	// ReadBufferSize specifies the ReadBufferSize for the HTTP
	// client.
	ReadBufferSize int

	// WriteBufferSize specifies the WriteBufferSize for the HTTP
	// client.
	WriteBufferSize int

	// Encoding specifies the encoding to use for the messages. Valid
	// options are `proto' and `json'.
	Encoding MessageEncoding

	// RetryOnFailure specifies the retry policy of the exporter.
	RetryOnFailure RetryOnFailureConfig

	// Compression specifies the compression to use.
	//
	// Possible options are gzip, zstd, snappy and none.
	Compression Compression
}

// IsEnabled is a predicate which returns whether the exporter is enabled or
// not.
func (cfg *OTLPHTTPExporterConfig) IsEnabled() bool {
	if cfg.Enabled != nil {
		return *cfg.Enabled
	}

	return false
}

// DebugExporterVerbosity specifies the verbosity level for the debug exporter.
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
	Enabled *bool

	// Verbosity specifies the verbosity level for the debug exporter.
	Verbosity DebugExporterVerbosity
}

// IsEnabled is a predicate which returns whether the exporter is enabled or
// not.
func (cfg *DebugExporterConfig) IsEnabled() bool {
	if cfg.Enabled != nil {
		return *cfg.Enabled
	}

	return false
}

// CollectorExportersConfig provides the OTLP exporter settings.
type CollectorExportersConfig struct {
	// HTTPExporter provides the OTLP HTTP Exporter settings.
	OTLPHTTPExporter OTLPHTTPExporterConfig

	// DebugExporter provides the settings for the debug exporter.
	DebugExporter DebugExporterConfig
}

// CollectorLogsConfig provides the settings for the collector internal logs.
//
// See [Configure internal logs] for more details.
//
// [Configure internal logs]: https://opentelemetry.io/docs/collector/internal-telemetry/#configure-internal-logs
type CollectorLogsConfig struct {
	// Level specifies the log level of the collector.
	Level LogLevel

	// Encoding specifies the encoding for logs of the collector.
	Encoding LogEncoding
}

// CollectorMetricsConfig provides the settings for the collector internal
// metrics.
//
// See [Metrics verbosity] for more details.
//
// [Metrics verbosity]: https://opentelemetry.io/docs/collector/internal-telemetry/#metric-verbosity
type CollectorMetricsConfig struct {
	// Level specifies the collector internal metrics verbosity level.
	Level MetricsVerbosityLevel
}

// CollectorConfigSpec specifies the desired state of [CollectorConfig]
type CollectorConfigSpec struct {
	// Exporters specifies the exporters configuration of the collector.
	Exporters CollectorExportersConfig

	// Logs specifies the settings for the collector logs.
	Logs CollectorLogsConfig

	// Metrics specifies the settings for the internal collector metrics.
	Metrics CollectorMetricsConfig
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CollectorConfig provides the OpenTelemetry Collector API configuration.
type CollectorConfig struct {
	metav1.TypeMeta

	// Spec provides the extension configuration spec.
	Spec CollectorConfigSpec
}

// TLSConfig provides the TLS settings used by exporters.
type TLSConfig struct {
	// InsecureSkipVerify specifies whether to skip verifying the
	// certificate or not.
	InsecureSkipVerify *bool
	// CA references the CA certificate to use for verifying the server certificate.
	// For a client this verifies the server certificate.
	// For a server this verifies client certificates.
	// If empty uses system root CA.
	CA *ResourceReference
	// Cert references the client certificate to use for TLS required connections.
	Cert *ResourceReference
	// Key references the client key to use for TLS required connections.
	Key *ResourceReference
}

// ResourceReference references data from a Secret.
type ResourceReference struct {
	// ResourceRef references a resource in the shoot.
	ResourceRef ResourceReferenceDetails
}

// ResourceReferenceDetails references a resource (e.g., a Secret) in the garden cluster.
type ResourceReferenceDetails struct {
	// Name is the name of the resource e reference in `.spec.resources` in
	// the Shoot resource.
	Name string
	// DataKey is the key in the resource data map.
	DataKey string
}
