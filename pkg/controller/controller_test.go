// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package controller_test

import (
	"context"
	"time"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	predicateutils "github.com/gardener/gardener/pkg/controllerutils/predicate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/rest"
	"k8s.io/utils/ptr"
	crctrl "sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/gardener/gardener-extension-otelcol/pkg/actuator"
	"github.com/gardener/gardener-extension-otelcol/pkg/controller"
)

var _ = Describe("Controller", Ordered, func() {
	var act *actuator.Actuator

	BeforeAll(func() {
		a, err := actuator.New()
		Expect(err).NotTo(HaveOccurred())
		Expect(a).NotTo(BeNil())
		act = a
	})

	It("should fail to create controller with missing actuator", func() {
		opts := []controller.Option{}
		c, err := controller.New(opts...)

		Expect(err).Should(HaveOccurred())
		Expect(err).To(MatchError(controller.ErrInvalidController))
		Expect(err).To(MatchError(ContainSubstring("missing actuator implementation")))
		Expect(c).To(BeNil())
	})

	It("should fail to create controller with missing name", func() {
		opts := []controller.Option{
			controller.WithActuator(act),
		}
		c, err := controller.New(opts...)

		Expect(err).Should(HaveOccurred())
		Expect(err).To(MatchError(controller.ErrInvalidController))
		Expect(err).To(MatchError(ContainSubstring("missing controller name")))
		Expect(c).To(BeNil())
	})

	It("should fail to create controller with missing extension type", func() {
		opts := []controller.Option{
			controller.WithActuator(act),
			controller.WithName("example"),
		}
		c, err := controller.New(opts...)

		Expect(err).Should(HaveOccurred())
		Expect(err).To(MatchError(controller.ErrInvalidController))
		Expect(err).To(MatchError(ContainSubstring("missing extension type")))
		Expect(c).To(BeNil())
	})

	It("should fail to create controller with missing extension class", func() {
		opts := []controller.Option{
			controller.WithActuator(act),
			controller.WithName("example"),
			controller.WithExtensionType("example"),
		}
		c, err := controller.New(opts...)

		Expect(err).Should(HaveOccurred())
		Expect(err).To(MatchError(controller.ErrInvalidController))
		Expect(err).To(MatchError(ContainSubstring("missing extension class")))
		Expect(c).To(BeNil())
	})

	It("should successfully create a controller and register it", func() {
		opts := []controller.Option{
			controller.WithActuator(act),
			controller.WithName("example"),
			controller.WithExtensionType("example"),
			controller.WithExtensionClass(v1alpha1.ExtensionClassShoot),
			controller.WithFinalizerSuffix("custom-finalizer-suffix"),
			controller.WithControllerOptions(crctrl.Options{
				RecoverPanic:            ptr.To(true),
				MaxConcurrentReconciles: 5,
			}),
			controller.WithIgnoreOperationAnnotation(true),
			controller.WithResyncInterval(30 * time.Second),
			controller.WithPredicate(predicateutils.HasName("example")),
			controller.WithWatchBuilder(extensionscontroller.NewWatchBuilder()),
		}
		c, err := controller.New(opts...)

		Expect(err).ShouldNot(HaveOccurred())
		Expect(c).NotTo(BeNil())

		m, err := manager.New(&rest.Config{}, manager.Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(m).NotTo(BeNil())
		Expect(c.SetupWithManager(context.TODO(), m)).To(Succeed())
	})
})
