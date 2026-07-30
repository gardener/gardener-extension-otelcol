// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package actuator

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gardener/gardener-extension-otelcol/pkg/apis/config"
)

var _ = Describe("parseShootNamespaceAttributes", func() {
	DescribeTable("should parse the namespace into OTel resource attributes",
		func(namespace, wantCluster, wantProject, wantShoot string) {
			cluster, project, shoot := parseShootNamespaceAttributes(namespace)

			Expect(cluster).To(Equal(wantCluster))
			Expect(project).To(Equal(wantProject))
			Expect(shoot).To(Equal(wantShoot))
		},
		Entry("standard shoot namespace",
			"shoot--my-project--my-shoot",
			"shoot--my-project--my-shoot", "my-project", "my-shoot",
		),
		Entry("shoot name containing hyphens",
			"shoot--local--my-complex-shoot-name",
			"shoot--local--my-complex-shoot-name", "local", "my-complex-shoot-name",
		),
		Entry("non-shoot namespace returns empty project and shoot",
			"kube-system",
			"kube-system", "", "",
		),
		Entry("only two segments returns empty project and shoot",
			"shoot--local",
			"shoot--local", "", "",
		),
	)
})

var _ = Describe("signal selection", func() {
	// configWithSignals builds a config with a single target serving the given
	// signals.
	configWithSignals := func(enabled ...config.SignalType) config.CollectorConfig {
		if len(enabled) == 0 {
			return config.CollectorConfig{}
		}

		return config.CollectorConfig{Spec: config.CollectorConfigSpec{
			Targets: []config.Target{{Signals: enabled}},
		}}
	}

	DescribeTable("buildPipelines includes exactly the pipelines for the served signals",
		func(cfg config.CollectorConfig, wantPipelines []string) {
			pipelines := buildPipelines(cfg, nil)
			Expect(pipelines).To(HaveLen(len(wantPipelines)))
			for _, name := range wantPipelines {
				Expect(pipelines).To(HaveKey(name))
			}
		},
		Entry("no target builds no pipelines",
			configWithSignals(),
			[]string{},
		),
		Entry("metrics only",
			configWithSignals(config.SignalMetrics),
			[]string{"metrics/0"},
		),
		Entry("logs and events only",
			configWithSignals(config.SignalLogs, config.SignalEvents),
			[]string{"logs/0", "logs/events/0"},
		),
		Entry("all signals",
			configWithSignals(config.SignalMetrics, config.SignalLogs, config.SignalEvents),
			[]string{
				"metrics/0",
				"logs/0",
				"logs/events/0",
			},
		),
		Entry("a target with no signals set defaults to all signals",
			config.CollectorConfig{Spec: config.CollectorConfigSpec{
				Targets: []config.Target{{Signals: nil}},
			}},
			[]string{
				"metrics/0",
				"logs/0",
				"logs/events/0",
			},
		),
	)

	It("wires the served receivers and per-target exporters into each pipeline", func() {
		cfg := configWithSignals(config.SignalMetrics, config.SignalLogs, config.SignalEvents)
		exporterNames := map[config.SignalType]map[int][]string{
			config.SignalMetrics: {0: {signalExporterName(config.SignalMetrics, 0, transportHTTP)}},
			config.SignalLogs:    {0: {signalExporterName(config.SignalLogs, 0, transportHTTP)}},
			config.SignalEvents:  {0: {signalExporterName(config.SignalEvents, 0, transportHTTP)}},
		}

		pipelines := buildPipelines(cfg, exporterNames)

		Expect(pipelines[signalPipelineName(config.SignalLogs, 0)].Receivers).
			To(Equal([]string{"otlp"}))
		Expect(pipelines[signalPipelineName(config.SignalEvents, 0)].Receivers).
			To(Equal([]string{"k8sobjects/events"}))
		Expect(pipelines[signalPipelineName(config.SignalMetrics, 0)].Receivers).
			To(Equal([]string{"prometheus"}))

		Expect(pipelines[signalPipelineName(config.SignalMetrics, 0)].Exporters).
			To(Equal([]string{"otlphttp/metrics/0"}))
		Expect(pipelines[signalPipelineName(config.SignalLogs, 0)].Exporters).
			To(Equal([]string{"otlphttp/logs/0"}))
		Expect(pipelines[signalPipelineName(config.SignalEvents, 0)].Exporters).
			To(Equal([]string{"otlphttp/events/0"}))
	})
})
