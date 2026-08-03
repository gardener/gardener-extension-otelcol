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
	configWithSignals := func(enabled ...config.SignalType) config.CollectorConfig {
		signals := config.SignalsConfig{}
		for _, sig := range enabled {
			s := config.SignalConfig{Enabled: new(true), Targets: []config.SignalTarget{{}}}
			switch sig {
			case config.SignalMetrics:
				signals.Metrics = s
			case config.SignalLogs:
				signals.Logs = s
			case config.SignalTraces:
				signals.Traces = s
			case config.SignalProfiles:
				signals.Profiles = s
			case config.SignalEvents:
				signals.Events = s
			default:
			}
		}

		return config.CollectorConfig{Spec: config.CollectorConfigSpec{Signals: signals}}
	}

	DescribeTable("buildPipelines includes exactly the pipelines for the enabled signals",
		func(cfg config.CollectorConfig, wantPipelines []string) {
			pipelines := buildPipelines(cfg, nil)
			Expect(pipelines).To(HaveLen(len(wantPipelines)))
			for _, name := range wantPipelines {
				Expect(pipelines).To(HaveKey(name))
			}
		},
		Entry("no signal enabled builds no pipelines",
			configWithSignals(),
			[]string{},
		),
		Entry("metrics only",
			configWithSignals(config.SignalMetrics),
			[]string{signalPipelineName(config.SignalMetrics, 0)},
		),
		Entry("logs and events only",
			configWithSignals(config.SignalLogs, config.SignalEvents),
			[]string{signalPipelineName(config.SignalLogs, 0), signalPipelineName(config.SignalEvents, 0)},
		),
		Entry("all signals",
			configWithSignals(config.SignalMetrics, config.SignalLogs, config.SignalTraces, config.SignalProfiles, config.SignalEvents),
			[]string{
				signalPipelineName(config.SignalMetrics, 0),
				signalPipelineName(config.SignalLogs, 0),
				signalPipelineName(config.SignalTraces, 0),
				signalPipelineName(config.SignalProfiles, 0),
				signalPipelineName(config.SignalEvents, 0),
			},
		),
	)

	It("wires the enabled receivers and per-target exporters into each pipeline", func() {
		cfg := configWithSignals(config.SignalMetrics, config.SignalLogs, config.SignalEvents)
		exporterNames := map[config.SignalType][]string{
			config.SignalMetrics: {signalExporterName(config.SignalMetrics, 0, config.ExporterProtocolHTTP)},
			config.SignalLogs:    {signalExporterName(config.SignalLogs, 0, config.ExporterProtocolHTTP)},
			config.SignalEvents:  {signalExporterName(config.SignalEvents, 0, config.ExporterProtocolHTTP)},
		}

		pipelines := buildPipelines(cfg, exporterNames)

		Expect(pipelines[signalPipelineName(config.SignalLogs, 0)].Receivers).To(Equal([]string{otlpReceiverName}))
		Expect(pipelines[signalPipelineName(config.SignalEvents, 0)].Receivers).To(Equal([]string{eventsReceiverName}))
		Expect(pipelines[signalPipelineName(config.SignalMetrics, 0)].Receivers).To(Equal([]string{prometheusReceiverName}))

		Expect(pipelines[signalPipelineName(config.SignalMetrics, 0)].Exporters).To(Equal(exporterNames[config.SignalMetrics]))
		Expect(pipelines[signalPipelineName(config.SignalLogs, 0)].Exporters).To(Equal(exporterNames[config.SignalLogs]))
		Expect(pipelines[signalPipelineName(config.SignalEvents, 0)].Exporters).To(Equal(exporterNames[config.SignalEvents]))
	})
})
