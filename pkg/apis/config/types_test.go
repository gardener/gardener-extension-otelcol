// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gardener/gardener-extension-otelcol/pkg/apis/config"
)

var _ = Describe("Target", func() {
	Describe("EffectiveSignals", func() {
		It("returns all signals when none are configured", func() {
			target := config.Target{}

			Expect(target.EffectiveSignals()).To(ConsistOf([]config.SignalType{"metrics", "logs", "events"}))
		})

		It("returns all signals for an empty (non-nil) signal list", func() {
			target := config.Target{Signals: []config.SignalType{}}

			Expect(target.EffectiveSignals()).To(ConsistOf([]config.SignalType{"metrics", "logs", "events"}))
		})

		It("returns the configured signals as-is", func() {
			target := config.Target{Signals: []config.SignalType{config.SignalLogs}}

			Expect(target.EffectiveSignals()).To(Equal([]config.SignalType{"logs"}))
		})
	})
})
