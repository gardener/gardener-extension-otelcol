# API Reference

## Packages
- [otelcol.extensions.gardener.cloud/v1alpha1](#otelcolextensionsgardenercloudv1alpha1)


## otelcol.extensions.gardener.cloud/v1alpha1

Package v1alpha1 provides the v1alpha1 version of the external API types.





#### CollectorConfigSpec



CollectorConfigSpec specifies the desired state of [CollectorConfig]



_Appears in:_
- [CollectorConfig](#collectorconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `targets` _[Target](#target) array_ | Targets is the list of exporter destinations. Each target pairs a<br />self-contained exporter with the signals it receives and its own filter<br />rules. Each (target, signal) pair becomes one collector service pipeline. |  | Optional: \{\} <br /> |
| `logs` _[CollectorLogsConfig](#collectorlogsconfig)_ | Logs specifies the settings for the collector's internal logs. |  | Optional: \{\} <br /> |
| `metrics` _[CollectorMetricsConfig](#collectormetricsconfig)_ | Metrics specifies the settings for the internal collector metrics. |  | Optional: \{\} <br /> |


#### CollectorLogsConfig



CollectorLogsConfig provides the settings for the collector internal logs.

See [Configure internal logs] for more details.

[Configure internal logs]: https://opentelemetry.io/docs/collector/internal-telemetry/#configure-internal-logs



_Appears in:_
- [CollectorConfigSpec](#collectorconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `level` _[LogLevel](#loglevel)_ | Level specifies the log level of the collector. | <nil> | Optional: \{\} <br /> |
| `encoding` _[LogEncoding](#logencoding)_ | Encoding specifies the encoding for logs of the collector. | <nil> | Optional: \{\} <br /> |


#### CollectorMetricsConfig



CollectorMetricsConfig provides the settings for the collector internal
metrics.

See [Metrics verbosity] for more details.

[Metrics verbosity]: https://opentelemetry.io/docs/collector/internal-telemetry/#metric-verbosity



_Appears in:_
- [CollectorConfigSpec](#collectorconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `level` _[MetricsVerbosityLevel](#metricsverbositylevel)_ | Level specifies the collector internal metrics verbosity level. | <nil> | Optional: \{\} <br /> |


#### Compression

_Underlying type:_ _string_

Compression specifies the compression used by the collector.



_Appears in:_
- [ExporterConfig](#exporterconfig)

| Field | Description |
| --- | --- |
| `gzip` | CompressionGzip specifies that gzip compression is used.<br /> |
| `zstd` | CompressionZstd specifies that zstd compression is used.<br /> |
| `snappy` | CompressionSnappy specifies that snappy compression is used.<br /> |
| `none` | CompressionNone specifies that no compression is used.<br /> |


#### ContextConditions



ContextConditions specifies a group of OTTL conditions for the filter
processor's context-inferred condition style (metric_conditions /
log_conditions).

When Context and ErrorMode are both empty, the group is rendered as a flat
list of condition strings (basic style) and the OTTL context is inferred
from each expression. Otherwise it is rendered as an explicit group
(advanced style).



_Appears in:_
- [FilterLogs](#filterlogs)
- [FilterMetrics](#filtermetrics)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `context` _string_ | Context specifies the OTTL context the conditions are evaluated against,<br />e.g. "resource", "metric", "datapoint", "log". If empty, the context is<br />inferred from each condition. |  | Optional: \{\} <br /> |
| `conditions` _string array_ | Conditions is the list of OTTL conditions. If any condition resolves to<br />true, the matching telemetry is dropped. |  | Required: \{\} <br /> |
| `error_mode` _[FilterErrorMode](#filtererrormode)_ | ErrorMode determines how the processor reacts to errors that occur while<br />processing this group of conditions. When set, it overrides the<br />top-level ErrorMode. Only honored when the group is rendered in the<br />advanced style. |  | Optional: \{\} <br /> |


#### DebugExporterVerbosity

_Underlying type:_ _string_

DebugExporterVerbosity specifies the verbosity level for the debug exporter.



_Appears in:_
- [ExporterConfig](#exporterconfig)

| Field | Description |
| --- | --- |
| `basic` | DebugExporterVerbosityBasic specifies basic level of verbosity.<br /> |
| `normal` | DebugExporterVerbosityNormal specifies normal level of verbosity.<br /> |
| `detailed` | DebugExporterVerbosityDetailed specifies detailed level of verbosity.<br /> |


#### ExporterConfig



ExporterConfig provides a full exporter configuration.

It folds the OTLP HTTP, OTLP gRPC and debug exporters into a single type,
selected by Protocol. Each signal target carries its own ExporterConfig.

When Protocol is [ExporterProtocolDebug] only Verbosity is honored; the
endpoint, TLS, token and buffer settings are ignored.

See [OTLP HTTP Exporter], [OTLP gRPC Exporter] and [Debug Exporter] for more
details.

[OTLP HTTP Exporter]: https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlphttpexporter
[OTLP gRPC Exporter]: https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlpexporter
[Debug Exporter]: https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/debugexporter



_Appears in:_
- [Target](#target)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `protocol` _[ExporterProtocol](#exporterprotocol)_ | Protocol selects the transport used by the exporter. The default value is<br />[ExporterProtocolHTTP]. Set it to [ExporterProtocolDebug] to write<br />telemetry to the collector's own logs instead of sending it to a remote<br />endpoint. | <nil> | Optional: \{\} <br /> |
| `endpoint` _string_ | Endpoint specifies the target endpoint to send data to. It is required<br />unless Protocol is [ExporterProtocolDebug].<br />For the HTTP protocol this is a base URL, e.g. https://example.com:4318;<br />the collector appends the per-signal path (e.g. "/v1/metrics"). For the<br />gRPC protocol this is a gRPC endpoint, see<br />https://github.com/grpc/grpc/blob/master/doc/naming.md. |  | Optional: \{\} <br /> |
| `tls` _[TLSConfig](#tlsconfig)_ | TLS specifies the TLS configuration settings for the exporter. |  | Optional: \{\} <br /> |
| `token` _[ResourceReference](#resourcereference)_ | Token references a bearer token for authentication. |  | Optional: \{\} <br /> |
| `timeout` _[Duration](#duration)_ | Timeout specifies the request time limit. Default value is<br />[DefaultHTTPExporterClientTimeout]. | <nil> | Optional: \{\} <br /> |
| `read_buffer_size` _integer_ | ReadBufferSize specifies the ReadBufferSize for the client. Default value<br />is [DefaultHTTPExporterClientReadBufferSize]. | <nil> | Optional: \{\} <br /> |
| `write_buffer_size` _integer_ | WriteBufferSize specifies the WriteBufferSize for the client. Default<br />value is [DefaultHTTPExporterClientWriteBufferSize]. | <nil> | Optional: \{\} <br /> |
| `encoding` _[MessageEncoding](#messageencoding)_ | Encoding specifies the encoding to use for the messages. It is only<br />honored by the HTTP protocol. The default value is [MessageEncodingProto]. | <nil> | Optional: \{\} <br /> |
| `retry_on_failure` _[RetryOnFailureConfig](#retryonfailureconfig)_ | RetryOnFailure specifies the retry policy of the exporter. |  | Optional: \{\} <br /> |
| `compression` _[Compression](#compression)_ | Compression specifies the compression to use. The default value is<br />[CompressionGzip]. | <nil> | Optional: \{\} <br /> |
| `verbosity` _[DebugExporterVerbosity](#debugexporterverbosity)_ | Verbosity specifies the verbosity level of the debug exporter. It is only<br />honored when Protocol is [ExporterProtocolDebug]. The default value is<br />[DebugExporterVerbosityBasic]. | <nil> | Optional: \{\} <br /> |


#### ExporterProtocol

_Underlying type:_ _string_

ExporterProtocol selects the OTLP transport used by an exporter.



_Appears in:_
- [ExporterConfig](#exporterconfig)

| Field | Description |
| --- | --- |
| `http` | ExporterProtocolHTTP selects the OTLP HTTP exporter.<br /> |
| `grpc` | ExporterProtocolGRPC selects the OTLP gRPC exporter.<br /> |
| `debug` | ExporterProtocolDebug selects the debug exporter, which writes telemetry<br />to the collector's own logs. It only honors the Verbosity field; the<br />endpoint, TLS, token and buffer settings are ignored.<br /> |


#### FilterAttribute



FilterAttribute specifies an attribute key/value pair that the filter
processor match properties evaluate against.



_Appears in:_
- [LogMatchProperties](#logmatchproperties)
- [MetricMatchProperties](#metricmatchproperties)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `key` _string_ | Key specifies the attribute key to match against. |  | Required: \{\} <br /> |
| `value` _string_ | Value specifies the attribute value to match against. If empty, only the<br />presence of the key is checked. |  | Optional: \{\} <br /> |


#### FilterErrorMode

_Underlying type:_ _string_

FilterErrorMode determines how the filter processor reacts to errors that
occur while processing an OTTL condition.

See the link below for more details.

https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/filterprocessor#error-modes



_Appears in:_
- [ContextConditions](#contextconditions)
- [FilterRule](#filterrule)

| Field | Description |
| --- | --- |
| `ignore` | FilterErrorModeIgnore means the processor ignores errors returned by<br />conditions, logs them, and continues on to the next condition.<br /> |
| `silent` | FilterErrorModeSilent means the processor ignores errors returned by<br />conditions, does not log them, and continues on to the next condition.<br /> |
| `propagate` | FilterErrorModePropagate means the processor returns the error up the<br />pipeline, which results in the payload being dropped from the collector.<br /> |


#### FilterLogs



FilterLogs specifies the logs filterprocessor block. It mirrors the "logs"
section of the OTel filter processor and is valid on targets that serve the
logs or events signals.



_Appears in:_
- [FilterRule](#filterrule)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `resource` _string array_ | Resource is a list of OTTL conditions for an ottlresource context. If any<br />condition resolves to true, the whole resource is dropped. |  | Optional: \{\} <br /> |
| `log_record` _string array_ | LogRecord is a list of OTTL conditions for an ottllog context. If any<br />condition resolves to true, the log record is dropped. |  | Optional: \{\} <br /> |
| `include` _[LogMatchProperties](#logmatchproperties)_ | Include specifies the logs that should be kept in the pipeline; all other<br />logs are dropped. If both Include and Exclude are specified, include<br />filtering occurs first. |  | Optional: \{\} <br /> |
| `exclude` _[LogMatchProperties](#logmatchproperties)_ | Exclude specifies the logs that should be dropped from the pipeline; all<br />other logs are kept. |  | Optional: \{\} <br /> |
| `log_conditions` _[ContextConditions](#contextconditions) array_ | LogConditions specifies the logs filter using the context-inferred<br />condition style. |  | Optional: \{\} <br /> |


#### FilterMatchType

_Underlying type:_ _string_

FilterMatchType specifies the type of string pattern matching used by the filter
processor include/exclude match properties.



_Appears in:_
- [LogMatchProperties](#logmatchproperties)
- [MetricMatchProperties](#metricmatchproperties)

| Field | Description |
| --- | --- |
| `strict` | MatchTypeStrict matches values by exact string equality.<br /> |
| `regexp` | MatchTypeRegexp matches values against regular expressions.<br /> |


#### FilterMetrics



FilterMetrics specifies the metrics filterprocessor block. It mirrors the
"metrics" section of the OTel filter processor and is valid on targets that
serve the metrics signal.



_Appears in:_
- [FilterRule](#filterrule)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `resource` _string array_ | Resource is a list of OTTL conditions for an ottlresource context. If any<br />condition resolves to true, the whole resource is dropped. |  | Optional: \{\} <br /> |
| `metric` _string array_ | Metric is a list of OTTL conditions for an ottlmetric context. If any<br />condition resolves to true, the metric is dropped. |  | Optional: \{\} <br /> |
| `datapoint` _string array_ | DataPoint is a list of OTTL conditions for an ottldatapoint context. If<br />any condition resolves to true, the datapoint is dropped. |  | Optional: \{\} <br /> |
| `include` _[MetricMatchProperties](#metricmatchproperties)_ | Include specifies the metrics that should be kept in the pipeline; all<br />other metrics are dropped. If both Include and Exclude are specified,<br />include filtering occurs first. |  | Optional: \{\} <br /> |
| `exclude` _[MetricMatchProperties](#metricmatchproperties)_ | Exclude specifies the metrics that should be dropped from the pipeline;<br />all other metrics are kept. |  | Optional: \{\} <br /> |
| `metric_conditions` _[ContextConditions](#contextconditions) array_ | MetricConditions specifies the metrics filter using the context-inferred<br />condition style. |  | Optional: \{\} <br /> |


#### FilterRule



FilterRule specifies a single filter processor instance, which drops
telemetry matching the given OTTL conditions or match properties. A target's
Filters list produces one filter processor per rule per signal, applied in
order.

Its blocks mirror the OTel filterprocessor, keyed by signal: the Metrics
block feeds a target's metrics pipeline, the Logs block its logs and events
pipelines. A block may only be set for a signal the enclosing target serves.

See [Filter Processor] for more details.

[Filter Processor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/filterprocessor



_Appears in:_
- [Target](#target)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `error_mode` _[FilterErrorMode](#filtererrormode)_ | ErrorMode determines how the processor reacts to errors that occur while<br />processing an OTTL condition. If empty, the processor default is used. |  | Optional: \{\} <br /> |
| `metrics` _[FilterMetrics](#filtermetrics)_ | Metrics specifies the metrics filterprocessor block. Valid only on targets<br />that serve the metrics signal. |  | Optional: \{\} <br /> |
| `logs` _[FilterLogs](#filterlogs)_ | Logs specifies the logs filterprocessor block. Valid on targets that serve<br />the logs or events signals. |  | Optional: \{\} <br /> |


#### LogEncoding

_Underlying type:_ _string_

LogEncoding specifies the encoding for the internal collector logger.

See the link below for more details.

https://opentelemetry.io/docs/collector/internal-telemetry/#configure-internal-logs



_Appears in:_
- [CollectorLogsConfig](#collectorlogsconfig)

| Field | Description |
| --- | --- |
| `console` | LogEncodingConsole sets the collector's internal logger with console<br />encoding.<br /> |
| `json` | LogEncodingJSON sets the collector's internal logger with JSON<br />encoding.<br /> |


#### LogLevel

_Underlying type:_ _string_

LogLevel specifies the minimum enabled logging level for the collector.

See the link below for more details.

https://opentelemetry.io/docs/collector/internal-telemetry/#configure-internal-logs



_Appears in:_
- [CollectorLogsConfig](#collectorlogsconfig)

| Field | Description |
| --- | --- |
| `INFO` | LogLevelInfo sets the collector's internal logger to INFO level.<br /> |
| `WARN` | LogLevelWarn sets the collector's internal logger to WARN level.<br /> |
| `ERROR` | LogLevelError sets the collector's internal logger to ERROR level.<br /> |
| `DEBUG` | LogLevelDebug sets the collector's internal logger to DEBUG level.<br /> |


#### LogMatchProperties



LogMatchProperties specifies the set of properties the filter processor
matches logs against, and the type of string pattern matching to use.



_Appears in:_
- [FilterLogs](#filterlogs)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `match_type` _[FilterMatchType](#filtermatchtype)_ | MatchType specifies the type of matching desired. |  | Required: \{\} <br /> |
| `resource_attributes` _[FilterAttribute](#filterattribute) array_ | ResourceAttributes specifies a list of resource attributes to match logs<br />against. A match occurs if any resource attribute matches all expressions<br />in this list. |  | Optional: \{\} <br /> |
| `record_attributes` _[FilterAttribute](#filterattribute) array_ | RecordAttributes specifies a list of record attributes to match logs<br />against. A match occurs if any record attribute matches at least one<br />expression in this list. |  | Optional: \{\} <br /> |
| `severity_texts` _string array_ | SeverityTexts specifies a list of strings that the log record's severity<br />text field must match against. |  | Optional: \{\} <br /> |
| `severity_number` _[LogSeverityNumberMatchProperties](#logseveritynumbermatchproperties)_ | SeverityNumber specifies how to match against a log record's severity<br />number, if defined. |  | Optional: \{\} <br /> |
| `bodies` _string array_ | Bodies specifies a list of strings that the log record's body field must<br />match against. |  | Optional: \{\} <br /> |


#### LogSeverityNumberMatchProperties



LogSeverityNumberMatchProperties specifies how the filter processor matches
against a log record's severity number.



_Appears in:_
- [LogMatchProperties](#logmatchproperties)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `min` _string_ | Min specifies the minimum severity a log record must have to match. This<br />corresponds to the severity short names, e.g. "INFO", "WARN", "ERROR".<br />The value is case-insensitive. |  | Required: \{\} <br /> |
| `match_undefined` _boolean_ | MatchUndefined specifies whether log records with an "unspecified"<br />severity match. |  | Optional: \{\} <br /> |


#### MessageEncoding

_Underlying type:_ _string_

MessageEncoding specifies the encoding used by the collector exporters.



_Appears in:_
- [ExporterConfig](#exporterconfig)

| Field | Description |
| --- | --- |
| `proto` | MessageEncodingProto specifies that proto encoding is used for<br />messages.<br /> |
| `json` | MessageEncodingJSON specifies that JSON is used for encoding<br />messages.<br /> |


#### MetricMatchProperties



MetricMatchProperties specifies the set of properties the filter processor
matches metrics against, and the type of string pattern matching to use.



_Appears in:_
- [FilterMetrics](#filtermetrics)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `match_type` _[FilterMatchType](#filtermatchtype)_ | MatchType specifies the type of matching desired. |  | Required: \{\} <br /> |
| `regexp` _[RegexpConfig](#regexpconfig)_ | Regexp specifies the options for the regexp match type. |  | Optional: \{\} <br /> |
| `metric_names` _string array_ | MetricNames specifies the list of string patterns to match metric names<br />against. A match occurs if the metric name matches at least one pattern<br />in this list. |  | Optional: \{\} <br /> |
| `expressions` _string array_ | Expressions specifies the list of expr expressions to match metrics<br />against. A match occurs if any datapoint in a metric matches at least one<br />expression in this list. |  | Optional: \{\} <br /> |
| `resource_attributes` _[FilterAttribute](#filterattribute) array_ | ResourceAttributes specifies a list of resource attributes to match<br />metrics against. A match occurs if any resource attribute matches all<br />expressions in this list. |  | Optional: \{\} <br /> |


#### MetricsVerbosityLevel

_Underlying type:_ _string_

MetricsVerbosityLevel specifies the verbosity of the internal collector
metrics.

See the link below for more details.

https://opentelemetry.io/docs/collector/internal-telemetry/#metric-verbosity



_Appears in:_
- [CollectorMetricsConfig](#collectormetricsconfig)

| Field | Description |
| --- | --- |
| `none` | MetricsVerbosityLevelNone disables the internal collector metrics.<br /> |
| `basic` | MetricsVerbosityLevelBasic configures the collector to emit basic<br />metrics only.<br /> |
| `normal` | MetricsVerbosityLevelNormal configures the collector with standard<br />indicators on top of the basic ones.<br /> |
| `detailed` | MetricsVerbosityLevelDetailed configures the collector with the most<br />verbose level, which includes dimensions and views.<br /> |


#### RegexpConfig



RegexpConfig specifies the options for the regexp match type used by the
filter processor include/exclude match properties.



_Appears in:_
- [MetricMatchProperties](#metricmatchproperties)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `cacheenabled` _boolean_ | CacheEnabled specifies whether match results are cached. |  | Optional: \{\} <br /> |
| `cachemaxnumentries` _integer_ | CacheMaxNumEntries specifies the maximum number of entries in the cache. |  | Optional: \{\} <br /> |


#### ResourceReference



ResourceReference references data from a Secret.



_Appears in:_
- [ExporterConfig](#exporterconfig)
- [TLSConfig](#tlsconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `resourceRef` _[ResourceReferenceDetails](#resourcereferencedetails)_ | ResourceRef references a resource in the shoot. |  | Required: \{\} <br /> |


#### ResourceReferenceDetails



ResourceReferenceDetails references a resource (e.g., a Secret) in the garden cluster.



_Appears in:_
- [ResourceReference](#resourcereference)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name is the name of thresource e reference in `.spec.resources` in the Shoot resource. |  | Required: \{\} <br /> |
| `dataKey` _string_ | DataKey is the key in the resource data map. |  | Required: \{\} <br /> |


#### RetryOnFailureConfig



RetryOnFailureConfig provides the retry policy for an exporter.



_Appears in:_
- [ExporterConfig](#exporterconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `enabled` _boolean_ | Enabled specifies whether retry on failure is enabled or not. Default<br />is true. | true | Optional: \{\} <br /> |
| `initial_interval` _[Duration](#duration)_ | InitialInterval specifies the time to wait after the first failure<br />before retrying. The default value is [DefaultRetryInitialInterval]. | <nil> | Optional: \{\} <br /> |
| `max_interval` _[Duration](#duration)_ | MaxInterval specifies the upper bound on backoff. Default value is<br />[DefaultRetryMaxInterval]. | <nil> | Optional: \{\} <br /> |
| `max_elapsed_time` _[Duration](#duration)_ | MaxElapsedTime specifies the maximum amount of time spent trying to<br />send a batch. If set to 0, the retries are never stopped. The default<br />value is [DefaultRetryMaxElapsedTime]. | <nil> | Optional: \{\} <br /> |
| `multiplier` _float_ | Multiplier specifies the factor by which the retry interval is<br />multiplied on each attempt. The default value is<br />[DefaultRetryMultiplier]. | <nil> | Optional: \{\} <br /> |


#### SignalType

_Underlying type:_ _string_

SignalType identifies a telemetry signal. It is used both as the value of a
target's Signals list and internally for pipeline and component naming.



_Appears in:_
- [Target](#target)

| Field | Description |
| --- | --- |
| `metrics` | SignalMetrics is the metrics signal, scraped via Prometheus.<br /> |
| `logs` | SignalLogs is the logs signal, received via OTLP.<br /> |
| `events` | SignalEvents is the Kubernetes events signal, collected from the shoot.<br /> |


#### TLSConfig



TLSConfig provides the TLS settings used by exporters.



_Appears in:_
- [ExporterConfig](#exporterconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `insecureSkipVerify` _boolean_ | InsecureSkipVerify specifies whether to skip verifying the<br />certificate or not. | false | Optional: \{\} <br /> |
| `ca` _[ResourceReference](#resourcereference)_ | CA references the CA certificate to use for verifying the server certificate.<br />For a client this verifies the server certificate.<br />For a server this verifies client certificates.<br />If empty uses system root CA. |  | Optional: \{\} <br /> |
| `cert` _[ResourceReference](#resourcereference)_ | Cert references the client certificate to use for TLS required connections. |  | Optional: \{\} <br /> |
| `key` _[ResourceReference](#resourcereference)_ | Key references the client key to use for TLS required connections. |  | Optional: \{\} <br /> |
| `reloadInterval` _[Duration](#duration)_ | ReloadInterval specifies mTLS key and cert reload interval<br />from mounted secret volume | <nil> | Optional: \{\} <br /> |


#### Target



Target pairs a self-contained exporter with the signals it receives and its
own ordered list of filter rules. Each (target, signal) pair produces one
collector service pipeline, so a target can fan out several signals to one
destination, each with independent filtering.



_Appears in:_
- [CollectorConfigSpec](#collectorconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `exporter` _[ExporterConfig](#exporterconfig)_ | Exporter is the exporter this target sends to. Set its Protocol to<br />[ExporterProtocolDebug] to make this target a debug destination. |  | Optional: \{\} <br /> |
| `signals` _[SignalType](#signaltype) array_ | Signals lists the telemetry signals sent to this target's exporter. Valid<br />values are "logs", "events" and "metrics". A signal is enabled iff at<br />least one target lists it. |  | Required: \{\} <br /> |
| `filters` _[FilterRule](#filterrule) array_ | Filters is an ordered list of filter rules applied to this target's<br />pipelines. Each rule becomes a filter processor instance per matching<br />signal. |  | Optional: \{\} <br /> |


