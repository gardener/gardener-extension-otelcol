.DEFAULT_GOAL := build

# Set SHELL to bash and configure options
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

GOCMD?= go
SRC_ROOT := $(shell git rev-parse --show-toplevel)
HACK_DIR := $(SRC_ROOT)/hack
SRC_DIRS := $(shell $(GOCMD) list -f '{{ .Dir }}' ./...)

GOOS := $(shell $(GOCMD) env GOOS)
GOARCH := $(shell $(GOCMD) env GOARCH)
TOOLS_MOD_DIR := $(SRC_ROOT)/internal/tools
TOOLS_MOD_FILE := $(TOOLS_MOD_DIR)/go.mod
GO_MODULE := $(shell $(GOCMD) list -m -f '{{ .Path }}' )
GO_TOOL := $(GOCMD) tool -modfile $(TOOLS_MOD_FILE)

LOCAL_BIN ?= $(SRC_ROOT)/bin
BINARY    ?= $(LOCAL_BIN)/extension

VERSION := $(shell cat VERSION)
REVISION := $(shell git rev-parse --short HEAD)
EFFECTIVE_VERSION := $(VERSION)-$(REVISION)
ifneq ($(strip $(shell git status --porcelain 2>/dev/null)),)
	EFFECTIVE_VERSION := $(EFFECTIVE_VERSION)-dirty
endif

# Name and version of the Gardener extension.
EXTENSION_NAME ?= gardener-extension-otelcol

# Name for the extension image
IMAGE ?= europe-docker.pkg.dev/gardener-project/public/gardener/extensions/$(EXTENSION_NAME)

# Registry used for local development
LOCAL_REGISTRY ?= registry.local.gardener.cloud:5001
# Name of the kind cluster for local development
GARDENER_DEV_CLUSTER ?= gardener-local
# Name of the kind cluster for local development (with gardener-operator)
GARDENER_DEV_OPERATOR_CLUSTER ?= gardener-operator-local
# Name of the kind cluster for local Gardener dev environment
KIND_CLUSTER ?= $(GARDENER_DEV_CLUSTER)

# Kubernetes code-generator tools
#
# https://github.com/kubernetes/code-generator
K8S_GEN_TOOLS := deepcopy-gen defaulter-gen register-gen conversion-gen
K8S_GEN_TOOLS_LOG_LEVEL ?= 0

# ENVTEST_K8S_VERSION configures the version of Kubernetes, which will be
# installed by setup-envtest.
#
# In order to configure the Kubernetes version to match the version used by the
# k8s.io/api package, use the following setting.
#
# ENVTEST_K8S_VERSION ?= $(shell go list -m -f "{{ .Version }}" k8s.io/api | awk -F'[v.]' '{ printf "1.%d.%d", $$3, $$4 }')
#
# Or set the version here explicitly.
ENVTEST_K8S_VERSION ?= 1.34.1

# Common options for the `kubeconform' tool
KUBECONFORM_OPTS ?= 	-strict \
			-verbose \
			-summary \
			-output pretty \
			-skip Kustomization \
			-schema-location default

# Common options for the `addlicense' tool
ADDLICENSE_OPTS ?= -f $(HACK_DIR)/LICENSE_BOILERPLATE.txt \
			-ignore "dev/**" \
			-ignore "**/*.md" \
			-ignore "**/*.html" \
			-ignore "**/*.yaml" \
			-ignore "**/*.yml" \
			-ignore "**/Dockerfile"

# Path in which to generate the API reference docs
API_REF_DOCS ?= $(SRC_ROOT)/docs/api-reference

# Run a command.
#
# When used with `foreach' the result is concatenated, so make sure to preserve
# the empty whitespace at the end of this function.
#
# https://www.gnu.org/software/make/manual/html_node/Foreach-Function.html
define run-command
$(1)

endef

##@ gardener-extension-otelcol

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-30s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

$(LOCAL_BIN):
	mkdir -p $(LOCAL_BIN)

