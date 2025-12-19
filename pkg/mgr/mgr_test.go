// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package mgr_test

import (
	"context"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/component-base/config/v1alpha1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	controllerconfig "sigs.k8s.io/controller-runtime/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/gardener/gardener-extension-otelcol/pkg/mgr"
)

var _ = Describe("Manager", Ordered, func() {
	It("should fail to create manager without rest.Config, because running out-of-cluster", func() {
		opts := []mgr.Option{}
		m, err := mgr.New(opts...)

		Expect(err).Should(HaveOccurred())
		Expect(err).To(MatchError(ContainSubstring("failed to get rest config")))
		Expect(m).To(BeNil())
	})

	It("should successfully create a manager", func() {
		extraHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		})

		testServer := webhook.NewServer(webhook.Options{Port: 9443})
		testRunnable := manager.RunnableFunc(func(ctx context.Context) error {
			<-ctx.Done()

			return nil
		})

		installSchemeFunc := func(s *runtime.Scheme) {
			// A dummy install scheme function
		}

		opts := []mgr.Option{
			mgr.WithConfig(cfg),
			mgr.WithScheme(runtime.NewScheme()),
			mgr.WithAddToScheme(corev1.AddToScheme),
			mgr.WithInstallScheme(installSchemeFunc),
			mgr.WithMetricsOptions(metricsserver.Options{SecureServing: true}),
			mgr.WithMetricsAddress(":9090"),
			mgr.WithExtraMetricsHandler("/test-handler", extraHandler),
			mgr.WithLeaderElection(true),
			mgr.WithLeaderElectionID("foobar"),
			mgr.WithLeaderElectionNamespace("default"),
			mgr.WithContext(ctx),
			mgr.WithMaxConcurrentReconciles(42),
			mgr.WithControllerOptions(controllerconfig.Controller{RecoverPanic: ptr.To(true)}),
			mgr.WithHealthzCheck("healthz", healthz.Ping),
			mgr.WithReadyzCheck("readyz", healthz.Ping),
			mgr.WithHealthProbeAddress(":9091"),
			mgr.WithWebhookServer(testServer),
			mgr.WithClientOptions(client.Options{HTTPClient: http.DefaultClient}),
			mgr.WithConnectionConfiguration(&v1alpha1.ClientConnectionConfiguration{QPS: 100.0, Burst: 130}),
			mgr.WithCacheOptions(cache.Options{HTTPClient: http.DefaultClient}),
			mgr.WithLogger(logger),
			mgr.WithPprofAddress(":7070"),
			mgr.WithRunnable(testRunnable),
		}

		m, err := mgr.New(opts...)

		Expect(err).ShouldNot(HaveOccurred())
		Expect(m).NotTo(BeNil())
	})
})
