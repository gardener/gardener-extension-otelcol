// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package actuator provides the implementation of a Gardener extension
// actuator.
package actuator

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/extension"
	v1beta1helper "github.com/gardener/gardener/pkg/api/core/v1beta1/helper"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	resourcesv1alpha1 "github.com/gardener/gardener/pkg/apis/resources/v1alpha1"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	gardenerfeatures "github.com/gardener/gardener/pkg/features"
	"github.com/gardener/gardener/pkg/utils"
	gardenerutils "github.com/gardener/gardener/pkg/utils/gardener"
	imagevectorutils "github.com/gardener/gardener/pkg/utils/imagevector"
	kubernetesutils "github.com/gardener/gardener/pkg/utils/kubernetes"
	"github.com/gardener/gardener/pkg/utils/managedresources"
	secretsutils "github.com/gardener/gardener/pkg/utils/secrets"
	secretsmanager "github.com/gardener/gardener/pkg/utils/secrets/manager"
	otelv1alpha1 "github.com/gardener/gardener/third_party/open-telemetry/opentelemetry-operator/apis/v1alpha1"
	otelv1beta1 "github.com/gardener/gardener/third_party/open-telemetry/opentelemetry-operator/apis/v1beta1"
	"github.com/go-logr/logr"
	"go.opentelemetry.io/collector/processor/batchprocessor"
	"go.opentelemetry.io/collector/processor/memorylimiterprocessor"
	"go.yaml.in/yaml/v4"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/component-base/featuregate"
	"k8s.io/utils/clock"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/gardener-extension-otelcol/pkg/apis/config"
	"github.com/gardener/gardener-extension-otelcol/pkg/apis/config/validation"
	"github.com/gardener/gardener-extension-otelcol/pkg/imagevector"
)

// ErrInvalidActuator is an error which is returned when creating an [Actuator]
// with invalid config settings.
var ErrInvalidActuator = errors.New("invalid actuator")

const (
	// Name is the name of the actuator
	Name = "otelcol"
	// ExtensionType is the type of the extension resources, which the
	// actuator reconciles.
	ExtensionType = "otelcol"
	// FinalizerSuffix is the finalizer suffix used by the actuator
	FinalizerSuffix = "gardener-extension-otelcol"

	// baseResourceName is the base name for resources.
	baseResourceName = "external-otelcol"

	// managedResourceName is the name of the managed resource created by
	// the actuator.
	managedResourceName = baseResourceName

	// otelCollectorName is the name of the
	// [otelv1beta1.OpenTelemetryCollector] resource created by the
	// extension.
	otelCollectorName = baseResourceName
	// otelCollectorMetricsPort is the port on which the OTel Collector
	// exposes it's internal metrics.
	otelCollectorMetricsPort = 8888
	// otelCollectorReplicas specifies the number of replicas of the OTel
	// Collector.
	otelCollectorReplicas int32 = 1
	// otelCollectorServiceAccountName is the name of the service account
	// for the OTel Collector.
	otelCollectorServiceAccountName = otelCollectorName + "-collector"
	// otelCollectorGRPCReceiverPort is the port on which the OTel collector
	// binds the gRPC receiver.
	otelCollectorGRPCReceiverPort = 4317

	// secretsManagerIdentity is the identity used for secrets management.
	secretsManagerIdentity = "gardener-extension-" + Name
	// secretNameCACertificate is the name of the CA certificate secret.
	secretNameCACertificate = "ca-" + Name
	// secretNameServerCertificate is the name of the server certificate of the Target Allocator.
	secretNameServerCertificate = Name + "-targetallocator-server"
	// secretNameClientCertificate is the name of the server certificate of the Target Allocator.
	secretNameClientCertificate = Name + "-collector-client"

	// targetAllocatorDeploymentName is the name of the deployment for the
	// Target Allocator.
	targetAllocatorDeploymentName = baseResourceName + "-targetallocator"
	// targetAllocatorHTTPSServiceName is the name of the Kubernetes service for
	// HTTPS communication of the Target Allocator.
	targetAllocatorHTTPSServiceName = baseResourceName + "-targetallocator-https"
	// targetAllocatorHTTPSPort is the port on which Target Allocator's
	// HTTPS service listens to.
	targetAllocatorHTTPSPort = 8443
	// targetAllocatorServiceAccountName is the name of the service account
	// for the Target Allocator.
	targetAllocatorServiceAccountName = baseResourceName + "-targetallocator"
	// targetAllocatorReplicas specifies the number of replicas of the Target Allocator.
	targetAllocatorReplicas int32 = 1
	// targetAllocatorRoleName is the name of the Role and RoleBinding
	// resource for the Target Allocator.
	targetAllocatorRoleName = baseResourceName + "-targetallocator"
	// targetAllocatorConfigMapName is the name of the ConfigMap for the
	// Target Allocator.
	targetAllocatorConfigMapName = baseResourceName + "-targetallocator-config"

	// transformEventsProcessorName is the name of the transform processor for
	// the k8sobjects/events pipeline.
	transformEventsProcessorName = "transform/events"

	// shootAccessSecretName is the name of the shoot access secret used by the
	// k8sobjects/events receiver to authenticate to the shoot cluster.
	shootAccessSecretName = "shoot-access-" + otelCollectorName // #nosec: G101

	// shootManagedResourceName is the name of the ManagedResource that deploys
	// RBAC into the shoot cluster for the k8sobjects/events receiver.
	shootManagedResourceName = baseResourceName + "-shoot"

	// volumeNameShootKubeconfig is the volume name for the shoot kubeconfig
	// projected into the OTel Collector pod for the k8sobjects/events receiver.
	volumeNameShootKubeconfig = "shoot-kubeconfig"

	// bearertokenauthextension base name used to derive per-signal names.
	baseBearerTokenAuthName = "bearertokenauth"

	// TLS volume base name used to derive per-signal volume names.
	baseVolumeNameTLS = "tls"

	// TLS volume mount base path used to derive per-signal mount paths.
	baseVolumeMountPathTLS = "/etc/ssl/tls"

	// batchProcessorName is the name of the OpenTelemetry Batch processor.
	batchProcessorName = "batch"

	// memoryLimiterProcessorName is the name of the OpenTelemetry Memory
	// Limiter processor name.
	memoryLimiterProcessorName = "memory_limiter"

	// resourceProcessorName is the name of the OpenTelemetry Resource processor.
	resourceProcessorName = "resource"

	// filterProcessorBaseName is the base name of the OpenTelemetry Filter
	// processor, used to derive per-signal and global filter processor names.
	filterProcessorBaseName = "filter"

	// otlpReceiverName is the name of the OTLP receiver. It feeds the logs,
	// traces and profiles signals.
	otlpReceiverName = "otlp"

	// eventsReceiverName is the name of the k8sobjects receiver for events.
	eventsReceiverName = "k8sobjects/events"

	// prometheusReceiverName is the name of the Prometheus receiver.
	prometheusReceiverName = "prometheus"

	// debugExporterBaseName is the base name of the debug exporter, used to
	// derive per-signal debug exporter names.
	debugExporterBaseName = "debug"

	// logsPipelineName is the name of the logs pipeline.
	logsPipelineName = "logs"

	// eventsPipelineName is the name of the events pipeline.
	eventsPipelineName = "logs/events"

	// metricsPipelineName is the name of the metrics pipeline.
	metricsPipelineName = "metrics"

	// tracesPipelineName is the name of the traces pipeline.
	tracesPipelineName = "traces"

	// profilesPipelineName is the name of the profiles pipeline.
	profilesPipelineName = "profiles"

	// telemetryMetricsKey is the telemetry config key for metrics settings.
	telemetryMetricsKey = "metrics"

	// telemetryLogsKey is the telemetry config key for logs settings.
	telemetryLogsKey = "logs"

	// labelKeyComponent is the standard kubernetes app component label key.
	labelKeyComponent = "app.kubernetes.io/component"
	// labelValueTargetAllocator is the component label value identifying the
	// Target Allocator workload.
	labelValueTargetAllocator = "opentelemetry-targetallocator"

	// keys used in OTel/Target Allocator config maps.
	configKeyEnabled    = "enabled"
	configKeyEndpoint   = "endpoint"
	configKeyPrometheus = "prometheus"
	configKeyKey        = "key"
	configKeyValue      = "value"
	configKeyAction     = "action"
	configKeyMatchType  = "match_type"
	// labelValuePrometheusShoot is the value used for the `prometheus` label on
	// service monitors that should be scraped in the shoot.
	labelValuePrometheusShoot = "shoot"
)

