// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package actuator_test

import (
	"encoding/json"

	corev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	gardenerfeatures "github.com/gardener/gardener/pkg/features"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/component-base/featuregate"
	"k8s.io/utils/ptr"

	"github.com/gardener/gardener-extension-otelcol/pkg/actuator"
	"github.com/gardener/gardener-extension-otelcol/pkg/apis/config"
)

const localName = "local"

var _ = Describe("Actuator", Ordered, func() {
	var (
		// The serialized objects
		providerConfigData, cloudProfileData, seedData, shootData []byte

		extResource  *extensionsv1alpha1.Extension
		cluster      *extensionsv1alpha1.Cluster
		decoder      = serializer.NewCodecFactory(scheme.Scheme, serializer.EnableStrict).UniversalDecoder()
		featureGates = map[featuregate.Feature]bool{
			gardenerfeatures.OpenTelemetryCollector: true,
		}
		actuatorOpts   []actuator.Option
		providerConfig = config.CollectorConfig{
			Spec: config.CollectorConfigSpec{
				Exporters: config.CollectorExportersConfig{
					DebugExporter: config.DebugExporterConfig{
						Enabled:   new(true),
						Verbosity: config.DebugExporterVerbosityNormal,
					},
				},
			},
		}

		projectNamespace = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "garden-local",
			},
		}
		shootNamespace = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "shoot--local--local",
			},
		}
		cloudProfile = &corev1beta1.CloudProfile{
			ObjectMeta: metav1.ObjectMeta{
				Name: localName,
			},
			Spec: corev1beta1.CloudProfileSpec{
				Type: localName,
			},
		}
		seed = &corev1beta1.Seed{
			ObjectMeta: metav1.ObjectMeta{
				Name: localName,
			},
			Spec: corev1beta1.SeedSpec{
				Ingress: &corev1beta1.Ingress{
					Domain: "ingress.local.seed.local.gardener.cloud",
				},
				Provider: corev1beta1.SeedProvider{
					Type:   localName,
					Region: localName,
					Zones:  []string{"0"},
				},
			},
		}
		shoot = &corev1beta1.Shoot{
			ObjectMeta: metav1.ObjectMeta{
				Name:      localName,
				Namespace: projectNamespace.Name,
			},
			Spec: corev1beta1.ShootSpec{
				SeedName: new(localName),
				Provider: corev1beta1.Provider{
					Type: localName,
				},
				Region: localName,
			},
		}
	)

	BeforeAll(func() {
		actuatorOpts = []actuator.Option{
			actuator.WithGardenerVersion("1.0.0"),
			actuator.WithDecoder(decoder),
			actuator.WithGardenletFeatures(featureGates),
		}

		// Serialize our test objects, so we can later re-use them.
		var err error
		cloudProfileData, err = json.Marshal(cloudProfile)
		Expect(err).NotTo(HaveOccurred())
		seedData, err = json.Marshal(seed)
		Expect(err).NotTo(HaveOccurred())
		shootData, err = json.Marshal(shoot)
		Expect(err).NotTo(HaveOccurred())
		providerConfigData, err = json.Marshal(providerConfig)
		Expect(err).NotTo(HaveOccurred())

		Expect(k8sClient.Create(ctx, projectNamespace)).To(Succeed())
		Expect(k8sClient.Create(ctx, shootNamespace)).To(Succeed())
	})

	BeforeEach(func() {
		extResource = &extensionsv1alpha1.Extension{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "example",
				Namespace: shootNamespace.Name,
			},
			Spec: extensionsv1alpha1.ExtensionSpec{
				DefaultSpec: extensionsv1alpha1.DefaultSpec{
					Type:  actuator.ExtensionType,
					Class: ptr.To(extensionsv1alpha1.ExtensionClassShoot),
				},
			},
		}

		cluster = &extensionsv1alpha1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: shootNamespace.Name,
			},
			Spec: extensionsv1alpha1.ClusterSpec{
				CloudProfile: runtime.RawExtension{
					Raw: cloudProfileData,
				},
				Seed: runtime.RawExtension{
					Raw: seedData,
				},
				Shoot: runtime.RawExtension{
					Raw: shootData,
				},
			},
		}

		Expect(k8sClient.Create(ctx, cluster)).To(Succeed())
	})

	AfterEach(func() {
		Expect(k8sClient.Delete(ctx, cluster)).To(Succeed())
	})

	It("should successfully create an actuator", func() {
		act, err := actuator.New(k8sClient, actuatorOpts...)

		Expect(err).NotTo(HaveOccurred())
		Expect(act).NotTo(BeNil())
		Expect(act.Name()).To(Equal(actuator.Name))
		Expect(act.ExtensionType()).To(Equal(actuator.ExtensionType))
		Expect(act.FinalizerSuffix()).To(Equal(actuator.FinalizerSuffix))
		Expect(act.ExtensionClass()).To(Equal(extensionsv1alpha1.ExtensionClassShoot))
	})

	It("should fail to reconcile when no cluster exists", func() {
		// Change namespace of the extension resource, so that a
		// non-existing cluster is looked up.
		extResource.Namespace = "non-existing-namespace"

		act, err := actuator.New(k8sClient, actuatorOpts...)
		Expect(err).NotTo(HaveOccurred())
		Expect(act).NotTo(BeNil())
		err = act.Reconcile(ctx, logger, extResource)
		Expect(err).Should(HaveOccurred())
		Expect(err).To(MatchError(ContainSubstring("failed to get cluster")))
	})

	It("should fail to reconcile without provider config", func() {
		act, err := actuator.New(k8sClient, actuatorOpts...)
		Expect(err).NotTo(HaveOccurred())
		Expect(act).NotTo(BeNil())

		err = act.Reconcile(ctx, logger, extResource)
		Expect(err).Should(HaveOccurred())
		Expect(err).To(MatchError(ContainSubstring("no provider config specified")))
	})

	It("should fail to reconcile with no exporters configured", func() {
		emptyProviderConfig := config.CollectorConfig{
			Spec: config.CollectorConfigSpec{
				Exporters: config.CollectorExportersConfig{},
			},
		}

		data, err := json.Marshal(emptyProviderConfig)
		Expect(err).NotTo(HaveOccurred())
		extResource.Spec.ProviderConfig = &runtime.RawExtension{
			Raw: data,
		}

		act, err := actuator.New(k8sClient, actuatorOpts...)
		Expect(err).NotTo(HaveOccurred())
		Expect(act).NotTo(BeNil())

		err = act.Reconcile(ctx, logger, extResource)
		Expect(err).Should(HaveOccurred())
		Expect(err).To(MatchError(ContainSubstring("no exporter enabled")))
	})

	It("should succeed on Reconcile", func() {
		// Ensure we have valid provider config
		extResource.Spec.ProviderConfig = &runtime.RawExtension{
			Raw: providerConfigData,
		}

		act, err := actuator.New(k8sClient, actuatorOpts...)
		Expect(err).NotTo(HaveOccurred())
		Expect(act).NotTo(BeNil())
		Expect(act.Reconcile(ctx, logger, extResource)).To(Succeed())

		// TODO(user): Add more tests
	})

	It("should succeed on Delete", func() {
		act, err := actuator.New(k8sClient, actuatorOpts...)
		Expect(err).NotTo(HaveOccurred())
		Expect(act).NotTo(BeNil())
		Expect(act.Delete(ctx, logger, extResource)).To(Succeed())

		// TODO(user): Add more tests
	})

	It("should succeed on ForceDelete", func() {
		act, err := actuator.New(k8sClient, actuatorOpts...)
		Expect(err).NotTo(HaveOccurred())
		Expect(act).NotTo(BeNil())
		Expect(act.ForceDelete(ctx, logger, extResource)).To(Succeed())

		// TODO(user): Add more tests
	})

	It("should succeed on Restore", func() {
		// Ensure we have valid provider config
		extResource.Spec.ProviderConfig = &runtime.RawExtension{
			Raw: providerConfigData,
		}

		act, err := actuator.New(k8sClient, actuatorOpts...)
		Expect(err).NotTo(HaveOccurred())
		Expect(act).NotTo(BeNil())
		Expect(act.Restore(ctx, logger, extResource)).To(Succeed())

		// TODO(user): Add more tests
	})

	It("should succeed on Migrate", func() {
		// Ensure we have valid provider config
		extResource.Spec.ProviderConfig = &runtime.RawExtension{
			Raw: providerConfigData,
		}

		act, err := actuator.New(k8sClient, actuatorOpts...)
		Expect(err).NotTo(HaveOccurred())
		Expect(act).NotTo(BeNil())
		Expect(act.Migrate(ctx, logger, extResource)).To(Succeed())

		// TODO(user): Add more tests
	})
})

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
			pipelines := buildPipelines(cfg, []string{debugExporterName})
			Expect(pipelines).To(HaveLen(len(wantPipelines)))
			for _, name := range wantPipelines {
				Expect(pipelines).To(HaveKey(name))
			}
		},
		Entry("empty selection enables all pipelines",
			configWithSignals(),
			[]string{logsPipelineName, eventsPipelineName, metricsPipelineName},
		),
		Entry("metrics only",
			configWithSignals(config.SignalMetrics),
			[]string{metricsPipelineName},
		),
		Entry("logs and events only",
			configWithSignals(config.SignalLogs, config.SignalEvents),
			[]string{logsPipelineName, eventsPipelineName},
		),
		Entry("events only",
			configWithSignals(config.SignalEvents),
			[]string{eventsPipelineName},
		),
	)

	It("wires the enabled receivers and exporters into each pipeline", func() {
		pipelines := buildPipelines(configWithSignals(), []string{debugExporterName, "otlp_http"})

		Expect(pipelines[logsPipelineName].Receivers).To(Equal([]string{otlpReceiverName}))
		Expect(pipelines[eventsPipelineName].Receivers).To(Equal([]string{eventsReceiverName}))
		Expect(pipelines[metricsPipelineName].Receivers).To(Equal([]string{prometheusReceiverName}))

		for _, p := range pipelines {
			Expect(p.Exporters).To(Equal([]string{debugExporterName, "otlp_http"}))
		}
	})
})

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
