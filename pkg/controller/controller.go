// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package controller provides utility wrappers for registering controllers with
// Gardener extension actuators.
package controller

import (
	"context"
	"errors"
	"fmt"
	"time"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/extension"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	crctrl "sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// ErrInvalidController is an error, which is returned when attempting to create
// a controller, but the configuration was found to be invalid.
var ErrInvalidController = errors.New("invalid controller config")

// Controller wraps an [extension.Actuator], which reconciles
// [extensionsv1alpha1.Extension] resources.
type Controller struct {
	// actuator is the [extension.Actuator] used by the controller.
	actuator extension.Actuator

	// name is the name of the controller.
	name string

	// finalizerSuffix is the suffix for the finalizer.
	finalizerSuffix string

	// controllerOptions are the controller options used for creating a
	// controller.  The options.Reconciler is always overridden with a
	// reconciler created from the given actuator.
	controllerOptions crctrl.Options

	// predicates are the predicates to use.
	predicates []predicate.Predicate

	// resync determines the requeue interval.
	resync time.Duration

	// extensionType is the type of the resource considered for
	// reconciliation.
	extensionType string

	// watchBuilder defines additional watches on controllers that should be
	// set up.
	watchBuilder extensionscontroller.WatchBuilder

	// IgnoreOperationAnnotation specifies whether to ignore the operation
	// annotation or not.  If the annotation is not ignored, the extension
	// controller will only reconcile with a present operation annotation
	// typically set during a reconcile (e.g. in the maintenance time) by
	// the Gardenlet.
	ignoreOperationAnnotation bool

	// extensionClasses defines the extension classes this extension is
	// responsible for.
	extensionClasses []extensionsv1alpha1.ExtensionClass
}

// New creates a new [Controller] with the given options.
func New(opts ...Option) (*Controller, error) {
	c := &Controller{
		predicates:       make([]predicate.Predicate, 0),
		extensionClasses: make([]extensionsv1alpha1.ExtensionClass, 0),
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	// Validate controller configuration and set defaults, if needed.
	if c.actuator == nil {
		return nil, fmt.Errorf("%w: missing actuator implementation", ErrInvalidController)
	}
	if c.name == "" {
		return nil, fmt.Errorf("%w: missing controller name", ErrInvalidController)
	}
	if c.extensionType == "" {
		return nil, fmt.Errorf("%w: missing extension type", ErrInvalidController)
	}
	if len(c.extensionClasses) == 0 {
		return nil, fmt.Errorf("%w: missing extension class", ErrInvalidController)
	}
	if c.finalizerSuffix == "" {
		c.finalizerSuffix = c.name
	}

	return c, nil
}

// SetupWithManager registers the [Controller] with the given [manager.Manager].
// Internally, this method uses [extension.Add], which builds a reconciler
// wrapper around the [extension.Actuator] used by the [Controller].
func (c *Controller) SetupWithManager(ctx context.Context, mgr manager.Manager) error {
	if len(c.predicates) == 0 {
		c.predicates = extension.DefaultPredicates(ctx, mgr, c.ignoreOperationAnnotation)
	}

	return extension.Add(
		mgr,
		extension.AddArgs{
			Actuator:                  c.actuator,
			Name:                      c.name,
			FinalizerSuffix:           c.finalizerSuffix,
			ControllerOptions:         c.controllerOptions,
			Predicates:                c.predicates,
			Resync:                    c.resync,
			Type:                      c.extensionType,
			WatchBuilder:              c.watchBuilder,
			IgnoreOperationAnnotation: c.ignoreOperationAnnotation,
			ExtensionClasses:          c.extensionClasses,
		},
	)
}

// Option is a function, which configures the [Controller].
type Option func(c *Controller) error

// WithActuator is an [Option], which configures the [Controller] to use the
// given [extension.Actuator].
func WithActuator(act extension.Actuator) Option {
	opt := func(c *Controller) error {
		c.actuator = act

		return nil
	}

	return opt
}

// WithName is an [Option], which configures the [Controller] with the given
// name.
func WithName(name string) Option {
	opt := func(c *Controller) error {
		c.name = name

		return nil
	}

	return opt
}

// WithFinalizerSuffix is an [Option], which configures the [Controller] to use
// the given finalizer suffix.
func WithFinalizerSuffix(suffix string) Option {
	opt := func(c *Controller) error {
		c.finalizerSuffix = suffix

		return nil
	}

	return opt
}

// WithControllerOptions is an [Option], which configures the [Controller] to
// use the given [crctrl.Options].
func WithControllerOptions(opts crctrl.Options) Option {
	opt := func(c *Controller) error {
		c.controllerOptions = opts

		return nil
	}

	return opt
}

// WithPredicate is an [Option], which configures the [Controller] to use the
// given [predicate.Predicate].
func WithPredicate(pred predicate.Predicate) Option {
	opt := func(c *Controller) error {
		c.predicates = append(c.predicates, pred)

		return nil
	}

	return opt
}

// WithExtensionType is an [Option], which configures the [Controller] to
// reconcile extension resources of the given type.
func WithExtensionType(extensionType string) Option {
	opt := func(c *Controller) error {
		c.extensionType = extensionType

		return nil
	}

	return opt
}

// WithWatchBuilder is an [Option], which configures the [Controller] to
// use the given [crutils.WatchBuilder].
func WithWatchBuilder(builder extensionscontroller.WatchBuilder) Option {
	opt := func(c *Controller) error {
		c.watchBuilder = builder

		return nil
	}

	return opt
}

// WithIgnoreOperationAnnotation is an [Option], which configures the
// [Controller] whether to ignore the operations annotation, or not.
func WithIgnoreOperationAnnotation(ignore bool) Option {
	opt := func(c *Controller) error {
		c.ignoreOperationAnnotation = ignore

		return nil
	}

	return opt
}

// WithExtensionClass is an [Option], which configures the [Controller] to be
// responsible for the given [extensionsv1alpha1.ExtensionClass].
func WithExtensionClass(item extensionsv1alpha1.ExtensionClass) Option {
	opt := func(c *Controller) error {
		c.extensionClasses = append(c.extensionClasses, item)

		return nil
	}

	return opt
}

// WithResyncInterval is an [Option], which configures the requeue interval of
// the [Controller].
func WithResyncInterval(duration time.Duration) Option {
	opt := func(c *Controller) error {
		c.resync = duration

		return nil
	}

	return opt
}