// readVerbs is the canonical RBAC verb set for read-only access to a resource.
var readVerbs = []string{"get", "list", "watch"}

// signalPipelineName returns the collector service pipeline name for the i-th
// target of a signal, e.g. "metrics/0".
func signalPipelineName(sig config.SignalType, i int) string {
	return fmt.Sprintf("%s/%d", signalPipelineBaseName(sig), i)
}

// signalPipelineBaseName returns the base pipeline name for a signal. Because
// the collector infers the pipeline signal type from this prefix, the events
// signal (collected as logs) maps to the "logs" base.
func signalPipelineBaseName(sig config.SignalType) string {
	switch sig {
	case config.SignalMetrics:
		return metricsPipelineName
	case config.SignalLogs:
		return logsPipelineName
	case config.SignalTraces:
		return tracesPipelineName
	case config.SignalProfiles:
		return profilesPipelineName
	case config.SignalEvents:
		return eventsPipelineName
	default:
		return string(sig)
	}
}

// signalReceiverName returns the receiver feeding a signal's pipeline.
func signalReceiverName(sig config.SignalType) string {
	switch sig {
	case config.SignalMetrics:
		return prometheusReceiverName
	case config.SignalEvents:
		return eventsReceiverName
	default:
		// Logs, traces and profiles are all received via OTLP.
		return otlpReceiverName
	}
}

// signalExporterName returns the exporter component name for the i-th target of
// a signal and the chosen protocol, e.g. "otlphttp/metrics/0", "otlp/events/0"
// or "debug/metrics/2".
func signalExporterName(sig config.SignalType, i int, proto config.ExporterProtocol) string {
	base := "otlphttp"
	switch proto {
	case config.ExporterProtocolGRPC:
		base = "otlp"
	case config.ExporterProtocolDebug:
		base = debugExporterBaseName
	default:
		// HTTP is the default protocol; keep the otlphttp base.
	}

	return fmt.Sprintf("%s/%s/%d", base, sig, i)
}

// signalFilterName returns the filter processor component name for the
// ruleIdx-th filter rule of a signal's i-th target, e.g. "filter/metrics/0/1".
func signalFilterName(sig config.SignalType, i, ruleIdx int) string {
	return fmt.Sprintf("%s/%s/%d/%d", filterProcessorBaseName, sig, i, ruleIdx)
}

// signalBearerTokenAuthName returns the bearertokenauth extension name for the
// i-th target of a signal, e.g. "bearertokenauth/metrics/0".
func signalBearerTokenAuthName(sig config.SignalType, i int) string {
	return fmt.Sprintf("%s/%s/%d", baseBearerTokenAuthName, sig, i)
}

// signalVolumeNameTLS returns the TLS volume name for the i-th target of a
// signal, e.g. "tls-metrics-0".
func signalVolumeNameTLS(sig config.SignalType, i int) string {
	return fmt.Sprintf("%s-%s-%d", baseVolumeNameTLS, sig, i)
}

// signalVolumeMountPathTLS returns the TLS volume mount path for the i-th
// target of a signal, e.g. "/etc/ssl/tls-metrics-0".
func signalVolumeMountPathTLS(sig config.SignalType, i int) string {
	return fmt.Sprintf("%s-%s-%d", baseVolumeMountPathTLS, sig, i)
}

// signalVolumeNameBearerToken returns the bearer token volume name for the i-th
// target of a signal.
func signalVolumeNameBearerToken(sig config.SignalType, i int) string {
	return fmt.Sprintf("bearer-token-auth-%s-%d", sig, i)
}

// signalVolumeMountPathBearerToken returns the bearer token mount path for the
// i-th target of a signal.
func signalVolumeMountPathBearerToken(sig config.SignalType, i int) string {
	return fmt.Sprintf("/etc/auth/bearer-%s-%d", sig, i)
}

// upsertAttribute returns an OTel resourceprocessor `attributes` entry that
// adds (or overwrites) the given key/value on the resource.
func upsertAttribute(key string, value any) map[string]any {
	return map[string]any{
		configKeyKey:    key,
		configKeyValue:  value,
		configKeyAction: "upsert",
	}
}

// Actuator is an implementation of [extension.Actuator].
type Actuator struct {
	client               client.Client
	decoder              runtime.Decoder
	memoryLimiterConfig  *memorylimiterprocessor.Config
	batchProcessorConfig *batchprocessor.Config

	// The following fields are usually derived from the list of extra Helm
	// values provided by gardenlet during the deployment of the extension.
	//
	// See the link below for more details about how gardenlet provides
	// extra values to Helm during the extension deployment.
	//
	// https://github.com/gardener/gardener/blob/d5071c800378616eb6bb2c7662b4b28f4cfe7406/pkg/gardenlet/controller/controllerinstallation/controllerinstallation/reconciler.go#L236-L263
	gardenerVersion       string
	gardenletFeatureGates map[featuregate.Feature]bool
}

var _ extension.Actuator = &Actuator{}

// Option is a function, which configures the [Actuator].
type Option func(a *Actuator) error

// New creates a new actuator with the given options.
func New(c client.Client, opts ...Option) (*Actuator, error) {
	if c == nil {
		return nil, fmt.Errorf("%w: no client specified", ErrInvalidActuator)
	}

	act := &Actuator{
		client:                c,
		gardenletFeatureGates: make(map[featuregate.Feature]bool),
		memoryLimiterConfig: &memorylimiterprocessor.Config{
			CheckInterval:         time.Second,
			MemoryLimitPercentage: 75,

			// Default from OTel
			//
			// https://github.com/open-telemetry/opentelemetry-collector/blob/168030d61d7db2a15176f3e52ab4fd1e96012f15/internal/memorylimiter/config.go#L61
			MinGCIntervalWhenSoftLimited: 10 * time.Second,
		},
		batchProcessorConfig: &batchprocessor.Config{
			Timeout:       5 * time.Second,
			SendBatchSize: 8192,
		},
	}

	for _, opt := range opts {
		if err := opt(act); err != nil {
			return nil, err
		}
	}

	if act.decoder == nil {
		act.decoder = serializer.NewCodecFactory(c.Scheme(), serializer.EnableStrict).UniversalDecoder()
	}

	return act, nil
}

// WithDecoder is an [Option], which configures the [Actuator] with the given
// [runtime.Decoder].
func WithDecoder(d runtime.Decoder) Option {
	opt := func(a *Actuator) error {
		a.decoder = d

		return nil
	}

	return opt
}

// WithGardenerVersion is an [Option], which configures the [Actuator] with the
// given version of Gardener. This version of Gardener is usually provided by
// the gardenlet as part of the extra Helm values during deployment of the
// extension.
func WithGardenerVersion(v string) Option {
	opt := func(a *Actuator) error {
		a.gardenerVersion = v

		return nil
	}

	return opt
}

// WithGardenletFeatures is an [Option], which configures the [Actuator] with
// the given gardenlet feature gates. These feature gates are usually provided
// by the gardenlet as part of the extra Helm values during deployment of the
// extension.
func WithGardenletFeatures(feats map[featuregate.Feature]bool) Option {
	opt := func(a *Actuator) error {
		a.gardenletFeatureGates = feats

		return nil
	}

	return opt
}

// WithMemoryLimiterProcessorConfig is an [Option], which configures the
// [Actuator] to create an OTel collector configured with the Memory Limiter
// Processor based on the provided configuration.
func WithMemoryLimiterProcessorConfig(cfg *memorylimiterprocessor.Config) Option {
	opt := func(a *Actuator) error {
		if cfg == nil {
			return errors.New("invalid memory limiter configuration specified")
		}

		// https://github.com/open-telemetry/opentelemetry-collector/blob/168030d61d7db2a15176f3e52ab4fd1e96012f15/internal/memorylimiter/config.go#L61
		cfg.MinGCIntervalWhenSoftLimited = 10 * time.Second
		a.memoryLimiterConfig = cfg

		return cfg.Validate()
	}

	return opt
}

