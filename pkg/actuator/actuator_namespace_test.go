// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package actuator

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
