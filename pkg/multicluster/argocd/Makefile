ROOT_DIR := ../../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/tools.mk

USER_ID := $(shell id -u)
GROUP_ID := $(shell id -g)

GIT_HOST := fybrik.io
GIT_USER_ID := argocd
GIT_REPO_ID := argocd-go

AUTO_GENERATED = auto-generated

IGNORE := $(shell bash -c "sed -n /=/p  ${ROOT_DIR}/hack/tools/requirements.env | sed 's/=/:=/' | sed 's/^/export /' > makeenv")
include makeenv

ABS_ROOT_DIR := $(abspath ${ROOT_DIR})
OPENAPI_GENERATOR_IMG := openapitools/openapi-generator-cli:v${OPENAPI_GENERATOR_VERSION}

.PHONY: patch
patch:
# workaround for https://github.com/argoproj/argo-cd/issues/4314
		jq 'setpath(["definitions", "v1Time"]; {"type":"string"}) | setpath(["definitions", "v1alpha1ClusterCacheInfo", "properties", "apisCount"]; {"type":"integer"}) | setpath(["definitions", "v1alpha1ClusterCacheInfo", "properties", "resourcesCount"]; {"type":"integer"}) | setpath(["definitions", "v1alpha1ClusterInfo", "properties", "applicationsCount"]; {"type":"integer"}) | setpath(["definitions", "v1ObjectMeta", "properties", "generation"]; {"type":"integer"})| setpath(["definitions", "v1alpha1RevisionHistory", "properties", "id"]; {"type":"integer"}) | setpath(["definitions", "v1alpha1RetryStrategy", "properties", "limit"]; {"type":"integer"})' swagger.json.orig > swagger.json

.PHONY: generate-client-argocd
generate-client-argocd: patch
	rm -Rf ${AUTO_GENERATED}/
	docker run --rm \
		-v ${PWD}:/local \
		-u "${USER_ID}:${GROUP_ID}" \
		${OPENAPI_GENERATOR_IMG} generate -g go \
		--additional-properties=packageName=openapiclient,isGoSubmodule=false,debugOpenAPI \
		--global-property=verbose=true,models,debugModels,apis="ApplicationService:ClusterService:SessionService",supportingFiles,modelTests=false,modelDocs=false,apiTests=false,apiDocs=false \
		--git-host=${GIT_HOST} \
		--git-user-id=${GIT_USER_ID} \
		--git-repo-id=${GIT_REPO_ID} \
		-o /local/${AUTO_GENERATED}/client \
		-i /local/swagger.json
	rm -f $(ROOT_DIR)/pkg/multicluster/argocd/auto-generated/client/go.mod $(ROOT_DIR)/pkg/multicluster/argocd/auto-generated/client/go.sum $(ROOT_DIR)/pkg/multicluster/argocd/auto-generated/client/.travis.yml
	rm -f $(ROOT_DIR)/pkg/multicluster/argocd/auto-generated/client/README.md $(ROOT_DIR)/pkg/multicluster/argocd/auto-generated/client/git_push.sh
	rm -f $(ROOT_DIR)/pkg/multicluster/argocd/auto-generated/client/.gitignore $(ROOT_DIR)/pkg/multicluster/argocd/auto-generated/client/.openapi-generator-ignore
	rm -rf $(ROOT_DIR)/pkg/multicluster/argocd/auto-generated/client/docs

.PHONY: check
check:
	go fmt ./...
	go vet ./...
	go mod tidy -compat=1.19

clean:
	rm makeenv
