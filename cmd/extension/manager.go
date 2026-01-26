// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"time"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	resourcesv1alpha1 "github.com/gardener/gardener/pkg/apis/resources/v1alpha1"
	glogger "github.com/gardener/gardener/pkg/logger"
	"github.com/urfave/cli/v3"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	componentbaseconfigv1alpha1 "k8s.io/component-base/config/v1alpha1"
	"k8s.io/component-base/featuregate"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/gardener/gardener-extension-otelcol/pkg/actuator"
	configinstall "github.com/gardener/gardener-extension-otelcol/pkg/apis/config/install"
	"github.com/gardener/gardener-extension-otelcol/pkg/controller"
	"github.com/gardener/gardener-extension-otelcol/pkg/heartbeat"
	"github.com/gardener/gardener-extension-otelcol/pkg/mgr"
)

// flags stores the manager flags as provided from the command-line
type flags struct {
	extensionName             string
	metricsBindAddr           string
	healthProbeBindAddr       string
	heartbeatRenewInterval    time.Duration
	heartbeatNamespace        string
	leaderElection            bool
	leaderElectionID          string
	leaderElectionNamespace   string
	ignoreOperationAnnotation bool
	maxConcurrentReconciles   int
	kubeconfig                string
	zapLogLevel               string
	zapLogFormat              string
	resyncInterval            time.Duration
	pprofBindAddr             string
	clientConnQPS             float32
	clientConnBurst           int32

	// The following flags are meant to be specified by the Helm chart,
	// which gardenlet will invoke during deployment. The value of each flag
	// is derived from a list of extra values, which gardenlet passes to
	// Helm.
	//
	// See the link below for more details.
	//
	// https://github.com/gardener/gardener/blob/d5071c800378616eb6bb2c7662b4b28f4cfe7406/pkg/gardenlet/controller/controllerinstallation/controllerinstallation/reconciler.go#L236-L263
	gardenerVersion       string
	gardenletFeatureGates map[featuregate.Feature]bool
}

// getManager creates a new [ctrl.Manager] based on the parsed [flags].
func (f *flags) getManager(ctx context.Context) (ctrl.Manager, error) {
	m, err := mgr.New(
		mgr.WithContext(ctx),
		mgr.WithAddToScheme(clientgoscheme.AddToScheme),
		mgr.WithAddToScheme(extensionscontroller.AddToScheme),
		mgr.WithAddToScheme(resourcesv1alpha1.AddToScheme),
		mgr.WithInstallScheme(configinstall.Install),
		mgr.WithMetricsAddress(f.metricsBindAddr),
		mgr.WithHealthProbeAddress(f.healthProbeBindAddr),
		mgr.WithLeaderElection(f.leaderElection),
		mgr.WithLeaderElectionID(f.leaderElectionID),
		mgr.WithLeaderElectionNamespace(f.leaderElectionNamespace),
		mgr.WithMaxConcurrentReconciles(f.maxConcurrentReconciles),
		mgr.WithHealthzCheck("healthz", healthz.Ping),
		mgr.WithReadyzCheck("readyz", healthz.Ping),
		mgr.WithPprofAddress(f.pprofBindAddr),
		mgr.WithConnectionConfiguration(&componentbaseconfigv1alpha1.ClientConnectionConfiguration{
			QPS:   f.clientConnQPS,
			Burst: f.clientConnBurst,
		}),
	)

	if err != nil {
		return nil, err
	}

	hb, err := heartbeat.New(
		heartbeat.WithExtensionName(f.extensionName),
		heartbeat.WithLeaseNamespace(f.heartbeatNamespace),
		heartbeat.WithRenewInterval(f.heartbeatRenewInterval),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create heartbeat controller: %w", err)
	}

	if err := hb.SetupWithManager(ctx, m); err != nil {
		return nil, fmt.Errorf("failed to setup heartbeat controller: %w", err)
	}

	return m, nil
}

// flagsKey is the key used to store the parsed command-line flags in a
// [context.Context].
type flagsKey struct{}

// getFlags extracts and returns the [flags] from the given [context.Context].
func getFlags(ctx context.Context) *flags {
	conf, ok := ctx.Value(flagsKey{}).(*flags)
	if !ok {
		return &flags{}
	}

	return conf
}