// WithBatchProcessorConfig is an [Option], which configures the [Actuator] to
// create an OTel collector configured with the Batch Processor based on the
// provided configuration.
func WithBatchProcessorConfig(cfg *batchprocessor.Config) Option {
	opt := func(a *Actuator) error {
		if cfg == nil {
			return errors.New("invalid batch processor configuration specified")
		}

		a.batchProcessorConfig = cfg

		return cfg.Validate()
	}

	return opt
}

// Name returns the name of the actuator. This name can be used when registering
// a controller for the actuator.
func (a *Actuator) Name() string {
	return Name
}

// FinalizerSuffix returns the finalizer suffix to use for the actuator. The
// result of this method may be used when registering a controller with the
// actuator.
func (a *Actuator) FinalizerSuffix() string {
	return FinalizerSuffix
}

// ExtensionType returns the type of extension resources the actuator
// reconciles. The result of this method may be used when registering a
// controller with the actuator.
func (a *Actuator) ExtensionType() string {
	return ExtensionType
}

// ExtensionClass returns the [extensionsv1alpha1.ExtensionClass] for the
// actuator. The result of this method may be used when registering a controller
// with the actuator.
func (a *Actuator) ExtensionClass() extensionsv1alpha1.ExtensionClass {
	return extensionsv1alpha1.ExtensionClassShoot
}

// Reconcile reconciles the [extensionsv1alpha1.Extension] resource by taking
// care of any resources managed by the [Actuator]. This method implements the
// [extension.Actuator] interface.
func (a *Actuator) Reconcile(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	otelcolFeature, ok := a.gardenletFeatureGates[gardenerfeatures.OpenTelemetryCollector]
	if !ok || !otelcolFeature {
		logger.Info("gardenlet feature gate OpenTelemetryCollector is either missing or disabled")

		return a.Delete(ctx, logger, ex)
	}

	// The cluster name is the same as the name of the namespace for our
	// [extensionsv1alpha1.Extension] resource.
	clusterName := ex.Namespace

	secretsManager, err := a.newSecretsManager(ctx, logger, ex.Namespace)
	if err != nil {
		return fmt.Errorf("failed creating a new secrets manager: %w", err)
	}

	logger.Info("reconciling extension", "name", ex.Name, "cluster", clusterName)

	cluster, err := extensionscontroller.GetCluster(ctx, a.client, clusterName)
	if err != nil {
		return fmt.Errorf("failed to get cluster: %w", err)
	}

	// Nothing to do here, if the shoot cluster is hibernated at the moment.
	if v1beta1helper.HibernationIsEnabled(cluster.Shoot) {
		return nil
	}

	// Parse and validate the provider config
	if ex.Spec.ProviderConfig == nil {
		return errors.New("no provider config specified")
	}

	var cfg config.CollectorConfig
	if err := runtime.DecodeInto(a.decoder, ex.Spec.ProviderConfig.Raw, &cfg); err != nil {
		return fmt.Errorf("invalid provider spec configuration: %w", err)
	}

	if err := validation.Validate(cfg); err != nil {
		return err
	}

	// Generate CA and server certificate for Target Allocator
	if _, err := secretsManager.Generate(ctx, &secretsutils.CertificateSecretConfig{
		Name:       secretNameCACertificate,
		CommonName: Name,
		CertType:   secretsutils.CACert,
		Validity:   ptr.To(30 * 24 * time.Hour),
	}, secretsmanager.Rotate(secretsmanager.KeepOld), secretsmanager.IgnoreOldSecretsAfter(24*time.Hour)); err != nil {
		return fmt.Errorf("failed generating CA certificate secret: %w", err)
	}
	caBundleSecret, _ := secretsManager.Get(secretNameCACertificate)

	serverSecret, err := secretsManager.Generate(ctx, &secretsutils.CertificateSecretConfig{
		Name:                        secretNameServerCertificate,
		CommonName:                  targetAllocatorHTTPSServiceName,
		DNSNames:                    kubernetesutils.DNSNamesForService(targetAllocatorHTTPSServiceName, ex.Namespace),
		CertType:                    secretsutils.ServerCert,
		SkipPublishingCACertificate: true,
	}, secretsmanager.SignedByCA(secretNameCACertificate), secretsmanager.Rotate(secretsmanager.InPlace))
	if err != nil {
		return fmt.Errorf("failed generating server certificate secret for target allocator: %w", err)
	}

	clientSecret, err := secretsManager.Generate(ctx, &secretsutils.CertificateSecretConfig{
		Name:                        secretNameClientCertificate,
		CommonName:                  secretNameClientCertificate,
		CertType:                    secretsutils.ClientCert,
		SkipPublishingCACertificate: true,
	}, secretsmanager.SignedByCA(secretNameCACertificate), secretsmanager.Rotate(secretsmanager.InPlace))
	if err != nil {
		return fmt.Errorf("failed generating server certificate secret for target allocator: %w", err)
	}

	taImage, err := imagevector.Images().FindImage(imagevector.ImageNameOTelTargetAllocator)
	if err != nil {
		return fmt.Errorf("failed to find image: %w", err)
	}

	collectorImage, err := imagevector.Images().FindImage(imagevector.ImageNameOTelCollector)
	if err != nil {
		return fmt.Errorf("failed to find image: %w", err)
	}

	// Bundle things up in a managed resource
	registry := managedresources.NewRegistry(
		kubernetes.SeedScheme,
		kubernetes.SeedCodec,
		kubernetes.SeedSerializer,
	)

	taConfigMap, err := a.getTargetAllocatorConfigMap(ex.Namespace)
	if err != nil {
		return err
	}

	shootKubeconfigSecretName := extensionscontroller.GenericTokenKubeconfigSecretNameFromCluster(cluster)

	shootAccessSecret := gardenerutils.NewShootAccessSecret(shootAccessSecretName, ex.Namespace)
	if err := shootAccessSecret.Reconcile(ctx, a.client); err != nil {
		return fmt.Errorf("failed reconciling shoot access secret: %w", err)
	}

	data, err := registry.AddAllAndSerialize(
		taConfigMap,
		a.getTargetAllocatorServiceAccount(ex.Namespace),
		a.getTargetAllocatorRole(ex.Namespace),
		a.getTargetAllocatorRoleBinding(ex.Namespace),
		a.getTargetAllocatorHTTPSService(ex.Namespace),
		a.getTargetAllocatorDeployment(ex.Namespace, caBundleSecret, serverSecret, taImage),
		a.getOtelCollectorServiceAccount(ex.Namespace),
		a.getOtelCollector(
			ex.Namespace,
			caBundleSecret,
			clientSecret,
			cfg,
			cluster.Shoot.Spec.Resources,
			shootKubeconfigSecretName,
			shootAccessSecret.Secret.Name,
			collectorImage,
		),
	)

	if err != nil {
		return err
	}

	shootRegistry := managedresources.NewRegistry(
		kubernetes.ShootScheme,
		kubernetes.ShootCodec,
		kubernetes.ShootSerializer,
	)

	shootData, err := shootRegistry.AddAllAndSerialize(
		a.getEventsClusterRole(),
		a.getEventsClusterRoleBinding(shootAccessSecret.ServiceAccountName),
	)
	if err != nil {
		return err
	}

	if err := managedresources.CreateForShoot(ctx, a.client, ex.Namespace, shootManagedResourceName, Name, false, shootData); err != nil {
		return fmt.Errorf("failed creating shoot managed resource: %w", err)
	}

	return managedresources.CreateForSeed(
		ctx,
		a.client,
		ex.Namespace,
		managedResourceName,
		false,
		data,
	)
}

// Delete deletes any resources managed by the [Actuator]. This method
// implements the [extension.Actuator] interface.
func (a *Actuator) Delete(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	secretsManager, err := a.newSecretsManager(ctx, logger, ex.Namespace)
	if err != nil {
		return fmt.Errorf("failed creating a new secrets manager: %w", err)
	}

	logger.Info("deleting resources managed by extension")

	if err := secretsManager.Cleanup(ctx); err != nil {
		return fmt.Errorf("failed cleaning up secrets managed by secrets manager: %w", err)
	}

	if err := client.IgnoreNotFound(managedresources.DeleteForShoot(ctx, a.client, ex.Namespace, shootManagedResourceName)); err != nil {
		return fmt.Errorf("failed deleting shoot managed resource: %w", err)
	}

	if err := managedresources.WaitUntilDeleted(ctx, a.client, ex.Namespace, shootManagedResourceName); err != nil {
		return fmt.Errorf("failed waiting for shoot managed resource to be deleted: %w", err)
	}

	if err := client.IgnoreNotFound(a.client.Delete(ctx, gardenerutils.NewShootAccessSecret(shootAccessSecretName, ex.Namespace).Secret)); err != nil {
		return fmt.Errorf("failed deleting shoot access secret: %w", err)
	}

	return client.IgnoreNotFound(managedresources.DeleteForSeed(ctx, a.client, ex.Namespace, managedResourceName))
}

// ForceDelete signals the [Actuator] to delete any resources managed by it,
// because of a force-delete event of the shoot cluster. This method implements
// the [extension.Actuator] interface.
func (a *Actuator) ForceDelete(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	logger.Info("shoot has been force-deleted, deleting resources managed by extension")

	return a.Delete(ctx, logger, ex)
}

// Restore restores the resources managed by the extension [Actuator]. This
// method implements the [extension.Actuator] interface.
func (a *Actuator) Restore(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	return a.Reconcile(ctx, logger, ex)
}

// Migrate signals the [Actuator] to migrate the resources managed by it,
// because of a shoot control-plane migration event. This method implements the
// [extension.Actuator] interface.
//
// Shoot-scoped resources (RBAC) must be preserved on the shoot cluster so the
// target seed can pick them up after migration. SetKeepObjects prevents the
// ManagedResource controller from deleting them when the ManagedResource is
// removed from the old seed.
func (a *Actuator) Migrate(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	if err := managedresources.SetKeepObjects(ctx, a.client, ex.Namespace, shootManagedResourceName, true); err != nil {
		return fmt.Errorf("failed setting keep-objects on shoot managed resource: %w", err)
	}

	return a.Delete(ctx, logger, ex)
}

func (a *Actuator) newSecretsManager(ctx context.Context, log logr.Logger, namespace string) (secretsmanager.Interface, error) {
	return secretsmanager.New(
		ctx,
		log,
		clock.RealClock{},
		a.client,
		secretsManagerIdentity,
		secretsmanager.WithCASecretAutoRotation(),
		secretsmanager.WithNamespaces(namespace),
	)
}

// getCommonLabels returns the common set of labels for the Collector and Target
// Allocator resources.
func (a *Actuator) getCommonLabels() map[string]string {
	items := map[string]string{
		v1beta1constants.LabelRole:                     v1beta1constants.LabelObservability,
		v1beta1constants.GardenRole:                    v1beta1constants.GardenRoleObservability,
		v1beta1constants.LabelObservabilityApplication: otelCollectorName,
	}

	return items
}

// getNetworkLabels returns the set of labels related to Gardener Network
// Policies.
func (a *Actuator) getNetworkLabels() map[string]string {
	// The `networking.resources.gardener.cloud/to-all-scrape-targets' label
	toAllScrapeTargetsLabel := resourcesv1alpha1.NetworkPolicyLabelKeyPrefix + "to-" + v1beta1constants.LabelNetworkPolicyScrapeTargets

	items := map[string]string{
		v1beta1constants.LabelNetworkPolicyToDNS:              v1beta1constants.LabelNetworkPolicyAllowed,
		v1beta1constants.LabelNetworkPolicyToRuntimeAPIServer: v1beta1constants.LabelNetworkPolicyAllowed,
		v1beta1constants.LabelNetworkPolicyToPrivateNetworks:  v1beta1constants.LabelNetworkPolicyAllowed,
		v1beta1constants.LabelNetworkPolicyToPublicNetworks:   v1beta1constants.LabelNetworkPolicyAllowed,
		resourcesv1alpha1.NetworkPolicyLabelKeyPrefix + "to-" + targetAllocatorHTTPSServiceName + "-tcp-" + strconv.Itoa(targetAllocatorHTTPSPort): v1beta1constants.LabelNetworkPolicyAllowed,
		toAllScrapeTargetsLabel: v1beta1constants.LabelNetworkPolicyAllowed,
	}

	return items
}

// getAnnotations returns the common set of annotations for the Collector and
// Target Allocator resources.
func (a *Actuator) getAnnotations() map[string]string {
	// The `networking.resources.gardener.cloud/from-all-scrape-targets-allowed-ports' annotation
	fromAllScrapeTargetsAnnotation := resourcesv1alpha1.NetworkPolicyLabelKeyPrefix + "from-all-scrape-targets-allowed-ports"

	items := map[string]string{
		fromAllScrapeTargetsAnnotation: fmt.Sprintf(`[{"protocol":"TCP","port":%d},{"protocol":"TCP","port":%d}]`, otelCollectorMetricsPort, otelCollectorGRPCReceiverPort),
	}

	return items
}

// getTargetAllocatorServiceAccount returns the [corev1.ServiceAccount] for the
// Target Allocator.
func (a *Actuator) getTargetAllocatorServiceAccount(namespace string) *corev1.ServiceAccount {
	obj := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      targetAllocatorServiceAccountName,
			Namespace: namespace,
			Labels:    a.getCommonLabels(),
		},
		AutomountServiceAccountToken: new(false),
	}

	return obj
}

// getTargetAllocatorHTTPSService returns the [corev1.Service] for the
// HTTPS communication of the Target Allocator.
func (a *Actuator) getTargetAllocatorHTTPSService(namespace string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      targetAllocatorHTTPSServiceName,
			Namespace: namespace,
			Labels:    a.getCommonLabels(),
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{{
				Port:       443,
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt32(targetAllocatorHTTPSPort),
			}},
			Selector: map[string]string{
				labelKeyComponent: labelValueTargetAllocator,
			},
		},
	}
}

// getTargetAllocatorConfigMap returns the [corev1.ConfigMap] for the Target
// Allocator.
func (a *Actuator) getTargetAllocatorConfigMap(namespace string) (*corev1.ConfigMap, error) {
	taConfig := map[string]any{
		"allocation_strategy":              otelv1alpha1.OpenTelemetryTargetAllocatorAllocationStrategyConsistentHashing,
		"collector_not_ready_grace_period": 30 * time.Second,
		"collector_namespace":              namespace,
		"collector_selector": map[string]any{
			"matchLabels": map[string]any{
				labelKeyComponent:              "opentelemetry-collector",
				"app.kubernetes.io/instance":   fmt.Sprintf("%s.%s", namespace, baseResourceName),
				"app.kubernetes.io/managed-by": "opentelemetry-operator",
				"app.kubernetes.io/name":       fmt.Sprintf("%s-collector", baseResourceName),
				"app.kubernetes.io/part-of":    "opentelemetry",
			},
		},
		"filter_strategy": "relabel-config",
		"prometheus_cr": map[string]any{
			configKeyEnabled:         true,
			"allow_namespaces":       []string{namespace},
			"scrape_interval":        30 * time.Second,
			"scrape_config_selector": nil,
			"probe_selector":         nil,
			"pod_monitor_selector":   nil,
			"deny_namespaces":        nil,
			"service_monitor_selector": map[string]any{
				"matchLabels": map[string]any{
					configKeyPrometheus: labelValuePrometheusShoot,
				},
			},
		},
	}

	data, err := yaml.Marshal(taConfig)
	if err != nil {
		return nil, err
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      targetAllocatorConfigMapName,
			Namespace: namespace,
			Labels:    a.getCommonLabels(),
		},
		Data: map[string]string{
			"targetallocator.yaml": string(data),
		},
	}

	return configMap, nil
}

// getTargetAllocatorRole returns the [rbacv1.Role] for the Target Allocator.
func (a *Actuator) getTargetAllocatorRole(namespace string) *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      targetAllocatorRoleName,
			Namespace: namespace,
			Labels:    a.getCommonLabels(),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "services", "endpoints", "secrets", "namespaces"},
				Verbs:     readVerbs,
			},
			{
				APIGroups: []string{"discovery.k8s.io"},
				Resources: []string{"endpointslices"},
				Verbs:     readVerbs,
			},
			{
				APIGroups: []string{"monitoring.coreos.com"},
				Resources: []string{"servicemonitors", "podmonitors", "scrapeconfigs", "probes"},
				Verbs:     readVerbs,
			},
		},
	}
}

