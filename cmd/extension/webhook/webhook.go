// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package webhook

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"slices"

	extensionscmdcontroller "github.com/gardener/gardener/extensions/pkg/controller/cmd"
	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	extensionscmdwebhook "github.com/gardener/gardener/extensions/pkg/webhook/cmd"
	gardencoreinstall "github.com/gardener/gardener/pkg/apis/core/install"
	gardenerhealthz "github.com/gardener/gardener/pkg/healthz"
	glogger "github.com/gardener/gardener/pkg/logger"
	"github.com/go-logr/logr"
	"github.com/urfave/cli/v3"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	componentbaseconfigv1alpha1 "k8s.io/component-base/config/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	admissionvalidator "github.com/gardener/gardener-extension-otelcol/pkg/admission/validator"
	configinstall "github.com/gardener/gardener-extension-otelcol/pkg/apis/config/install"
	"github.com/gardener/gardener-extension-otelcol/pkg/mgr"
)

// flags stores the webhook flags as provided from the command-line
type flags struct {
	extensionName               string
	metricsBindAddr             string
	healthProbeBindAddr         string
	leaderElection              bool
	leaderElectionID            string
	leaderElectionNamespace     string
	kubeconfig                  string
	gardenKubeconfig            string
	zapLogLevel                 string
	zapLogFormat                string
	pprofBindAddr               string
	clientConnQPS               float32
	clientConnBurst             int32
	webhookServerHost           string
	webhookServerPort           int
	webhookServerCertDir        string
	webhookServerCertName       string
	webhookServerKeyName        string
	webhookConfigNamespace      string
	webhookConfigMode           string
	webhookConfigURL            string
	webhookConfigServicePort    int
	webhookConfigOwnerNamespace string
	gardenerVersion             string
	selfHostedShootCluster      bool
	sourceCluster               cluster.Cluster
}

// getLogger returns a [logr.Logger] based on the specified command-line
// options.
func (f *flags) getLogger() logr.Logger {
	return glogger.MustNewZapLogger(f.zapLogLevel, f.zapLogFormat)
}

// getManager creates a new [ctrl.Manager] based on the parsed [flags].
func (f *flags) getManager(ctx context.Context) (ctrl.Manager, error) {
	logger := f.getLogger()
	webhookOpts := webhook.Options{
		Host:     f.webhookServerHost,
		Port:     f.webhookServerPort,
		CertDir:  f.webhookServerCertDir,
		CertName: f.webhookServerCertName,
		KeyName:  f.webhookServerKeyName,
	}
	webhookServer := webhook.NewServer(webhookOpts)

	// Gardener extension webhooks are usually deployed in one Kubernetes
	// cluster, called `source cluster' here, and are validating/mutating
	// resources located in another cluster, called `target cluster' here.
	//
	// The controller manager serving the extension webhooks will be
	// configured to use the `source cluster' for leader election, as well
	// as storing required TLS secrets for the webhook server. TLS secrets
	// are managed by a separate reconciler, which is installed by the
	// Gardener extension webhook utility package.
	//
	// Since admission webhooks are deployed in the `runtime cluster', then
	// the `source cluster' is the runtime cluster itself.
	//
	// The `target cluster' is the (virtual) Garden cluster, where resources
	// validated/mutated by webhooks reside.
	sourceClusterConfig, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load source cluster config: %w", err)
	}

	sourceCluster, err := cluster.New(sourceClusterConfig, func(opts *cluster.Options) {
		opts.Logger = logger
		opts.Cache.DefaultNamespaces = map[string]cache.Config{f.webhookConfigNamespace: {}}
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create source cluster from config: %w", err)
	}
	// Persist the source cluster, since we will need it later when
	// registering our webhooks with the Gardener extension webhook utility
	// package.
	f.sourceCluster = sourceCluster

	targetClusterConfig, err := clientcmd.BuildConfigFromFlags("", f.gardenKubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load garden cluster config: %w", err)
	}

	// Base set of controller manager options
	managerOpts := []mgr.Option{
		mgr.WithContext(ctx),
		mgr.WithConfig(targetClusterConfig),
		mgr.WithAddToScheme(clientgoscheme.AddToScheme),
		mgr.WithInstallScheme(gardencoreinstall.Install),
		mgr.WithInstallScheme(configinstall.Install),
		mgr.WithMetricsAddress(f.metricsBindAddr),
		mgr.WithHealthProbeAddress(f.healthProbeBindAddr),
		mgr.WithLeaderElection(f.leaderElection),
		mgr.WithLeaderElectionID(f.leaderElectionID),
		mgr.WithLeaderElectionNamespace(f.leaderElectionNamespace),
		mgr.WithLeaderElectionConfig(sourceClusterConfig),
		mgr.WithHealthzCheck("healthz", healthz.Ping),
		mgr.WithReadyzCheck("readyz", healthz.Ping),
		mgr.WithPprofAddress(f.pprofBindAddr),
		mgr.WithConnectionConfiguration(&componentbaseconfigv1alpha1.ClientConnectionConfiguration{
			QPS:   f.clientConnQPS,
			Burst: f.clientConnBurst,
		}),
		mgr.WithWebhookServer(webhookServer),
		mgr.WithReadyzCheck("webhook-server", webhookServer.StartedChecker()),
		mgr.WithReadyzCheck("source-informer-sync", gardenerhealthz.NewCacheSyncHealthz(sourceCluster.GetCache())),
		mgr.WithRunnable(sourceCluster),
	}

	m, err := mgr.New(managerOpts...)
	if err != nil {
		return nil, err
	}

	if err := m.AddReadyzCheck("informer-sync", gardenerhealthz.NewCacheSyncHealthz(m.GetCache())); err != nil {
		return nil, fmt.Errorf("failed to setup ready check: %w", err)
	}

	return m, nil
}

