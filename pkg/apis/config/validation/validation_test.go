// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validation_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/gardener/gardener-extension-otelcol/pkg/apis/config"
	"github.com/gardener/gardener-extension-otelcol/pkg/apis/config/validation"
)

const (
	testEndpoint = "https://example.com:4318"
)

// filterBody returns a RawExtension wrapping the given raw filterprocessor body.
func filterBody(body string) runtime.RawExtension {
	return runtime.RawExtension{Raw: []byte(body)}
}

var _ = Describe("Validate", func() {
	// baseConfig returns a config with a single target serving the logs signal
	// with an endpoint, so it passes validation.
	baseConfig := func() config.CollectorConfig {
		return config.CollectorConfig{
			Spec: config.CollectorConfigSpec{
				Targets: []config.Target{{
					Signals: []config.SignalType{config.SignalLogs},
					Exporter: config.CollectorExportersConfig{
						OTLPHTTPExporter: &config.OTLPHTTPExporterConfig{
							Endpoint: testEndpoint,
						},
					},
				}},
			},
		}
	}

	It("accepts a valid config", func() {
		Expect(validation.Validate(baseConfig())).NotTo(HaveOccurred())
	})

	It("rejects a config with no target", func() {
		cfg := config.CollectorConfig{}

		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("at least one target must be defined"))
	})

	It("rejects a target that serves no signal", func() {
		cfg := baseConfig()
		cfg.Spec.Targets[0].Signals = nil

		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.targets[0].signals"))
	})

	It("rejects an unknown signal", func() {
		cfg := baseConfig()
		cfg.Spec.Targets[0].Signals = []config.SignalType{"traces"}

		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.targets[0].signals[0]"))
	})

	It("rejects a duplicated signal", func() {
		cfg := baseConfig()
		cfg.Spec.Targets[0].Signals = []config.SignalType{config.SignalLogs, config.SignalLogs}

		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.targets[0].signals[1]"))
	})

	It("rejects a target without an endpoint", func() {
		cfg := baseConfig()
		cfg.Spec.Targets[0].Exporter.OTLPHTTPExporter.Endpoint = ""

		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.targets[0].exporter.otlp_http.endpoint"))
	})

	It("rejects a grpc exporter without an endpoint", func() {
		cfg := baseConfig()
		cfg.Spec.Targets[0].Exporter = config.CollectorExportersConfig{OTLPGRPCExporter: &config.OTLPGRPCExporterConfig{}}

		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.targets[0].exporter.otlp_grpc.endpoint"))
	})

	It("accepts a debug exporter without any endpoint", func() {
		cfg := baseConfig()
		cfg.Spec.Targets[0].Exporter = config.CollectorExportersConfig{
			DebugExporter: &config.DebugExporterConfig{Verbosity: config.DebugExporterVerbosityBasic},
		}

		Expect(validation.Validate(cfg)).NotTo(HaveOccurred())
	})

	It("rejects a target with no exporter transport enabled", func() {
		cfg := baseConfig()
		cfg.Spec.Targets[0].Exporter = config.CollectorExportersConfig{}

		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.targets[0].exporter"))
	})

	It("accepts a target exporting to both HTTP and gRPC", func() {
		cfg := baseConfig()
		cfg.Spec.Targets[0].Exporter = config.CollectorExportersConfig{
			OTLPHTTPExporter: &config.OTLPHTTPExporterConfig{Endpoint: testEndpoint},
			OTLPGRPCExporter: &config.OTLPGRPCExporterConfig{Endpoint: "https://example.com:4317"},
		}

		Expect(validation.Validate(cfg)).NotTo(HaveOccurred())
	})

	It("rejects a negative buffer size on a target exporter", func() {
		cfg := baseConfig()
		cfg.Spec.Targets[0].Exporter.OTLPHTTPExporter.ReadBufferSize = -1

		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.targets[0].exporter.otlp_http.read_buffer_size"))
	})

	It("rejects an incomplete resource reference on a target exporter token", func() {
		cfg := baseConfig()
		cfg.Spec.Targets[0].Exporter.OTLPHTTPExporter.Token = &config.ResourceReference{
			ResourceRef: config.ResourceReferenceDetails{Name: "only-name"},
		}

		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.targets[0].exporter.otlp_http.token"))
	})

	It("accepts a metrics filter body on a target serving the metrics signal", func() {
		cfg := baseConfig()
		cfg.Spec.Targets[0].Signals = []config.SignalType{config.SignalMetrics}
		cfg.Spec.Targets[0].Filters = filterBody(`{"metrics":{"metric":["metric.name == \"foo\""],"datapoint":["value_int == 0"]}}`)

		Expect(validation.Validate(cfg)).NotTo(HaveOccurred())
	})

	It("accepts a logs filter body on a target serving the events signal", func() {
		cfg := baseConfig()
		cfg.Spec.Targets[0].Signals = []config.SignalType{config.SignalEvents}
		cfg.Spec.Targets[0].Filters = filterBody(`{"logs":{"log_record":["true"]}}`)

		Expect(validation.Validate(cfg)).NotTo(HaveOccurred())
	})

	It("accepts an empty filter body", func() {
		cfg := baseConfig()
		cfg.Spec.Targets[0].Filters = runtime.RawExtension{}

		Expect(validation.Validate(cfg)).NotTo(HaveOccurred())
	})

	It("accepts a filter body covering multiple signals of the target", func() {
		cfg := baseConfig()
		cfg.Spec.Targets[0].Signals = []config.SignalType{config.SignalLogs, config.SignalMetrics}
		cfg.Spec.Targets[0].Filters = filterBody(`{"metrics":{"metric":["metric.name == \"foo\""]},"logs":{"log_record":["true"]}}`)

		Expect(validation.Validate(cfg)).NotTo(HaveOccurred())
	})

	It("rejects a filter with an invalid OTTL condition", func() {
		cfg := baseConfig()
		cfg.Spec.Targets[0].Signals = []config.SignalType{config.SignalMetrics}
		cfg.Spec.Targets[0].Filters = filterBody(`{"metrics":{"metric":["this is ~~ not ottl"]}}`)

		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.targets[0].filters"))
	})

	It("rejects a filter with an unknown severity number", func() {
		cfg := baseConfig()
		cfg.Spec.Targets[0].Filters = filterBody(`{"logs":{"include":{"match_type":"strict","severity_number":{"min":"NOTASEV"}}}}`)

		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.targets[0].filters"))
	})

	It("rejects a filter with an unknown top-level key", func() {
		cfg := baseConfig()
		cfg.Spec.Targets[0].Filters = filterBody(`{"totally_unknown_key":true}`)

		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.targets[0].filters"))
	})

	It("rejects a filter with a malformed JSON body", func() {
		cfg := baseConfig()
		cfg.Spec.Targets[0].Filters = filterBody(`{not json`)

		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.targets[0].filters"))
	})

	It("accepts basic-style context-inferred conditions", func() {
		cfg := baseConfig()
		cfg.Spec.Targets[0].Signals = []config.SignalType{config.SignalMetrics}
		cfg.Spec.Targets[0].Filters = filterBody(`{"metric_conditions":["metric.name == \"foo\""]}`)

		Expect(validation.Validate(cfg)).NotTo(HaveOccurred())
	})

	It("accepts advanced-style context-inferred conditions", func() {
		cfg := baseConfig()
		cfg.Spec.Targets[0].Signals = []config.SignalType{config.SignalMetrics}
		cfg.Spec.Targets[0].Filters = filterBody(`{"metric_conditions":[{"context":"metric","conditions":["metric.name == \"foo\""]}]}`)

		Expect(validation.Validate(cfg)).NotTo(HaveOccurred())
	})

	It("rejects mixing basic and advanced condition styles", func() {
		cfg := baseConfig()
		cfg.Spec.Targets[0].Signals = []config.SignalType{config.SignalMetrics}
		cfg.Spec.Targets[0].Filters = filterBody(`{"metric_conditions":["metric.name == \"a\"",{"context":"metric","conditions":["metric.name == \"b\""]}]}`)

		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.targets[0].filters"))
	})
})
