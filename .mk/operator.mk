# Image URL to use all building/pushing image targets
#IMG ?= controller:latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

.PHONY: all
all: manager

# Run tests
test: generate fmt vet manifests
	go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install CRDs into a cluster
.PHONY: install
install: $(TOOLBIN)/kustomize $(TOOLBIN)/kubectl manifests
	$(TOOLBIN)/kustomize build config/crd | $(TOOLBIN)/kubectl apply -f -

# Uninstall CRDs from a cluster
.PHONY: uninstall
uninstall: $(TOOLBIN)/kustomize $(TOOLBIN)/kubectl manifests
	$(TOOLBIN)/kustomize build config/crd | $(TOOLBIN)/kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: docker-secret manifests $(TOOLBIN)/kustomize $(TOOLBIN)/kubectl
	$(TOOLBIN)/kubectl create namespace m4d-system || true
	cd config/manager && $(ABSTOOLBIN)/kustomize edit set image controller=${IMG}
	$(TOOLBIN)/kustomize build config/default | $(TOOLBIN)/kubectl apply -f -

# Delete controller in the configured Kubernetes cluster in ~/.kube/config
undeploy: $(TOOLBIN)/kustomize $(TOOLBIN)/kubectl manifests
	$(TOOLBIN)/kustomize build config/default | $(TOOLBIN)/kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy-local: manifests $(TOOLBIN)/kustomize $(TOOLBIN)/kubectl
	$(TOOLBIN)/kubectl create namespace m4d-system || true
	cd config/manager && $(ABSTOOLBIN)/kustomize edit set image controller=manager:latest
	$(TOOLBIN)/kustomize build config/minikube | $(TOOLBIN)/kubectl apply -f -

deploy-test: manifests $(TOOLBIN)/kustomize $(TOOLBIN)/kubectl
	$(TOOLBIN)/kubectl create namespace m4d-system || true
	cd config/manager && $(ABSTOOLBIN)/kustomize edit set image controller=${IMG}
	$(TOOLBIN)/kustomize build config/test | $(TOOLBIN)/kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
.PHONY: generate
generate: $(TOOLBIN)/controller-gen
	$(TOOLBIN)/controller-gen object:headerFile=$(ROOT_DIR)/hack/boilerplate.go.txt,year=$(shell date +%Y) paths="./..."

# Generate code
.PHONY: manifests
manifests: $(TOOLBIN)/controller-gen
	$(TOOLBIN)/controller-gen $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases
