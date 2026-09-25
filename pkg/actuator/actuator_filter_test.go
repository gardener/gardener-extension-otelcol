// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package actuator

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/gardener/gardener-extension-otelcol/pkg/apis/config"
)

// metricFilterBody is a filterprocessor body that drops a metric by name.
const metricFilterBody = `{"metrics":{"metric":["metric.name == \"foo\""]}}`

// logFilterBody is a filterprocessor body that drops a log record.
const logFilterBody = `{"logs":{"log_record":["true"]}}`

// target returns a Target serving the given signals with a default exporter.
func target(signals []config.SignalType) config.Target {
	return config.Target{
		Signals: signals,
	}
}

// filterTarget returns a Target serving the given signals and carrying the given
// raw filterprocessor body.
func filterTarget(body string, signals ...config.SignalType) config.Target {
	return config.Target{
		Signals: signals,
		Filters: runtime.RawExtension{Raw: []byte(body)},
	}
}

// exporterNamesFor builds the exporter-name map keyed by signal and target
// index for the given (signal -> target indices) mapping, using the default
// HTTP transport.
func exporterNamesFor(m map[config.SignalType][]int) exporterNamesBySignal {
	out := exporterNamesBySignal{}
	for sig, idxs := range m {
		out[sig] = map[int][]string{}
		for _, i := range idxs {
			out[sig][i] = []string{signalExporterName(sig, i, transportHTTP)}
		}
	}

	return out
}

