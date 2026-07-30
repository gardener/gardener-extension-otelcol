// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package actuator

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gardener/gardener-extension-otelcol/pkg/apis/config"
)

var _ = Describe("filter processor", func() {
	Describe("buildPipelines pipeline wiring", func() {
		It("does not wire the filter processor when the filter is unset", func() {
			pipelines := buildPipelines(config.CollectorConfig{}, []string{debugExporterName})

			for name, p := range pipelines {
				Expect(p.Processors).NotTo(ContainElement(filterProcessorName), "pipeline %q", name)
			}
		})

		It("wires the filter processor into all pipelines after memory_limiter and before batch", func() {
			cfg := config.CollectorConfig{
				Spec: config.CollectorConfigSpec{
					Filter: &config.FilterConfig{
						Metrics: &config.MetricFilters{Metric: []string{`metric.name == "foo"`}},
					},
				},
			}

			pipelines := buildPipelines(cfg, []string{debugExporterName})

			Expect(pipelines[logsPipelineName].Processors).To(Equal(
				[]string{resourceProcessorName, memoryLimiterProcessorName, filterProcessorName, batchProcessorName}))
			Expect(pipelines[eventsPipelineName].Processors).To(Equal(
				[]string{resourceProcessorName, memoryLimiterProcessorName, transformEventsProcessorName, filterProcessorName, batchProcessorName}))
			Expect(pipelines[metricsPipelineName].Processors).To(Equal(
				[]string{resourceProcessorName, memoryLimiterProcessorName, filterProcessorName, batchProcessorName}))
		})
	})

	Describe("getFilterProcessorConfig rendering", func() {
		var a *Actuator

		BeforeEach(func() {
			a = &Actuator{}
		})

		It("renders the error_mode only when set", func() {
			Expect(a.getFilterProcessorConfig(config.FilterConfig{})).NotTo(HaveKey("error_mode"))
			Expect(a.getFilterProcessorConfig(config.FilterConfig{ErrorMode: config.FilterErrorModeIgnore})).
				To(HaveKeyWithValue("error_mode", "ignore"))
		})

		It("renders the metrics OTTL condition lists", func() {
			out := a.getFilterProcessorConfig(config.FilterConfig{
				Metrics: &config.MetricFilters{
					Resource:  []string{`resource.attributes["env"] == "dev"`},
					Metric:    []string{`metric.name == "foo"`},
					DataPoint: []string{`datapoint.value_int == 0`},
				},
			})

			Expect(out["metrics"]).To(Equal(map[string]any{
				"resource":  []string{`resource.attributes["env"] == "dev"`},
				"metric":    []string{`metric.name == "foo"`},
				"datapoint": []string{`datapoint.value_int == 0`},
			}))
		})

		It("renders the metrics include/exclude match properties", func() {
			out := a.getFilterProcessorConfig(config.FilterConfig{
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
						map[string]any{"key": "service.name", "value": "my-service"},
						map[string]any{"key": "present-only"},
					},
				},
			}))
		})

		It("renders the logs OTTL conditions and include match properties", func() {
			out := a.getFilterProcessorConfig(config.FilterConfig{
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
					"record_attributes": []any{map[string]any{"key": "foo", "value": "bar"}},
					"severity_number": map[string]any{
						"min":             "WARN",
						"match_undefined": false,
					},
				},
			}))
		})

		It("omits signal keys when their filters are empty", func() {
			out := a.getFilterProcessorConfig(config.FilterConfig{
				ErrorMode: config.FilterErrorModePropagate,
				Metrics:   &config.MetricFilters{},
				Logs:      &config.LogFilters{},
			})

			Expect(out).To(Equal(map[string]any{"error_mode": "propagate"}))
		})

		It("renders basic context-inferred conditions as bare strings", func() {
			out := a.getFilterProcessorConfig(config.FilterConfig{
				MetricConditions: []config.ContextConditions{
					{Conditions: []string{`metric.name == "foo"`, `metric.name == "bar"`}},
				},
			})

			Expect(out["metric_conditions"]).To(Equal([]any{
				`metric.name == "foo"`,
				`metric.name == "bar"`,
			}))
		})

		It("renders advanced context-inferred conditions as objects", func() {
			out := a.getFilterProcessorConfig(config.FilterConfig{
				LogConditions: []config.ContextConditions{
					{
						Context:    "resource",
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
	})
})
