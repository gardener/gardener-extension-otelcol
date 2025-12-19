// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package metrics specifies various metrics provided by the extension.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

// Namespace is the namespace component of the fully qualified metric name.
const Namespace = "gardener_extension_otelcol"

var (
	// ActuatorOperationTotal is an example metric, which increments each
	// time our extension actuator is being called.
	ActuatorOperationTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Name:      "actuator_operation_total",
			Help:      "Total number of times our extension actuator did something",
		},
		[]string{"cluster", "operation"},
	)

	// ActuatorOperationDurationSeconds is an example metric, which tracks
	// the duration of execution for our extension actuator.
	ActuatorOperationDurationSeconds = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: Namespace,
			Name:      "actuator_operation_duration_seconds",
			Help:      "Duration of execution for our extension actuator",
		},
		[]string{"cluster", "operation"},
	)
)

// init registers our custom metrics with the default controller-runtime registry.
func init() {
	ctrlmetrics.Registry.MustRegister(
		ActuatorOperationTotal,
		ActuatorOperationDurationSeconds,
	)
}
