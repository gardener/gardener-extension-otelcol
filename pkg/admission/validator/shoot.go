// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"context"
	"errors"
	"fmt"
	"slices"

	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	"github.com/gardener/gardener/pkg/apis/core"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/gardener/gardener-extension-otelcol/pkg/actuator"
	"github.com/gardener/gardener-extension-otelcol/pkg/apis/config"
	"github.com/gardener/gardener-extension-otelcol/pkg/apis/config/validation"
)

// ErrExtensionNotFound is an error, which is returned when the extension was
// not found in the [core.Shoot] spec.
var ErrExtensionNotFound = errors.New("extension not found")

// IgnoreExtensionNotFound returns nil if err is [ErrExtensionNotFound],
// otherwise it returns err.
func IgnoreExtensionNotFound(err error) error {
	if errors.Is(err, ErrExtensionNotFound) {
		return nil
	}

	return err
}

// shootValidator is an implementation of [extensionswebhook.Validator], which
// validates the provider configuration of the extension from a [core.Shoot]
// spec.
type shootValidator struct {
	decoder       runtime.Decoder
	extensionType string
}

var _ extensionswebhook.Validator = &shootValidator{}

// newShootValidator returns a new [shootValidator], which implements the
// [extensionswebhook.Validator] interface.
func newShootValidator(decoder runtime.Decoder) (*shootValidator, error) {
	validator := &shootValidator{
		decoder:       decoder,
		extensionType: actuator.ExtensionType,
	}

	if decoder == nil {
		return nil, fmt.Errorf("invalid decoder specified for shoot validator %s", validator.extensionType)
	}

	return validator, nil
}

// NewShootValidator returns a new [extensionswebhook.Validator] for
// [core.Shoot] objects.
func NewShootValidator(decoder runtime.Decoder) (extensionswebhook.Validator, error) {
	return newShootValidator(decoder)
}

// Validate implements the [extensionswebhook.Validator] interface.
func (v *shootValidator) Validate(ctx context.Context, newObj, oldObj client.Object) error {
	newShoot, ok := newObj.(*core.Shoot)
	if !ok {
		return fmt.Errorf("invalid object type: %T", newObj)
	}
	oldShoot, ok := oldObj.(*core.Shoot)
	if !ok {
		oldShoot = nil
	}

	if newShoot.DeletionTimestamp != nil {
		return nil
	}

	return v.validateExtension(newShoot, oldShoot)
}

// getExtension returns the [core.Extension] by extracting it from the given
// [core.Shoot] object.
func (v *shootValidator) getExtension(obj *core.Shoot) (core.Extension, error) {
	if obj == nil {
		return core.Extension{}, errors.New("invalid shoot resource provided")
	}

	idx := slices.IndexFunc(obj.Spec.Extensions, func(ext core.Extension) bool {
		return ext.Type == v.extensionType
	})

	if idx == -1 {
		return core.Extension{}, fmt.Errorf("%w: %s", ErrExtensionNotFound, v.extensionType)
	}

	return obj.Spec.Extensions[idx], nil
}

// validateExtension validates the extension configuration from the given
// [core.Shoot] specs.
func (v *shootValidator) validateExtension(newObj *core.Shoot, _ *core.Shoot) error {
	ext, err := v.getExtension(newObj)
	if err != nil {
		return IgnoreExtensionNotFound(err)
	}

	// Extension is disabled, nothing to validate
	if ext.Disabled != nil && *ext.Disabled {
		return nil
	}

	if ext.ProviderConfig == nil {
		return fmt.Errorf("no provider config specified for %s", v.extensionType)
	}

	var cfg config.CollectorConfig
	if err := runtime.DecodeInto(v.decoder, ext.ProviderConfig.Raw, &cfg); err != nil {
		return fmt.Errorf("invalid provider spec configuration for %s: %w", v.extensionType, err)
	}

	if err := validation.Validate(cfg); err != nil {
		return fmt.Errorf("invalid extension configuration for %s: %w", v.extensionType, err)
	}

	// TODO: additional validation checks, referenced secrets, etc.

	return nil
}

// NewShootValidatorWebhook returns a new validating [extensionswebhook.Webhook]
// for [core.Shoot] objects.
func NewShootValidatorWebhook(mgr manager.Manager) (*extensionswebhook.Webhook, error) {
	decoder := serializer.NewCodecFactory(mgr.GetScheme(), serializer.EnableStrict).UniversalDecoder()
	validator, err := newShootValidator(decoder)
	if err != nil {
		return nil, err
	}

	name := fmt.Sprintf("validator.%s", validator.extensionType)
	extensionLabel := fmt.Sprintf("extensions.extensions.gardener.cloud/%s", validator.extensionType)
	path := fmt.Sprintf("/webhooks/validate/%s", validator.extensionType)

	logger := mgr.GetLogger()
	logger.Info("setting up webhook", "name", name, "path", path, "label", extensionLabel)

	args := extensionswebhook.Args{
		Provider: validator.extensionType,
		Name:     name,
		Path:     path,
		Validators: map[extensionswebhook.Validator][]extensionswebhook.Type{
			validator: {{Obj: &core.Shoot{}}},
		},
		Target: extensionswebhook.TargetSeed,
		ObjectSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				extensionLabel: "true",
			},
		},
	}

	return extensionswebhook.New(mgr, args)
}
