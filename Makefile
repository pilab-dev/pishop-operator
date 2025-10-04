# Image URL to use all building/pushing image targets
IMG ?= pishop-operator:latest

# GHCR Configuration
GHCR_REGISTRY ?= ghcr.io
GHCR_NAMESPACE ?= pilab-dev
GHCR_IMAGE_NAME ?= pishop-operator
GHCR_TAG ?= latest
GHCR_FULL_IMAGE ?= $(GHCR_REGISTRY)/$(GHCR_NAMESPACE)/$(GHCR_IMAGE_NAME):$(GHCR_TAG)

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT_SHA ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "generateEmbeddedObjectMeta=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: generate manifests

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands don't modify the actual
# target descriptions, just add extra blank lines to make things more readable.
.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=pishop-operator-manager-role crd:$(CRD_OPTIONS) paths="./..." output:crd:dir=./config/crd/bases

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: manifests generate fmt vet ## Run tests.
	go test ./... -coverprofile cover.out

##@ Build

.PHONY: build
build: manifests generate fmt vet ## Build manager binary.
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags="-X 'main.Version=$(VERSION)' -X 'main.Commit=$(COMMIT_SHA)' -X 'main.BuildDate=$(BUILD_DATE)' -w -s" \
		-o bin/pishop-operator ./operator/

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./operator/main.go


##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	kubectl apply -f config/crd/bases

.PHONY: uninstall
uninstall: manifests ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	kubectl delete --ignore-not-found=$(ignore-not-found) -f config/crd/bases

.PHONY: deploy
deploy: build ## Build static binary, create container, push to GHCR, and deploy to K8s.
	@echo "🚀 Starting deployment workflow..."
	@echo "📋 Build Information:"
	@make info
	@echo ""
	@echo "🐳 Building Docker image with pre-built binary..."
	@docker build -f Dockerfile.deploy -t $(GHCR_FULL_IMAGE) -t $(GHCR_REGISTRY)/$(GHCR_NAMESPACE)/$(GHCR_IMAGE_NAME):$(VERSION) .
	@echo ""
	@echo "🔐 Checking GHCR authentication..."
	@if [ -n "$$GITHUB_TOKEN" ]; then \
		echo "Logging in to GHCR..."; \
		echo "$$GITHUB_TOKEN" | docker login $(GHCR_REGISTRY) -u $(GHCR_NAMESPACE) --password-stdin; \
	else \
		echo "⚠️  GITHUB_TOKEN not set, attempting push without login..."; \
	fi
	@echo ""
	@echo "📤 Pushing to GHCR..."
	@docker push $(GHCR_FULL_IMAGE)
	@docker push $(GHCR_REGISTRY)/$(GHCR_NAMESPACE)/$(GHCR_IMAGE_NAME):$(VERSION)
	@echo ""
	@echo "🔄 Updating deployment manifests..."
	@sed -i.bak 's|image: .*|image: $(GHCR_FULL_IMAGE)|g' config/manager/manager.yaml
	@rm -f config/manager/*.bak
	@echo ""
	@echo "📦 Creating namespace and applying CRDs..."
	@kubectl create namespace pishop-operator-system --dry-run=client -o yaml | kubectl apply -f -
	@kubectl apply -f config/crd/bases
	@echo ""
	@echo "🔐 Applying secrets..."
	@kubectl apply -f config/manager/mongodb-credentials-secret.yaml
	@kubectl apply -f config/manager/github-credentials-secret.yaml
	@echo ""
	@echo "🚀 Deploying to Kubernetes..."
	@kubectl apply -f config/manager/manager.yaml
	@echo ""
	@echo "⏳ Waiting for rollout to complete..."
	@kubectl rollout status deployment/pishop-operator -n pishop-operator-system --timeout=300s
	@echo ""
	@echo "✅ Deployment completed successfully!"
	@echo "📊 Pod status:"
	@kubectl get pods -n pishop-operator-system -l control-plane=controller-manager
	@echo ""
	@echo "📝 To view logs: kubectl logs -f deployment/pishop-operator -n pishop-operator-system"


.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster.
	kubectl delete -f config/samples/ --ignore-not-found=$(ignore-not-found)
	kubectl delete -f config/crd/bases --ignore-not-found=$(ignore-not-found)

.PHONY: clean
clean: ## Clean build artifacts and temporary files.
	rm -rf bin/
	rm -f cover.out
	docker rmi $(IMG) 2>/dev/null || true
	docker rmi $(GHCR_FULL_IMAGE) 2>/dev/null || true

.PHONY: info
info: ## Display build information.
	@echo "Build Information:"
	@echo "  Version: $(VERSION)"
	@echo "  Commit:  $(COMMIT_SHA)"
	@echo "  Date:    $(BUILD_DATE)"
	@echo "  Image:   $(GHCR_FULL_IMAGE)"

.PHONY: status
status: ## Check deployment status.
	@echo "📊 Deployment Status:"
	@echo ""
	@echo "🔍 CRDs:"
	@kubectl get crd | grep shop.pilab.hu || echo "  No CRDs found"
	@echo ""
	@echo "🚀 Deployments:"
	@kubectl get deployment pishop-operator -n pishop-operator-system 2>/dev/null || echo "  Deployment not found"
	@echo ""
	@echo "📦 Pods:"
	@kubectl get pods -n pishop-operator-system -l control-plane=controller-manager 2>/dev/null || echo "  No pods found"
	@echo ""
	@echo "📋 PRStacks:"
	@kubectl get prstacks 2>/dev/null || echo "  No PRStacks found"

.PHONY: logs
logs: ## Show operator logs.
	@echo "📝 Operator logs:"
	@kubectl logs -f deployment/pishop-operator -n pishop-operator-system

.PHONY: restart
restart: ## Restart the operator deployment.
	@echo "🔄 Restarting operator deployment..."
	@kubectl rollout restart deployment/pishop-operator -n pishop-operator-system
	@echo "⏳ Waiting for rollout to complete..."
	@kubectl rollout status deployment/pishop-operator -n pishop-operator-system --timeout=300s
	@echo "✅ Restart completed!"

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
## Tool Versions
CONTROLLER_TOOLS_VERSION ?= v0.17.0

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)
