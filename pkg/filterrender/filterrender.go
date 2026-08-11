// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package filterrender turns a target's opaque filter configuration into the
// map[string]any settings consumed by the OpenTelemetry Filter processor.
package filterrender

import (
	"encoding/json"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor"
	"go.opentelemetry.io/collector/confmap"

	"github.com/gardener/gardener-extension-otelcol/pkg/apis/config"
)

// FilterProcessorConfig decodes a target's opaque filter configuration into the
// map[string]any consumed by the OTel filter processor. A target with no filter
// yields an empty map. It returns an error if the body is not valid JSON.
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

// FilterProcessorConfigsForSignals renders the target's filter and returns the
// filter config only for the wanted signals. Only metrics and logs are considered.
// Events are collected as logs and run in log-type pipelines, so they use the
// filter's log section.
func FilterProcessorConfigsForSignals(
	target config.Target,
	signals []config.SignalType,
) (map[config.SignalType]map[string]any, error) {
	rendered, err := FilterProcessorConfig(target)
	if err != nil {
		return nil, err
	}

	out := map[config.SignalType]map[string]any{}
	if len(rendered) == 0 {
		return out, nil
	}

	cfg := filterprocessor.
		NewFactory().
		CreateDefaultConfig().(*filterprocessor.Config)
	if err := confmap.NewFromStringMap(rendered).Unmarshal(cfg); err != nil {
		return nil, err
	}

	for _, sig := range signals {
		switch sig {
		case config.SignalMetrics:
			m := cfg.Metrics //nolint:staticcheck // deprecated block still supported as valid input
			if m.Include != nil ||
				m.Exclude != nil ||
				m.RegexpConfig != nil ||
				len(m.ResourceConditions) > 0 ||
				len(m.MetricConditions) > 0 ||
				len(m.DataPointConditions) > 0 ||
				len(cfg.MetricConditions) > 0 {
				out[sig] = rendered
			}
		case config.SignalLogs, config.SignalEvents:
			l := cfg.Logs //nolint:staticcheck // deprecated block still supported as valid input
			if l.Include != nil ||
				l.Exclude != nil ||
				len(l.ResourceConditions) > 0 ||
				len(l.LogConditions) > 0 ||
				len(cfg.LogConditions) > 0 {
				out[sig] = rendered
			}
		}
	}

	return out, nil
}
