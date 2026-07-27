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
	configWithSignals := func(signals ...config.SignalType) config.CollectorConfig {
		return config.CollectorConfig{
			Spec: config.CollectorConfigSpec{
				Signals: signals,
			},
		}
	}

	DescribeTable("buildPipelines includes exactly the pipelines for the enabled signals",
		func(cfg config.CollectorConfig, wantPipelines []string) {
			pipelines := buildPipelines(cfg, []string{"debug"})
			Expect(pipelines).To(HaveLen(len(wantPipelines)))
			for _, name := range wantPipelines {
				Expect(pipelines).To(HaveKey(name))
			}
		},
		Entry("empty selection enables all pipelines",
			configWithSignals(),
			[]string{"logs", "logs/events", "metrics"},
		),
		Entry("metrics only",
			configWithSignals(config.SignalMetrics),
			[]string{"metrics"},
		),
		Entry("logs and events only",
			configWithSignals(config.SignalLogs, config.SignalEvents),
			[]string{"logs", "logs/events"},
		),
		Entry("events only",
			configWithSignals(config.SignalEvents),
			[]string{"logs/events"},
		),
	)

	It("wires the enabled receivers and exporters into each pipeline", func() {
		pipelines := buildPipelines(configWithSignals(), []string{"debug", "otlp_http"})

		Expect(pipelines["logs"].Receivers).To(Equal([]string{"otlp"}))
		Expect(pipelines["logs/events"].Receivers).To(Equal([]string{"k8sobjects/events"}))
		Expect(pipelines["metrics"].Receivers).To(Equal([]string{"prometheus"}))

		for _, p := range pipelines {
			Expect(p.Exporters).To(Equal([]string{"debug", "otlp_http"}))
		}
	})
})
