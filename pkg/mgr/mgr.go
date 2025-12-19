// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package mgr provides utility functions for creating [manager.Manager] using a
// functional options API.
package mgr

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gardener/gardener/extensions/pkg/util"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	componentbaseconfigv1alpha1 "k8s.io/component-base/config/v1alpha1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	controllerconfig "sigs.k8s.io/controller-runtime/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// mgr is a wrapper around [manager.Manager] with functional options API.
type mgr struct {
	scheme                  *runtime.Scheme
	addToSchemes            []func(s *runtime.Scheme) error
	installSchemes          []func(s *runtime.Scheme)
	restConfig              *rest.Config
	metricsServerOpts       metricsserver.Options
	healthProbeAddr         string
	pprofAddr               string
	leaderElectionEnabled   bool
	leaderElectionID        string
	leaderElectionNamespace string
	webhookServer           webhook.Server
	baseCtxFunc             manager.BaseContextFunc
	controllerOpts          controllerconfig.Controller
	logger                  logr.Logger
	runnables               []manager.Runnable
	extraMetricsHandlers    map[string]http.Handler
	healthzChecks           map[string]healthz.Checker
	readyzChecks            map[string]healthz.Checker
	clientOpts              client.Options
	cacheOpts               cache.Options
	clientConnConfig        *componentbaseconfigv1alpha1.ClientConnectionConfiguration
}

// New creates a new [manager.Manager] with the given options.
func New(opts ...Option) (manager.Manager, error) {
	m := &mgr{
		scheme:            runtime.NewScheme(),
		addToSchemes:      make([]func(s *runtime.Scheme) error, 0),
		installSchemes:    make([]func(s *runtime.Scheme), 0),
		metricsServerOpts: metricsserver.Options{},
		baseCtxFunc:       context.Background,
		controllerOpts: controllerconfig.Controller{
			MaxConcurrentReconciles: 5,
			RecoverPanic:            ptr.To(true),
		},
		runnables:            make([]manager.Runnable, 0),
		extraMetricsHandlers: make(map[string]http.Handler),
		healthzChecks:        make(map[string]healthz.Checker),
		readyzChecks:         make(map[string]healthz.Checker),
	}

	for _, opt := range opts {
		if err := opt(m); err != nil {
			return nil, err
		}
	}

	// Register additional schemes
	for _, addToScheme := range m.addToSchemes {
		if err := addToScheme(m.scheme); err != nil {
			return nil, fmt.Errorf("failed to add scheme: %w", err)
		}
	}

	for _, installScheme := range m.installSchemes {
		installScheme(m.scheme)
	}

	// Get rest.Config, unless we have one already
	if m.restConfig == nil {
		restConfig, err := config.GetConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get rest config: %w", err)
		}
		m.restConfig = restConfig
	}

	// Apply any connection config settings, if we have such
	util.ApplyClientConnectionConfigurationToRESTConfig(m.clientConnConfig, m.restConfig)

	crMgr, err := manager.New(
		m.restConfig,
		manager.Options{
			Scheme:                     m.scheme,
			Metrics:                    m.metricsServerOpts,
			HealthProbeBindAddress:     m.healthProbeAddr,
			LeaderElection:             m.leaderElectionEnabled,
			LeaderElectionID:           m.leaderElectionID,
			LeaderElectionNamespace:    m.leaderElectionNamespace,
			LeaderElectionResourceLock: resourcelock.LeasesResourceLock,
			BaseContext:                m.baseCtxFunc,
			Controller:                 m.controllerOpts,
			WebhookServer:              m.webhookServer,
			Logger:                     m.logger,
			PprofBindAddress:           m.pprofAddr,
			Client:                     m.clientOpts,
			Cache:                      m.cacheOpts,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create manager: %w", err)
	}

	// Configure health and ready checks
	for name, checker := range m.healthzChecks {
		if err := crMgr.AddHealthzCheck(name, checker); err != nil {
			return nil, fmt.Errorf("failed to setup health check: %w", err)
		}
	}
	for name, checker := range m.readyzChecks {
		if err := crMgr.AddReadyzCheck(name, checker); err != nil {
			return nil, fmt.Errorf("failed to setup ready check: %w", err)
		}
	}

	// Configure extra handlers via the metrics server
	for path, handler := range m.extraMetricsHandlers {
		if err := crMgr.AddMetricsServerExtraHandler(path, handler); err != nil {
			return nil, fmt.Errorf("failed to setup extra handler: %w", err)
		}
	}

	// Configure runnables
	for _, runnable := range m.runnables {
		if err := crMgr.Add(runnable); err != nil {
			return nil, fmt.Errorf("failed to setup runnable: %w", err)
		}
	}

	return crMgr, nil
}

// Option is a function, which configures the [manager.Manager].
type Option func(m *mgr) error

// WithConfig is an [Option], which configures the [manager.Manager] with the
// given [rest.Config].
func WithConfig(c *rest.Config) Option {
	opt := func(m *mgr) error {
		m.restConfig = c

		return nil
	}

	return opt
}

// WithScheme is an [Option], which configures the [manager.Manager] with the
// given [runtime.Scheme].
func WithScheme(s *runtime.Scheme) Option {
	opt := func(m *mgr) error {
		m.scheme = s

		return nil
	}

	return opt
}

// WithAddToScheme is an [Option], which adds the given function for registering
// additional schemes to the [manager.Manager].
func WithAddToScheme(f func(s *runtime.Scheme) error) Option {
	opt := func(m *mgr) error {
		m.addToSchemes = append(m.addToSchemes, f)

		return nil
	}

	return opt
}

// WithInstallScheme is an [Option], which registers an API group and adds types
// to the scheme.
func WithInstallScheme(f func(s *runtime.Scheme)) Option {
	opt := func(m *mgr) error {
		m.installSchemes = append(m.installSchemes, f)

		return nil
	}

	return opt
}

