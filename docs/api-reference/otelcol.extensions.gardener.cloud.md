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
| `exporters` _[CollectorExportersConfig](#collectorexportersconfig)_ | Exporters specifies the exporters configuration of the collector. |  |  |
| `logs` _[CollectorLogsConfig](#collectorlogsconfig)_ | Logs specifies the settings for the collector logs. |  |  |
| `metrics` _[CollectorMetricsConfig](#collectormetricsconfig)_ | Metrics specifies the settings for the internal collector metrics. |  |  |


#### CollectorExportersConfig



CollectorExportersConfig provides the OTLP exporter settings.



_Appears in:_
- [CollectorConfigSpec](#collectorconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `otlphttp` _[OTLPHTTPExporterConfig](#otlphttpexporterconfig)_ | HTTPExporter provides the OTLP HTTP Exporter settings. |  |  |
| `debug` _[DebugExporterConfig](#debugexporterconfig)_ | DebugExporter provides the settings for the debug exporter. |  |  |


#### CollectorLogsConfig



CollectorLogsConfig provides the settings for the collector internal logs.

See [Configure internal logs] for more details.

[Configure internal logs]: https://opentelemetry.io/docs/collector/internal-telemetry/#configure-internal-logs



_Appears in:_
- [CollectorConfigSpec](#collectorconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `level` _[LogLevel](#loglevel)_ | Level specifies the log level of the collector. |  |  |
| `encoding` _[LogEncoding](#logencoding)_ | Encoding specifies the encoding for logs of the collector. |  |  |


#### CollectorMetricsConfig



CollectorMetricsConfig provides the settings for the collector internal
metrics.

See [Metrics verbosity] for more details.

[Metrics verbosity]: https://opentelemetry.io/docs/collector/internal-telemetry/#metric-verbosity



_Appears in:_
- [CollectorConfigSpec](#collectorconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `level` _[MetricsVerbosityLevel](#metricsverbositylevel)_ | Level specifies the collector internal metrics verbosity level. |  |  |


#### Compression

_Underlying type:_ _string_

Compression specifies the compression used by the collector.



_Appears in:_
- [OTLPHTTPExporterConfig](#otlphttpexporterconfig)

| Field | Description |
| --- | --- |
| `gzip` | CompressionGzip specifies that gzip compression is used.<br /> |
| `zstd` | CompressionZstd specifies that zstd compression is used.<br /> |
| `snappy` | CompressionSnappy specifies that snappy compression is used.<br /> |
| `none` | CompressionNone specifies that no compression is used.<br /> |


#### DebugExporterConfig



DebugExporterConfig provides the settings for the debug exporter



_Appears in:_
- [CollectorExportersConfig](#collectorexportersconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `enabled` _boolean_ | Enabled specifies whether the debug exporter is enabled or not. |  |  |
| `verbosity` _[DebugExporterVerbosity](#debugexporterverbosity)_ | Verbosity specifies the verbosity level for the debug exporter. |  |  |


#### DebugExporterVerbosity

_Underlying type:_ _string_

DebugExporterVerbosity specifies the verbosity level for the debug exporter.



_Appears in:_
- [DebugExporterConfig](#debugexporterconfig)

| Field | Description |
| --- | --- |
| `basic` | DebugExporterVerbosityBasic specifies basic level of verbosity.<br /> |
| `normal` | DebugExporterVerbosityNormal specifies normal level of verbosity.<br /> |
| `detailed` | DebugExporterVerbosityDetailed specifies detailed level of verbosity.<br /> |


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


#### MessageEncoding

_Underlying type:_ _string_

MessageEncoding specifies the encoding used by the collector exporters.



_Appears in:_
- [OTLPHTTPExporterConfig](#otlphttpexporterconfig)

| Field | Description |
| --- | --- |
| `proto` | MessageEncodingProto specifies that proto encoding is used for<br />messages.<br /> |
| `json` | MessageEncodingJSON specifies that JSON is used for encoding<br />messages.<br /> |


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
| `detailed` | MetricsVerbosityDetailed configures the collector with the most<br />verbose level, which includes dimensions and views.<br /> |


#### OTLPHTTPExporterConfig



OTLPHTTPExporterConfig provides the OTLP HTTP Exporter configuration settings.

See [OTLP HTTP Exporter] for more details.

[OTLP HTTP Exporter]: https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlphttpexporter



_Appears in:_
- [CollectorExportersConfig](#collectorexportersconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `enabled` _boolean_ | Enabled specifies whether the OTLP HTTP exporter is enabled or not. |  |  |
| `endpoint` _string_ | Endpoint specifies the target base URL to send data to, e.g. https://example.com:4318<br />To send each signal a corresponding path will be added to this base<br />URL, i.e. for traces "/v1/traces" will appended, for metrics<br />"/v1/metrics" will be appended, for logs "/v1/logs" will be appended. |  |  |
| `traces_endpoint` _string_ | TracesEndpoint specifies the target URL to send trace data to, e.g. https://example.com:4318/v1/traces.<br />When this setting is present the base endpoint setting is ignored for<br />traces. |  |  |
| `metrics_endpoint` _string_ | MetricsEndpoint specifies the target URL to send metric data to, e.g. https://example.com:4318/v1/metrics.<br />When this setting is present the base endpoint setting is ignored for<br />metrics. |  |  |
| `logs_endpoint` _string_ | LogsEndpoint specifies the target URL to send log data to, e.g. https://example.com:4318/v1/logs<br />When this setting is present the base endpoint setting is ignored for<br />logs. |  |  |
| `profiles_endpoint` _string_ | ProfilesEndpoint specifies the target URL to send profile data to, e.g. https://example.com:4318/v1development/profiles.<br />When this setting is present the endpoint setting is ignored for<br />profile data. |  |  |
| `tls` _[TLSConfig](#tlsconfig)_ | TLS specifies the TLS configuration settings for the exporter. |  |  |
| `token` _[ResourceReference](#resourcereference)_ | Token references a bearer token for authentication. |  |  |
| `timeout` _[Duration](#duration)_ | Timeout specifies the HTTP request time limit. Default value is<br />[DefaultExporterClientTimeout]. |  |  |
| `read_buffer_size` _integer_ | ReadBufferSize specifies the ReadBufferSize for the HTTP<br />client. Default value is [DefaultExporterClientReadBufferSize]. |  |  |
| `write_buffer_size` _integer_ | WriteBufferSize specifies the WriteBufferSize for the HTTP<br />client. Default value is [DefaultExporterClientWriteBufferSize]. |  |  |
| `encoding` _[MessageEncoding](#messageencoding)_ | Encoding specifies the encoding to use for the messages. The default<br />value is [MessageEncodingProto]. |  |  |
| `retry_on_failure` _[RetryOnFailureConfig](#retryonfailureconfig)_ | RetryOnFailure specifies the retry policy of the exporter. |  |  |
| `compression` _[Compression](#compression)_ | Compression specifies the compression to use. The default value is<br />[CompressionGzip]. |  |  |


#### ResourceReference



ResourceReference references data from a Secret.



_Appears in:_
- [OTLPHTTPExporterConfig](#otlphttpexporterconfig)
- [TLSConfig](#tlsconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `resourceRef` _[ResourceReferenceDetails](#resourcereferencedetails)_ | ResourceRef references a resource in the shoot. |  |  |


#### ResourceReferenceDetails



ResourceReferenceDetails references a resource (e.g., a Secret) in the garden cluster.



_Appears in:_
- [ResourceReference](#resourcereference)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name is the name of thresource e reference in `.spec.resources` in the Shoot resource. |  |  |
| `dataKey` _string_ | DataKey is the key in the resource data map. |  |  |


#### RetryOnFailureConfig



RetryOnFailureConfig provides the retry policy for an exporter.



_Appears in:_
- [OTLPHTTPExporterConfig](#otlphttpexporterconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `enabled` _boolean_ | Enabled specifies whether retry on failure is enabled or not. Default<br />is true. |  |  |
| `initial_interval` _[Duration](#duration)_ | InitialInterval specifies the time to wait after the first failure<br />before retrying. The default value is [DefaultRetryInitialInterval]. |  |  |
| `max_interval` _[Duration](#duration)_ | MaxInterval specifies the upper bound on backoff. Default value is<br />[DefaultRetryMaxInterval]. |  |  |
| `max_elapsed_time` _[Duration](#duration)_ | MaxElapsedTime specifies the maximum amount of time spent trying to<br />send a batch. If set to 0, the retries are never stopped. The default<br />value is [DefaultRetryMaxElapsedTime]. |  |  |
| `multiplier` _float_ | Multiplier specifies the factor by which the retry interval is<br />multiplied on each attempt. The default value is<br />[DefaultRetryMultiplier]. |  |  |


#### TLSConfig



TLSConfig provides the TLS settings used by exporters.



_Appears in:_
- [OTLPHTTPExporterConfig](#otlphttpexporterconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `insecureSkipVerify` _boolean_ | InsecureSkipVerify specifies whether to skip verifying the<br />certificate or not. |  |  |
| `ca` _[ResourceReference](#resourcereference)_ | CA references the CA certificate to use for verifying the server certificate.<br />For a client this verifies the server certificate.<br />For a server this verifies client certificates.<br />If empty uses system root CA. |  |  |
| `cert` _[ResourceReference](#resourcereference)_ | Cert references the client certificate to use for TLS required connections. |  |  |
| `key` _[ResourceReference](#resourcereference)_ | Key references the client key to use for TLS required connections. |  |  |