$(BINARY): $(SRC_DIRS) | $(LOCAL_BIN)
	$(GOCMD) build \
		-o $(LOCAL_BIN)/ \
		-ldflags="-X '$(GO_MODULE)/pkg/version.Version=${VERSION}'" \
		./cmd/extension

.PHONY: goimports-reviser
goimports-reviser:  ## Run goimports-reviser.
	@$(GO_TOOL) goimports-reviser -set-exit-status -rm-unused ./...

.PHONY: lint
lint:  ## Run linters.
	@$(GO_TOOL) golangci-lint run --config=$(SRC_ROOT)/.golangci.yaml ./...

.PHONY: govulncheck
govulncheck:  ## Run vulnerability scan.
	@$(GO_TOOL) govulncheck -show verbose ./...

.PHONY: api-ref-docs
api-ref-docs:  ## Generate API reference docs.
	@mkdir -p $(API_REF_DOCS)
	@$(GO_TOOL) crd-ref-docs \
		--config $(SRC_ROOT)/api-ref-docs.yaml \
		--output-mode group \
		--output-path $(API_REF_DOCS) \
		--renderer markdown \
		--source-path $(SRC_ROOT)/pkg/apis

.PHONY: build
build: $(BINARY)  ## Build the extension binary.

.PHONY: clean
clean:  ## Clean up binary and test utils.
	rm -f $(BINARY)
	$(GO_TOOL) setup-envtest --bin-dir $(LOCAL_BIN) cleanup

.PHONY: run
run: $(BINARY)  ## Run the extension binary.
	$(BINARY) manager

.PHONY: get
get:  ## Download Go modules and run go mod tidy.
	@$(GOCMD) mod download
	@$(GOCMD) mod tidy

.PHONY: gotidy
gotidy:  ## Run go mod tidy in main and tools modules.
	@$(GOCMD) mod tidy
	@cd $(TOOLS_MOD_DIR) && $(GOCMD) mod tidy

.PHONY: test
test:  ## Start envtest and run the unit tests.
	@echo "Setting up envtest for Kubernetes version v$(ENVTEST_K8S_VERSION) ..."
	@KUBEBUILDER_ASSETS="$$( $(GO_TOOL) setup-envtest use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCAL_BIN) -p path )" \
		$(GOCMD) test \
			-v \
			-race \
			-coverprofile=coverage.txt \
			-covermode=atomic \
			$(shell $(GOCMD) list ./pkg/... )
	@sed -i \
		-e '/generated.*\.go/d' \
		-e '/pkg\/apis\/config\/install/d' \
		-e '/pkg\/apis\/config\/register\.go/d' \
		-e '/pkg\/imagevector/d' \
		-e '/pkg\/apis\/config\/types\.go/d' \
		-e '/pkg\/apis\/config\/v1alpha1\/register\./d' \
		-e '/pkg\/metrics/d' coverage.txt

.PHONY: docker-build
docker-build:  ## Build the extension Docker image.
	@docker build \
		--build-arg BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ') \
		--build-arg VERSION=$(VERSION) \
		--build-arg REVISION=$(REVISION) \
		-t $(IMAGE):$(VERSION) \
		-t $(IMAGE):$(EFFECTIVE_VERSION) \
		-t $(IMAGE):latest .

.PHONY: docker-push
docker-push:  ## Push the extension Docker image.
	@docker push --quiet $(IMAGE):$(VERSION)
	@docker push --quiet $(IMAGE):$(EFFECTIVE_VERSION)
	@docker push --quiet $(IMAGE):latest

.PHONY: update-tools
update-tools:  ## Update Go tools.
	$(GOCMD) get -u -modfile $(TOOLS_MOD_FILE) tool

.PHONY: addlicense
addlicense:  ## Add license headers to all source files.
	@$(GO_TOOL) addlicense $(ADDLICENSE_OPTS) .

.PHONY: checklicense
checklicense:  ## Check source files for license headers.
	@files=$$( $(GO_TOOL) addlicense -check $(ADDLICENSE_OPTS) .) || { \
		echo "Missing license headers in the following files:"; \
		echo "$${files}"; \
		echo "Run 'make addlicense' in order to fix them."; \
		exit 1; \
	}