// getTargetAllocatorRoleBinding returns the [rbacv1.RoleBinding] for the Target
// Allocator.
func (a *Actuator) getTargetAllocatorRoleBinding(namespace string) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      targetAllocatorRoleName,
			Namespace: namespace,
			Labels:    a.getCommonLabels(),
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     targetAllocatorRoleName,
		},
		Subjects: []rbacv1.Subject{{
			Kind:      rbacv1.ServiceAccountKind,
			Name:      targetAllocatorServiceAccountName,
			Namespace: namespace,
		}},
	}
}

// getTargetAllocator returns the [appsv1.Deployment] resource for the Target
// Allocator.
//
// We are creating a deployment here, instead of using the upstream OTel
// TargetAllocator custom resource, because the OTel Operator expects that mTLS
// between the Target Allocator and the Collector is handled via Cert Manager
// only. However, Gardener does not use Cert Manager, so we can't configure mTLS
// easily.
//
// mTLS between the TA and the Collector is required, otherwise the TA will
// return invalid secrets for scrape targets which require authentication.
//
// Currently the mTLS between TA and Collector cannot be done in a generic way
// when using the OTel Operator, because upon start up the OTel Operator looks
// for Cert Manager. If it doesn't find Cert Manager, it will always configure
// the communication between the TA and Collector to happen via HTTP, which in
// turn results in invalid secrets being delivered to the Collector. As a result
// scraping will always fail.
//
// The following upstream issue tracks the progress of allowing clients to
// configure mTLS between TA and Collector without having to rely on Cert
// Manager.
//
// https://github.com/open-telemetry/opentelemetry-operator/issues/3982
//
// Once the issue above is fixed we can drop the following resources, which we
// are now explicitly managing, and instead use the TargetAllocator custom
// resource only.
//
// - Deployment for the TargetAllocator (getTargetAllocatorDeployment)
// - ConfigMap for the TargetAllocator (getTargetAllocatorConfigMap)
// - HTTPS Service for the Target Allocator (getTargetAllocatorHTTPSService)
func (a *Actuator) getTargetAllocatorDeployment(namespace string, caSecret, serverSecret *corev1.Secret, image *imagevectorutils.Image) *appsv1.Deployment {
	const (
		volumeNameCACertificate      = "ca-cert"
		volumeMountPathCACertificate = "/etc/ssl/certs/ca"

		volumeNameServerCertificate      = "server-cert"
		volumeMountPathServerCertificate = "/etc/ssl/certs/server"

		volumeNameTargetAllocatorConfig  = "targetallocator-config"
		volumeMountTargetAllocatorConfig = "/app/targetallocator"
	)

	allLabels := utils.MergeStringMaps(
		a.getCommonLabels(),
		a.getNetworkLabels(),
		map[string]string{
			labelKeyComponent: labelValueTargetAllocator,
		},
	)

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      targetAllocatorDeploymentName,
			Namespace: namespace,
			Labels:    a.getCommonLabels(),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas:             new(targetAllocatorReplicas),
			RevisionHistoryLimit: ptr.To[int32](2),
			Selector: &metav1.LabelSelector{
				MatchLabels: allLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: allLabels,
				},
				Spec: corev1.PodSpec{
					PriorityClassName:  v1beta1constants.PriorityClassNameShootControlPlane100,
					ServiceAccountName: targetAllocatorServiceAccountName,
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: new(true),
						RunAsUser:    ptr.To[int64](65532),
						RunAsGroup:   ptr.To[int64](65532),
						FSGroup:      ptr.To[int64](65532),
					},
					Containers: []corev1.Container{
						{
							Name:  "ta-container",
							Image: image.String(),
							Args: []string{
								"--enable-https-server=true",
								fmt.Sprintf("--config-file=%s/targetallocator.yaml", volumeMountTargetAllocatorConfig),
								fmt.Sprintf("--https-ca-file=%s/%s", volumeMountPathCACertificate, secretsutils.DataKeyCertificateBundle),
								fmt.Sprintf("--https-tls-cert-file=%s/%s", volumeMountPathServerCertificate, secretsutils.DataKeyCertificate),
								fmt.Sprintf("--https-tls-key-file=%s/%s", volumeMountPathServerCertificate, secretsutils.DataKeyPrivateKey),
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("10m"),
									corev1.ResourceMemory: resource.MustParse("50Mi"),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{Name: volumeNameCACertificate, MountPath: volumeMountPathCACertificate, ReadOnly: true},
								{Name: volumeNameServerCertificate, MountPath: volumeMountPathServerCertificate, ReadOnly: true},
								{Name: volumeNameTargetAllocatorConfig, MountPath: volumeMountTargetAllocatorConfig, ReadOnly: true},
							},
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: new(false),
							},
						},
					},
					Volumes: []corev1.Volume{
						{Name: volumeNameCACertificate, VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: caSecret.Name}}},
						{Name: volumeNameServerCertificate, VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: serverSecret.Name}}},
						{Name: volumeNameTargetAllocatorConfig, VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{LocalObjectReference: corev1.LocalObjectReference{Name: targetAllocatorConfigMapName}}}},
					},
				},
			},
		},
	}
}

// getOtelCollectorServiceAccount returns the [corev1.ServiceAccount] for the
// the OTel Collector.
func (a *Actuator) getOtelCollectorServiceAccount(namespace string) *corev1.ServiceAccount {
	obj := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      otelCollectorServiceAccountName,
			Namespace: namespace,
			Labels:    a.getCommonLabels(),
		},
		AutomountServiceAccountToken: new(false),
	}

	return obj
}

// getDebugExporterConfig returns the OTel settings for the debug exporter,
// derived from a debug-protocol [config.ExporterConfig].
func (a *Actuator) getDebugExporterConfig(cfg config.ExporterConfig) map[string]any {
	// See the link below for more details about each config setting for the
	// debug exporter.
	//
	// https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/debugexporter
	exporter := map[string]any{
		"verbosity": cfg.Verbosity,
	}

	return exporter
}

// filterAttributes converts the given filter attributes into the OTel
// filterprocessor `key`/`value` attribute maps.
func filterAttributes(attrs []config.FilterAttribute) []any {
	out := make([]any, 0, len(attrs))
	for _, attr := range attrs {
		entry := map[string]any{configKeyKey: attr.Key}
		if attr.Value != "" {
			entry[configKeyValue] = attr.Value
		}
		out = append(out, entry)
	}

	return out
}

// getMetricMatchProperties renders the OTel filterprocessor include/exclude
// match properties for the metrics signal.
func getMetricMatchProperties(props *config.MetricMatchProperties) map[string]any {
	match := map[string]any{
		configKeyMatchType: string(props.MatchType),
	}

	if len(props.MetricNames) > 0 {
		match["metric_names"] = props.MetricNames
	}
	if len(props.Expressions) > 0 {
		match["expressions"] = props.Expressions
	}
	if len(props.ResourceAttributes) > 0 {
		match["resource_attributes"] = filterAttributes(props.ResourceAttributes)
	}
	if props.Regexp != nil {
		regexpConfig := map[string]any{}
		if props.Regexp.CacheEnabled != nil {
			regexpConfig["cacheenabled"] = *props.Regexp.CacheEnabled
		}
		if props.Regexp.CacheMaxNumEntries != 0 {
			regexpConfig["cachemaxnumentries"] = props.Regexp.CacheMaxNumEntries
		}
		match["regexp"] = regexpConfig
	}

	return match
}

// getLogMatchProperties renders the OTel filterprocessor include/exclude match
// properties for the logs signal.
func getLogMatchProperties(props *config.LogMatchProperties) map[string]any {
	match := map[string]any{
		configKeyMatchType: string(props.MatchType),
	}

	if len(props.ResourceAttributes) > 0 {
		match["resource_attributes"] = filterAttributes(props.ResourceAttributes)
	}
	if len(props.RecordAttributes) > 0 {
		match["record_attributes"] = filterAttributes(props.RecordAttributes)
	}
	if len(props.SeverityTexts) > 0 {
		match["severity_texts"] = props.SeverityTexts
	}
	if len(props.Bodies) > 0 {
		match["bodies"] = props.Bodies
	}
	if props.SeverityNumber != nil {
		severityNumber := map[string]any{
			"min": props.SeverityNumber.Min,
		}
		if props.SeverityNumber.MatchUndefined != nil {
			severityNumber["match_undefined"] = *props.SeverityNumber.MatchUndefined
		}
		match["severity_number"] = severityNumber
	}

	return match
}