// getExtensionWebhookOpts returns [extensionscmdwebhook.AddToManagerOptions]
// based on the specified command-line flags.
func (f *flags) getExtensionWebhookOpts() *extensionscmdwebhook.AddToManagerOptions {
	serverOpts := &extensionscmdwebhook.ServerOptions{
		Mode:           f.webhookConfigMode,
		URL:            f.webhookConfigURL,
		ServicePort:    f.webhookConfigServicePort,
		Namespace:      f.webhookConfigNamespace,
		OwnerNamespace: f.webhookConfigOwnerNamespace,
	}

	generalOpts := &extensionscmdcontroller.GeneralOptions{
		GardenerVersion:        f.gardenerVersion,
		SelfHostedShootCluster: f.selfHostedShootCluster,
	}

	addToManagerOpts := extensionscmdwebhook.NewAddToManagerOptions(
		f.extensionName,
		"",
		nil,
		generalOpts,
		serverOpts,
		&extensionscmdwebhook.SwitchOptions{},
	)

	return addToManagerOpts
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

// New creates a new [cli.Command] for running the webhook server.
func New() *cli.Command {
	flags := flags{}

	cmd := &cli.Command{
		Name:    "webhook",
		Aliases: []string{"w"},
		Usage:   "start extension webhook server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "extension-name",
				Usage:       "name of the gardener extension",
				Value:       "gardener-extension-otelcol-admission",
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
				Value:       "gardener-extension-otelcol-admission",
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
				Name:        "garden-kubeconfig",
				Required:    true,
				Aliases:     []string{"target-kubeconfig"},
				Usage:       "path to a kubeconfig for the garden cluster",
				Sources:     cli.EnvVars("GARDEN_KUBECONFIG"),
				Destination: &flags.gardenKubeconfig,
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
			&cli.Float32Flag{
				Name:        "client-conn-qps",
				Usage:       "allowed client queries per second for the connection",
				Value:       -1.0,
				Sources:     cli.EnvVars("CLIENT_CONNECTION_QPS"),
				Destination: &flags.clientConnQPS,
			},
			&cli.Int32Flag{
				Name:        "client-conn-burst",
				Usage:       "client connection burst size",
				Value:       0,
				Sources:     cli.EnvVars("CLIENT_CONNECTION_BURST"),
				Destination: &flags.clientConnBurst,
			},
			&cli.StringFlag{
				Name:        "gardener-version",
				Usage:       "version of gardener provided by gardenlet or gardener-operator",
				Sources:     cli.EnvVars("GARDENER_VERSION"),
				Destination: &flags.gardenerVersion,
			},
			&cli.BoolFlag{
				Name:        "self-hosted-shoot-cluster",
				Usage:       "set to true, if the extension runs in a self-hosted shoot cluster",
				Sources:     cli.EnvVars("SELF_HOSTED_SHOOT_CLUSTER"),
				Destination: &flags.selfHostedShootCluster,
			},
			&cli.StringFlag{
				Name:        "webhook-server-host",
				Usage:       "address on which the webhook server listens on",
				Sources:     cli.EnvVars("WEBHOOK_SERVER_HOST"),
				Destination: &flags.webhookServerHost,
			},
			&cli.IntFlag{
				Name:        "webhook-server-port",
				Value:       9443,
				Usage:       "port on which the webhook server listens on",
				Sources:     cli.EnvVars("WEBHOOK_SERVER_PORT"),
				Destination: &flags.webhookServerPort,
			},
			&cli.StringFlag{
				Name:        "webhook-server-cert-dir",
				Usage:       "path to directory, which contains the server key and cert",
				Sources:     cli.EnvVars("WEBHOOK_SERVER_CERT_DIR"),
				Destination: &flags.webhookServerCertDir,
			},
			&cli.StringFlag{
				Name:        "webhook-server-cert-name",
				Value:       "tls.crt",
				Usage:       "the server certificate file name",
				Sources:     cli.EnvVars("WEBHOOK_SERVER_CERT_NAME"),
				Destination: &flags.webhookServerCertName,
			},
			&cli.StringFlag{
				Name:        "webhook-server-key-name",
				Value:       "tls.key",
				Usage:       "the server certificate key file name",
				Sources:     cli.EnvVars("WEBHOOK_SERVER_KEY_NAME"),
				Destination: &flags.webhookServerKeyName,
			},
			&cli.StringFlag{
				Name:        "webhook-config-namespace",
				Value:       "garden",
				Usage:       "namespace where the webhook CA bundle, services, etc. are created",
				Sources:     cli.EnvVars("WEBHOOK_CONFIG_NAMESPACE"),
				Destination: &flags.webhookConfigNamespace,
			},
			&cli.StringFlag{
				Name:    "webhook-config-mode",
				Value:   string(extensionswebhook.ModeService),
				Usage:   "one of service, url or url-service",
				Sources: cli.EnvVars("WEBHOOK_CONFIG_MODE"),
				Validator: func(val string) error {
					supportedModes := []string{
						string(extensionswebhook.ModeService),
						string(extensionswebhook.ModeURL),
						string(extensionswebhook.ModeURLWithServiceName),
					}
					if !slices.Contains(supportedModes, val) {
						return errors.New("invalid webhook config mode specified")
					}

					return nil
				},
				Destination: &flags.webhookConfigMode,
			},
			&cli.StringFlag{
				Name:    "webhook-config-url",
				Usage:   "URL at which to find the webhook server, used with `url' mode only",
				Sources: cli.EnvVars("WEBHOOK_CONFIG_URL"),
				Validator: func(val string) error {
					_, err := url.Parse(val)

					return err
				},
				Destination: &flags.webhookConfigURL,
			},
			&cli.IntFlag{
				Name:    "webhook-config-service-port",
				Usage:   "service port for the webhook when running in `service' mode",
				Sources: cli.EnvVars("WEBHOOK_CONFIG_SERVICE_PORT"),
				Validator: func(val int) error {
					if val <= 0 {
						return errors.New("port cannot be negative")
					}

					return nil
				},
				Destination: &flags.webhookConfigServicePort,
			},
			&cli.StringFlag{
				Name:        "webhook-config-owner-namespace",
				Usage:       "namespace which is used as the owner reference for webhook registration",
				Sources:     cli.EnvVars("WEBHOOK_CONFIG_OWNER_NAMESPACE"),
				Destination: &flags.webhookConfigOwnerNamespace,
			},
		},
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			ctrllog.SetLogger(flags.getLogger())
			newCtx := context.WithValue(ctx, flagsKey{}, &flags)

			return newCtx, nil
		},
		Action: runWebhookServer,
	}

	return cmd
}

