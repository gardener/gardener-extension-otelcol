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
)

var _ = Describe("Validate", func() {
	// baseConfig returns a config with the logs signal enabled and a single
	// target with an endpoint, so it passes validation.
	baseConfig := func() config.CollectorConfig {
		return config.CollectorConfig{
			Spec: config.CollectorConfigSpec{
				Signals: config.SignalsConfig{
					Logs: config.SignalConfig{
						Enabled: new(true),
						Targets: []config.SignalTarget{{
							Exporter: config.ExporterConfig{
								Protocol: config.ExporterProtocolHTTP,
								Endpoint: testEndpoint,
							},
						}},
					},
				},
			},
		}
	}

	It("accepts a valid config", func() {
		Expect(validation.Validate(baseConfig())).NotTo(HaveOccurred())
	})

	It("rejects a config with no signal enabled", func() {
		cfg := baseConfig()
		cfg.Spec.Signals.Logs.Enabled = new(false)

		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("no signal enabled"))
	})

	It("rejects an enabled signal without any target", func() {
		cfg := baseConfig()
		cfg.Spec.Signals.Logs.Targets = nil

		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.signals.logs.targets"))
	})

	It("rejects a target without an endpoint", func() {
		cfg := baseConfig()
		cfg.Spec.Signals.Logs.Targets[0].Exporter.Endpoint = ""

		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.signals.logs.targets[0].exporter.endpoint"))
	})

	It("rejects a grpc target without an endpoint", func() {
		cfg := baseConfig()
		cfg.Spec.Signals.Logs.Targets = []config.SignalTarget{{
			Exporter: config.ExporterConfig{Protocol: config.ExporterProtocolGRPC},
		}}

		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.signals.logs.targets[0].exporter.endpoint"))
	})

	It("accepts a debug target without any endpoint", func() {
		cfg := baseConfig()
		cfg.Spec.Signals.Logs.Targets = []config.SignalTarget{{
			Exporter: config.ExporterConfig{
				Protocol:  config.ExporterProtocolDebug,
				Verbosity: config.DebugExporterVerbosityBasic,
			},
		}}

		Expect(validation.Validate(cfg)).NotTo(HaveOccurred())
	})

	It("rejects a negative buffer size on a target exporter", func() {
		cfg := baseConfig()
		cfg.Spec.Signals.Logs.Targets[0].Exporter.ReadBufferSize = -1

		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.signals.logs.targets[0].exporter.read_buffer_size"))
	})

	It("rejects an incomplete resource reference on a target exporter token", func() {
		cfg := baseConfig()
		cfg.Spec.Signals.Logs.Targets[0].Exporter.Token = &config.ResourceReference{
			ResourceRef: config.ResourceReferenceDetails{Name: "only-name"},
		}

		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.signals.logs.targets[0].exporter.token"))
	})

	It("accepts metric filter fields on the metrics signal", func() {
		cfg := baseConfig()
		cfg.Spec.Signals.Metrics = config.SignalConfig{
			Enabled: new(true),
			Targets: []config.SignalTarget{{
				Exporter: config.ExporterConfig{Endpoint: testEndpoint},
				Filters: []config.FilterRule{{
					Metric:    []string{metricCondition},
					DataPoint: []string{`value_int == 0`},
				}},
			}},
		}

		Expect(validation.Validate(cfg)).NotTo(HaveOccurred())
	})

	It("rejects log filter fields on the metrics signal", func() {
		cfg := baseConfig()
		cfg.Spec.Signals.Metrics = config.SignalConfig{
			Enabled: new(true),
			Targets: []config.SignalTarget{{
				Exporter: config.ExporterConfig{Endpoint: testEndpoint},
				Filters:  []config.FilterRule{{LogRecord: []string{`true`}}},
			}},
		}

		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.signals.metrics.targets[0].filters[0].log_record"))
	})

	It("rejects metric filter fields on the logs signal", func() {
		cfg := baseConfig()
		cfg.Spec.Signals.Logs.Targets[0].Filters = []config.FilterRule{{
			Metric: []string{metricCondition},
		}}

		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.signals.logs.targets[0].filters[0].metric"))
	})

	It("rejects the signal-agnostic conditions form on the metrics signal", func() {
		cfg := baseConfig()
		cfg.Spec.Signals.Metrics = config.SignalConfig{
			Enabled: new(true),
			Targets: []config.SignalTarget{{
				Exporter: config.ExporterConfig{Endpoint: testEndpoint},
				Filters: []config.FilterRule{{
					Conditions: []config.ContextConditions{{Conditions: []string{`true`}}},
				}},
			}},
		}

		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.signals.metrics.targets[0].filters[0].conditions"))
	})

	It("accepts the conditions form on the traces signal and rejects metric fields", func() {
		cfg := baseConfig()
		cfg.Spec.Signals.Traces = config.SignalConfig{
			Enabled: new(true),
			Targets: []config.SignalTarget{{
				Exporter: config.ExporterConfig{Endpoint: testEndpoint},
				Filters: []config.FilterRule{{
					Conditions: []config.ContextConditions{{Conditions: []string{`name == "ping"`}}},
				}},
			}},
		}
		Expect(validation.Validate(cfg)).NotTo(HaveOccurred())

		cfg.Spec.Signals.Traces.Targets[0].Filters = []config.FilterRule{{
			Metric: []string{metricCondition},
		}}
		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.signals.traces.targets[0].filters[0].metric"))
	})
})
