// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Encoding specifies the encoding used by the collector exporters.
type Encoding string

const (
	// EncodingProto specifies that proto encoding is used for messages.
	EncodingProto Encoding = "proto"
	// EncodingJSON specifies that JSON is used for encoding messages.
	EncodingJSON Encoding = "json"
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
	Encoding Encoding

	// RetryOnFailure specifies the retry policy of the exporter.
	RetryOnFailure RetryOnFailureConfig

	// Compression specifies the compression to use.
	//
	// Possible options are gzip, zstd, snappy and none.
	Compression Compression
}

// CollectorExportersConfig provides the OTLP exporter settings.
type CollectorExportersConfig struct {
	// HTTPExporter provides the OTLP HTTP Exporter settings.
	OTLPHTTPExporter OTLPHTTPExporterConfig
}

// CollectorConfigSpec specifies the desired state of [CollectorConfig]
type CollectorConfigSpec struct {
	// Exporters specify exporters configuration of the collector.
	Exporters CollectorExportersConfig
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
	// Name is the name of thresource e reference in `.spec.resources` in the Shoot resource.
	Name string
	// DataKey is the key in the resource data map.
	DataKey string
}
