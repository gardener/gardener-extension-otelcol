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

// ExporterProtocol selects the OTLP transport used by an exporter.
type ExporterProtocol string

const (
	// ExporterProtocolHTTP selects the OTLP HTTP exporter.
	ExporterProtocolHTTP ExporterProtocol = "http"
	// ExporterProtocolGRPC selects the OTLP gRPC exporter.
	ExporterProtocolGRPC ExporterProtocol = "grpc"
	// ExporterProtocolDebug selects the debug exporter, which writes telemetry
	// to the collector's own logs. It only honors the Verbosity field; the
	// endpoint, TLS, token and buffer settings are ignored.
	ExporterProtocolDebug ExporterProtocol = "debug"
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

// ExporterConfig provides a full exporter configuration.
//
// It folds the OTLP HTTP, OTLP gRPC and debug exporters into a single type,
// selected by Protocol. Each signal target carries its own ExporterConfig.
//
// When Protocol is [ExporterProtocolDebug] only Verbosity is honored; the
// endpoint, TLS, token and buffer settings are ignored.
//
// See [OTLP HTTP Exporter], [OTLP gRPC Exporter] and [Debug Exporter] for more
// details.
//
// [OTLP HTTP Exporter]: https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlphttpexporter
// [OTLP gRPC Exporter]: https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlpexporter
// [Debug Exporter]: https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/debugexporter
type ExporterConfig struct {
	// Protocol selects the transport used by the exporter.
	Protocol ExporterProtocol

	// Endpoint specifies the target endpoint to send data to. For HTTP this is
	// a base URL; for gRPC this is a gRPC endpoint. It is ignored when Protocol
	// is [ExporterProtocolDebug].
	Endpoint string

	// TLS specifies the TLS configuration settings for the exporter. It is
	// ignored when Protocol is [ExporterProtocolDebug].
	TLS *TLSConfig

	// Token references a bearer token for authentication. It is ignored when
	// Protocol is [ExporterProtocolDebug].
	Token *ResourceReference

	// Timeout specifies the request time limit.
	Timeout time.Duration

	// ReadBufferSize specifies the ReadBufferSize for the client.
	ReadBufferSize int

	// WriteBufferSize specifies the WriteBufferSize for the client.
	WriteBufferSize int

	// Encoding specifies the encoding to use for the messages. It is only
	// honored by the HTTP protocol.
	Encoding MessageEncoding

	// RetryOnFailure specifies the retry policy of the exporter.
	RetryOnFailure RetryOnFailureConfig

	// Compression specifies the compression to use.
	Compression Compression

	// Verbosity specifies the verbosity level of the debug exporter. It is only
	// honored when Protocol is [ExporterProtocolDebug].
	Verbosity DebugExporterVerbosity
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

// FilterRule specifies a single filter processor instance, which drops metrics
// and logs matching the given OTTL conditions or match properties.
//
// See [Filter Processor] for more details.
//
// [Filter Processor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/filterprocessor
type FilterRule struct {
	// ErrorMode determines how the processor reacts to errors that occur while
	// processing an OTTL condition. If empty, the processor default is used.
	ErrorMode FilterErrorMode

	// Metrics specifies the filter settings for the metrics signal.
	Metrics *MetricFilters

	// Logs specifies the filter settings for the logs and events signal.
	Logs *LogFilters

	// MetricConditions specifies the metrics filter using the context-inferred
	// condition style. It is mutually exclusive with Metrics.
	MetricConditions []ContextConditions

	// LogConditions specifies the logs filter using the context-inferred
	// condition style. It is mutually exclusive with Logs.
	LogConditions []ContextConditions

	// Conditions specifies a signal-agnostic filter using the context-inferred
	// condition style.
	Conditions []ContextConditions
}

// SignalTarget pairs an exporter with its own ordered list of filter rules.
// Each target produces one collector service pipeline for the signal, so a
// signal can fan out to multiple destinations, each with its own filtering.
type SignalTarget struct {
	// Exporter is the exporter this target sends to.
	Exporter ExporterConfig

	// Filters is an ordered list of filter rules applied to this target's
	// pipeline.
	Filters []FilterRule
}

// SignalConfig configures a single telemetry signal end to end.
type SignalConfig struct {
	// Enabled turns the signal's pipelines on or off.
	Enabled *bool

	// Targets is the list of exporter/filters pairs for this signal. Each
	// target becomes its own collector service pipeline.
	Targets []SignalTarget
}

// IsEnabled is a predicate which returns whether the signal's pipelines should
// be created.
func (s SignalConfig) IsEnabled() bool {
	if s.Enabled != nil {
		return *s.Enabled
	}

	return false
}

// SignalsConfig groups the per-signal configuration sections.
type SignalsConfig struct {
	// Metrics configures the metrics signal, which is scraped via Prometheus.
	Metrics SignalConfig

	// Logs configures the logs signal, which is received via OTLP.
	Logs SignalConfig

	// Traces configures the traces signal, which is received via OTLP.
	Traces SignalConfig

	// Profiles configures the profiles signal, which is received via OTLP.
	Profiles SignalConfig

	// Events configures the Kubernetes events signal, which is collected from
	// the shoot cluster.
	Events SignalConfig
}

// SignalType identifies a telemetry signal. It is used internally to iterate
// the per-signal configuration sections in a stable order; it is not part of
// the serialized configuration.
type SignalType string

const (
	// SignalMetrics is the metrics signal, scraped via Prometheus.
	SignalMetrics SignalType = "metrics"
	// SignalLogs is the logs signal, received via OTLP.
	SignalLogs SignalType = "logs"
	// SignalTraces is the traces signal, received via OTLP.
	SignalTraces SignalType = "traces"
	// SignalProfiles is the profiles signal, received via OTLP.
	SignalProfiles SignalType = "profiles"
	// SignalEvents is the Kubernetes events signal, collected from the shoot.
	SignalEvents SignalType = "events"
)

// CollectorConfigSpec specifies the desired state of [CollectorConfig]
type CollectorConfigSpec struct {
	// Signals groups the per-signal configuration sections.
	Signals SignalsConfig

	// Logs specifies the settings for the collector's internal logs.
	Logs CollectorLogsConfig

	// Metrics specifies the settings for the internal collector metrics.
	Metrics CollectorMetricsConfig
}

// Signal returns the [SignalConfig] for the given signal type.
func (s SignalsConfig) Signal(sig SignalType) SignalConfig {
	switch sig {
	case SignalMetrics:
		return s.Metrics
	case SignalLogs:
		return s.Logs
	case SignalTraces:
		return s.Traces
	case SignalProfiles:
		return s.Profiles
	case SignalEvents:
		return s.Events
	default:
		return SignalConfig{}
	}
}

// AllSignals lists the signal types in a stable iteration order.
func AllSignals() []SignalType {
	return []SignalType{SignalMetrics, SignalLogs, SignalTraces, SignalProfiles, SignalEvents}
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