var _ = Describe("filter processor", func() {
	Describe("buildPipelines pipeline wiring", func() {
		It("builds no pipelines when there are no targets", func() {
			pipelines := buildPipelines(config.CollectorConfig{}, nil)

			Expect(pipelines).To(BeEmpty())
		})

		It("wires each target signal with its receiver", func() {
			cfg := config.CollectorConfig{
				Spec: config.CollectorConfigSpec{
					Targets: []config.Target{
						target([]config.SignalType{
							config.SignalMetrics,
							config.SignalLogs,
							config.SignalEvents,
						}),
					},
				},
			}
			exporterNames := exporterNamesFor(map[config.SignalType][]int{
				config.SignalMetrics: {0},
				config.SignalLogs:    {0},
				config.SignalEvents:  {0},
			})
			pipelines := buildPipelines(cfg, exporterNames)

			Expect(pipelines).To(HaveLen(3))
			Expect(pipelines[signalPipelineName(config.SignalMetrics, 0)].Receivers).
				To(Equal([]string{"prometheus"}))
			Expect(pipelines[signalPipelineName(config.SignalLogs, 0)].Receivers).
				To(Equal([]string{"otlp"}))
			Expect(pipelines[signalPipelineName(config.SignalEvents, 0)].Receivers).
				To(Equal([]string{"k8sobjects/events"}))
		})

		It("builds one pipeline per target and wires each to its own exporter", func() {
			cfg := config.CollectorConfig{
				Spec: config.CollectorConfigSpec{
					Targets: []config.Target{
						{
							Signals: []config.SignalType{
								config.SignalMetrics,
							},
							Exporter: config.CollectorExportersConfig{
								OTLPHTTPExporter: &config.OTLPHTTPExporterConfig{
									Endpoint: "https://a:4318",
								},
							},
						},
						{
							Signals: []config.SignalType{
								config.SignalMetrics,
							},
							Exporter: config.CollectorExportersConfig{
								OTLPGRPCExporter: &config.OTLPGRPCExporterConfig{
									Endpoint: "https://b:4317",
								},
							},
						},
						{
							Signals: []config.SignalType{
								config.SignalMetrics,
							},
							Exporter: config.CollectorExportersConfig{
								DebugExporter: &config.DebugExporterConfig{},
							},
						},
					},
				},
			}
			exporterNames := map[config.SignalType]map[int][]string{
				config.SignalMetrics: {
					0: {signalExporterName(config.SignalMetrics, 0, transportHTTP)},
					1: {signalExporterName(config.SignalMetrics, 1, transportGRPC)},
					2: {signalExporterName(config.SignalMetrics, 2, transportDebug)},
				},
			}
			pipelines := buildPipelines(cfg, exporterNames)

			Expect(pipelines).To(HaveLen(3))
			Expect(pipelines[signalPipelineName(config.SignalMetrics, 0)].Exporters).
				To(Equal([]string{"otlphttp/metrics_0"}))
			Expect(pipelines[signalPipelineName(config.SignalMetrics, 1)].Exporters).
				To(Equal([]string{"otlp/metrics_1"}))
			Expect(pipelines[signalPipelineName(config.SignalMetrics, 2)].Exporters).
				To(Equal([]string{"debug/metrics_2"}))
		})

		It("does not wire a filter processor when the target has no filters", func() {
			cfg := config.CollectorConfig{
				Spec: config.CollectorConfigSpec{
					Targets: []config.Target{
						target(
							[]config.SignalType{
								config.SignalMetrics,
							},
						),
					},
				},
			}
			pipelines := buildPipelines(cfg, exporterNamesFor(map[config.SignalType][]int{
				config.SignalMetrics: {0},
			}))

			Expect(pipelines[signalPipelineName(config.SignalMetrics, 0)].Processors).
				To(Equal([]string{
					"resource",
					"memory_limiter",
					"batch",
				}))
		})

		It("wires the target filter after memory_limiter and before batch", func() {
			cfg := config.CollectorConfig{
				Spec: config.CollectorConfigSpec{
					Targets: []config.Target{
						filterTarget(metricFilterBody, config.SignalMetrics),
						filterTarget(logFilterBody, config.SignalEvents),
					},
				},
			}
			pipelines := buildPipelines(
				cfg,
				exporterNamesFor(map[config.SignalType][]int{
					config.SignalMetrics: {0},
					config.SignalEvents:  {1},
				}),
			)

			Expect(pipelines[signalPipelineName(config.SignalMetrics, 0)].Processors).
				To(Equal([]string{
					"resource",
					"memory_limiter",
					"filter/metrics_0",
					"batch",
				}))
			Expect(pipelines[signalPipelineName(config.SignalEvents, 1)].Processors).
				To(Equal([]string{
					"resource",
					"memory_limiter",
					"transform/events",
					"filter/events_1",
					"batch",
				}))
		})

		It("wires the target filter only into signals its body targets", func() {
			// A target serving both logs and metrics with a metrics-only filter:
			// the filter is wired into the metrics pipeline only, and the logs
			// pipeline stays free of a dead filter reference.
			cfg := config.CollectorConfig{
				Spec: config.CollectorConfigSpec{
					Targets: []config.Target{
						filterTarget(metricFilterBody, config.SignalLogs, config.SignalMetrics),
					},
				},
			}
			pipelines := buildPipelines(cfg, exporterNamesFor(map[config.SignalType][]int{
				config.SignalLogs:    {0},
				config.SignalMetrics: {0},
			}))

			Expect(pipelines[signalPipelineName(config.SignalMetrics, 0)].Processors).
				To(ContainElement("filter/metrics_0"))
			Expect(pipelines[signalPipelineName(config.SignalLogs, 0)].Processors).
				NotTo(ContainElement("filter/logs_0"))
			Expect(pipelines[signalPipelineName(config.SignalLogs, 0)].Processors).
				To(Equal([]string{"resource", "memory_limiter", "batch"}))
		})

		It("wires a logs-only filter into the logs pipeline only", func() {
			// The mirror of the metrics-only case: a logs-only filter on a target
			// serving both logs and metrics wires filter/logs/0 into the logs
			// pipeline and leaves the metrics pipeline untouched.
			cfg := config.CollectorConfig{
				Spec: config.CollectorConfigSpec{
					Targets: []config.Target{
						filterTarget(logFilterBody, config.SignalLogs, config.SignalMetrics),
					},
				},
			}
			pipelines := buildPipelines(cfg, exporterNamesFor(map[config.SignalType][]int{
				config.SignalLogs:    {0},
				config.SignalMetrics: {0},
			}))

			Expect(pipelines[signalPipelineName(config.SignalLogs, 0)].Processors).
				To(ContainElement("filter/logs_0"))
			Expect(pipelines[signalPipelineName(config.SignalMetrics, 0)].Processors).
				NotTo(ContainElement("filter/metrics_0"))
			Expect(pipelines[signalPipelineName(config.SignalMetrics, 0)].Processors).
				To(Equal([]string{"resource", "memory_limiter", "batch"}))
		})

		It("does not wire a filter processor when the filter body is empty", func() {
			// A target with no filter body produces no processor, so the pipeline
			// stays free of a dangling reference.
			cfg := config.CollectorConfig{
				Spec: config.CollectorConfigSpec{
					Targets: []config.Target{
						target([]config.SignalType{
							config.SignalMetrics,
						}),
					},
				},
			}
			pipelines := buildPipelines(cfg, exporterNamesFor(map[config.SignalType][]int{
				config.SignalMetrics: {0},
			}))

			Expect(pipelines[signalPipelineName(config.SignalMetrics, 0)].Processors).To(Equal(
				[]string{"resource", "memory_limiter", "batch"}))
		})
	})

	Describe("getOtelExporters transport fan-out", func() {
		It("fans a signal pipeline out to all of a target's enabled transports", func() {
			cfg := config.CollectorConfig{
				Spec: config.CollectorConfigSpec{
					Targets: []config.Target{
						{
							Signals: []config.SignalType{config.SignalMetrics},
							Exporter: config.CollectorExportersConfig{
								OTLPHTTPExporter: &config.OTLPHTTPExporterConfig{
									Endpoint: "https://a:4318",
								},
								OTLPGRPCExporter: &config.OTLPGRPCExporterConfig{
									Endpoint: "https://a:4317",
								},
							},
						},
					},
				},
			}
			exporters, exporterNames := (&Actuator{}).getOtelExporters(cfg)

			Expect(exporterNames[config.SignalMetrics][0]).
				To(Equal([]string{
					"otlphttp/metrics_0",
					"otlp/metrics_0",
				}))
			Expect(exporters).
				To(HaveKey("otlphttp/metrics_0"))
			Expect(exporters).
				To(HaveKey("otlp/metrics_0"))

			pipelines := buildPipelines(cfg, exporterNames)
			Expect(pipelines[signalPipelineName(config.SignalMetrics, 0)].Exporters).
				To(Equal([]string{
					"otlphttp/metrics_0", "otlp/metrics_0",
				}))
		})
	})
})