// renderContextConditions renders the filter processor context-inferred
// condition groups (metric_conditions / log_conditions).
//
// A group with no explicit context and no per-group error_mode is rendered as
// bare condition strings (basic style); the filter processor infers the OTTL
// context from each expression. A group that sets either is rendered as an
// explicit object (advanced style). The two styles must not be mixed within a
// single list, which the validation enforces.
func renderContextConditions(groups []config.ContextConditions) []any {
	out := make([]any, 0, len(groups))
	for _, group := range groups {
		if group.Context == "" && group.ErrorMode == "" {
			for _, cond := range group.Conditions {
				out = append(out, cond)
			}

			continue
		}

		entry := map[string]any{
			"conditions": group.Conditions,
		}
		if group.Context != "" {
			entry["context"] = group.Context
		}
		if group.ErrorMode != "" {
			entry["error_mode"] = string(group.ErrorMode)
		}
		out = append(out, entry)
	}

	return out
}

// getFilterProcessorConfig returns the OTel settings for a single filter
// processor instance built from one [config.FilterRule].
//
// See the link below for more details about each config setting of the filter
// processor.
//
// https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/filterprocessor
func (a *Actuator) getFilterProcessorConfig(rule config.FilterRule) map[string]any {
	processor := map[string]any{}

	if rule.ErrorMode != "" {
		processor["error_mode"] = string(rule.ErrorMode)
	}

	if m := rule.Metrics; m != nil {
		metrics := map[string]any{}
		if len(m.Resource) > 0 {
			metrics["resource"] = m.Resource
		}
		if len(m.Metric) > 0 {
			metrics["metric"] = m.Metric
		}
		if len(m.DataPoint) > 0 {
			metrics["datapoint"] = m.DataPoint
		}
		if m.Include != nil {
			metrics["include"] = getMetricMatchProperties(m.Include)
		}
		if m.Exclude != nil {
			metrics["exclude"] = getMetricMatchProperties(m.Exclude)
		}
		if len(metrics) > 0 {
			processor["metrics"] = metrics
		}
	}

	if l := rule.Logs; l != nil {
		logs := map[string]any{}
		if len(l.Resource) > 0 {
			logs["resource"] = l.Resource
		}
		if len(l.LogRecord) > 0 {
			logs["log_record"] = l.LogRecord
		}
		if l.Include != nil {
			logs["include"] = getLogMatchProperties(l.Include)
		}
		if l.Exclude != nil {
			logs["exclude"] = getLogMatchProperties(l.Exclude)
		}
		if len(logs) > 0 {
			processor["logs"] = logs
		}
	}

	if len(rule.MetricConditions) > 0 {
		processor["metric_conditions"] = renderContextConditions(rule.MetricConditions)
	}

	if len(rule.LogConditions) > 0 {
		processor["log_conditions"] = renderContextConditions(rule.LogConditions)
	}

	if len(rule.Conditions) > 0 {
		processor["conditions"] = renderContextConditions(rule.Conditions)
	}

	return processor
}

// getOTLPHTTPExporterConfig returns the OTel settings for the OTLP HTTP
// exporter of the given signal's i-th target.
func (a *Actuator) getOTLPHTTPExporterConfig(cfg config.ExporterConfig, sig config.SignalType, i int) map[string]any {
	exporter := map[string]any{}

	// See the link below for more details about each config setting of the
	// OTLP HTTP exporter.
	//
	// https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlphttpexporter
	if cfg.Endpoint != "" {
		exporter[configKeyEndpoint] = cfg.Endpoint
	}

	exporter["read_buffer_size"] = cfg.ReadBufferSize
	exporter["write_buffer_size"] = cfg.WriteBufferSize
	exporter["timeout"] = cfg.Timeout.String()
	exporter["compression"] = string(cfg.Compression)
	exporter["encoding"] = string(cfg.Encoding)

	// Retry on Failure settings
	if cfg.RetryOnFailure.Enabled != nil {
		exporter["retry_on_failure"] = map[string]any{
			configKeyEnabled:   *cfg.RetryOnFailure.Enabled,
			"initial_interval": cfg.RetryOnFailure.InitialInterval.String(),
			"max_interval":     cfg.RetryOnFailure.MaxInterval.String(),
			"max_elapsed_time": cfg.RetryOnFailure.MaxElapsedTime.String(),
			"multiplier":       cfg.RetryOnFailure.Multiplier,
		}
	}

	// TLS settings
	if tls := cfg.TLS; tls != nil {
		exporter["tls"] = a.getExporterTLSConfig(tls, signalVolumeMountPathTLS(sig, i))
	}

	// Bearer Token Authentication settings
	if cfg.Token != nil {
		exporter["auth"] = map[string]any{
			"authenticator": signalBearerTokenAuthName(sig, i),
		}
	}

	return exporter
}

// getExporterTLSConfig renders the TLS block for an exporter, resolving the
// CA/cert/key file paths against the given per-signal mount path.
func (a *Actuator) getExporterTLSConfig(tls *config.TLSConfig, mountPath string) map[string]any {
	tlsConfig := map[string]any{}
	if tls.InsecureSkipVerify != nil {
		tlsConfig["insecure_skip_verify"] = *tls.InsecureSkipVerify
	}
	if tls.CA != nil {
		tlsConfig["ca_file"] = filepath.Join(mountPath, tls.CA.ResourceRef.DataKey)
	}
	if tls.Cert != nil {
		tlsConfig["cert_file"] = filepath.Join(mountPath, tls.Cert.ResourceRef.DataKey)
	}
	if tls.Key != nil {
		tlsConfig["key_file"] = filepath.Join(mountPath, tls.Key.ResourceRef.DataKey)
	}

	tlsConfig["reload_interval"] = tls.ReloadInterval.String()

	return tlsConfig
}

// getOTLPGRPCExporterConfig returns the OTel settings for the OTLP gRPC
// exporter of the given signal's i-th target.
func (a *Actuator) getOTLPGRPCExporterConfig(cfg config.ExporterConfig, sig config.SignalType, i int) map[string]any {
	// See the link below for more details about each config setting of the
	// OTLP gRPC exporter.
	//
	// https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlpexporter
	exporter := map[string]any{
		configKeyEndpoint:   cfg.Endpoint,
		"read_buffer_size":  cfg.ReadBufferSize,
		"write_buffer_size": cfg.WriteBufferSize,
		"timeout":           cfg.Timeout.String(),
		"compression":       string(cfg.Compression),
	}

	// Retry on Failure settings
	if cfg.RetryOnFailure.Enabled != nil {
		exporter["retry_on_failure"] = map[string]any{
			configKeyEnabled:   *cfg.RetryOnFailure.Enabled,
			"initial_interval": cfg.RetryOnFailure.InitialInterval.String(),
			"max_interval":     cfg.RetryOnFailure.MaxInterval.String(),
			"max_elapsed_time": cfg.RetryOnFailure.MaxElapsedTime.String(),
			"multiplier":       cfg.RetryOnFailure.Multiplier,
		}
	}

	// TLS settings
	if tls := cfg.TLS; tls != nil {
		exporter["tls"] = a.getExporterTLSConfig(tls, signalVolumeMountPathTLS(sig, i))
	}

	// Bearer Token Authentication settings
	if cfg.Token != nil {
		exporter["auth"] = map[string]any{
			"authenticator": signalBearerTokenAuthName(sig, i),
		}
	}

	return exporter
}

// getSignalExporterConfig renders the exporter config for a signal's i-th
// target based on its effective (merged) protocol.
func (a *Actuator) getSignalExporterConfig(cfg config.ExporterConfig, sig config.SignalType, i int) map[string]any {
	switch cfg.Protocol {
	case config.ExporterProtocolGRPC:
		return a.getOTLPGRPCExporterConfig(cfg, sig, i)
	case config.ExporterProtocolDebug:
		return a.getDebugExporterConfig(cfg)
	default:
		return a.getOTLPHTTPExporterConfig(cfg, sig, i)
	}
}

