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
func exporterNamesFor(m map[config.SignalType][]int) map[config.SignalType]map[int][]string {
	out := map[config.SignalType]map[int][]string{}
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
						target([]config.SignalType{config.SignalMetrics, config.SignalLogs, config.SignalEvents}),
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
			Expect(pipelines[signalPipelineName(config.SignalMetrics, 0)].Receivers).To(Equal([]string{prometheusReceiverName}))
			Expect(pipelines[signalPipelineName(config.SignalLogs, 0)].Receivers).To(Equal([]string{otlpReceiverName}))
			Expect(pipelines[signalPipelineName(config.SignalEvents, 0)].Receivers).To(Equal([]string{eventsReceiverName}))
			Expect(pipelines[signalPipelineName(config.SignalMetrics, 0)].Exporters).To(Equal([]string{signalExporterName(config.SignalMetrics, 0, transportHTTP)}))
		})

		It("builds one pipeline per target and wires each to its own exporter", func() {
			cfg := config.CollectorConfig{
				Spec: config.CollectorConfigSpec{
					Targets: []config.Target{
						{Signals: []config.SignalType{config.SignalMetrics}, Exporter: config.CollectorExportersConfig{OTLPHTTPExporter: &config.OTLPHTTPExporterConfig{Endpoint: "https://a:4318"}}},
						{Signals: []config.SignalType{config.SignalMetrics}, Exporter: config.CollectorExportersConfig{OTLPGRPCExporter: &config.OTLPGRPCExporterConfig{Endpoint: "https://b:4317"}}},
						{Signals: []config.SignalType{config.SignalMetrics}, Exporter: config.CollectorExportersConfig{DebugExporter: &config.DebugExporterConfig{}}},
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
			Expect(pipelines[signalPipelineName(config.SignalMetrics, 0)].Exporters).To(Equal([]string{"otlphttp/metrics/0"}))
			Expect(pipelines[signalPipelineName(config.SignalMetrics, 1)].Exporters).To(Equal([]string{"otlp/metrics/1"}))
			Expect(pipelines[signalPipelineName(config.SignalMetrics, 2)].Exporters).To(Equal([]string{"debug/metrics/2"}))
		})

		It("does not wire a filter processor when the target has no filters", func() {
			cfg := config.CollectorConfig{
				Spec: config.CollectorConfigSpec{
					Targets: []config.Target{target([]config.SignalType{config.SignalMetrics})},
				},
			}

			pipelines := buildPipelines(cfg, exporterNamesFor(map[config.SignalType][]int{
				config.SignalMetrics: {0},
			}))

			Expect(pipelines[signalPipelineName(config.SignalMetrics, 0)].Processors).To(Equal(
				[]string{resourceProcessorName, memoryLimiterProcessorName, batchProcessorName}))
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

			pipelines := buildPipelines(cfg, exporterNamesFor(map[config.SignalType][]int{
				config.SignalMetrics: {0},
				config.SignalEvents:  {1},
			}))

			Expect(pipelines[signalPipelineName(config.SignalMetrics, 0)].Processors).To(Equal([]string{
				resourceProcessorName, memoryLimiterProcessorName,
				signalFilterName(config.SignalMetrics, 0), batchProcessorName,
			}))
			Expect(pipelines[signalPipelineName(config.SignalEvents, 1)].Processors).To(Equal([]string{
				resourceProcessorName, memoryLimiterProcessorName, transformEventsProcessorName,
				signalFilterName(config.SignalEvents, 1), batchProcessorName,
			}))
		})

		It("wires the target filter into every signal the target serves", func() {
			// A target serving both logs and metrics: its single filter is wired
			// into both pipelines.
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

			Expect(pipelines[signalPipelineName(config.SignalMetrics, 0)].Processors).To(ContainElement(signalFilterName(config.SignalMetrics, 0)))
			Expect(pipelines[signalPipelineName(config.SignalLogs, 0)].Processors).To(ContainElement(signalFilterName(config.SignalLogs, 0)))
		})

		It("does not wire a filter processor when the filter body is empty", func() {
			// A target with no filter body produces no processor, so the pipeline
			// stays free of a dangling reference.
			cfg := config.CollectorConfig{
				Spec: config.CollectorConfigSpec{
					Targets: []config.Target{target([]config.SignalType{config.SignalMetrics})},
				},
			}

			pipelines := buildPipelines(cfg, exporterNamesFor(map[config.SignalType][]int{
				config.SignalMetrics: {0},
			}))

			Expect(pipelines[signalPipelineName(config.SignalMetrics, 0)].Processors).To(Equal(
				[]string{resourceProcessorName, memoryLimiterProcessorName, batchProcessorName}))
		})
	})

	Describe("getOtelExporters transport fan-out", func() {
		It("fans a signal pipeline out to all of a target's enabled transports", func() {
			cfg := config.CollectorConfig{
				Spec: config.CollectorConfigSpec{
					Targets: []config.Target{{
						Signals: []config.SignalType{config.SignalMetrics},
						Exporter: config.CollectorExportersConfig{
							OTLPHTTPExporter: &config.OTLPHTTPExporterConfig{Endpoint: "https://a:4318"},
							OTLPGRPCExporter: &config.OTLPGRPCExporterConfig{Endpoint: "https://a:4317"},
						},
					}},
				},
			}

			exporters, exporterNames := (&Actuator{}).getOtelExporters(cfg)

			Expect(exporterNames[config.SignalMetrics][0]).To(Equal([]string{
				signalExporterName(config.SignalMetrics, 0, transportHTTP),
				signalExporterName(config.SignalMetrics, 0, transportGRPC),
			}))
			Expect(exporters).To(HaveKey(signalExporterName(config.SignalMetrics, 0, transportHTTP)))
			Expect(exporters).To(HaveKey(signalExporterName(config.SignalMetrics, 0, transportGRPC)))

			pipelines := buildPipelines(cfg, exporterNames)
			Expect(pipelines[signalPipelineName(config.SignalMetrics, 0)].Exporters).To(Equal([]string{
				"otlphttp/metrics/0", "otlp/metrics/0",
			}))
		})
	})
})