var _ = Describe("filterProcessorConfigsPerSignals", func() {
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
		out, err := filterProcessorConfigsPerSignals(filterTarget(""), allSignals)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(BeEmpty())
	})

	It("returns only the metrics signal for a metrics-only filter", func() {
		out, err := filterProcessorConfigsPerSignals(filterTarget(metricFilterBody), allSignals)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(HaveKey(config.SignalMetrics))
		Expect(out).NotTo(HaveKey(config.SignalLogs))
		Expect(out).NotTo(HaveKey(config.SignalEvents))
	})

	It("returns only the logs and events signals for a logs-only filter", func() {
		out, err := filterProcessorConfigsPerSignals(filterTarget(logFilterBody), allSignals)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(HaveKey(config.SignalLogs))
		Expect(out).To(HaveKey(config.SignalEvents))
		Expect(out).NotTo(HaveKey(config.SignalMetrics))
	})

	It("returns each applicable signal for a filter targeting both", func() {
		out, err := filterProcessorConfigsPerSignals(filterTarget(bothFilterBody), allSignals)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(HaveKey(config.SignalMetrics))
		Expect(out).To(HaveKey(config.SignalLogs))
		Expect(out).To(HaveKey(config.SignalEvents))
	})

	It("detects metrics via the metric_conditions field", func() {
		out, err := filterProcessorConfigsPerSignals(filterTarget(metricConditionFilterBody), allSignals)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(HaveKey(config.SignalMetrics))
		Expect(out).NotTo(HaveKey(config.SignalLogs))
	})

	It("detects logs via the log_conditions field", func() {
		out, err := filterProcessorConfigsPerSignals(filterTarget(logConditionFilterBody), allSignals)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(HaveKey(config.SignalLogs))
		Expect(out).NotTo(HaveKey(config.SignalMetrics))
	})

	It("returns an empty map for an unsupported (traces) filter", func() {
		out, err := filterProcessorConfigsPerSignals(filterTarget(traceFilterBody), allSignals)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(BeEmpty())
	})

	It("omits a targeted signal that is not in the requested signals", func() {
		// The filter targets metrics, but metrics is not requested, so it is
		// dropped.
		out, err := filterProcessorConfigsPerSignals(
			filterTarget(metricFilterBody),
			[]config.SignalType{config.SignalLogs},
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(BeEmpty())
	})

	It("maps each applicable signal to the full rendered body", func() {
		out, err := filterProcessorConfigsPerSignals(filterTarget(metricFilterBody), allSignals)
		Expect(err).NotTo(HaveOccurred())
		Expect(out[config.SignalMetrics]).To(Equal(map[string]any{
			"metrics": map[string]any{
				"metric": []any{`metric.name == "foo"`},
			},
		}))
	})

	It("returns an error for a malformed body", func() {
		_, err := filterProcessorConfigsPerSignals(filterTarget(`{not json`), allSignals)
		Expect(err).To(HaveOccurred())
	})
})
