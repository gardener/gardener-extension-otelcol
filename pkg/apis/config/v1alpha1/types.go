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

// ExporterProtocol selects the OTLP transport used by an exporter.
//
// +k8s:enum
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
	// Protocol selects the transport used by the exporter. The default value is
	// [ExporterProtocolHTTP]. Set it to [ExporterProtocolDebug] to write
	// telemetry to the collector's own logs instead of sending it to a remote
	// endpoint.
	//
	// +k8s:optional
	// +default=ref(ExporterProtocolHTTP)
	Protocol ExporterProtocol `json:"protocol,omitzero"`

	// Endpoint specifies the target endpoint to send data to. It is required
	// unless Protocol is [ExporterProtocolDebug].
	//
	// For the HTTP protocol this is a base URL, e.g. https://example.com:4318;
	// the collector appends the per-signal path (e.g. "/v1/metrics"). For the
	// gRPC protocol this is a gRPC endpoint, see
	// https://github.com/grpc/grpc/blob/master/doc/naming.md.
	//
	// +k8s:optional
	Endpoint string `json:"endpoint,omitzero"`

	// TLS specifies the TLS configuration settings for the exporter.
	//
	// +k8s:optional
	TLS *TLSConfig `json:"tls,omitzero"`

	// Token references a bearer token for authentication.
	//
	// +k8s:optional
	Token *ResourceReference `json:"token,omitzero"`

	// Timeout specifies the request time limit. Default value is
	// [DefaultHTTPExporterClientTimeout].
	//
	// +k8s:optional
	// +default=ref(DefaultHTTPExporterClientTimeout)
	Timeout time.Duration `json:"timeout,omitzero"`

	// ReadBufferSize specifies the ReadBufferSize for the client. Default value
	// is [DefaultHTTPExporterClientReadBufferSize].
	//
	// +k8s:optional
	// +default=ref(DefaultHTTPExporterClientReadBufferSize)
	ReadBufferSize int `json:"read_buffer_size,omitzero"`

	// WriteBufferSize specifies the WriteBufferSize for the client. Default
	// value is [DefaultHTTPExporterClientWriteBufferSize].
	//
	// +k8s:optional
	// +default=ref(DefaultHTTPExporterClientWriteBufferSize)
	WriteBufferSize int `json:"write_buffer_size,omitzero"`

	// Encoding specifies the encoding to use for the messages. It is only
	// honored by the HTTP protocol. The default value is [MessageEncodingProto].
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

	// Verbosity specifies the verbosity level of the debug exporter. It is only
	// honored when Protocol is [ExporterProtocolDebug]. The default value is
	// [DebugExporterVerbosityBasic].
	//
	// +k8s:optional
	// +default=ref(DebugExporterVerbosityBasic)
	Verbosity DebugExporterVerbosity `json:"verbosity,omitzero"`
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

// FilterMetrics specifies the metrics filterprocessor block. It mirrors the
// "metrics" section of the OTel filter processor and is valid on targets that
// serve the metrics signal.
type FilterMetrics struct {
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

	// Include specifies the metrics that should be kept in the pipeline; all
	// other metrics are dropped. If both Include and Exclude are specified,
	// include filtering occurs first.
	//
	// +k8s:optional
	Include *MetricMatchProperties `json:"include,omitzero"`

	// Exclude specifies the metrics that should be dropped from the pipeline;
	// all other metrics are kept.
	//
	// +k8s:optional
	Exclude *MetricMatchProperties `json:"exclude,omitzero"`

	// MetricConditions specifies the metrics filter using the context-inferred
	// condition style.
	//
	// +k8s:optional
	MetricConditions []ContextConditions `json:"metric_conditions,omitempty"`
}

