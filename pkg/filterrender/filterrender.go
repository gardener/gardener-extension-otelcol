// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package filterrender turns a target's opaque filter configuration into the
// map[string]any settings consumed by the OpenTelemetry filter processor.
//
// The filter body is carried verbatim as opaque JSON, so rendering is a plain
// JSON decode: the extension does not mirror the upstream filterprocessor
// schema. It is shared by the actuator (which embeds the result into the
// OpenTelemetryCollector resource) and the validation package (which round-trips
// the result through the upstream filterprocessor.Config).
package filterrender

import (
	"encoding/json"

	"github.com/gardener/gardener-extension-otelcol/pkg/apis/config"
)

// FilterProcessorConfig decodes a target's opaque filter configuration into the
// map[string]any consumed by the OTel filter processor. A target with no filter
// yields an empty map. It returns an error if the body is not valid JSON.
//
// The map is passed through unchanged to both the collector and the upstream
// filterprocessor.Config validation, so both the basic (flat condition list) and
// advanced (context-inferred) OTTL styles are supported without any special
// handling here.
func FilterProcessorConfig(target config.Target) (map[string]any, error) {
	if len(target.Filters.Raw) == 0 {
		return map[string]any{}, nil
	}

	out := map[string]any{}
	if err := json.Unmarshal(target.Filters.Raw, &out); err != nil {
		return nil, err
	}

	return out, nil
}