// NewManagerCommand creates a new [cli.Command] for running the controller manager.
func NewManagerCommand() *cli.Command {
	flags := flags{
		gardenletFeatureGates: make(map[featuregate.Feature]bool),
	}

	cmd := &cli.Command{
		Name:    "manager",
		Aliases: []string{"m"},
		Usage:   "start controller manager",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "extension-name",
				Usage:       "name of the gardener extension",
				Value:       "gardener-extension-otelcol",
				Sources:     cli.EnvVars("EXTENSION_NAME"),
				Destination: &flags.extensionName,
			},
			&cli.StringFlag{
				Name:        "metrics-bind-address",
				Usage:       "the address the metrics endpoint binds to",
				Value:       ":8080",
				Sources:     cli.EnvVars("METRICS_BIND_ADDRESS"),
				Destination: &flags.metricsBindAddr,
			},
			&cli.StringFlag{
				Name:        "pprof-bind-address",
				Usage:       "the address at which pprof binds to",
				Sources:     cli.EnvVars("PPROF_BIND_ADDRESS"),
				Destination: &flags.pprofBindAddr,
			},
			&cli.StringFlag{
				Name:        "health-probe-bind-address",
				Usage:       "the address the probe endpoint binds to",
				Value:       ":8081",
				Sources:     cli.EnvVars("HEALTH_PROBE_BIND_ADDRESS"),
				Destination: &flags.healthProbeBindAddr,
			},
			&cli.DurationFlag{
				Name:        "heartbeat-renew-interval",
				Usage:       "renew heartbeat lease on specified interval",
				Value:       30 * time.Second,
				Sources:     cli.EnvVars("HEARTBEAT_RENEW_INTERVAL"),
				Destination: &flags.heartbeatRenewInterval,
			},
			&cli.StringFlag{
				Name:        "heartbeat-namespace",
				Usage:       "namespace to use for the heartbeat lease",
				Value:       "gardener-extension-otelcol",
				Sources:     cli.EnvVars("HEARTBEAT_NAMESPACE"),
				Destination: &flags.heartbeatNamespace,
			},
			&cli.BoolFlag{
				Name:        "leader-election",
				Usage:       "enable leader election for controller manager",
				Value:       false,
				Sources:     cli.EnvVars("LEADER_ELECTION"),
				Destination: &flags.leaderElection,
			},
			&cli.StringFlag{
				Name:        "leader-election-id",
				Usage:       "the leader election id to use, if leader election is enabled",
				Value:       "gardener-extension-otelcol-leader-election",
				Sources:     cli.EnvVars("LEADER_ELECTION_ID"),
				Destination: &flags.leaderElectionID,
			},
			&cli.StringFlag{
				Name:        "leader-election-namespace",
				Usage:       "namespace to use for the leader election lease",
				Value:       "gardener-extension-otelcol",
				Sources:     cli.EnvVars("LEADER_ELECTION_NAMESPACE"),
				Destination: &flags.leaderElectionNamespace,
			},
			&cli.BoolFlag{
				Name:        "ignore-operation-annotation",
				Usage:       "specifies whether to ignore operation annotation",
				Value:       false,
				Sources:     cli.EnvVars("IGNORE_OPERATION_ANNOTATION"),
				Destination: &flags.ignoreOperationAnnotation,
			},
			&cli.IntFlag{
				Name:        "max-concurrent-reconciles",
				Usage:       "max number of concurrent reconciliations",
				Value:       5,
				Sources:     cli.EnvVars("MAX_CONCURRENT_RECONCILES"),
				Destination: &flags.maxConcurrentReconciles,
			},
			&cli.StringFlag{
				Name:        "kubeconfig",
				Usage:       "path to a kubeconfig when running out-of-cluster",
				Sources:     cli.EnvVars("KUBECONFIG"),
				Destination: &flags.kubeconfig,
				Action: func(ctx context.Context, c *cli.Command, val string) error {
					return os.Setenv(clientcmd.RecommendedConfigPathEnvVar, val)
				},
			},
			&cli.StringFlag{
				Name:  "log-level",
				Usage: "Zap Level to configure the verbosity of logging",
				Value: glogger.InfoLevel,
				Validator: func(val string) error {
					if !slices.Contains(glogger.AllLogLevels, val) {
						return errors.New("invalid log level specified")
					}

					return nil
				},
				Destination: &flags.zapLogLevel,
			},
			&cli.StringFlag{
				Name:  "log-format",
				Usage: "Zap log encoding format, json or text",
				Value: glogger.FormatText,
				Validator: func(val string) error {
					if !slices.Contains(glogger.AllLogFormats, val) {
						return errors.New("invalid log level format specified")
					}

					return nil
				},
				Destination: &flags.zapLogFormat,
			},
			&cli.DurationFlag{
				Name:        "resync-interval",
				Usage:       "requeue interval of the controllers",
				Value:       30 * time.Second,
				Sources:     cli.EnvVars("RESYNC_INTERVAL"),
				Destination: &flags.resyncInterval,
			},
			&cli.Float32Flag{
				Name:        "client-conn-qps",
				Usage:       "allowed client queries per second for the connection",
				Value:       100.0,
				Sources:     cli.EnvVars("CLIENT_CONNECTION_QPS"),
				Destination: &flags.clientConnQPS,
			},
			&cli.Int32Flag{
				Name:        "client-conn-burst",
				Usage:       "client connection burst size",
				Value:       130,
				Sources:     cli.EnvVars("CLIENT_CONNECTION_BURST"),
				Destination: &flags.clientConnBurst,
			},
			// The following flags are meant to be specified by the
			// Helm chart, which is rendered and deployed by the
			// gardenlet.
			//
			// During deployment the gardenlet will provide extra
			// values to Helm, which our CLI app can pick up.
			&cli.StringFlag{
				Name:        "gardener-version",
				Usage:       "version of gardener provided by gardenlet during deployment",
				Destination: &flags.gardenerVersion,
			},
			&cli.StringMapFlag{
				Name:  "gardenlet-feature-gate",
				Usage: "gardenlet feature gate provided by gardenlet during deployment",
				Action: func(ctx context.Context, c *cli.Command, items map[string]string) error {
					for feat, val := range items {
						enabled, err := strconv.ParseBool(val)
						if err != nil {
							return fmt.Errorf("invalid value for gardenlet feature gate: %w", err)
						}
						flags.gardenletFeatureGates[featuregate.Feature(feat)] = enabled
					}

					return nil
				},
			},
		},
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			ctrllog.SetLogger(glogger.MustNewZapLogger(flags.zapLogLevel, flags.zapLogFormat))
			newCtx := context.WithValue(ctx, flagsKey{}, &flags)

			return newCtx, nil
		},
		Action: runManager,
	}

	return cmd
}

