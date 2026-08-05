// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validation_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gardener/gardener-extension-otelcol/pkg/apis/config"
	"github.com/gardener/gardener-extension-otelcol/pkg/apis/config/validation"
)

const (
	testEndpoint    = "https://example.com:4318"
	metricCondition = `name == "foo"`
	trueCondition   = `true`
)

var _ = Describe("Validate", func() {
	// baseConfig returns a config with a single target serving the logs signal
	// with an endpoint, so it passes validation.
	baseConfig := func() config.CollectorConfig {
		return config.CollectorConfig{
			Spec: config.CollectorConfigSpec{
				Targets: []config.Target{{
					Signals: []config.SignalType{config.SignalLogs},
					Exporter: config.ExporterConfig{
						Protocol: config.ExporterProtocolHTTP,
						Endpoint: testEndpoint,
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
		cfg.Spec.Targets[0].Exporter.Endpoint = ""

		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.targets[0].exporter.endpoint"))
	})

	It("rejects a grpc target without an endpoint", func() {
		cfg := baseConfig()
		cfg.Spec.Targets[0].Exporter = config.ExporterConfig{Protocol: config.ExporterProtocolGRPC}

		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.targets[0].exporter.endpoint"))
	})

	It("accepts a debug target without any endpoint", func() {
		cfg := baseConfig()
		cfg.Spec.Targets[0].Exporter = config.ExporterConfig{
			Protocol:  config.ExporterProtocolDebug,
			Verbosity: config.DebugExporterVerbosityBasic,
		}

		Expect(validation.Validate(cfg)).NotTo(HaveOccurred())
	})

	It("rejects a negative buffer size on a target exporter", func() {
		cfg := baseConfig()
		cfg.Spec.Targets[0].Exporter.ReadBufferSize = -1

		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.targets[0].exporter.read_buffer_size"))
	})

	It("rejects an incomplete resource reference on a target exporter token", func() {
		cfg := baseConfig()
		cfg.Spec.Targets[0].Exporter.Token = &config.ResourceReference{
			ResourceRef: config.ResourceReferenceDetails{Name: "only-name"},
		}

		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.targets[0].exporter.token"))
	})

	It("accepts a metrics filter block on a target serving the metrics signal", func() {
		cfg := baseConfig()
		cfg.Spec.Targets[0].Signals = []config.SignalType{config.SignalMetrics}
		cfg.Spec.Targets[0].Filters = []config.FilterRule{{
			Metrics: &config.FilterMetrics{
				Metric:    []string{metricCondition},
				DataPoint: []string{`value_int == 0`},
			},
		}}

		Expect(validation.Validate(cfg)).NotTo(HaveOccurred())
	})

	It("accepts a logs filter block on a target serving the events signal", func() {
		cfg := baseConfig()
		cfg.Spec.Targets[0].Signals = []config.SignalType{config.SignalEvents}
		cfg.Spec.Targets[0].Filters = []config.FilterRule{{
			Logs: &config.FilterLogs{LogRecord: []string{trueCondition}},
		}}

		Expect(validation.Validate(cfg)).NotTo(HaveOccurred())
	})

	It("rejects a metrics filter block on a target that does not serve metrics", func() {
		cfg := baseConfig()
		cfg.Spec.Targets[0].Filters = []config.FilterRule{{
			Metrics: &config.FilterMetrics{Metric: []string{metricCondition}},
		}}

		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.targets[0].filters[0].metrics"))
	})

	It("rejects a logs filter block on a metrics-only target", func() {
		cfg := baseConfig()
		cfg.Spec.Targets[0].Signals = []config.SignalType{config.SignalMetrics}
		cfg.Spec.Targets[0].Filters = []config.FilterRule{{
			Logs: &config.FilterLogs{LogRecord: []string{trueCondition}},
		}}

		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.targets[0].filters[0].logs"))
	})

	It("accepts a filter with both blocks on a target serving both signals", func() {
		cfg := baseConfig()
		cfg.Spec.Targets[0].Signals = []config.SignalType{config.SignalLogs, config.SignalMetrics}
		cfg.Spec.Targets[0].Filters = []config.FilterRule{{
			Metrics: &config.FilterMetrics{Metric: []string{metricCondition}},
			Logs:    &config.FilterLogs{LogRecord: []string{trueCondition}},
		}}

		Expect(validation.Validate(cfg)).NotTo(HaveOccurred())
	})
})
