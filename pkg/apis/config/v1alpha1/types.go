// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

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
//
// +k8s:enum
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
	// MetricsVerbosityLevelDetailed configures the collector with the most
	// verbose level, which includes dimensions and views.
	MetricsVerbosityLevelDetailed MetricsVerbosityLevel = "detailed"
)

// SignalType identifies a telemetry signal the collector can collect and
// export.
//
// +k8s:enum
type SignalType string

const (
	// SignalLogs is the signal for logs received via OTLP.
	SignalLogs SignalType = "logs"
	// SignalEvents is the signal for Kubernetes events collected from the
	// shoot cluster.
	SignalEvents SignalType = "events"
	// SignalMetrics is the signal for metrics scraped via Prometheus.
	SignalMetrics SignalType = "metrics"
)

// LogLevel specifies the minimum enabled logging level for the collector.
//
// See the link below for more details.
//
// https://opentelemetry.io/docs/collector/internal-telemetry/#configure-internal-logs
//
// +k8s:enum
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
//
// +k8s:enum
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
//
// +k8s:enum
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

// FilterErrorMode determines how the filter processor reacts to errors that
// occur while processing an OTTL condition.
//
// See the link below for more details.
//
// https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/filterprocessor#error-modes
//
// +k8s:enum
type FilterErrorMode string

const (
	// FilterErrorModeIgnore means the processor ignores errors returned by
	// conditions, logs them, and continues on to the next condition.
	FilterErrorModeIgnore FilterErrorMode = "ignore"
	// FilterErrorModeSilent means the processor ignores errors returned by
	// conditions, does not log them, and continues on to the next condition.
	FilterErrorModeSilent FilterErrorMode = "silent"
	// FilterErrorModePropagate means the processor returns the error up the
	// pipeline, which results in the payload being dropped from the collector.
	FilterErrorModePropagate FilterErrorMode = "propagate"
)

// FilterMatchType specifies the type of string pattern matching used by the filter
// processor include/exclude match properties.
//
// +k8s:enum
type FilterMatchType string