// runManager starts the controller manager
func runManager(ctx context.Context, cmd *cli.Command) error {
	logger := ctrllog.Log.WithName("manager-setup")
	logger.Info("creating manager")

	flags := getFlags(ctx)
	m, err := flags.getManager(ctx)
	if err != nil {
		return err
	}

	logger.Info("creating actuators")
	decoder := serializer.NewCodecFactory(m.GetScheme(), serializer.EnableStrict).UniversalDecoder()
	act, err := actuator.New(
		m.GetClient(),
		actuator.WithDecoder(decoder),
		actuator.WithGardenerVersion(flags.gardenerVersion),
		actuator.WithGardenletFeatures(flags.gardenletFeatureGates),
	)
	if err != nil {
		return fmt.Errorf("failed to create actuator: %w", err)
	}

	logger.Info("creating controllers")
	c, err := controller.New(
		controller.WithActuator(act),
		controller.WithName(act.Name()),
		controller.WithExtensionType(act.ExtensionType()),
		controller.WithFinalizerSuffix(act.FinalizerSuffix()),
		controller.WithExtensionClass(act.ExtensionClass()),
		controller.WithIgnoreOperationAnnotation(flags.ignoreOperationAnnotation),
		controller.WithResyncInterval(flags.resyncInterval),
	)
	if err != nil {
		return fmt.Errorf("failed to create a controller: %w", err)
	}

	if err := c.SetupWithManager(ctx, m); err != nil {
		return fmt.Errorf("failed to setup controller with manager: %w", err)
	}

	if flags.gardenerVersion != "" {
		logger.Info("configured gardener version", "version", flags.gardenerVersion)
	}
	for feat, enabled := range flags.gardenletFeatureGates {
		logger.Info("configured gardenlet feature gate", "feature", feat, "enabled", enabled)
	}

	logger.Info("starting manager")
	if err := m.Start(ctx); err != nil {
		return fmt.Errorf("failed to start manager: %w", err)
	}

	return nil
}
