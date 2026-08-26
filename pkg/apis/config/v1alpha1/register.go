// SPDX-FileCopyrightText: Copyright Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

func init() {
	// Manually registered functions.
	localSchemeBuilder.Register(RegisterDefaults)
}
