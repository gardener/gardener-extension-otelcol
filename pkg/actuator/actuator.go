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
	"maps"
	"slices"
	"strconv"
	"time"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/extension"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	v1beta1helper "github.com/gardener/gardener/pkg/apis/core/v1beta1/helper"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	resourcesv1alpha1 "github.com/gardener/gardener/pkg/apis/resources/v1alpha1"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	gardenerfeatures "github.com/gardener/gardener/pkg/features"
	"github.com/gardener/gardener/pkg/utils"
	kubernetesutils "github.com/gardener/gardener/pkg/utils/kubernetes"
	"github.com/gardener/gardener/pkg/utils/managedresources"
	secretsutils "github.com/gardener/gardener/pkg/utils/secrets"
	secretsmanager "github.com/gardener/gardener/pkg/utils/secrets/manager"
	"github.com/go-logr/logr"
	otelv1alpha1 "github.com/open-telemetry/opentelemetry-operator/apis/v1alpha1"
	otelv1beta1 "github.com/open-telemetry/opentelemetry-operator/apis/v1beta1"
	"go.yaml.in/yaml/v4"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/component-base/featuregate"
	"k8s.io/utils/clock"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/gardener-extension-otelcol/pkg/apis/config"
	"github.com/gardener/gardener-extension-otelcol/pkg/apis/config/validation"
)

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
)

