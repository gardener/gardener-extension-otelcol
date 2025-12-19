// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package heartbeat_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/rest"
	"k8s.io/utils/clock"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/gardener/gardener-extension-otelcol/pkg/heartbeat"
)

var _ = Describe("Heartbeat Controller", Ordered, func() {
	It("should fail to create heartbeat controller with missing extension name", func() {
		opts := []heartbeat.Option{}
		c, err := heartbeat.New(opts...)

		Expect(err).Should(HaveOccurred())
		Expect(err).To(MatchError(heartbeat.ErrInvalidHeartbeat))
		Expect(err).To(MatchError(ContainSubstring("missing extension name")))
		Expect(c).To(BeNil())
	})

	It("should fail to create heartbeat controller with missing lease namespace", func() {
		opts := []heartbeat.Option{
			heartbeat.WithExtensionName("example"),
		}
		c, err := heartbeat.New(opts...)

		Expect(err).Should(HaveOccurred())
		Expect(err).To(MatchError(heartbeat.ErrInvalidHeartbeat))
		Expect(err).To(MatchError(ContainSubstring("missing lease namespace")))
		Expect(c).To(BeNil())
	})

	It("should successfully create heartbeat controller and register it", func() {
		opts := []heartbeat.Option{
			heartbeat.WithExtensionName("example"),
			heartbeat.WithClock(clock.RealClock{}),
			heartbeat.WithLeaseNamespace("default"),
			heartbeat.WithRenewInterval(1 * time.Minute),
		}
		h, err := heartbeat.New(opts...)

		Expect(err).ShouldNot(HaveOccurred())
		Expect(h).NotTo(BeNil())

		m, err := manager.New(&rest.Config{}, manager.Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(m).NotTo(BeNil())
		Expect(h.SetupWithManager(context.TODO(), m)).To(Succeed())
	})
})
