// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package heartbeat provides utilities for creating heartbeat reconcilers for
// extensions.
package heartbeat

import (
	"context"
	"errors"
	"fmt"
	"time"

	heartbeatcontroller "github.com/gardener/gardener/extensions/pkg/controller/heartbeat"
	"k8s.io/utils/clock"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// ErrInvalidHeartbeat is an error, which is returned when attempting to create
// a [Heartbeat], but the configuration was found to be invalid.
var ErrInvalidHeartbeat = errors.New("invalid heartbeat config")

// Heartbeat is a wrapper for a reconciler, which periodically renews heartbeat
// leases.
type Heartbeat struct {
	extensionName string
	namespace     string
	renewInterval time.Duration
	clock         clock.Clock
}

// Option is a function, which configures the [Heartbeat].
type Option func(a *Heartbeat) error

// New creates a new [Heartbeat] with the given options.
func New(opts ...Option) (*Heartbeat, error) {
	h := &Heartbeat{
		clock:         clock.RealClock{},
		renewInterval: 30 * time.Second,
	}

	for _, opt := range opts {
		if err := opt(h); err != nil {
			return nil, err
		}
	}

	if h.extensionName == "" {
		return nil, fmt.Errorf("%w: missing extension name", ErrInvalidHeartbeat)
	}
	if h.namespace == "" {
		return nil, fmt.Errorf("%w: missing lease namespace", ErrInvalidHeartbeat)
	}

	return h, nil
}

// SetupWithManager registers the [Heartbeat] controller with the given [manager.Manager].
func (h *Heartbeat) SetupWithManager(ctx context.Context, mgr manager.Manager) error {
	return heartbeatcontroller.Add(
		mgr,
		heartbeatcontroller.AddArgs{
			ExtensionName:        h.extensionName,
			Namespace:            h.namespace,
			RenewIntervalSeconds: int32(h.renewInterval.Seconds()),
			Clock:                h.clock,
		},
	)
}

// WithExtensionName is an [Option], which configures the [Heartbeat] to use the
// given extension name.
func WithExtensionName(name string) Option {
	opt := func(h *Heartbeat) error {
		h.extensionName = name

		return nil
	}

	return opt
}

// WithLeaseNamespace is an [Option], which configures the [Heartbeat] to create
// a lease in the given namespace.
func WithLeaseNamespace(namespace string) Option {
	opt := func(h *Heartbeat) error {
		h.namespace = namespace

		return nil
	}

	return opt
}

// WithRenewInterval is an [Option], which configures the [Heartbeat] to renew
// the lease on the given interval.
func WithRenewInterval(interval time.Duration) Option {
	opt := func(h *Heartbeat) error {
		h.renewInterval = interval

		return nil
	}

	return opt
}

// WithClock is an [Option], which configures the [Heartbeat] to use the given
// [clock.Clock].
func WithClock(clk clock.Clock) Option {
	opt := func(h *Heartbeat) error {
		h.clock = clk

		return nil
	}

	return opt
}