// WithMetricsOptions is an [Option], which configures the [manager.Manager]
// with the given [metricsserver.Options].
func WithMetricsOptions(opts metricsserver.Options) Option {
	opt := func(m *mgr) error {
		m.metricsServerOpts = opts

		return nil
	}

	return opt
}

// WithMetricsAddress is an [Option], which configures the [manager.Manager] to
// serve metrics on the given address. In order to disable metrics server,
// specify "0" as the address value.
func WithMetricsAddress(addr string) Option {
	opt := func(m *mgr) error {
		m.metricsServerOpts.BindAddress = addr

		return nil
	}

	return opt
}

// WithExtraMetricsHandler is an [Option], which configures the
// [manager.Manager] to serve an extra handler via the metrics server.
func WithExtraMetricsHandler(path string, handler http.Handler) Option {
	opt := func(m *mgr) error {
		m.extraMetricsHandlers[path] = handler

		return nil
	}

	return opt
}

// WithLeaderElection is an [Option], which configures leader election for the
// [manager.Manager], if set to true.
func WithLeaderElection(enable bool) Option {
	opt := func(m *mgr) error {
		m.leaderElectionEnabled = enable

		return nil
	}

	return opt
}

// WithLeaderElectionID is an [Option], which configures the leader election ID
// to be used, if leader election has been enabled.
func WithLeaderElectionID(id string) Option {
	opt := func(m *mgr) error {
		m.leaderElectionID = id

		return nil
	}

	return opt
}

// WithLeaderElectionNamespace is an [Option], which configures the namespace to
// use for the lease, if leader election has been enabled.
func WithLeaderElectionNamespace(ns string) Option {
	opt := func(m *mgr) error {
		m.leaderElectionNamespace = ns

		return nil
	}

	return opt
}

// WithContext is an [Option], which configures the [manager.Manager] to use the
// given [context.Context] as the base context.
func WithContext(ctx context.Context) Option {
	opt := func(m *mgr) error {
		m.baseCtxFunc = func() context.Context { return ctx }

		return nil
	}

	return opt
}

// WithControllerOptions is an [Option], which configures the [manager.Manager] to use
// the given [controllerconfig.Controller] options.
func WithControllerOptions(opts controllerconfig.Controller) Option {
	opt := func(m *mgr) error {
		m.controllerOpts = opts

		return nil
	}

	return opt
}

// WithMaxConcurrentReconciles is an [Option], which configures the
// [manager.Manager] with the given max concurrent reconciles.
func WithMaxConcurrentReconciles(val int) Option {
	opt := func(m *mgr) error {
		m.controllerOpts.MaxConcurrentReconciles = val

		return nil
	}

	return opt
}

// WithHealthzCheck is an [Option], which configures the [manager.Manager] to
// use the given [healthz.Checker] for health checks.
func WithHealthzCheck(name string, checker healthz.Checker) Option {
	opt := func(m *mgr) error {
		m.healthzChecks[name] = checker

		return nil
	}

	return opt
}

// WithReadyzCheck is an [Option], which configures the [manager.Manager] to use
// the given [healthz.Checker] for readiness checks.
func WithReadyzCheck(name string, checker healthz.Checker) Option {
	opt := func(m *mgr) error {
		m.readyzChecks[name] = checker

		return nil
	}

	return opt
}

// WithHealthProbeAddress is an [Option], which configures the [manager.Manager]
// to use the given address for health probes. In order to disable health probes
// specify value for the address of "0".
func WithHealthProbeAddress(addr string) Option {
	opt := func(m *mgr) error {
		m.healthProbeAddr = addr

		return nil
	}

	return opt
}

// WithWebhookServer is an [Option], which configures the [manager.Manager] with
// the given [webhook.Server].
func WithWebhookServer(server webhook.Server) Option {
	opt := func(m *mgr) error {
		m.webhookServer = server

		return nil
	}

	return opt
}

// WithLogger is an [Option], which configures the [manager.Manager] with
// the given [logr.Logger].
func WithLogger(logger logr.Logger) Option {
	opt := func(m *mgr) error {
		m.logger = logger

		return nil
	}

	return opt
}

// WithPprofAddress is an [Option], which configures the [manager.Manager] to
// serve pprof data on the given address. In order to disable pprof, specify "0"
// as the address value.
func WithPprofAddress(addr string) Option {
	opt := func(m *mgr) error {
		m.pprofAddr = addr

		return nil
	}

	return opt
}

// WithRunnable is an [Option], which adds the given [manager.Runnable] to the
// [manager.Manager].
func WithRunnable(r manager.Runnable) Option {
	opt := func(m *mgr) error {
		m.runnables = append(m.runnables, r)

		return nil
	}

	return opt
}

// WithClientOptions is an [Option], which configures the [manager.Manager]
// with the given [client.Options].
func WithClientOptions(opts client.Options) Option {
	opt := func(m *mgr) error {
		m.clientOpts = opts

		return nil
	}

	return opt
}

// WithCacheOptions is an [Option], which configures the [manager.Manager] with
// the given [cache.Options].
func WithCacheOptions(opts cache.Options) Option {
	opt := func(m *mgr) error {
		m.cacheOpts = opts

		return nil
	}

	return opt
}

// WithConnectionConfiguration is an [Option], which configures the client
// connection options used by the [manager.Manager] with the given
// [componentbaseconfigv1alpha1.ClientConnectionConfiguration] settings.
func WithConnectionConfiguration(cfg *componentbaseconfigv1alpha1.ClientConnectionConfiguration) Option {
	opt := func(m *mgr) error {
		m.clientConnConfig = cfg

		return nil
	}

	return opt
}
