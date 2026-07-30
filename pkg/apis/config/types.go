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
	// MetricsVerbosityLevelDetailed configures the collector with the most
	// verbose level, which includes dimensions and views.
	MetricsVerbosityLevelDetailed MetricsVerbosityLevel = "detailed"
)

// SignalType identifies a telemetry signal the collector can collect and
// export.
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

// FilterErrorMode determines how the filter processor reacts to errors that
// occur while processing an OTTL condition.
//
// See the link below for more details.
//
// https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/filterprocessor#error-modes
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

// MatchType specifies the type of string pattern matching used by the filter
// processor include/exclude match properties.
type MatchType string

const (
	// MatchTypeStrict matches values by exact string equality.
	MatchTypeStrict MatchType = "strict"
	// MatchTypeRegexp matches values against regular expressions.
	MatchTypeRegexp MatchType = "regexp"
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
func (cfg OTLPHTTPExporterConfig) IsEnabled() bool {
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
func (cfg DebugExporterConfig) IsEnabled() bool {
	if cfg.Enabled != nil {
		return *cfg.Enabled
	}

	return false
}

// OTLPGRPCExporterConfig provides the OTLP gRPC Exporter config settings.
//
// See [OTLP gRPC Exporter] for more details.
//
// [OTLP gRPC Exporter]: https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlpexporter
type OTLPGRPCExporterConfig struct {
	// Enabled specifies whether the OTLP gRPC exporter is enabled or not.
	Enabled *bool

	// Endpoint specifies the gRPC endpoint to which signals will be exported.
	//
	// Check the link below for more details about the format of this field.
	//
	// https://github.com/grpc/grpc/blob/master/doc/naming.md
	Endpoint string

	// TLS specifies the TLS configuration settings for the exporter.
	TLS *TLSConfig

	// Token references a bearer token for authentication.
	Token *ResourceReference

	// Timeout specifies the time to wait per individual attempt to send
	// data to the backend.
	Timeout time.Duration

	// ReadBufferSize specifies the ReadBufferSize for the gRPC
	// client. Default value is [DefaultGRPCExporterClientReadBufferSize].
	ReadBufferSize int

	// WriteBufferSize specifies the WriteBufferSize for the gRPC
	// client. Default value is [DefaultGRPCExporterClientWriteBufferSize].
	WriteBufferSize int

	// RetryOnFailure specifies the retry policy of the exporter.
	RetryOnFailure RetryOnFailureConfig

	// Compression specifies the compression to use. The default value is
	// [CompressionGzip].
	Compression Compression
}

// IsEnabled is a predicate which returns whether the exporter is enabled or
// not.
func (cfg OTLPGRPCExporterConfig) IsEnabled() bool {
	if cfg.Enabled != nil {
		return *cfg.Enabled
	}

	return false
}

// CollectorExportersConfig provides the OTLP exporter settings.
type CollectorExportersConfig struct {
	// OTLPGRPCExporter provides the OTLP gRPC Exporter settings.
	OTLPGRPCExporter OTLPGRPCExporterConfig

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

// FilterAttribute specifies an attribute key/value pair that the filter
// processor match properties evaluate against.
type FilterAttribute struct {
	// Key specifies the attribute key to match against.
	Key string

	// Value specifies the attribute value to match against. If empty, only the
	// presence of the key is checked.
	Value string
}

// RegexpConfig specifies the options for the regexp match type used by the
// filter processor include/exclude match properties.
type RegexpConfig struct {
	// CacheEnabled specifies whether match results are cached.
	CacheEnabled *bool

	// CacheMaxNumEntries specifies the maximum number of entries in the cache.
	CacheMaxNumEntries int
}

// LogSeverityNumberMatchProperties specifies how the filter processor matches
// against a log record's severity number.
type LogSeverityNumberMatchProperties struct {
	// Min specifies the minimum severity a log record must have to match.
	Min string

	// MatchUndefined specifies whether log records with an "unspecified"
	// severity match.
	MatchUndefined *bool
}

// MetricMatchProperties specifies the set of properties the filter processor
// matches metrics against, and the type of string pattern matching to use.
type MetricMatchProperties struct {
	// MatchType specifies the type of matching desired.
	MatchType MatchType

	// Regexp specifies the options for the regexp match type.
	Regexp *RegexpConfig

	// MetricNames specifies the list of string patterns to match metric names
	// against.
	MetricNames []string

	// Expressions specifies the list of expr expressions to match metrics
	// against.
	Expressions []string

	// ResourceAttributes specifies a list of resource attributes to match
	// metrics against.
	ResourceAttributes []FilterAttribute
}

// LogMatchProperties specifies the set of properties the filter processor
// matches logs against, and the type of string pattern matching to use.
type LogMatchProperties struct {
	// MatchType specifies the type of matching desired.
	MatchType MatchType

	// ResourceAttributes specifies a list of resource attributes to match logs
	// against.
	ResourceAttributes []FilterAttribute

	// RecordAttributes specifies a list of record attributes to match logs
	// against.
	RecordAttributes []FilterAttribute

	// SeverityTexts specifies a list of strings that the log record's severity
	// text field must match against.
	SeverityTexts []string

	// SeverityNumber specifies how to match against a log record's severity
	// number, if defined.
	SeverityNumber *LogSeverityNumberMatchProperties

	// Bodies specifies a list of strings that the log record's body field must
	// match against.
	Bodies []string
}

// MetricFilters specifies the filter processor settings for the metrics signal.
//
// The OTTL condition lists (Resource, Metric, DataPoint) and the object-based
// match properties (Include, Exclude) are mutually exclusive.
type MetricFilters struct {
	// Include specifies the metrics that should be included in the collector
	// service pipeline; all other metrics are dropped.
	Include *MetricMatchProperties

	// Exclude specifies the metrics that should be excluded from the collector
	// service pipeline; all other metrics are included.
	Exclude *MetricMatchProperties

	// Resource is a list of OTTL conditions for an ottlresource context.
	Resource []string

	// Metric is a list of OTTL conditions for an ottlmetric context.
	Metric []string

	// DataPoint is a list of OTTL conditions for an ottldatapoint context.
	DataPoint []string
}

// LogFilters specifies the filter processor settings for the logs signal.
//
// The OTTL condition lists (Resource, LogRecord) and the object-based match
// properties (Include, Exclude) are mutually exclusive.
type LogFilters struct {
	// Include specifies the logs that should be included in the collector
	// service pipeline; all other logs are dropped.
	Include *LogMatchProperties

	// Exclude specifies the logs that should be excluded from the collector
	// service pipeline; all other logs are included.
	Exclude *LogMatchProperties

	// Resource is a list of OTTL conditions for an ottlresource context.
	Resource []string

	// LogRecord is a list of OTTL conditions for an ottllog context.
	LogRecord []string
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
	// Context specifies the OTTL context the conditions are evaluated against.
	// If empty, the context is inferred from each condition.
	Context string

	// Conditions is the list of OTTL conditions.
	Conditions []string

	// ErrorMode overrides the top-level ErrorMode for this group of conditions.
	ErrorMode FilterErrorMode
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
	ErrorMode FilterErrorMode

	// Metrics specifies the filter settings for the metrics signal.
	Metrics *MetricFilters

	// Logs specifies the filter settings for the logs signal.
	Logs *LogFilters

	// MetricConditions specifies the metrics filter using the context-inferred
	// condition style. It is mutually exclusive with Metrics.
	MetricConditions []ContextConditions

	// LogConditions specifies the logs filter using the context-inferred
	// condition style. It is mutually exclusive with Logs.
	LogConditions []ContextConditions
}

// CollectorConfigSpec specifies the desired state of [CollectorConfig]
type CollectorConfigSpec struct {
	// Exporters specifies the exporters configuration of the collector.
	Exporters CollectorExportersConfig

	// Logs specifies the settings for the collector logs.
	Logs CollectorLogsConfig

	// Metrics specifies the settings for the internal collector metrics.
	Metrics CollectorMetricsConfig

	// Signals lists the telemetry signals the collector should collect and
	// export. Valid values are "logs", "events" and "metrics". If empty, all
	// signals are enabled.
	Signals []SignalType

	// Filter specifies the filter processor settings used to drop unwanted
	// metrics and logs.
	Filter *FilterConfig
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
	// ReloadInterval specifies the duration after which the certificate will be reloaded.
	// If not set, it will never be reloaded
	ReloadInterval time.Duration
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