.PHONY: generate
generate:  ## Run code-generator tools.
	@echo "Running code-generator tools ..."
	$(foreach gen_tool,$(K8S_GEN_TOOLS),$(call run-command,$(GO_TOOL) $(gen_tool) -v=$(K8S_GEN_TOOLS_LOG_LEVEL) ./pkg/apis/...))

.PHONY: generate-operator-extension
generate-operator-extension:  ## Generate operator extension example resources.
	@$(GO_TOOL) extension-generator \
		--name $(EXTENSION_NAME) \
		--component-category extension \
		--provider-type otelcol \
		--destination $(SRC_ROOT)/examples/operator-extension/base/extension.yaml \
		--extension-oci-repository $(IMAGE):$(VERSION)
	@$(GO_TOOL) kustomize build $(SRC_ROOT)/examples/operator-extension

.PHONY: check-helm
check-helm:  ## Lint helm charts and validate rendered templates.
	@for chart in controller admission-runtime admission-virtual; do \
		$(GO_TOOL) helm lint $(SRC_ROOT)/charts/$${chart}; \
		$(GO_TOOL) helm template $(SRC_ROOT)/charts/$${chart} | \
			$(GO_TOOL) kubeconform $(KUBECONFORM_OPTS) \
				-schema-location 'https://raw.githubusercontent.com/datreeio/CRDs-catalog/main/{{.Group}}/{{.ResourceKind}}_{{.ResourceAPIVersion}}.json'; \
	done

.PHONY: check-examples
check-examples:  ## Lint the generated example resources.
	@echo "Checking example resources ..."
	@$(GO_TOOL) kubeconform \
		$(KUBECONFORM_OPTS) \
		-schema-location "$(SRC_ROOT)/test/schemas/{{.Group}}/{{.ResourceAPIVersion}}/{{.ResourceKind}}.json" \
		./examples
	@echo "Checking operator extension resource ..."
	@$(GO_TOOL) kustomize build $(SRC_ROOT)/examples/operator-extension | \
		$(GO_TOOL) kubeconform \
			$(KUBECONFORM_OPTS) \
			-schema-location "$(SRC_ROOT)/test/schemas/{{.Group}}/{{.ResourceAPIVersion}}/{{.ResourceKind}}.json"

.PHONY: kind-load-image
kind-load-image:  ## Load extension images to target cluster.
	@kind load docker-image --name $(KIND_CLUSTER) $(IMAGE):$(VERSION)
	@kind load docker-image --name $(KIND_CLUSTER) $(IMAGE):$(EFFECTIVE_VERSION)
	@kind load docker-image --name $(KIND_CLUSTER) $(IMAGE):latest

.PHONY: helm-load-chart
helm-load-chart:  ## Load helm chart to local registry.
	@for chart in controller admission-runtime admission-virtual; do \
		chart_name=$$( $(GO_TOOL) yq '.name' $(SRC_ROOT)/charts/$${chart}/Chart.yaml ); \
		$(GO_TOOL) helm package $(SRC_ROOT)/charts/$${chart} --version $(VERSION); \
		$(GO_TOOL) helm push --plain-http $${chart_name}-$(VERSION).tgz oci://$(LOCAL_REGISTRY)/helm-charts; \
		rm -f $${chart_name}-$(VERSION).tgz; \
	done

