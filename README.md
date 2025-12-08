# gardener-extension-otelcol

The `gardener-extension-otelcol` repo provides Gardener Extension for OpenTelemetry Collector.

# Requirements

- [Go 1.25.x](https://go.dev/) or later
- [GNU Make](https://www.gnu.org/software/make/)
- [Docker](https://www.docker.com/) for local development
- [Gardener Local Setup](https://gardener.cloud/docs/gardener/local_setup/) for local development

# Code structure

The project repo uses the following code structure.

| Package           | Description                                                                              |
|-------------------|------------------------------------------------------------------------------------------|
| `cmd`             | Command-line application of the extension                                                |
| `pkg/apis`        | Extension API types, e.g. configuration spec, etc.                                       |
| `pkg/actuator`    | Implementations for the Gardener Extension Actuator interfaces                           |
| `pkg/controller`  | Utility wrappers for creating Kubernetes reconcilers for Gardener Actuators              |
| `pkg/heartbeat`   | Utility wrappers for creating heartbeat reconcilers for Gardener extensions              |
| `pkg/metrics`     | Metrics emitted by the extension                                                         |
| `pkg/mgr`         | Utility wrappers for creating `controller-runtime` managers using functional options API |
| `pkg/version`     | Version metadata information about the extension                                         |
| `internal/tools`  | Go-based tools used for testing and linting the project                                  |
| `charts`          | Helm charts for deploying the extension                                                  |
| `examples`        | Example Kubernetes resources, which can be used in a dev environment                     |
| `test`            | Various files (e.g. schemas, CRDs, etc.), used during testing                            |

# Usage

You can enable the extension for a [Gardener Shoot
cluster](https://gardener.cloud/docs/glossary/_index#gardener-glossary) by
updating the `.spec.extensions` of your shoot manifest.

``` yaml
...

spec:
  extensions:
    - type: otelcol
      providerConfig:
        apiVersion: otelcol.extensions.gardener.cloud/v1alpha1
        kind: ExampleConfig
        spec:
          foo: bar
```

# Development

In order to build a binary of the extension, you can use the following command.

``` shell
make build
```

The resulting binary can be found in `bin/extension`.

In order to build a Docker image of the extension, you can use the following
command.

``` shell
make docker-build
```

For local development of the `gardener-extension-otelcol` it is recommended that
you setup a [development Gardener environment](https://gardener.cloud/docs/gardener/local_setup/).

Please refer to the next sections for more information about deploying and
testing the extension in a Gardener development environment.

## Development Environment without Gardener Operator

The following documents describe how to create a Gardener development
environment locally. Please make sure to read them in order to familiarize
yourself with the setup, and also to install any prerequisites.

- [Gardener: Local setup requirements](https://gardener.cloud/docs/gardener/local_setup/)
- [Gardener: Getting Started Locally](https://gardener.cloud/docs/gardener/deployment/getting_started_locally/)

The steps from this section describe how to deploy and develop the extension
against a local development environment, without the
[Gardener Operator](https://gardener.cloud/docs/gardener/concepts/operator/).

In summary, these are the steps you need to follow in order to start a local
development Gardener environment, however, please make sure that you read the
documents above for additional details.

``` shell
make kind-up gardener-up
```

Before you continue with the next steps, make sure that you configure your
`KUBECONFIG` to point to the kubeconfig file created by Gardener for you.

This file will be located in the
`/path/to/gardener/example/gardener-local/kind/local/kubeconfig` path after
creating the dev environment.

``` shell
export KUBECONFIG=/path/to/gardener/example/gardener-local/kind/local/kubeconfig
```

You can use the following command in order to load the OCI image to the nodes of
your local Gardener cluster, which is running in
[kind](https://kind.sigs.k8s.io/).

``` shell
make kind-load-image
```

The Helm charts, which are used by the
[gardenlet](https://gardener.cloud/docs/gardener/concepts/gardenlet/) for
deploying the extension can be pushed to the local OCI registry using the
following command.

``` shell
make helm-load-chart
```

In the [examples/dev-setup](./examples/dev-setup) directory you can find
[kustomize](https://kustomize.io/]) resources, which can be used to create the
`ControllerDeployment` and `ControllerRegistration` resources.

For more information about `ControllerDeployment` and `ControllerRegistration`
resources, please make sure to check the
[Registering Extension Controllers](https://gardener.cloud/docs/gardener/extensions/registration/)
documentation.

The `deploy` target takes care of deploying your extension in a local Gardener
environment. It does the following.

1. Builds a Docker image of the extension
2. Loads the image into the `kind` cluster nodes
3. Packages the Helm charts and pushes them to the local registry
4. Deploys the `ControllerDeployment` and `ControllerRegistration` resources

``` shell
make deploy
```

Verify that we have successfully created the `ControllerDeployment` and
`ControllerRegistration` resources.

``` shell
$ kubectl get controllerregistrations,controllerdeployments gardener-extension-otelcol
NAME                                                                    RESOURCES           AGE
controllerregistration.core.gardener.cloud/gardener-extension-otelcol   Extension/otelcol   40s

NAME                                                                  AGE
controllerdeployment.core.gardener.cloud/gardener-extension-otelcol   40s
```

Finally, we can create an example shoot with our extension enabled. The
[examples/shoot.yaml](./examples/shoot.yaml) file provides a ready-to-use shoot
manifest with the extension enabled and configured.

``` shell
kubectl apply -f examples/shoot.yaml
```

Once we create the shoot cluster, `gardenlet` will start deploying our
`gardener-extension-otelcol`, since it is required by our shoot.

Verify that the extension has been successfully installed by checking the
corresponding `ControllerInstallation` resource.

``` shell
$ kubectl get controllerinstallations.core.gardener.cloud
NAME                               REGISTRATION                 SEED    VALID   INSTALLED   HEALTHY   PROGRESSING   AGE
gardener-extension-otelcol-tktwt   gardener-extension-otelcol   local   True    True        True      False         103s
```

After your shoot cluster has been successfully created and reconciled, verify
that the extension is healthy.

``` shell
$ kubectl --namespace shoot--local--local get extensions
NAME      TYPE      STATUS      AGE
otelcol   otelcol   Succeeded   85m
```

In order to trigger reconciliation of the extension you can annotate the
extension resource.

``` shell
kubectl --namespace shoot--local--local annotate extensions otelcol gardener.cloud/operation=reconcile
```

## Development Environment with Gardener Operator

The extension can also be deployed via the
[Gardener Operator](https://gardener.cloud/docs/gardener/concepts/operator/).

In order to start a local development environment with the Gardener Operator,
please refer to the following documentations.

- [Gardener Operator](https://gardener.cloud/docs/gardener/concepts/operator/)
- [Gardener: Local setup with gardener-operator](https://gardener.cloud/docs/gardener/deployment/getting_started_locally/#alternative-way-to-set-up-garden-and-seed-leveraging-gardener-operator)

In summary, these are the steps you need to follow in order to start a local
development environment with the [Gardener Operator](https://gardener.cloud/docs/gardener/concepts/operator/),
however, please make sure that you read the documents above for additional details.

``` shell
make kind-multi-zone-up operator-up operator-seed-up
```

Before you continue with the next steps, make sure that you configure your
`KUBECONFIG` to point to the kubeconfig file of the cluster, which runs the
Gardener Operator.

There will be two kubeconfig files created for you, after the dev environment
has been created.

| Path                                                                  | Description                                                         |
|-----------------------------------------------------------------------|---------------------------------------------------------------------|
| `/path/to/gardener/example/gardener-local/kind/multi-zone/kubeconfig` | Cluster in which `gardener-operator` runs (a.k.a _runtime_ cluster) |
| `/path/to/gardener/dev-setup/kubeconfigs/virtual-garden/kubeconfig`   | The _virtual_ garden cluster                                        |

Throughout this document we will refer to the kubeconfigs for _runtime_ and
_virtual_ clusters as `$KUBECONFIG_RUNTIME` and `$KUBECONFIG_VIRTUAL`
respectively.

Before deploying the extension we need to target the _runtime_ cluster, since
this is where the extension resources for `gardener-operator` reside.

``` shell
export KUBECONFIG=$KUBECONFIG_RUNTIME
```

In order to deploy the extension, execute the following command.

``` shell
make deploy-operator
```

The `deploy-operator` target takes care of the following.

1. Builds a Docker image of the extension
2. Loads the image into the `kind` cluster nodes
3. Packages the Helm charts and pushes them to the local registry
4. Deploys the `Extension` (from group `operator.gardener.cloud/v1alpha1`) to
   the _runtime_ cluster

Verify that we have successfully created the
`Extension` (from group `operator.gardener.cloud/v1alpha1`) resource.

``` shell
$ kubectl --kubeconfig $KUBECONFIG_RUNTIME get extop gardener-extension-otelcol
NAME                         INSTALLED   REQUIRED RUNTIME   REQUIRED VIRTUAL   AGE
gardener-extension-otelcol   True        False              False              85s
```

Verify that the respective `ControllerRegistration` and `ControllerDeployment`
resources have been created by the `gardener-operator` in the _virtual_ garden
cluster.

``` shell
> kubectl --kubeconfig $KUBECONFIG_VIRTUAL get controllerregistrations,controllerdeployments gardener-extension-otelcol
NAME                                                                    RESOURCES           AGE
controllerregistration.core.gardener.cloud/gardener-extension-otelcol   Extension/otelcol  3m50s

NAME                                                                  AGE
controllerdeployment.core.gardener.cloud/gardener-extension-otelcol   3m50s
```

Now we can create an example shoot with our extension enabled. The
[examples/shoot.yaml](./examples/shoot.yaml) file provides a ready-to-use shoot
manifest, which we will use.

``` shell
kubectl --kubeconfig $KUBECONFIG_VIRTUAL apply -f examples/shoot.yaml
```

Once we create the shoot cluster, `gardenlet` will start deploying our
`gardener-extension-otelcol`, since it is required by our shoot.

Verify that the extension has been successfully installed by checking the
corresponding `ControllerInstallation` resource for our extension.

``` shell
$ kubectl --kubeconfig $KUBECONFIG_VIRTUAL get controllerinstallations.core.gardener.cloud
NAME                               REGISTRATION                 SEED    VALID   INSTALLED   HEALTHY   PROGRESSING   AGE
gardener-extension-otelcol-ng4r8   gardener-extension-otelcol   local   True    True        True      False         2m9s
```

After your shoot cluster has been successfully created and reconciled, verify
that the extension is healthy.

``` shell
$ kubectl --kubeconfig $KUBECONFIG_RUNTIME --namespace shoot--local--local get extensions
NAME      TYPE      STATUS      AGE
otelcol   otelcol   Succeeded   2m37s
```

In order to trigger reconciliation of the extension you can annotate the
extension resource.

``` shell
kubectl --kubeconfig $KUBECONFIG_RUNTIME --namespace shoot--local--local annotate extensions otelcol gardener.cloud/operation=reconcile
```

# Tests

In order to run the tests use the command below:

``` shell
make test
```

In order to test the Helm chart and the manifests provided by it you can run the
following command.

``` shell
make check-helm
```

In order to test the example resources from the `examples/` directory you can
run the following command.

``` shell
make check-examples
```
# Documentation

Make sure to check the following documents for more information about Gardener
Extensions and the available extensions API.

- [Gardener: Extensibility Overview](https://gardener.cloud/docs/gardener/extensions/)
- [Gardener: Registering Extension Controllers](https://gardener.cloud/docs/gardener/extensions/registration/)
- [Gardener: Extension Resources](https://github.com/gardener/gardener/tree/master/docs/extensions/resources)
- [Gardener: Extensions API Contract](https://github.com/gardener/gardener/blob/master/docs/extensions/resources/extension.md)
- [Gardener: How to Set Up a Gardener Landscape](https://gardener.cloud/docs/gardener/deployment/setup_gardener/)
- [Gardener: Extension Packages (Go)](https://github.com/gardener/gardener/tree/master/extensions/pkg)

# Contributing

`gardener-extension-otelcol` is hosted on
[Github](https://github.com/gardener/gardener-extension-otelcol).

Please contribute by reporting issues, suggesting features or by sending patches
using pull requests.

# License

This project is Open Source and licensed under [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).
