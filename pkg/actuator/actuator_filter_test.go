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

// enabledSignal returns a SignalConfig that is enabled with a single target
// whose exporter inherits the default, optionally carrying the given filters.
func enabledSignal(filters ...config.FilterRule) config.SignalConfig {
	return config.SignalConfig{
		Enabled: new(true),
		Targets: []config.SignalTarget{{Filters: filters}},
	}
}

var _ = Describe("filter processor", func() {
	Describe("buildPipelines pipeline wiring", func() {
		It("builds no pipelines when no signal is enabled", func() {
			pipelines := buildPipelines(config.CollectorConfig{}, nil)

			Expect(pipelines).To(BeEmpty())
		})

		It("wires only the enabled signals with their receivers", func() {
			cfg := config.CollectorConfig{
				Spec: config.CollectorConfigSpec{
					Signals: config.SignalsConfig{
						Metrics: enabledSignal(),
						Logs:    enabledSignal(),
						Events:  enabledSignal(),
					},
				},
			}
			exporterNames := map[config.SignalType][]string{
				config.SignalMetrics: {signalExporterName(config.SignalMetrics, 0, config.ExporterProtocolHTTP)},
				config.SignalLogs:    {signalExporterName(config.SignalLogs, 0, config.ExporterProtocolHTTP)},
				config.SignalEvents:  {signalExporterName(config.SignalEvents, 0, config.ExporterProtocolHTTP)},
			}

			pipelines := buildPipelines(cfg, exporterNames)

			Expect(pipelines).To(HaveLen(3))
			Expect(pipelines[signalPipelineName(config.SignalMetrics, 0)].Receivers).To(Equal([]string{prometheusReceiverName}))
			Expect(pipelines[signalPipelineName(config.SignalLogs, 0)].Receivers).To(Equal([]string{otlpReceiverName}))
			Expect(pipelines[signalPipelineName(config.SignalEvents, 0)].Receivers).To(Equal([]string{eventsReceiverName}))
			Expect(pipelines[signalPipelineName(config.SignalMetrics, 0)].Exporters).To(Equal(exporterNames[config.SignalMetrics]))
		})

		It("uses the OTLP receiver for traces and profiles", func() {
			cfg := config.CollectorConfig{
				Spec: config.CollectorConfigSpec{
					Signals: config.SignalsConfig{
						Traces:   enabledSignal(),
						Profiles: enabledSignal(),
					},
				},
			}

			pipelines := buildPipelines(cfg, nil)

			Expect(pipelines[signalPipelineName(config.SignalTraces, 0)].Receivers).To(Equal([]string{otlpReceiverName}))
			Expect(pipelines[signalPipelineName(config.SignalProfiles, 0)].Receivers).To(Equal([]string{otlpReceiverName}))
		})

		It("builds one pipeline per target and wires each to its own exporter", func() {
			cfg := config.CollectorConfig{
				Spec: config.CollectorConfigSpec{
					Signals: config.SignalsConfig{
						Metrics: config.SignalConfig{
							Enabled: new(true),
							Targets: []config.SignalTarget{
								{Exporter: config.ExporterConfig{Endpoint: "https://a:4318"}},
								{Exporter: config.ExporterConfig{Protocol: config.ExporterProtocolGRPC, Endpoint: "https://b:4317"}},
								{Exporter: config.ExporterConfig{Protocol: config.ExporterProtocolDebug}},
							},
						},
					},
				},
			}
			exporterNames := map[config.SignalType][]string{
				config.SignalMetrics: {
					signalExporterName(config.SignalMetrics, 0, config.ExporterProtocolHTTP),
					signalExporterName(config.SignalMetrics, 1, config.ExporterProtocolGRPC),
					signalExporterName(config.SignalMetrics, 2, config.ExporterProtocolDebug),
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
					Signals: config.SignalsConfig{Metrics: enabledSignal()},
				},
			}

			pipelines := buildPipelines(cfg, map[config.SignalType][]string{
				config.SignalMetrics: {signalExporterName(config.SignalMetrics, 0, config.ExporterProtocolHTTP)},
			})

			Expect(pipelines[signalPipelineName(config.SignalMetrics, 0)].Processors).To(Equal(
				[]string{resourceProcessorName, memoryLimiterProcessorName, batchProcessorName}))
		})

		It("wires per-target filters after memory_limiter and before batch", func() {
			cfg := config.CollectorConfig{
				Spec: config.CollectorConfigSpec{
					Signals: config.SignalsConfig{
						Metrics: enabledSignal(config.FilterRule{
							Metrics: &config.MetricFilters{Metric: []string{metricNameFooCondition}},
						}),
						Events: enabledSignal(config.FilterRule{
							Logs: &config.LogFilters{LogRecord: []string{`true`}},
						}),
					},
				},
			}

			pipelines := buildPipelines(cfg, map[config.SignalType][]string{
				config.SignalMetrics: {signalExporterName(config.SignalMetrics, 0, config.ExporterProtocolHTTP)},
				config.SignalEvents:  {signalExporterName(config.SignalEvents, 0, config.ExporterProtocolHTTP)},
			})

			Expect(pipelines[signalPipelineName(config.SignalMetrics, 0)].Processors).To(Equal([]string{
				resourceProcessorName, memoryLimiterProcessorName,
				signalFilterName(config.SignalMetrics, 0, 0), batchProcessorName,
			}))
			Expect(pipelines[signalPipelineName(config.SignalEvents, 0)].Processors).To(Equal([]string{
				resourceProcessorName, memoryLimiterProcessorName, transformEventsProcessorName,
				signalFilterName(config.SignalEvents, 0, 0), batchProcessorName,
			}))
		})
	})

	Describe("getFilterProcessorConfig rendering", func() {
		var a *Actuator

		BeforeEach(func() {
			a = &Actuator{}
		})

		It("renders the error_mode only when set", func() {
			Expect(a.getFilterProcessorConfig(config.FilterRule{})).NotTo(HaveKey("error_mode"))
			Expect(a.getFilterProcessorConfig(config.FilterRule{ErrorMode: config.FilterErrorModeIgnore})).
				To(HaveKeyWithValue("error_mode", "ignore"))
		})

		It("renders the metrics OTTL condition lists", func() {
			out := a.getFilterProcessorConfig(config.FilterRule{
				Metrics: &config.MetricFilters{
					Resource:  []string{`resource.attributes["env"] == "dev"`},
					Metric:    []string{metricNameFooCondition},
					DataPoint: []string{`datapoint.value_int == 0`},
				},
			})

			Expect(out["metrics"]).To(Equal(map[string]any{
				resourceProcessorName: []string{`resource.attributes["env"] == "dev"`},
				"metric":              []string{metricNameFooCondition},
				"datapoint":           []string{`datapoint.value_int == 0`},
			}))
		})

		It("renders the metrics include/exclude match properties", func() {
			out := a.getFilterProcessorConfig(config.FilterRule{
				Metrics: &config.MetricFilters{
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
			})

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
				Logs: &config.LogFilters{
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
			})

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

		It("omits signal keys when their filters are empty", func() {
			out := a.getFilterProcessorConfig(config.FilterRule{
				ErrorMode: config.FilterErrorModePropagate,
				Metrics:   &config.MetricFilters{},
				Logs:      &config.LogFilters{},
			})

			Expect(out).To(Equal(map[string]any{"error_mode": "propagate"}))
		})

		It("renders basic context-inferred conditions as bare strings", func() {
			out := a.getFilterProcessorConfig(config.FilterRule{
				MetricConditions: []config.ContextConditions{
					{Conditions: []string{metricNameFooCondition, `metric.name == "bar"`}},
				},
			})

			Expect(out["metric_conditions"]).To(Equal([]any{
				metricNameFooCondition,
				`metric.name == "bar"`,
			}))
		})

		It("renders advanced context-inferred conditions as objects", func() {
			out := a.getFilterProcessorConfig(config.FilterRule{
				LogConditions: []config.ContextConditions{
					{
						Context:    resourceProcessorName,
						Conditions: []string{`attributes["k"] == "v"`},
						ErrorMode:  config.FilterErrorModeSilent,
					},
				},
			})

			Expect(out["log_conditions"]).To(Equal([]any{
				map[string]any{
					"context":    "resource",
					"conditions": []string{`attributes["k"] == "v"`},
					"error_mode": "silent",
				},
			}))
		})

		It("renders the signal-agnostic conditions form", func() {
			out := a.getFilterProcessorConfig(config.FilterRule{
				Conditions: []config.ContextConditions{
					{Conditions: []string{`resource.attributes["k8s.namespace.name"] == "kube-system"`}},
				},
			})

			Expect(out["conditions"]).To(Equal([]any{
				`resource.attributes["k8s.namespace.name"] == "kube-system"`,
			}))
		})
	})
})