// getOtelExporters returns the OpenTelemetry exporters based on the given
// [config.CollectorConfig] spec, along with the exporter component names
// grouped per signal so pipelines can reference them. For each signal the
// returned slice is indexed by target, i.e. element i is the exporter name for
// that signal's i-th target.
func (a *Actuator) getOtelExporters(cfg config.CollectorConfig) (map[string]any, map[config.SignalType][]string) {
	exporters := make(map[string]any)
	perSignalTarget := make(map[config.SignalType][]string)

	for _, sig := range config.AllSignals() {
		signal := cfg.Spec.Signals.Signal(sig)
		if !signal.IsEnabled() {
			continue
		}

		names := make([]string, len(signal.Targets))
		for i, target := range signal.Targets {
			name := signalExporterName(sig, i, target.Exporter.Protocol)
			exporters[name] = a.getSignalExporterConfig(target.Exporter, sig, i)
			names[i] = name
		}
		perSignalTarget[sig] = names
	}

	return exporters, perSignalTarget
}

// parseShootNamespaceAttributes extracts OTel resource attributes from a shoot
// namespace name of the form "shoot--<project>--<shoot>".
// The full namespace name maps to k8s.cluster.name; the two segments map to
// gardener.project.name and gardener.shoot.name respectively.
// For namespaces that do not follow the pattern, projectName and shootName are
// returned as empty strings.
func parseShootNamespaceAttributes(namespace string) (clusterName, projectName, shootName string) {
	clusterName = namespace
	parts := strings.SplitN(namespace, "--", 3)
	if len(parts) == 3 {
		projectName = parts[1]
		shootName = parts[2]
	}

	return clusterName, projectName, shootName
}

// buildPipelines returns the collector service pipelines for the signals
// enabled by cfg. Each enabled signal produces one pipeline per target, wired
// to that target's own exporter and filter processor instances.
//
// The processor chain is: resource, memory_limiter, (transform/events for the
// events signal), the target's own filter processors, and finally batch. The
// filter processors are inserted after memory_limiter and before batch so
// unwanted telemetry is dropped before batching.
func buildPipelines(cfg config.CollectorConfig, perSignalTargetExporterNames map[config.SignalType][]string) map[string]*otelv1beta1.Pipeline {
	pipelines := map[string]*otelv1beta1.Pipeline{}

	for _, sig := range config.AllSignals() {
		signal := cfg.Spec.Signals.Signal(sig)
		if !signal.IsEnabled() {
			continue
		}

		for i, target := range signal.Targets {
			processors := []string{resourceProcessorName, memoryLimiterProcessorName}
			if sig == config.SignalEvents {
				processors = append(processors, transformEventsProcessorName)
			}
			for ruleIdx := range target.Filters {
				processors = append(processors, signalFilterName(sig, i, ruleIdx))
			}
			processors = append(processors, batchProcessorName)

			var exporters []string
			if names := perSignalTargetExporterNames[sig]; i < len(names) {
				exporters = []string{names[i]}
			}

			pipelines[signalPipelineName(sig, i)] = &otelv1beta1.Pipeline{
				Receivers:  []string{signalReceiverName(sig)},
				Processors: processors,
				Exporters:  exporters,
			}
		}
	}

	return pipelines
}

// getOTelCollector returns the [otelv1beta1.OpenTelemetryCollector]
// resource, which the extension manages.
func (a *Actuator) getOtelCollector(
	namespace string,
	caSecret, clientSecret *corev1.Secret,
	cfg config.CollectorConfig,
	resources []gardencorev1beta1.NamedResourceReference,
	shootKubeconfigSecretName string,
	accessSecretName string,
	image *imagevectorutils.Image,
) *otelv1beta1.OpenTelemetryCollector {
	const (
		volumeNameCACertificate      = "ca-cert"
		volumeMountPathCACertificate = "/etc/ssl/certs/ca"

		volumeNameClientCertificate      = "client-cert"
		volumeMountPathClientCertificate = "/etc/ssl/certs/client"
	)

	exporters, perSignalExporterNames := a.getOtelExporters(cfg)
	clusterName, projectName, shootName := parseShootNamespaceAttributes(namespace)
	allLabels := utils.MergeStringMaps(
		a.getCommonLabels(),
		a.getNetworkLabels(),
	)

	obj := &otelv1beta1.OpenTelemetryCollector{
		ObjectMeta: metav1.ObjectMeta{
			Name:      otelCollectorName,
			Namespace: namespace,
			Labels:    allLabels,
			Annotations: utils.MergeStringMaps(
				a.getAnnotations(),
				map[string]string{
					resourcesv1alpha1.NetworkPolicyLabelKeyPrefix + "pod-label-selector-namespace-alias": "all-shoots",
					resourcesv1alpha1.NetworkPolicyLabelKeyPrefix + "namespace-selectors":                `[{"matchExpressions":[{"key":"kubernetes.io/metadata.name","operator":"In","values":["garden"]}]},{"matchExpressions":[{"key":"gardener.cloud/role","operator":"In","values":["extension"]}]}]`,
				}),
		},
		Spec: otelv1beta1.OpenTelemetryCollectorSpec{
			// Note that the Target Allocator expects either a
			// statefulset or a daemonset deployment mode, because
			// it provides load-balancing of scrape targets between
			// multiple OTel Collectors. In order to achieve this,
			// the respective OTel collectors must have
			// deterministic and stable IDs, hence the requirement
			// for running in statefulset mode.
			//
			// https://github.com/open-telemetry/opentelemetry-operator/tree/main/cmd/otel-allocator
			Mode:            otelv1beta1.ModeStatefulSet,
			UpgradeStrategy: otelv1beta1.UpgradeStrategyNone,
			OpenTelemetryCommonFields: otelv1beta1.OpenTelemetryCommonFields{
				Image:    image.String(),
				Replicas: new(otelCollectorReplicas),
				VolumeMounts: []corev1.VolumeMount{
					{Name: volumeNameCACertificate, MountPath: volumeMountPathCACertificate, ReadOnly: true},
					{Name: volumeNameClientCertificate, MountPath: volumeMountPathClientCertificate, ReadOnly: true},
					{Name: volumeNameShootKubeconfig, MountPath: gardenerutils.VolumeMountPathGenericKubeconfig, ReadOnly: true},
				},
				Volumes: []corev1.Volume{
					{Name: volumeNameCACertificate, VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: caSecret.Name}}},
					{Name: volumeNameClientCertificate, VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: clientSecret.Name}}},
					gardenerutils.GenerateGenericKubeconfigVolume(shootKubeconfigSecretName, accessSecretName, volumeNameShootKubeconfig),
				},
				Env: []corev1.EnvVar{{
					Name:  "KUBECONFIG",
					Value: gardenerutils.PathGenericKubeconfig,
				}},
				PriorityClassName: v1beta1constants.PriorityClassNameShootControlPlane100,
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("10m"),
						corev1.ResourceMemory: resource.MustParse("50Mi"),
					},
				},
				SecurityContext: &corev1.SecurityContext{
					AllowPrivilegeEscalation: new(false),
				},
				ServiceAccount: otelCollectorServiceAccountName,
			},
			// Explicitly configure the Prometheus receiver to point
			// at an existing Target Allocator.
			Config: otelv1beta1.Config{
				Receivers: otelv1beta1.AnyConfig{
					Object: map[string]any{
						otlpReceiverName: map[string]any{
							"protocols": map[string]any{
								"grpc": map[string]any{
									configKeyEndpoint: fmt.Sprintf("0.0.0.0:%d", otelCollectorGRPCReceiverPort),
								},
							},
						},
						configKeyPrometheus: map[string]any{
							"target_allocator": map[string]any{
								"collector_id":    "${POD_NAME}",
								configKeyEndpoint: "https://" + targetAllocatorHTTPSServiceName,
								"interval":        "30s",
								"tls": map[string]any{
									"ca_file":   filepath.Join(volumeMountPathCACertificate, secretsutils.DataKeyCertificateBundle),
									"cert_file": filepath.Join(volumeMountPathClientCertificate, secretsutils.DataKeyCertificate),
									"key_file":  filepath.Join(volumeMountPathClientCertificate, secretsutils.DataKeyPrivateKey),
								},
							},
							"config": map[string]any{
								"scrape_configs": []any{
									map[string]any{
										"job_name":        otelCollectorName,
										"scrape_interval": "15s",
									},
								},
							},
						},
						eventsReceiverName: map[string]any{
							"auth_type": "kubeConfig",
							"objects": []any{
								map[string]any{
									"name":  "events",
									"group": "events.k8s.io",
									"mode":  "watch",
								},
							},
						},
					},
				},
				Processors: &otelv1beta1.AnyConfig{
					Object: map[string]any{
						batchProcessorName: map[string]any{
							"timeout":             a.batchProcessorConfig.Timeout.String(),
							"send_batch_size":     a.batchProcessorConfig.SendBatchSize,
							"send_batch_max_size": a.batchProcessorConfig.SendBatchMaxSize,
						},
						memoryLimiterProcessorName: map[string]any{
							"check_interval":         a.memoryLimiterConfig.CheckInterval.String(),
							"limit_mib":              a.memoryLimiterConfig.MemoryLimitMiB,
							"spike_limit_mib":        a.memoryLimiterConfig.MemorySpikeLimitMiB,
							"limit_percentage":       a.memoryLimiterConfig.MemoryLimitPercentage,
							"spike_limit_percentage": a.memoryLimiterConfig.MemorySpikePercentage,
						},
						resourceProcessorName: map[string]any{
							"attributes": []any{
								upsertAttribute("k8s.cluster.name", clusterName),
								upsertAttribute("gardener.project.name", projectName),
								upsertAttribute("gardener.shoot.name", shootName),
							},
						},
						transformEventsProcessorName: map[string]any{
							"log_statements": []any{
								map[string]any{
									"context": "log",
									"statements": []any{
										`delete_key(body["object"]["metadata"], "managedFields")`,
									},
								},
							},
						},
					},
				},
				Exporters: otelv1beta1.AnyConfig{
					Object: exporters,
				},
				Service: otelv1beta1.Service{
					Telemetry: &otelv1beta1.AnyConfig{
						Object: map[string]any{
							telemetryMetricsKey: map[string]any{
								"level": string(cfg.Spec.Metrics.Level),
								"readers": []any{
									map[string]any{
										"pull": map[string]any{
											"exporter": map[string]any{
												configKeyPrometheus: map[string]any{
													"host": "0.0.0.0",
													"port": otelCollectorMetricsPort,
												},
											},
										},
									},
								},
							},
							telemetryLogsKey: map[string]any{
								"level":    string(cfg.Spec.Logs.Level),
								"encoding": string(cfg.Spec.Logs.Encoding),
							},
						},
					},
					Pipelines: buildPipelines(cfg, perSignalExporterNames),
				},
			},
		},
	}

	// Register the per-target filter processors and configure the per-target
	// exporter TLS and bearer token authentication volumes. Only enabled
	// signals are processed so the collector does not report unused components.
	for _, sig := range config.AllSignals() {
		signal := cfg.Spec.Signals.Signal(sig)
		if !signal.IsEnabled() {
			continue
		}

		for i, target := range signal.Targets {
			for ruleIdx, rule := range target.Filters {
				obj.Spec.Config.Processors.Object[signalFilterName(sig, i, ruleIdx)] = a.getFilterProcessorConfig(rule)
			}

			// A debug target writes to the collector's own logs; it has no
			// endpoint, TLS or token, so there are no volumes to configure.
			if target.Exporter.Protocol == config.ExporterProtocolDebug {
				continue
			}

			// Exporter TLS settings
			a.configureVolumeForTLS(
				obj,
				target.Exporter.TLS,
				signalVolumeNameTLS(sig, i),
				signalVolumeMountPathTLS(sig, i),
				resources,
			)

			// Exporter Bearer Token Authentication settings
			//
			// https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/bearertokenauthextension
			a.configureVolumeForBearerTokenAuthExtension(
				obj,
				target.Exporter.Token,
				signalBearerTokenAuthName(sig, i),
				signalVolumeMountPathBearerToken(sig, i),
				signalVolumeNameBearerToken(sig, i),
				signalVolumeMountPathBearerToken(sig, i),
				resources,
			)
		}
	}

	return obj
}

