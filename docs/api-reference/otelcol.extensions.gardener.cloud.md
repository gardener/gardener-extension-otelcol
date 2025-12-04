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
| `exporters` _[CollectorExportersConfig](#collectorexportersconfig)_ | Exporters specify exporters configuration of the collector. |  |  |


#### CollectorExportersConfig



CollectorExportersConfig provides the OTLP exporter settings.



_Appears in:_
- [CollectorConfigSpec](#collectorconfigspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `otlphttp` _[OTLPHTTPExporterConfig](#otlphttpexporterconfig)_ | HTTPExporter provides the OTLP HTTP Exporter settings. |  |  |


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


#### Encoding

_Underlying type:_ _string_

Encoding specifies the encoding used by the collector exporters.



_Appears in:_
- [OTLPHTTPExporterConfig](#otlphttpexporterconfig)

| Field | Description |
| --- | --- |
| `proto` | EncodingProto specifies that proto encoding is used for messages.<br /> |
| `json` | EncodingJSON specifies that JSON is used for encoding messages.<br /> |


#### OTLPHTTPExporterConfig



OTLPHTTPExporterConfig provides the OTLP HTTP Exporter configuration settings.

See [OTLP HTTP Exporter] for more details.

[OTLP HTTP Exporter]: https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlphttpexporter



_Appears in:_
- [CollectorExportersConfig](#collectorexportersconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `endpoint` _string_ | Endpoint specifies the target base URL to send data to, e.g. https://example.com:4318<br />To send each signal a corresponding path will be added to this base<br />URL, i.e. for traces "/v1/traces" will appended, for metrics<br />"/v1/metrics" will be appended, for logs "/v1/logs" will be appended. |  |  |
| `traces_endpoint` _string_ | TracesEndpoint specifies the target URL to send trace data to, e.g. https://example.com:4318/v1/traces.<br />When this setting is present the base endpoint setting is ignored for<br />traces. |  |  |
| `metrics_endpoint` _string_ | MetricsEndpoint specifies the target URL to send metric data to, e.g. https://example.com:4318/v1/metrics.<br />When this setting is present the base endpoint setting is ignored for<br />metrics. |  |  |
| `logs_endpoint` _string_ | LogsEndpoint specifies the target URL to send log data to, e.g. https://example.com:4318/v1/logs<br />When this setting is present the base endpoint setting is ignored for<br />logs. |  |  |
| `profiles_endpoint` _string_ | ProfilesEndpoint specifies the target URL to send profile data to, e.g. https://example.com:4318/v1development/profiles.<br />When this setting is present the endpoint setting is ignored for<br />profile data. |  |  |
| `tls` _[TLSConfig](#tlsconfig)_ | TLS specifies the TLS configuration settings for the exporter. |  |  |
| `timeout` _[Duration](#duration)_ | Timeout specifies the HTTP request time limit. Default value is<br />[DefaultExporterClientTimeout]. |  |  |
| `read_buffer_size` _integer_ | ReadBufferSize specifies the ReadBufferSize for the HTTP<br />client. Default value is [DefaultExporterClientReadBufferSize]. |  |  |
| `write_buffer_size` _integer_ | WriteBufferSize specifies the WriteBufferSize for the HTTP<br />client. Default value is [DefaultExporterClientWriteBufferSize]. |  |  |
| `encoding` _[Encoding](#encoding)_ | Encoding specifies the encoding to use for the messages. The default<br />value is [EncodingProto]. |  |  |
| `retry_on_failure` _[RetryOnFailureConfig](#retryonfailureconfig)_ | RetryOnFailure specifies the retry policy of the exporter. |  |  |
| `compression` _[Compression](#compression)_ | Compression specifies the compression to use. The default value is<br />[CompressionGzip]. |  |  |


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



TLSConfig provides the TLS settings used by exporters and receivers.

See [OpenTelemetry TLS Configuration Settings] for more details.

[OpenTelemetry TLS Configuration Settings]: https://github.com/open-telemetry/opentelemetry-collector/blob/main/config/configtls/README.md



_Appears in:_
- [OTLPHTTPExporterConfig](#otlphttpexporterconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `insecure` _boolean_ | Insecure specifies whether to disable client transport security for<br />the exporter's HTTPs or gRPC connection. Default is false. |  |  |
| `curve_preferences` _string array_ | CurvePreferences specifies the curve preferences that will be used in<br />an ECDHE handshake, in preference order.<br />Accepted values by OTLP are: X25519, P521, P256, and P384. |  |  |
| `cert_file` _string_ | CertFile specifies the path to the TLS cert to use for TLS required connections. |  |  |
| `cert_pem` _string_ | CertPEM is an alternative to CertFile, which provides the certificate<br />contents as a string instead of a filepath. |  |  |
| `key_file` _string_ | KeyFile specifies the path to the TLS key to use for TLS required<br />connections. |  |  |
| `key_pem` _string_ | KeyPEM is an alternative to KeyFile, which provides the key contents<br />as a string instead of a filepath. |  |  |
| `ca_file` _string_ | CAFile specifies the path to the CA cert. For a client this verifies<br />the server certificate. For a server this verifies client<br />certificates. If empty uses system root CA. |  |  |
| `ca_pem` _string_ | CAPEM is an alternative to CAFile, which provides the CA cert<br />contents as a string instead of a filepath. |  |  |
| `include_system_ca_certs_pool` _boolean_ | IncludeSystemCACertsPool specifies whether to load the system<br />certificate authorities pool alongside the certificate authority. |  |  |
| `insecure_skip_verify` _boolean_ | InsecureSkipVerify specifies whether to skip verifying the<br />certificate or not.<br />Additionally you can configure TLS to be enabled but skip verifying<br />the server's certificate chain. This cannot be combined with `Insecure'<br />since `Insecure' won't use TLS at all. |  |  |
| `min_version` _string_ | MinVersion specifies the minimum acceptable TLS version.<br />Valid values are 1.0, 1.1, 1.2, 1.3.<br />The default value for this field is 1.2.<br />Note, that TLS 1.0 and 1.1 are deprecated due to known<br />vulnerabilities and should be avoided.<br />default="1.2" |  |  |
| `max_version` _string_ | MaxVersion specifies the maximum acceptable TLS version. |  |  |
| `cipher_suites` _string array_ | CipherSuites specifies the list of cipher suites to use.<br />Explicit cipher suites can be set. If left blank, a safe default list<br />is used. See https://go.dev/src/crypto/tls/cipher_suites.go for a<br />list of supported cipher suites. |  |  |
| `reload_interval` _[Duration](#duration)_ | ReloadInterval specifies the duration after which the certificate<br />will be reloaded. If not set, it will never be reloaded. |  |  |


