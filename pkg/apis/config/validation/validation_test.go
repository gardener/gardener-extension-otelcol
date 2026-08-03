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

var _ = Describe("Validate", func() {
	// baseConfig returns a config with a default exporter endpoint and the logs
	// signal enabled with a single default-inheriting target, so it passes
	// validation.
	baseConfig := func() config.CollectorConfig {
		return config.CollectorConfig{
			Spec: config.CollectorConfigSpec{
				DefaultExporter: config.ExporterConfig{
					Protocol: config.ExporterProtocolHTTP,
					Endpoint: "https://example.com:4318",
				},
				Signals: config.SignalsConfig{
					Logs: config.SignalConfig{
						Enabled: new(true),
						Targets: []config.SignalTarget{{}},
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

	It("rejects a target without an effective endpoint", func() {
		cfg := baseConfig()
		cfg.Spec.DefaultExporter.Endpoint = ""

		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.signals.logs.targets[0].exporter.endpoint"))
	})

	It("accepts a target that supplies its own endpoint override", func() {
		cfg := baseConfig()
		cfg.Spec.DefaultExporter.Endpoint = ""
		cfg.Spec.Signals.Logs.Targets = []config.SignalTarget{{
			Exporter: config.ExporterConfig{Endpoint: "https://logs.example.com:4318"},
		}}

		Expect(validation.Validate(cfg)).NotTo(HaveOccurred())
	})

	It("rejects a grpc target without an effective endpoint", func() {
		cfg := baseConfig()
		cfg.Spec.DefaultExporter.Endpoint = ""
		cfg.Spec.Signals.Logs.Targets = []config.SignalTarget{{
			Exporter: config.ExporterConfig{Protocol: config.ExporterProtocolGRPC},
		}}

		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.signals.logs.targets[0].exporter.endpoint"))
	})

	It("accepts a debug target without any endpoint", func() {
		cfg := baseConfig()
		cfg.Spec.DefaultExporter.Endpoint = ""
		cfg.Spec.Signals.Logs.Targets = []config.SignalTarget{{
			Exporter: config.ExporterConfig{
				Protocol:  config.ExporterProtocolDebug,
				Verbosity: config.DebugExporterVerbosityBasic,
			},
		}}

		Expect(validation.Validate(cfg)).NotTo(HaveOccurred())
	})

	It("rejects a negative buffer size on the default exporter", func() {
		cfg := baseConfig()
		cfg.Spec.DefaultExporter.ReadBufferSize = -1

		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.defaultExporter.read_buffer_size"))
	})

	It("rejects an incomplete resource reference on the default exporter token", func() {
		cfg := baseConfig()
		cfg.Spec.DefaultExporter.Token = &config.ResourceReference{
			ResourceRef: config.ResourceReferenceDetails{Name: "only-name"},
		}

		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.defaultExporter.token"))
	})

	It("rejects a global filter that uses a signal-specific form", func() {
		cfg := baseConfig()
		cfg.Spec.GlobalFilters = []config.FilterRule{
			{Metrics: &config.MetricFilters{Metric: []string{`true`}}},
		}

		err := validation.Validate(cfg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.globalFilters[0]"))
	})

	It("accepts a global filter that uses the conditions form", func() {
		cfg := baseConfig()
		cfg.Spec.GlobalFilters = []config.FilterRule{
			{Conditions: []config.ContextConditions{{Conditions: []string{`true`}}}},
		}

		Expect(validation.Validate(cfg)).NotTo(HaveOccurred())
	})
})