// FilterLogs specifies the logs filterprocessor block. It mirrors the "logs"
// section of the OTel filter processor and is valid on targets that serve the
// logs or events signals.
type FilterLogs struct {
	// Resource is a list of OTTL conditions for an ottlresource context. If any
	// condition resolves to true, the whole resource is dropped.
	//
	// +k8s:optional
	Resource []string `json:"resource,omitempty"`

	// LogRecord is a list of OTTL conditions for an ottllog context. If any
	// condition resolves to true, the log record is dropped.
	//
	// +k8s:optional
	LogRecord []string `json:"log_record,omitempty"`

	// Include specifies the logs that should be kept in the pipeline; all other
	// logs are dropped. If both Include and Exclude are specified, include
	// filtering occurs first.
	//
	// +k8s:optional
	Include *LogMatchProperties `json:"include,omitzero"`

	// Exclude specifies the logs that should be dropped from the pipeline; all
	// other logs are kept.
	//
	// +k8s:optional
	Exclude *LogMatchProperties `json:"exclude,omitzero"`

	// LogConditions specifies the logs filter using the context-inferred
	// condition style.
	//
	// +k8s:optional
	LogConditions []ContextConditions `json:"log_conditions,omitempty"`
}

// FilterRule specifies a single filter processor instance, which drops
// telemetry matching the given OTTL conditions or match properties. A target's
// Filters list produces one filter processor per rule per signal, applied in
// order.
//
// Its blocks mirror the OTel filterprocessor, keyed by signal: the Metrics
// block feeds a target's metrics pipeline, the Logs block its logs and events
// pipelines. A block may only be set for a signal the enclosing target serves.
//
// See [Filter Processor] for more details.
//
// [Filter Processor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/filterprocessor
type FilterRule struct {
	// ErrorMode determines how the processor reacts to errors that occur while
	// processing an OTTL condition. If empty, the processor default is used.
	//
	// +k8s:optional
	ErrorMode FilterErrorMode `json:"error_mode,omitzero"`

	// Metrics specifies the metrics filterprocessor block. Valid only on targets
	// that serve the metrics signal.
	//
	// +k8s:optional
	Metrics *FilterMetrics `json:"metrics,omitzero"`

	// Logs specifies the logs filterprocessor block. Valid on targets that serve
	// the logs or events signals.
	//
	// +k8s:optional
	Logs *FilterLogs `json:"logs,omitzero"`
}

// Target pairs a self-contained exporter with the signals it receives and its
// own ordered list of filter rules. Each (target, signal) pair produces one
// collector service pipeline, so a target can fan out several signals to one
// destination, each with independent filtering.
type Target struct {
	// Exporter is the exporter this target sends to. Set its Protocol to
	// [ExporterProtocolDebug] to make this target a debug destination.
	//
	// +k8s:optional
	Exporter ExporterConfig `json:"exporter,omitzero"`

	// Signals lists the telemetry signals sent to this target's exporter. Valid
	// values are "logs", "events" and "metrics". A signal is enabled iff at
	// least one target lists it.
	//
	// +k8s:required
	Signals []SignalType `json:"signals"`

	// Filters is an ordered list of filter rules applied to this target's
	// pipelines. Each rule becomes a filter processor instance per matching
	// signal.
	//
	// +k8s:optional
	Filters []FilterRule `json:"filters,omitempty"`
}

// SignalType identifies a telemetry signal. It is used both as the value of a
// target's Signals list and internally for pipeline and component naming.
//
// +k8s:enum
type SignalType string

const (
	// SignalMetrics is the metrics signal, scraped via Prometheus.
	SignalMetrics SignalType = "metrics"
	// SignalLogs is the logs signal, received via OTLP.
	SignalLogs SignalType = "logs"
	// SignalEvents is the Kubernetes events signal, collected from the shoot.
	SignalEvents SignalType = "events"
)

// CollectorConfigSpec specifies the desired state of [CollectorConfig]
type CollectorConfigSpec struct {
	// Targets is the list of exporter destinations. Each target pairs a
	// self-contained exporter with the signals it receives and its own filter
	// rules. Each (target, signal) pair becomes one collector service pipeline.
	//
	// +k8s:optional
	Targets []Target `json:"targets,omitempty"`

	// Logs specifies the settings for the collector's internal logs.
	//
	// +k8s:optional
	Logs CollectorLogsConfig `json:"logs,omitzero"`

	// Metrics specifies the settings for the internal collector metrics.
	//
	// +k8s:optional
	Metrics CollectorMetricsConfig `json:"metrics,omitzero"`
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
