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
	// validConfig returns a config that passes validation (one exporter
	// enabled), so signal-specific assertions are not masked by other errors.
	validConfig := func(signals ...config.SignalType) config.CollectorConfig {
		return config.CollectorConfig{
			Spec: config.CollectorConfigSpec{
				Exporters: config.CollectorExportersConfig{
					DebugExporter: config.DebugExporterConfig{
						Enabled: new(true),
					},
				},
				Signals: signals,
			},
		}
	}

	DescribeTable("should validate the selected signals",
		func(signals []config.SignalType, wantErrSubstrings []string) {
			err := validation.Validate(validConfig(signals...))

			if len(wantErrSubstrings) == 0 {
				Expect(err).NotTo(HaveOccurred())
				return
			}

			Expect(err).To(HaveOccurred())
			for _, sub := range wantErrSubstrings {
				Expect(err.Error()).To(ContainSubstring(sub))
			}
		},
		Entry("empty selection is allowed",
			nil,
			nil,
		),
		Entry("a valid subset is allowed",
			[]config.SignalType{config.SignalMetrics, config.SignalEvents},
			nil,
		),
		Entry("all known signals are allowed",
			[]config.SignalType{config.SignalLogs, config.SignalEvents, config.SignalMetrics},
			nil,
		),
		Entry("an unknown signal is rejected at its index",
			[]config.SignalType{config.SignalLogs, "traces"},
			[]string{`spec.signals[1]: Unsupported value: "traces"`},
		),
		Entry("a duplicate signal is rejected at its second occurrence",
			[]config.SignalType{config.SignalLogs, config.SignalLogs},
			[]string{`spec.signals[1]: Duplicate value: "logs"`},
		),
		Entry("both a duplicate and an unknown are reported",
			[]config.SignalType{config.SignalLogs, config.SignalLogs, "traces"},
			[]string{
				`spec.signals[1]: Duplicate value: "logs"`,
				`spec.signals[2]: Unsupported value: "traces"`,
			},
		),
	)
})
