// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package filterrender_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/gardener/gardener-extension-otelcol/pkg/apis/config"
	"github.com/gardener/gardener-extension-otelcol/pkg/filterrender"
)

// filterTarget returns a Target with the given raw filter body.
func filterTarget(body string) config.Target {
	target := config.Target{}
	if body != "" {
		target.Filters = runtime.RawExtension{Raw: []byte(body)}
	}

	return target
}

var _ = Describe("FilterProcessorConfig", func() {
	It("returns an empty map for a target with no filter", func() {
		out, err := filterrender.FilterProcessorConfig(filterTarget(""))
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(BeEmpty())
	})

	It("decodes the opaque body into a map verbatim", func() {
		out, err := filterrender.FilterProcessorConfig(
			filterTarget(
				`{"error_mode":"ignore","metrics":{"metric":["metric.name == \"foo\""]}}`,
			),
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal(map[string]any{
			"error_mode": "ignore",
			"metrics": map[string]any{
				"metric": []any{`metric.name == "foo"`},
			},
		}))
	})

	It("returns an error for a malformed body", func() {
		_, err := filterrender.FilterProcessorConfig(filterTarget(`{not json`))
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("FilterProcessorConfigsForSignals", func() {
	const (
		metricFilterBody          = `{"metrics":{"metric":["metric.name == \"foo\""]}}`
		logFilterBody             = `{"logs":{"log_record":["true"]}}`
		metricConditionFilterBody = `{"metric_conditions":[{"context":"metric","conditions":["metric.name == \"foo\""]}]}`
		logConditionFilterBody    = `{"log_conditions":[{"context":"log","conditions":["true"]}]}`
		bothFilterBody            = `{"metrics":{"metric":["metric.name == \"foo\""]},"logs":{"log_record":["true"]}}`
		traceFilterBody           = `{"traces":{"span":["true"]}}`
	)

	allSignals := []config.SignalType{config.SignalMetrics, config.SignalLogs, config.SignalEvents}

	It("returns an empty map for a target with no filter", func() {
		out, err := filterrender.FilterProcessorConfigsForSignals(filterTarget(""), allSignals)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(BeEmpty())
	})

	It("returns only the metrics signal for a metrics-only filter", func() {
		out, err := filterrender.FilterProcessorConfigsForSignals(filterTarget(metricFilterBody), allSignals)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(HaveKey(config.SignalMetrics))
		Expect(out).NotTo(HaveKey(config.SignalLogs))
		Expect(out).NotTo(HaveKey(config.SignalEvents))
	})

	It("returns only the logs and events signals for a logs-only filter", func() {
		out, err := filterrender.FilterProcessorConfigsForSignals(filterTarget(logFilterBody), allSignals)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(HaveKey(config.SignalLogs))
		Expect(out).To(HaveKey(config.SignalEvents))
		Expect(out).NotTo(HaveKey(config.SignalMetrics))
	})

	It("returns each applicable signal for a filter targeting both", func() {
		out, err := filterrender.FilterProcessorConfigsForSignals(filterTarget(bothFilterBody), allSignals)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(HaveKey(config.SignalMetrics))
		Expect(out).To(HaveKey(config.SignalLogs))
		Expect(out).To(HaveKey(config.SignalEvents))
	})

	It("detects metrics via the metric_conditions field", func() {
		out, err := filterrender.FilterProcessorConfigsForSignals(filterTarget(metricConditionFilterBody), allSignals)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(HaveKey(config.SignalMetrics))
		Expect(out).NotTo(HaveKey(config.SignalLogs))
	})

	It("detects logs via the log_conditions field", func() {
		out, err := filterrender.FilterProcessorConfigsForSignals(filterTarget(logConditionFilterBody), allSignals)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(HaveKey(config.SignalLogs))
		Expect(out).NotTo(HaveKey(config.SignalMetrics))
	})

	It("returns an empty map for an unsupported (traces) filter", func() {
		out, err := filterrender.FilterProcessorConfigsForSignals(filterTarget(traceFilterBody), allSignals)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(BeEmpty())
	})

	It("omits a targeted signal that is not in the requested signals", func() {
		// The filter targets metrics, but metrics is not requested, so it is
		// dropped.
		out, err := filterrender.FilterProcessorConfigsForSignals(
			filterTarget(metricFilterBody),
			[]config.SignalType{config.SignalLogs},
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(BeEmpty())
	})

	It("maps each applicable signal to the full rendered body", func() {
		out, err := filterrender.FilterProcessorConfigsForSignals(filterTarget(metricFilterBody), allSignals)
		Expect(err).NotTo(HaveOccurred())
		Expect(out[config.SignalMetrics]).To(Equal(map[string]any{
			"metrics": map[string]any{
				"metric": []any{`metric.name == "foo"`},
			},
		}))
	})

	It("returns an error for a malformed body", func() {
		_, err := filterrender.FilterProcessorConfigsForSignals(filterTarget(`{not json`), allSignals)
		Expect(err).To(HaveOccurred())
	})
})
