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

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/extension"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	v1beta1helper "github.com/gardener/gardener/pkg/apis/core/v1beta1/helper"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	resourcesv1alpha1 "github.com/gardener/gardener/pkg/apis/resources/v1alpha1"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	gardenerfeatures "github.com/gardener/gardener/pkg/features"
	gardenerutils "github.com/gardener/gardener/pkg/utils/gardener"
	"github.com/gardener/gardener/pkg/utils/managedresources"
	"github.com/go-logr/logr"
	otelv1alpha1 "github.com/open-telemetry/opentelemetry-operator/apis/v1alpha1"
	otelv1beta1 "github.com/open-telemetry/opentelemetry-operator/apis/v1beta1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/component-base/featuregate"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/gardener-extension-otelcol/pkg/apis/config"
	"github.com/gardener/gardener-extension-otelcol/pkg/apis/config/validation"
	"github.com/gardener/gardener-extension-otelcol/pkg/metrics"
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

	// targetAllocatorName is the name of the [otelv1alpha1.TargetAllocator]
	// resource created by the extension.
	targetAllocatorName = baseResourceName
	// targetAllocatorServiceName is the name of the Kubernetes service for
	// the Target Allocator.
	targetAllocatorServiceName = baseResourceName + "-targetallocator"
	// targetAllocatorServicePort is the port on which the Target Allocator
	// service listens to.
	targetAllocatorServicePort = 80
	// targetAllocatorServiceAccountName is the name of the service account
	// for the Target Allocator.
	targetAllocatorServiceAccountName = baseResourceName + "-targetallocator"
	// targetAllocatorReplicas specifies the number of replicas of the Target Allocator.
	targetAllocatorReplicas int32 = 1
	// targetAllocatorRoleName is the name of the Role and RoleBinding
	// resource for the Target Allocator.
	targetAllocatorRoleName = baseResourceName + "-targetallocator"
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

	// Increment our example metrics counter
	defer func() {
		metrics.ActuatorOperationTotal.WithLabelValues(clusterName, "reconcile").Inc()
	}()

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

	// Bundle things up in a managed resource
	registry := managedresources.NewRegistry(
		kubernetes.SeedScheme,
		kubernetes.SeedCodec,
		kubernetes.SeedSerializer,
	)
	data, err := registry.AddAllAndSerialize(
		a.getTargetAllocatorServiceAccount(ex.Namespace),
		a.getTargetAllocatorRole(ex.Namespace),
		a.getTargetAllocatorRoleBinding(ex.Namespace),
		a.getTargetAllocator(ex.Namespace),
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
	// Increment our example metrics counter
	defer func() {
		metrics.ActuatorOperationTotal.WithLabelValues(ex.Namespace, "delete").Inc()
	}()

	logger.Info("deleting resources managed by extension")

	return client.IgnoreNotFound(managedresources.DeleteForSeed(ctx, a.client, ex.Namespace, managedResourceName))
}

// ForceDelete signals the [Actuator] to delete any resources managed by it,
// because of a force-delete event of the shoot cluster. This method implements
// the [extension.Actuator] interface.
func (a *Actuator) ForceDelete(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	// Increment our example metrics counter
	defer func() {
		metrics.ActuatorOperationTotal.WithLabelValues(ex.Namespace, "force_delete").Inc()
	}()

	logger.Info("shoot has been force-deleted, deleting resources managed by extension")

	return a.Delete(ctx, logger, ex)
}

// Restore restores the resources managed by the extension [Actuator]. This
// method implements the [extension.Actuator] interface.
func (a *Actuator) Restore(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	// Increment our example metrics counter
	defer func() {
		metrics.ActuatorOperationTotal.WithLabelValues(ex.Namespace, "restore").Inc()
	}()

	return a.Reconcile(ctx, logger, ex)
}

// Migrate signals the [Actuator] to reconcile the resources managed by it,
// because of a shoot control-plane migration event. This method implements the
// [extension.Actuator] interface.
func (a *Actuator) Migrate(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	// Increment our example metrics counter
	defer func() {
		metrics.ActuatorOperationTotal.WithLabelValues(ex.Namespace, "migrate").Inc()
	}()

	return a.Reconcile(ctx, logger, ex)
}

// getLabels returns the common set of labels for the Collector and Target
// Allocator resources.
func (a *Actuator) getLabels() map[string]string {
	// The `networking.resources.gardener.cloud/to-all-scrape-targets' label
	toAllScrapeTargetsLabel := resourcesv1alpha1.NetworkPolicyLabelKeyPrefix + "to" + v1beta1constants.LabelNetworkPolicyScrapeTargets

	items := map[string]string{
		v1beta1constants.LabelRole:                                                               v1beta1constants.LabelObservability,
		v1beta1constants.GardenRole:                                                              v1beta1constants.GardenRoleObservability,
		v1beta1constants.LabelObservabilityApplication:                                           otelCollectorName,
		v1beta1constants.LabelNetworkPolicyToDNS:                                                 v1beta1constants.LabelNetworkPolicyAllowed,
		v1beta1constants.LabelNetworkPolicyToRuntimeAPIServer:                                    v1beta1constants.LabelNetworkPolicyAllowed,
		v1beta1constants.LabelNetworkPolicyToPrivateNetworks:                                     v1beta1constants.LabelNetworkPolicyAllowed,
		v1beta1constants.LabelNetworkPolicyToPublicNetworks:                                      v1beta1constants.LabelNetworkPolicyAllowed,
		gardenerutils.NetworkPolicyLabel(targetAllocatorServiceName, targetAllocatorServicePort): v1beta1constants.LabelNetworkPolicyAllowed,
		toAllScrapeTargetsLabel:                                                                  v1beta1constants.LabelNetworkPolicyAllowed,
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
			Labels:    a.getLabels(),
		},
		AutomountServiceAccountToken: ptr.To(false),
	}

	return obj
}

// getTargetAllocatorRole returns the [rbacv1.Role] for the Target Allocator.
func (a *Actuator) getTargetAllocatorRole(namespace string) *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      targetAllocatorRoleName,
			Namespace: namespace,
			Labels:    a.getLabels(),
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
			Labels:    a.getLabels(),
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

// getTargetAllocator returns the [otelv1alpha1.TargetAllocator] resource.
func (a *Actuator) getTargetAllocator(namespace string) *otelv1alpha1.TargetAllocator {
	obj := &otelv1alpha1.TargetAllocator{
		ObjectMeta: metav1.ObjectMeta{
			Name:      targetAllocatorName,
			Namespace: namespace,
			Labels:    a.getLabels(),
		},
		// TODO(dnaeon): finish the rest of the spec
		Spec: otelv1alpha1.TargetAllocatorSpec{
			OpenTelemetryCommonFields: otelv1beta1.OpenTelemetryCommonFields{
				// TODO(dnaeon): add args, ports and volumes for TLS config

				Image:             "otel/target-allocator:v0.140.0", // TODO(dnaeon): this image should be configurable and vendored
				Replicas:          ptr.To(targetAllocatorReplicas),
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
				ServiceAccount: targetAllocatorServiceAccountName,
			},
			PrometheusCR: otelv1beta1.TargetAllocatorPrometheusCR{
				Enabled:         true,
				AllowNamespaces: []string{namespace},
				ServiceMonitorSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						// TODO(dnaeon): additional labels
						"prometheus": "shoot",
					},
				},
			},
		},
	}

	return obj
}