const (
	// MatchTypeStrict matches values by exact string equality.
	MatchTypeStrict FilterMatchType = "strict"
	// MatchTypeRegexp matches values against regular expressions.
	MatchTypeRegexp FilterMatchType = "regexp"
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

	// DefaultHTTPExporterClientTimeout specifies the default client timeout for
	// HTTP requests made by exporters.
	DefaultHTTPExporterClientTimeout = 30 * time.Second
	// DefaultHTTPExporterClientReadBufferSize specifies the default
	// ReadBufferSize for the HTTP client used by exporters.
	DefaultHTTPExporterClientReadBufferSize = 0
	// DefaultHTTPExporterClientWriteBufferSize specifies the default
	// WriteBufferSize for the HTTP client used by the exporters.
	DefaultHTTPExporterClientWriteBufferSize = 512 * 1024

	// DefaultGRPCExporterClientTimeout specifies the default client timeout
	// of the OTLP gRPC exporter.
	DefaultGRPCExporterClientTimeout = 5 * time.Second
	// DefaultGRPCExporterClientReadBufferSize specifies the default
	// ReadBufferSize for the gRPC client used by exporters.
	DefaultGRPCExporterClientReadBufferSize = 32 * 1024
	// DefaultGRPCExporterClientWriteBufferSize specifies the default
	// WriteBufferSize for the gRPC client used by the exporters.
	DefaultGRPCExporterClientWriteBufferSize = 32 * 1024

	// DefaultTLSReloadInterval specifies the default interval at which the
	// OTel Collector re-reads TLS material (CA, client cert, client key)
	// from disk. Without it, the collector loads the certs once at startup
	// and keeps using the in-memory copy even after the backing Secret is
	// rotated, leading to handshake failures with an expired client cert
	// until the pod is restarted.
	DefaultTLSReloadInterval = 30 * time.Second
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
	TLS *TLSConfig `json:"tls,omitzero"`

	// Token references a bearer token for authentication.
	//
	// +k8s:optional
	Token *ResourceReference `json:"token,omitempty"`

	// Timeout specifies the HTTP request time limit. Default value is
	// [DefaultHTTPExporterClientTimeout].
	//
	// +k8s:optional
	// +default=ref(DefaultHTTPExporterClientTimeout)
	Timeout time.Duration `json:"timeout,omitzero"`

	// ReadBufferSize specifies the ReadBufferSize for the HTTP
	// client. Default value is [DefaultHTTPExporterClientReadBufferSize].
	//
	// +k8s:optional
	// +default=ref(DefaultHTTPExporterClientReadBufferSize)
	ReadBufferSize int `json:"read_buffer_size,omitzero"`

	// WriteBufferSize specifies the WriteBufferSize for the HTTP
	// client. Default value is [DefaultHTTPExporterClientWriteBufferSize].
	//
	// +k8s:optional
	// +default=ref(DefaultHTTPExporterClientWriteBufferSize)
	WriteBufferSize int `json:"write_buffer_size,omitzero"`

	// Encoding specifies the encoding to use for the messages. The default
	// value is [MessageEncodingProto].
	//
	// +k8s:optional
	// +default=ref(MessageEncodingProto)
	Encoding MessageEncoding `json:"encoding,omitzero"`

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

// OTLPGRPCExporterConfig provides the OTLP gRPC Exporter config settings.
//
// See [OTLP gRPC Exporter] for more details.
//
// [OTLP gRPC Exporter]: https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlpexporter
type OTLPGRPCExporterConfig struct {
	// Enabled specifies whether the OTLP gRPC exporter is enabled or not.
	//
	// +k8s:optional
	// +default=false
	Enabled *bool `json:"enabled,omitzero"`

	// Endpoint specifies the gRPC endpoint to which signals will be exported.
	//
	// Check the link below for more details about the format of this field.
	//
	// https://github.com/grpc/grpc/blob/master/doc/naming.md
	//
	// +k8s:required
	Endpoint string `json:"endpoint,omitzero"`

	// TLS specifies the TLS configuration settings for the exporter.
	//
	// +k8s:optional
	TLS *TLSConfig `json:"tls,omitzero"`

	// Token references a bearer token for authentication.
	Token *ResourceReference `json:"token,omitzero"`

	// Timeout specifies the time to wait per individual attempt to send
	// data to the backend.
	//
	// +k8s:optional
	// +default=ref(DefaultGRPCExporterClientTimeout)
	Timeout time.Duration `json:"timeout,omitzero"`

	// ReadBufferSize specifies the ReadBufferSize for the gRPC
	// client. Default value is [DefaultGRPCExporterClientReadBufferSize].
	//
	// +k8s:optional
	// +default=ref(DefaultGRPCExporterClientReadBufferSize)
	ReadBufferSize int `json:"read_buffer_size,omitzero"`

	// WriteBufferSize specifies the WriteBufferSize for the gRPC
	// client. Default value is [DefaultGRPCExporterClientWriteBufferSize].
	//
	// +k8s:optional
	// +default=ref(DefaultGRPCExporterClientWriteBufferSize)
	WriteBufferSize int `json:"write_buffer_size,omitzero"`

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

// CollectorExportersConfig provides the OTLP exporter settings.
type CollectorExportersConfig struct {
	// OTLPGRPCExporter provides the OTLP gRPC Exporter settings.
	//
	// +k8s:optional
	OTLPGRPCExporter OTLPGRPCExporterConfig `json:"otlp_grpc,omitzero"`

	// HTTPExporter provides the OTLP HTTP Exporter settings.
	//
	// +k8s:optional
	OTLPHTTPExporter OTLPHTTPExporterConfig `json:"otlp_http,omitzero"`

	// DebugExporter provides the settings for the debug exporter.
	//
	// +k8s:optional
	DebugExporter DebugExporterConfig `json:"debug,omitzero"`
}

// CollectorLogsConfig provides the settings for the collector internal logs.
//
// See [Configure internal logs] for more details.
//
// [Configure internal logs]: https://opentelemetry.io/docs/collector/internal-telemetry/#configure-internal-logs
type CollectorLogsConfig struct {
	// Level specifies the log level of the collector.
	//
	// +k8s:optional
	// +default=ref(LogLevelInfo)
	Level LogLevel `json:"level,omitzero"`

	// Encoding specifies the encoding for logs of the collector.
	//
	// +k8s:optional
	// +default=ref(LogEncodingConsole)
	Encoding LogEncoding `json:"encoding,omitzero"`
}

// CollectorMetricsConfig provides the settings for the collector internal
// metrics.
//
// See [Metrics verbosity] for more details.
//
// [Metrics verbosity]: https://opentelemetry.io/docs/collector/internal-telemetry/#metric-verbosity
type CollectorMetricsConfig struct {
	// Level specifies the collector internal metrics verbosity level.
	//
	// +k8s:optional
	// +default=ref(MetricsVerbosityLevelNormal)
	Level MetricsVerbosityLevel `json:"level,omitzero"`
}

// FilterAttribute specifies an attribute key/value pair that the filter
// processor match properties evaluate against.
type FilterAttribute struct {
	// Key specifies the attribute key to match against.
	//
	// +k8s:required
	Key string `json:"key"`

	// Value specifies the attribute value to match against. If empty, only the
	// presence of the key is checked.
	//
	// +k8s:optional
	Value string `json:"value,omitzero"`
}

// RegexpConfig specifies the options for the regexp match type used by the
// filter processor include/exclude match properties.
type RegexpConfig struct {
	// CacheEnabled specifies whether match results are cached.
	//
	// +k8s:optional
	CacheEnabled *bool `json:"cacheenabled,omitzero"`

	// CacheMaxNumEntries specifies the maximum number of entries in the cache.
	//
	// +k8s:optional
	CacheMaxNumEntries int `json:"cachemaxnumentries,omitzero"`
}

// LogSeverityNumberMatchProperties specifies how the filter processor matches
// against a log record's severity number.
type LogSeverityNumberMatchProperties struct {
	// Min specifies the minimum severity a log record must have to match. This
	// corresponds to the severity short names, e.g. "INFO", "WARN", "ERROR".
	// The value is case-insensitive.
	//
	// +k8s:required
	Min string `json:"min"`

	// MatchUndefined specifies whether log records with an "unspecified"
	// severity match.
	//
	// +k8s:optional
	MatchUndefined *bool `json:"match_undefined,omitzero"`
}

// MetricMatchProperties specifies the set of properties the filter processor
// matches metrics against, and the type of string pattern matching to use.
type MetricMatchProperties struct {
	// MatchType specifies the type of matching desired.
	//
	// +k8s:required
	MatchType FilterMatchType `json:"match_type"`

	// Regexp specifies the options for the regexp match type.
	//
	// +k8s:optional
	Regexp *RegexpConfig `json:"regexp,omitzero"`

	// MetricNames specifies the list of string patterns to match metric names
	// against. A match occurs if the metric name matches at least one pattern
	// in this list.
	//
	// +k8s:optional
	MetricNames []string `json:"metric_names,omitempty"`

	// Expressions specifies the list of expr expressions to match metrics
	// against. A match occurs if any datapoint in a metric matches at least one
	// expression in this list.
	//
	// +k8s:optional
	Expressions []string `json:"expressions,omitempty"`

	// ResourceAttributes specifies a list of resource attributes to match
	// metrics against. A match occurs if any resource attribute matches all
	// expressions in this list.
	//
	// +k8s:optional
	ResourceAttributes []FilterAttribute `json:"resource_attributes,omitempty"`
}

// LogMatchProperties specifies the set of properties the filter processor
// matches logs against, and the type of string pattern matching to use.
type LogMatchProperties struct {
	// MatchType specifies the type of matching desired.
	//
	// +k8s:required
	MatchType FilterMatchType `json:"match_type"`

	// ResourceAttributes specifies a list of resource attributes to match logs
	// against. A match occurs if any resource attribute matches all expressions
	// in this list.
	//
	// +k8s:optional
	ResourceAttributes []FilterAttribute `json:"resource_attributes,omitempty"`

	// RecordAttributes specifies a list of record attributes to match logs
	// against. A match occurs if any record attribute matches at least one
	// expression in this list.
	//
	// +k8s:optional
	RecordAttributes []FilterAttribute `json:"record_attributes,omitempty"`

	// SeverityTexts specifies a list of strings that the log record's severity
	// text field must match against.
	//
	// +k8s:optional
	SeverityTexts []string `json:"severity_texts,omitempty"`

	// SeverityNumber specifies how to match against a log record's severity
	// number, if defined.
	//
	// +k8s:optional
	SeverityNumber *LogSeverityNumberMatchProperties `json:"severity_number,omitzero"`

	// Bodies specifies a list of strings that the log record's body field must
	// match against.
	//
	// +k8s:optional
	Bodies []string `json:"bodies,omitempty"`
}

// MetricFilters specifies the filter processor settings for the metrics signal.
//
// The OTTL condition lists (Resource, Metric, DataPoint) and the object-based
// match properties (Include, Exclude) are mutually exclusive.
type MetricFilters struct {
	// Include specifies the metrics that should be included in the collector
	// service pipeline; all other metrics are dropped. If both Include and
	// Exclude are specified, Include filtering occurs first.
	//
	// +k8s:optional
	Include *MetricMatchProperties `json:"include,omitzero"`

	// Exclude specifies the metrics that should be excluded from the collector
	// service pipeline; all other metrics are included. If both Include and
	// Exclude are specified, Include filtering occurs first.
	//
	// +k8s:optional
	Exclude *MetricMatchProperties `json:"exclude,omitzero"`

	// Resource is a list of OTTL conditions for an ottlresource context. If any
	// condition resolves to true, the whole resource is dropped.
	//
	// +k8s:optional
	Resource []string `json:"resource,omitempty"`

	// Metric is a list of OTTL conditions for an ottlmetric context. If any
	// condition resolves to true, the metric is dropped.
	//
	// +k8s:optional
	Metric []string `json:"metric,omitempty"`

	// DataPoint is a list of OTTL conditions for an ottldatapoint context. If
	// any condition resolves to true, the datapoint is dropped.
	//
	// +k8s:optional
	DataPoint []string `json:"datapoint,omitempty"`
}

// LogFilters specifies the filter processor settings for the logs signal.
//
// The OTTL condition lists (Resource, LogRecord) and the object-based match
// properties (Include, Exclude) are mutually exclusive.
type LogFilters struct {
	// Include specifies the logs that should be included in the collector
	// service pipeline; all other logs are dropped. If both Include and Exclude
	// are specified, Include filtering occurs first.
	//
	// +k8s:optional
	Include *LogMatchProperties `json:"include,omitzero"`

	// Exclude specifies the logs that should be excluded from the collector
	// service pipeline; all other logs are included. If both Include and
	// Exclude are specified, Include filtering occurs first.
	//
	// +k8s:optional
	Exclude *LogMatchProperties `json:"exclude,omitzero"`

	// Resource is a list of OTTL conditions for an ottlresource context. If any
	// condition resolves to true, the whole resource is dropped. Supports `and`,
	// `or`, and `()`.
	//
	// +k8s:optional
	Resource []string `json:"resource,omitempty"`

	// LogRecord is a list of OTTL conditions for an ottllog context. If any
	// condition resolves to true, the log record is dropped. Supports `and`,
	// `or`, and `()`.
	//
	// +k8s:optional
	LogRecord []string `json:"log_record,omitempty"`
}

// ContextConditions specifies a group of OTTL conditions for the filter
// processor's context-inferred condition style (metric_conditions /
// log_conditions).
//
// When Context and ErrorMode are both empty, the group is rendered as a flat
// list of condition strings (basic style) and the OTTL context is inferred
// from each expression. Otherwise it is rendered as an explicit group
// (advanced style).
type ContextConditions struct {
	// Context specifies the OTTL context the conditions are evaluated against,
	// e.g. "resource", "metric", "datapoint", "log". If empty, the context is
	// inferred from each condition.
	//
	// +k8s:optional
	Context string `json:"context,omitzero"`

	// Conditions is the list of OTTL conditions. If any condition resolves to
	// true, the matching telemetry is dropped.
	//
	// +k8s:required
	Conditions []string `json:"conditions"`

	// ErrorMode determines how the processor reacts to errors that occur while
	// processing this group of conditions. When set, it overrides the
	// top-level ErrorMode. Only honored when the group is rendered in the
	// advanced style.
	//
	// +k8s:optional
	ErrorMode FilterErrorMode `json:"error_mode,omitzero"`
}

// FilterConfig specifies the settings for the filter processor, which drops
// metrics and logs matching the given OTTL conditions or match properties.
//
// See [Filter Processor] for more details.
//
// [Filter Processor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/filterprocessor
type FilterConfig struct {
	// ErrorMode determines how the processor reacts to errors that occur while
	// processing an OTTL condition. If empty, the processor default is used.
	//
	// +k8s:optional
	ErrorMode FilterErrorMode `json:"error_mode,omitzero"`

	// Metrics specifies the filter settings for the metrics signal.
	//
	// +k8s:optional
	Metrics *MetricFilters `json:"metrics,omitzero"`

	// Logs specifies the filter settings for the logs signal. Because events
	// are collected as the logs signal, this also applies to the events
	// pipeline.
	//
	// +k8s:optional
	Logs *LogFilters `json:"logs,omitzero"`

	// MetricConditions specifies the metrics filter using the context-inferred
	// condition style. It is mutually exclusive with Metrics.
	//
	// +k8s:optional
	MetricConditions []ContextConditions `json:"metric_conditions,omitempty"`

	// LogConditions specifies the logs filter using the context-inferred
	// condition style. It is mutually exclusive with Logs.
	//
	// +k8s:optional
	LogConditions []ContextConditions `json:"log_conditions,omitempty"`
}

// CollectorConfigSpec specifies the desired state of [CollectorConfig]
type CollectorConfigSpec struct {
	// Exporters specifies the exporters configuration of the collector.
	//
	// +k8s:required
	Exporters CollectorExportersConfig `json:"exporters,omitzero"`

	// Logs specifies the settings for the collector logs.
	//
	// +k8s:optional
	Logs CollectorLogsConfig `json:"logs,omitzero"`

	// Metrics specifies the settings for the internal collector metrics.
	//
	// +k8s:optional
	Metrics CollectorMetricsConfig `json:"metrics,omitzero"`

	// Signals lists the telemetry signals the collector should collect and
	// export. Valid values are "logs", "events" and "metrics". If empty, all
	// signals are enabled.
	//
	// When a signal is omitted, its pipeline is not created. The corresponding
	// receiver is still defined, which the collector reports as an unused
	// component in its logs; this is expected and harmless.
	//
	// +k8s:optional
	Signals []SignalType `json:"signals,omitempty"`

	// Filter specifies the filter processor settings used to drop unwanted signals.
	//
	// +k8s:optional
	Filter *FilterConfig `json:"filter,omitzero"`
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
	// ReloadInterval specifies mTLS key and cert reload interval
	// from mounted secret volume
	//
	// +k8s:optional
	// +default=ref(DefaultTLSReloadInterval)
	ReloadInterval time.Duration `json:"reloadInterval,omitzero"`
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