// runWebhookServer starts the webhook server.
func runWebhookServer(ctx context.Context, cmd *cli.Command) error {
	logger := ctrllog.Log.WithName("manager-setup")
	logger.Info("creating manager")

	flags := getFlags(ctx)
	m, err := flags.getManager(ctx)
	if err != nil {
		return err
	}

	logger.Info("setting up admission webhooks")

	// Webhooks to be registered
	webhooks := make([]*extensionswebhook.Webhook, 0)
	webhookFuncs := []func(m ctrl.Manager) (*extensionswebhook.Webhook, error){
		admissionvalidator.NewShootValidatorWebhook,
	}

	for _, webhookFunc := range webhookFuncs {
		wh, err := webhookFunc(m)
		if err != nil {
			return fmt.Errorf("failed to create admission webhook: %w", err)
		}
		webhooks = append(webhooks, wh)
	}

	extensionWebhookOpts := flags.getExtensionWebhookOpts()
	if err := extensionWebhookOpts.Complete(); err != nil {
		return err
	}
	extensionWebhookConfig := extensionWebhookOpts.Completed()
	extensionWebhookConfig.Switch = extensionscmdwebhook.SwitchConfig{
		Disabled: false,
		WebhooksFactory: func(m manager.Manager) ([]*extensionswebhook.Webhook, error) {
			return webhooks, nil
		},
	}

	if _, err := extensionWebhookConfig.AddToManager(ctx, m, flags.sourceCluster); err != nil {
		return fmt.Errorf("failed to setup extension webhook with manager: %w", err)
	}

	logger.Info("starting manager")

	return m.Start(ctx)
}