.PHONY: update-version-tags
update-version-tags:  ## Update version tags in helm charts and example resources based on VERSION file.
	@for chart in controller admission-runtime admission-virtual; do \
		env version=$(VERSION) $(GO_TOOL) yq -i '.version = env(version)' $(SRC_ROOT)/charts/$${chart}/Chart.yaml; \
	done

	@for chart in controller admission-runtime; do \
		env image=$(IMAGE) tag=$(VERSION) \
			$(GO_TOOL) yq -i '(.image.repository = env(image)) | (.image.tag = env(tag))' $(SRC_ROOT)/charts/$${chart}/values.yaml; \
	done

	@env oci_charts=$(LOCAL_REGISTRY)/helm-charts/$(EXTENSION_NAME):$(VERSION) \
		$(GO_TOOL) yq -i '.helm.ociRepository.ref = env(oci_charts)' $(SRC_ROOT)/examples/dev-setup/controllerdeployment.yaml
	@env oci_charts=$(LOCAL_REGISTRY)/helm-charts/$(EXTENSION_NAME):$(VERSION) \
		$(GO_TOOL) yq -i '.spec.deployment.extension.helm.ociRepository.ref = env(oci_charts)' $(SRC_ROOT)/examples/operator-extension/base/extension.yaml
	@env oci_charts=$(LOCAL_REGISTRY)/helm-charts/$(EXTENSION_NAME)-admission-runtime:$(VERSION) \
		$(GO_TOOL) yq -i '.spec.deployment.admission.runtimeCluster.helm.ociRepository.ref = env(oci_charts)' $(SRC_ROOT)/examples/operator-extension/base/extension.yaml
	@env oci_charts=$(LOCAL_REGISTRY)/helm-charts/$(EXTENSION_NAME)-admission-virtual:$(VERSION) \
		$(GO_TOOL) yq -i '.spec.deployment.admission.virtualCluster.helm.ociRepository.ref = env(oci_charts)' $(SRC_ROOT)/examples/operator-extension/base/extension.yaml

deploy deploy-operator: export IMAGE=$(LOCAL_REGISTRY)/extensions/$(EXTENSION_NAME)

.PHONY: deploy
deploy: generate update-version-tags docker-build docker-push helm-load-chart  ## Deploy to local dev cluster.
	@env WITH_GARDENER_OPERATOR=false EXTENSION_IMAGE=$(IMAGE):$(VERSION) $(HACK_DIR)/deploy-dev-setup.sh

.PHONY: undeploy
undeploy:  ## Cleanup the deployed extension.
	@$(GO_TOOL) kustomize build $(SRC_ROOT)/examples/dev-setup | \
		kubectl delete --ignore-not-found=true --wait=false -f -

.PHONY: deploy-operator
deploy-operator: generate update-version-tags docker-build docker-push helm-load-chart  ## Deploy to local dev cluster with Gardener Operator.
	@env WITH_GARDENER_OPERATOR=true EXTENSION_IMAGE=$(IMAGE):$(VERSION) $(HACK_DIR)/deploy-dev-setup.sh

.PHONY: undeploy-operator
undeploy-operator:  ## Cleanup the deployed operator extension.
	@$(GO_TOOL) kustomize build $(SRC_ROOT)/examples/operator-extension | \
		kubectl delete --ignore-not-found=true --wait=false -f -

.PHONY: create-dev-shoot
create-dev-shoot:  ## Create a dev shoot cluster with enabled extension.
	@kubectl apply -f $(SRC_ROOT)/examples/secret-tls.yaml
	@kubectl apply -f $(SRC_ROOT)/examples/secret-bearer-token.yaml
	@kubectl apply -f $(SRC_ROOT)/examples/opentelemetry-receiver.yaml
	@kubectl apply -f $(SRC_ROOT)/examples/shoot.yaml

.PHONY: delete-dev-shoot
delete-dev-shoot:  ## Delete the dev shoot cluster.
	@kubectl --namespace garden-local annotate shoot local confirmation.gardener.cloud/deletion=true --overwrite
	@kubectl delete -f $(SRC_ROOT)/examples/shoot.yaml --ignore-not-found=true --wait=false
	@kubectl delete -f $(SRC_ROOT)/examples/secret-tls.yaml --ignore-not-found=true --wait=false
	@kubectl delete -f $(SRC_ROOT)/examples/secret-bearer-token.yaml --ignore-not-found=true --wait=false
	@kubectl delete -f $(SRC_ROOT)/examples/opentelemetry-receiver.yaml --ignore-not-found=true --wait=false
