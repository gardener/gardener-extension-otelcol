// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gardener/gardener-extension-otelcol/pkg/apis/config"
)

var _ = Describe("SignalConfig.IsEnabled", func() {
	It("is false when Enabled is nil", func() {
		Expect(config.SignalConfig{}.IsEnabled()).To(BeFalse())
	})

	It("reflects the Enabled pointer", func() {
		t, f := true, false
		Expect(config.SignalConfig{Enabled: &t}.IsEnabled()).To(BeTrue())
		Expect(config.SignalConfig{Enabled: &f}.IsEnabled()).To(BeFalse())
	})
})