// Actuator is an implementation of [extension.Actuator].
type Actuator struct {
	reader  client.Reader
	client  client.Client
	decoder runtime.Decoder

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
func New(opts ...Option) (*Actuator, error) {
	act := &Actuator{
		gardenletFeatureGates: make(map[featuregate.Feature]bool),
	}

	for _, opt := range opts {
		if err := opt(act); err != nil {
			return nil, err
		}
	}

	return act, nil
}

// WithClient is an [Option], which configures the [Actuator] with the given
// [client.Client].
func WithClient(c client.Client) Option {
	opt := func(a *Actuator) error {
		a.client = c

		return nil
	}

	return opt
}

// WithReader is an [Option], which configures the [Actuator] with the given
// [client.Reader].
func WithReader(r client.Reader) Option {
	opt := func(a *Actuator) error {
		a.reader = r

		return nil
	}

	return opt
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

	data, err := registry.AddAllAndSerialize(
		taConfigMap,
		a.getTargetAllocatorServiceAccount(ex.Namespace),
		a.getTargetAllocatorRole(ex.Namespace),
		a.getTargetAllocatorRoleBinding(ex.Namespace),
		a.getTargetAllocatorHTTPSService(ex.Namespace),
		a.getTargetAllocatorDeployment(ex.Namespace, caBundleSecret, serverSecret),
		a.getOtelCollectorServiceAccount(ex.Namespace),
		a.getOtelCollector(ex.Namespace, caBundleSecret, clientSecret, cfg),
	)
	if err != nil {
		return err
	}

	if err := managedresources.CreateForSeed(
		ctx,
		a.client,
		ex.Namespace,
		managedResourceName,
		false,
		data,
	); err != nil {
		return err
	}

	return nil
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

// Migrate signals the [Actuator] to reconcile the resources managed by it,
// because of a shoot control-plane migration event. This method implements the
// [extension.Actuator] interface.
func (a *Actuator) Migrate(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	return a.Reconcile(ctx, logger, ex)
}

func (a *Actuator) newSecretsManager(ctx context.Context, log logr.Logger, namespace string) (secretsmanager.Interface, error) {
	return secretsmanager.New(
		ctx,
		log,
		clock.RealClock{},
		a.client,
		namespace,
		secretsManagerIdentity,
		secretsmanager.Config{CASecretAutoRotation: true},
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
		fromAllScrapeTargetsAnnotation: fmt.Sprintf(`[{"protocol":"TCP","port":%d}]`, otelCollectorMetricsPort),
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
		AutomountServiceAccountToken: ptr.To(false),
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
				"app.kubernetes.io/component": "opentelemetry-targetallocator",
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
				"app.kubernetes.io/component":  "opentelemetry-collector",
				"app.kubernetes.io/instance":   fmt.Sprintf("%s.%s", namespace, baseResourceName),
				"app.kubernetes.io/managed-by": "opentelemetry-operator",
				"app.kubernetes.io/name":       fmt.Sprintf("%s-collector", baseResourceName),
				"app.kubernetes.io/part-of":    "opentelemetry",
			},
		},
		"filter_strategy": "relabel-config",
		"prometheus_cr": map[string]any{
			"enabled":                true,
			"allow_namespaces":       []string{namespace},
			"scrape_interval":        30 * time.Second,
			"scrape_config_selector": nil,
			"probe_selector":         nil,
			"pod_monitor_selector":   nil,
			"deny_namespaces":        nil,
			"service_monitor_selector": map[string]any{
				"matchLabels": map[string]any{
					// TODO(dnaeon): additional labels
					"prometheus": "shoot",
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
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"discovery.k8s.io"},
				Resources: []string{"endpointslices"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"monitoring.coreos.com"},
				Resources: []string{"servicemonitors", "podmonitors", "scrapeconfigs", "probes"},
				Verbs:     []string{"get", "list", "watch"},
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

// getTargetAllocator returns the [appsv1.Deployment] resource.
func (a *Actuator) getTargetAllocatorDeployment(namespace string, caSecret, serverSecret *corev1.Secret) *appsv1.Deployment {
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
			"app.kubernetes.io/component": "opentelemetry-targetallocator",
		},
	)

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      targetAllocatorDeploymentName,
			Namespace: namespace,
			Labels:    a.getCommonLabels(),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas:             ptr.To(targetAllocatorReplicas),
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
						RunAsNonRoot: ptr.To(true),
						RunAsUser:    ptr.To[int64](65532),
						RunAsGroup:   ptr.To[int64](65532),
						FSGroup:      ptr.To[int64](65532),
					},
					Containers: []corev1.Container{
						{
							Name:  "ta-container",
							Image: "otel/target-allocator:v0.140.0", // TODO(dnaeon): this image should be configurable and vendored
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
								AllowPrivilegeEscalation: ptr.To(false),
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
		AutomountServiceAccountToken: ptr.To(false),
	}

	return obj
}

// getOtelExporters returns the OpenTelemetry exporters based on the given
// [config.CollectorConfig] spec.
func (a *Actuator) getOtelExporters(cfg config.CollectorConfig) map[string]any {
	// TODO(dnaeon): authentication, tls, etc.
	httpExporter := map[string]any{
		"endpoint": cfg.Spec.Exporters.OTLPHTTPExporter.Endpoint,
	}

	// TODO(dnaeon): debug exporter should be configurable via the shoot
	// provider config
	exporters := map[string]any{
		"debug": map[string]any{
			"verbosity": "basic", // basic, normal or detailed levels
		},
	}

	if cfg.Spec.Exporters.OTLPHTTPExporter.Endpoint != "" {
		exporters["otlphttp"] = httpExporter
	}

	// TODO(dnaeon): add OTLP gRPC exporter

	return exporters
}

// getOTelCollector returns the [otelv1beta1.OpenTelemetryCollector]
// resource, which the extension manages.
func (a *Actuator) getOtelCollector(namespace string, caSecret, clientSecret *corev1.Secret, cfg config.CollectorConfig) *otelv1beta1.OpenTelemetryCollector {
	const (
		volumeNameCACertificate      = "ca-cert"
		volumeMountPathCACertificate = "/etc/ssl/certs/ca"

		volumeNameClientCertificate      = "client-cert"
		volumeMountPathClientCertificate = "/etc/ssl/certs/client"
	)

	exporters := a.getOtelExporters(cfg)
	exporterNames := slices.Sorted(maps.Keys(exporters))

	allLabels := utils.MergeStringMaps(a.getCommonLabels(), a.getNetworkLabels())

	return &otelv1beta1.OpenTelemetryCollector{
		ObjectMeta: metav1.ObjectMeta{
			Name:        otelCollectorName,
			Namespace:   namespace,
			Labels:      allLabels,
			Annotations: a.getAnnotations(),
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
				Image:    "otel/opentelemetry-collector:0.141.0", // TODO(dnaeon): this image should be configurable
				Replicas: ptr.To(otelCollectorReplicas),
				VolumeMounts: []corev1.VolumeMount{
					{Name: volumeNameCACertificate, MountPath: volumeMountPathCACertificate, ReadOnly: true},
					{Name: volumeNameClientCertificate, MountPath: volumeMountPathClientCertificate, ReadOnly: true},
				},
				Volumes: []corev1.Volume{
					{Name: volumeNameCACertificate, VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: caSecret.Name}}},
					{Name: volumeNameClientCertificate, VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: clientSecret.Name}}},
				},
				PriorityClassName: v1beta1constants.PriorityClassNameShootControlPlane100,
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("10m"),
						corev1.ResourceMemory: resource.MustParse("50Mi"),
					},
				},
				SecurityContext: &corev1.SecurityContext{
					AllowPrivilegeEscalation: ptr.To(false),
				},
				ServiceAccount: otelCollectorServiceAccountName,
			},
			// Explicitly configure the Prometheus receiver to point
			// at an existing Target Allocator.
			Config: otelv1beta1.Config{
				Receivers: otelv1beta1.AnyConfig{
					Object: map[string]any{
						// TODO(dnaeon): enable OTLP gRPC receiver for logs
						"prometheus": map[string]any{
							"target_allocator": map[string]any{
								"collector_id": "${POD_NAME}",
								"endpoint":     "https://" + targetAllocatorHTTPSServiceName,
								"interval":     "30s",
								"tls": map[string]any{
									"ca_file":   volumeMountPathCACertificate + "/" + secretsutils.DataKeyCertificateBundle,
									"cert_file": volumeMountPathClientCertificate + "/" + secretsutils.DataKeyCertificate,
									"key_file":  volumeMountPathClientCertificate + "/" + secretsutils.DataKeyPrivateKey,
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
					},
				},
				Processors: &otelv1beta1.AnyConfig{
					Object: map[string]any{
						"batch": map[string]any{
							"timeout": "15s",
						},
					},
				},
				Exporters: otelv1beta1.AnyConfig{
					Object: exporters,
				},
				Service: otelv1beta1.Service{
					Telemetry: &otelv1beta1.AnyConfig{
						Object: map[string]any{
							"metrics": map[string]any{
								"level": "detailed", // none, basic, normal and detailed levels
								"readers": []any{
									map[string]any{
										"pull": map[string]any{
											"exporter": map[string]any{
												"prometheus": map[string]any{
													"host": "0.0.0.0",
													"port": otelCollectorMetricsPort,
												},
											},
										},
									},
								},
							},
							"logs": map[string]any{
								"level":    "DEBUG", // INFO, WARN, DEBUG and ERROR levels
								"encoding": "json",
							},
						},
					},
					Pipelines: map[string]*otelv1beta1.Pipeline{
						// TODO(dnaeon): add a pipeline for logs, once we have them enabled
						"metrics": {
							Receivers: []string{"prometheus"},
							// TODO: enable debug
							Exporters:  exporterNames,
							Processors: []string{"batch"},
						},
					},
				},
			},
		},
	}
}
