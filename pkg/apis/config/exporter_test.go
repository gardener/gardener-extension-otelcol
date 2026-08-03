// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gardener/gardener-extension-otelcol/pkg/apis/config"
)

var _ = Describe("ExporterConfig merge semantics", func() {
	newDefault := func() config.ExporterConfig {
		return config.ExporterConfig{
			Protocol:    config.ExporterProtocolHTTP,
			Endpoint:    "https://default:4318",
			Timeout:     30_000_000_000, // 30s
			Compression: config.CompressionGzip,
			TLS:         &config.TLSConfig{},
			Token:       &config.ResourceReference{ResourceRef: config.ResourceReferenceDetails{Name: "def", DataKey: "token"}},
		}
	}

	It("returns a copy of the default when the override is nil", func() {
		Expect(newDefault().MergeWith(nil)).To(Equal(newDefault()))
	})

	It("overrides only the non-zero fields of the override", func() {
		merged := newDefault().MergeWith(&config.ExporterConfig{
			Endpoint: "https://override:4318",
		})

		Expect(merged.Endpoint).To(Equal("https://override:4318"))
		Expect(merged.Protocol).To(Equal(config.ExporterProtocolHTTP))
		Expect(merged.Compression).To(Equal(config.CompressionGzip))
		Expect(merged.TLS).To(Equal(newDefault().TLS))
		Expect(merged.Token).To(Equal(newDefault().Token))
	})

	It("overrides the protocol when set", func() {
		merged := newDefault().MergeWith(&config.ExporterConfig{
			Protocol: config.ExporterProtocolGRPC,
			Endpoint: "https://override:4317",
		})

		Expect(merged.Protocol).To(Equal(config.ExporterProtocolGRPC))
	})

	It("overrides the verbosity when set", func() {
		merged := newDefault().MergeWith(&config.ExporterConfig{
			Protocol:  config.ExporterProtocolDebug,
			Verbosity: config.DebugExporterVerbosityDetailed,
		})

		Expect(merged.Protocol).To(Equal(config.ExporterProtocolDebug))
		Expect(merged.Verbosity).To(Equal(config.DebugExporterVerbosityDetailed))
	})

	It("replaces pointer fields wholesale when the override sets them", func() {
		overrideToken := &config.ResourceReference{
			ResourceRef: config.ResourceReferenceDetails{Name: "sig", DataKey: "token"},
		}
		merged := newDefault().MergeWith(&config.ExporterConfig{Token: overrideToken})

		Expect(merged.Token).To(Equal(overrideToken))
	})

	It("is exposed through SignalTarget.EffectiveExporter", func() {
		target := config.SignalTarget{
			Exporter: config.ExporterConfig{Endpoint: "https://target:4318"},
		}

		Expect(target.EffectiveExporter(newDefault()).Endpoint).To(Equal("https://target:4318"))
	})

	It("returns the default unchanged when the target has no override", func() {
		Expect(config.SignalTarget{}.EffectiveExporter(newDefault())).To(Equal(newDefault()))
	})
})

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