// getEventsClusterRole returns the [rbacv1.ClusterRole] granting the OTel
// Collector's service account in the shoot cluster permission to list and watch
// events from the events.k8s.io API group.
func (a *Actuator) getEventsClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: otelCollectorName,
		},
		Rules: []rbacv1.PolicyRule{{
			APIGroups: []string{"events.k8s.io"},
			Resources: []string{"events"},
			Verbs:     readVerbs,
		}},
	}
}

// getEventsClusterRoleBinding returns the [rbacv1.ClusterRoleBinding] that
// binds the events ClusterRole to the OTel Collector's service account in the
// shoot cluster's kube-system namespace.
func (a *Actuator) getEventsClusterRoleBinding(serviceAccountName string) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: otelCollectorName,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     otelCollectorName,
		},
		Subjects: []rbacv1.Subject{{
			Kind:      rbacv1.ServiceAccountKind,
			Name:      serviceAccountName,
			Namespace: metav1.NamespaceSystem,
		}},
	}
}

func secretNameForResource(resourceName string, resources []gardencorev1beta1.NamedResourceReference) string {
	for _, r := range resources {
		if r.Name == resourceName &&
			r.ResourceRef.APIVersion == corev1.SchemeGroupVersion.String() && r.ResourceRef.Kind == "Secret" {
			return v1beta1constants.ReferencedResourcesPrefix + r.ResourceRef.Name
		}
	}

	return ""
}

// configureVolumeForTLS configures a volume for the OpenTelemetry collector for
// TLS secrets.
func (a *Actuator) configureVolumeForTLS(
	obj *otelv1beta1.OpenTelemetryCollector,
	tls *config.TLSConfig,
	volumeName string,
	volumeMount string,
	resources []gardencorev1beta1.NamedResourceReference,
) {
	if obj == nil || tls == nil {
		return
	}

	volume := corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			Projected: &corev1.ProjectedVolumeSource{},
		},
	}

	addSecretToProjectedVolume := func(resourceRef config.ResourceReferenceDetails) {
		volume.Projected.Sources = append(
			volume.Projected.Sources,
			corev1.VolumeProjection{
				Secret: &corev1.SecretProjection{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretNameForResource(resourceRef.Name, resources),
					},
					Items: []corev1.KeyToPath{{Key: resourceRef.DataKey, Path: resourceRef.DataKey}},
				},
			},
		)
	}

	if tls.CA != nil {
		addSecretToProjectedVolume(tls.CA.ResourceRef)
	}
	if tls.Cert != nil {
		addSecretToProjectedVolume(tls.Cert.ResourceRef)
	}
	if tls.Key != nil {
		addSecretToProjectedVolume(tls.Key.ResourceRef)
	}

	obj.Spec.Volumes = append(obj.Spec.Volumes, volume)
	obj.Spec.VolumeMounts = append(
		obj.Spec.VolumeMounts,
		corev1.VolumeMount{
			Name:      volumeName,
			MountPath: volumeMount,
		},
	)
}

// configureVolumeForBearerTokenAuthExtension configures a volume for the
// OpenTelemetry collector for the bearertokenauth extension.
func (a *Actuator) configureVolumeForBearerTokenAuthExtension(
	obj *otelv1beta1.OpenTelemetryCollector,
	ref *config.ResourceReference,
	authExtensionName string,
	tokenBasePath string,
	volumeName string,
	volumeMount string,
	resources []gardencorev1beta1.NamedResourceReference,
) {
	if obj == nil || ref == nil {
		return
	}

	if obj.Spec.Config.Extensions == nil {
		obj.Spec.Config.Extensions = &otelv1beta1.AnyConfig{}
	}

	if obj.Spec.Config.Extensions.Object == nil {
		obj.Spec.Config.Extensions.Object = make(map[string]any)
	}

	obj.Spec.Config.Extensions.Object[authExtensionName] = map[string]any{
		"filename": filepath.Join(tokenBasePath, ref.ResourceRef.DataKey),
	}

	obj.Spec.Config.Service.Extensions = append(obj.Spec.Config.Service.Extensions, authExtensionName)

	obj.Spec.Volumes = append(
		obj.Spec.Volumes,
		corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: secretNameForResource(ref.ResourceRef.Name, resources),
				},
			},
		},
	)

	obj.Spec.VolumeMounts = append(
		obj.Spec.VolumeMounts,
		corev1.VolumeMount{
			Name:      volumeName,
			MountPath: volumeMount,
		},
	)
}
