// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package actuator

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gardener/gardener-extension-otelcol/pkg/apis/config"
)

const metricNameFooCondition = `metric.name == "foo"`

// trueCondition is a trivially-true OTTL condition used across filter specs.
const trueCondition = `true`

// target returns a Target serving the given signals, with a default exporter
// and optionally carrying the given filters.
func target(signals []config.SignalType, filters ...config.FilterRule) config.Target {
	return config.Target{
		Signals: signals,
		Filters: filters,
	}
}

// exporterNamesFor builds the exporter-name map keyed by signal and target
// index for the given (signal -> target indices) mapping, using the default
// HTTP protocol.
func exporterNamesFor(m map[config.SignalType][]int) map[config.SignalType]map[int]string {
	out := map[config.SignalType]map[int]string{}
	for sig, idxs := range m {
		out[sig] = map[int]string{}
		for _, i := range idxs {
			out[sig][i] = signalExporterName(sig, i, config.ExporterProtocolHTTP)
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
			Expect(pipelines[signalPipelineName(config.SignalMetrics, 0)].Exporters).To(Equal([]string{signalExporterName(config.SignalMetrics, 0, config.ExporterProtocolHTTP)}))
		})

		It("builds one pipeline per target and wires each to its own exporter", func() {
			cfg := config.CollectorConfig{
				Spec: config.CollectorConfigSpec{
					Targets: []config.Target{
						{Signals: []config.SignalType{config.SignalMetrics}, Exporter: config.ExporterConfig{Endpoint: "https://a:4318"}},
						{Signals: []config.SignalType{config.SignalMetrics}, Exporter: config.ExporterConfig{Protocol: config.ExporterProtocolGRPC, Endpoint: "https://b:4317"}},
						{Signals: []config.SignalType{config.SignalMetrics}, Exporter: config.ExporterConfig{Protocol: config.ExporterProtocolDebug}},
					},
				},
			}
			exporterNames := map[config.SignalType]map[int]string{
				config.SignalMetrics: {
					0: signalExporterName(config.SignalMetrics, 0, config.ExporterProtocolHTTP),
					1: signalExporterName(config.SignalMetrics, 1, config.ExporterProtocolGRPC),
					2: signalExporterName(config.SignalMetrics, 2, config.ExporterProtocolDebug),
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

		It("wires per-target filters after memory_limiter and before batch", func() {
			cfg := config.CollectorConfig{
				Spec: config.CollectorConfigSpec{
					Targets: []config.Target{
						target([]config.SignalType{config.SignalMetrics}, config.FilterRule{
							Metrics: &config.FilterMetrics{Metric: []string{metricNameFooCondition}},
						}),
						target([]config.SignalType{config.SignalEvents}, config.FilterRule{
							Logs: &config.FilterLogs{LogRecord: []string{trueCondition}},
						}),
					},
				},
			}

			pipelines := buildPipelines(cfg, exporterNamesFor(map[config.SignalType][]int{
				config.SignalMetrics: {0},
				config.SignalEvents:  {1},
			}))

			Expect(pipelines[signalPipelineName(config.SignalMetrics, 0)].Processors).To(Equal([]string{
				resourceProcessorName, memoryLimiterProcessorName,
				signalFilterName(config.SignalMetrics, 0, 0), batchProcessorName,
			}))
			Expect(pipelines[signalPipelineName(config.SignalEvents, 1)].Processors).To(Equal([]string{
				resourceProcessorName, memoryLimiterProcessorName, transformEventsProcessorName,
				signalFilterName(config.SignalEvents, 1, 0), batchProcessorName,
			}))
		})

		It("only wires filters whose block matches the pipeline signal", func() {
			// A target serving both logs and metrics with a metrics-only filter
			// rule: the rule must appear in the metrics pipeline only.
			cfg := config.CollectorConfig{
				Spec: config.CollectorConfigSpec{
					Targets: []config.Target{
						target([]config.SignalType{config.SignalLogs, config.SignalMetrics}, config.FilterRule{
							Metrics: &config.FilterMetrics{Metric: []string{metricNameFooCondition}},
						}),
					},
				},
			}

			pipelines := buildPipelines(cfg, exporterNamesFor(map[config.SignalType][]int{
				config.SignalLogs:    {0},
				config.SignalMetrics: {0},
			}))

			Expect(pipelines[signalPipelineName(config.SignalMetrics, 0)].Processors).To(ContainElement(signalFilterName(config.SignalMetrics, 0, 0)))
			Expect(pipelines[signalPipelineName(config.SignalLogs, 0)].Processors).To(Equal(
				[]string{resourceProcessorName, memoryLimiterProcessorName, batchProcessorName}))
		})
	})

	Describe("getFilterProcessorConfig rendering", func() {
		var a *Actuator

		BeforeEach(func() {
			a = &Actuator{}
		})

		It("renders the error_mode only when set", func() {
			Expect(a.getFilterProcessorConfig(config.FilterRule{}, config.SignalMetrics)).NotTo(HaveKey("error_mode"))
			Expect(a.getFilterProcessorConfig(config.FilterRule{ErrorMode: config.FilterErrorModeIgnore}, config.SignalMetrics)).
				To(HaveKeyWithValue("error_mode", "ignore"))
		})

		It("renders the metrics OTTL condition lists", func() {
			out := a.getFilterProcessorConfig(config.FilterRule{
				Metrics: &config.FilterMetrics{
					Resource:  []string{`resource.attributes["env"] == "dev"`},
					Metric:    []string{metricNameFooCondition},
					DataPoint: []string{`datapoint.value_int == 0`},
				},
			}, config.SignalMetrics)

			Expect(out["metrics"]).To(Equal(map[string]any{
				resourceProcessorName: []string{`resource.attributes["env"] == "dev"`},
				"metric":              []string{metricNameFooCondition},
				"datapoint":           []string{`datapoint.value_int == 0`},
			}))
		})

		It("renders the metrics include/exclude match properties", func() {
			out := a.getFilterProcessorConfig(config.FilterRule{
				Metrics: &config.FilterMetrics{
					Include: &config.MetricMatchProperties{
						MatchType:   config.MatchTypeStrict,
						MetricNames: []string{"metric.a", "metric.b"},
						Regexp: &config.RegexpConfig{
							CacheEnabled:       new(true),
							CacheMaxNumEntries: 10,
						},
						ResourceAttributes: []config.FilterAttribute{
							{Key: "service.name", Value: "my-service"},
							{Key: "present-only"},
						},
					},
				},
			}, config.SignalMetrics)

			Expect(out["metrics"]).To(Equal(map[string]any{
				"include": map[string]any{
					"match_type":   "strict",
					"metric_names": []string{"metric.a", "metric.b"},
					"regexp": map[string]any{
						"cacheenabled":       true,
						"cachemaxnumentries": 10,
					},
					"resource_attributes": []any{
						map[string]any{configKeyKey: "service.name", "value": "my-service"},
						map[string]any{configKeyKey: "present-only"},
					},
				},
			}))
		})

		It("renders the logs OTTL conditions and include match properties", func() {
			out := a.getFilterProcessorConfig(config.FilterRule{
				Logs: &config.FilterLogs{
					LogRecord: []string{`IsMatch(log.body, ".*password.*")`},
					Include: &config.LogMatchProperties{
						MatchType:     config.MatchTypeStrict,
						SeverityTexts: []string{"INFO", "DEBUG"},
						Bodies:        []string{"exact body"},
						SeverityNumber: &config.LogSeverityNumberMatchProperties{
							Min:            "WARN",
							MatchUndefined: new(false),
						},
						RecordAttributes: []config.FilterAttribute{{Key: "foo", Value: "bar"}},
					},
				},
			}, config.SignalLogs)

			Expect(out["logs"]).To(Equal(map[string]any{
				"log_record": []string{`IsMatch(log.body, ".*password.*")`},
				"include": map[string]any{
					"match_type":        "strict",
					"severity_texts":    []string{"INFO", "DEBUG"},
					"bodies":            []string{"exact body"},
					"record_attributes": []any{map[string]any{configKeyKey: "foo", "value": "bar"}},
					"severity_number": map[string]any{
						"min":             "WARN",
						"match_undefined": false,
					},
				},
			}))
		})

		It("renders only the metrics block on the metrics signal, ignoring the logs block", func() {
			out := a.getFilterProcessorConfig(config.FilterRule{
				Metrics: &config.FilterMetrics{Metric: []string{metricNameFooCondition}},
				Logs:    &config.FilterLogs{LogRecord: []string{trueCondition}}, // ignored on the metrics signal
			}, config.SignalMetrics)

			Expect(out).To(HaveKey("metrics"))
			Expect(out).NotTo(HaveKey("logs"))
		})

		It("renders the logs block for the events signal", func() {
			out := a.getFilterProcessorConfig(config.FilterRule{
				Logs: &config.FilterLogs{LogRecord: []string{trueCondition}},
			}, config.SignalEvents)

			Expect(out).To(HaveKey("logs"))
			Expect(out).NotTo(HaveKey("metrics"))
		})

		It("omits signal keys when the blocks are empty", func() {
			out := a.getFilterProcessorConfig(config.FilterRule{
				ErrorMode: config.FilterErrorModePropagate,
			}, config.SignalMetrics)

			Expect(out).To(Equal(map[string]any{"error_mode": "propagate"}))
		})

		It("renders basic context-inferred conditions as bare strings", func() {
			out := a.getFilterProcessorConfig(config.FilterRule{
				Metrics: &config.FilterMetrics{
					MetricConditions: []config.ContextConditions{
						{Conditions: []string{metricNameFooCondition, `metric.name == "bar"`}},
					},
				},
			}, config.SignalMetrics)

			Expect(out["metric_conditions"]).To(Equal([]any{
				metricNameFooCondition,
				`metric.name == "bar"`,
			}))
		})

		It("renders advanced context-inferred conditions as objects", func() {
			out := a.getFilterProcessorConfig(config.FilterRule{
				Logs: &config.FilterLogs{
					LogConditions: []config.ContextConditions{
						{
							Context:    resourceProcessorName,
							Conditions: []string{`attributes["k"] == "v"`},
							ErrorMode:  config.FilterErrorModeSilent,
						},
					},
				},
			}, config.SignalLogs)

			Expect(out["log_conditions"]).To(Equal([]any{
				map[string]any{
					"context":    "resource",
					"conditions": []string{`attributes["k"] == "v"`},
					"error_mode": "silent",
				},
			}))
		})
	})
})
