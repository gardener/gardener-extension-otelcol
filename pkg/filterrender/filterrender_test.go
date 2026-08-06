// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package filterrender_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/gardener/gardener-extension-otelcol/pkg/apis/config"
	"github.com/gardener/gardener-extension-otelcol/pkg/filterrender"
)

// filterTarget returns a Target with the given raw filter body.
func filterTarget(body string) config.Target {
	target := config.Target{}
	if body != "" {
		target.Filters = runtime.RawExtension{Raw: []byte(body)}
	}

	return target
}

var _ = Describe("FilterProcessorConfig", func() {
	It("returns an empty map for a target with no filter", func() {
		out, err := filterrender.FilterProcessorConfig(filterTarget(""))
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(BeEmpty())
	})

	It("decodes the opaque body into a map verbatim", func() {
		out, err := filterrender.FilterProcessorConfig(filterTarget(
			`{"error_mode":"ignore","metrics":{"metric":["metric.name == \"foo\""]}}`,
		))
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal(map[string]any{
			"error_mode": "ignore",
			"metrics": map[string]any{
				"metric": []any{`metric.name == "foo"`},
			},
		}))
	})

	It("returns an error for a malformed body", func() {
		_, err := filterrender.FilterProcessorConfig(filterTarget(`{not json`))
		Expect(err).To(HaveOccurred())
	})
})
