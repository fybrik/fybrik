#!/bin/bash
set -x
set +e

#<<<<<<< upstreamscoping
export cluster_scoped=${cluster_scoped:-false}
=======
#>>>>>>> upstreamscoping-master
export run_tkn=${run_tkn:-0}
export skip_tests=${skip_tests:-false}
export GH_TOKEN=${GH_TOKEN:-fake}
export image_source_repo_password=${image_source_repo_password:-fake}
export cluster_scoped=${cluster_scoped:-false}
export git_user=${git_user:-fake@fake.com}
export github=${github:-github.com}
export github_workspace=${github_workspace}
export image_source_repo_username=${image_source_repo_username}
export image_repo="${image_repo:-kind-registry:5000}"
export image_source_repo="${image_source_repo:-fake.com}"
export dockerhub_hostname="${dockerhub_hostname:-docker.io}"
#<<<<<<< upstreamscoping
export cpd_url="${cpd_url:-https://cpd.fake.com}"
export git_url="${git_url:-https://github.com/fybrik/fybrik.git}"
export wkc_connector_git_url="${wkc_connector_git_url}"
export vault_plugin_secrets_wkc_reader_url="${vault_plugin_secrets_wkc_reader_url}"
export use_application_namespace=${use_application_namespace:-false}
#=======
export git_url="${git_url:-https://github.com/fybrik/fybrik.git}"
export is_kind="${is_kind:-false}"
#>>>>>>> upstreamscoping-master

helper_text=""
realpath() {
    [[ $1 = /* ]] && echo "$1" || echo "$PWD/${1#./}"
}
kube_os="$(uname | tr '[:upper:]' '[:lower:]')"
base64_arg="-w 0"
if [[ $kube_os == "darwin" ]]; then
    base64_arg="-b 0"
fi

repo_root=$(realpath $(dirname $(realpath $0)))/..

. ${repo_root}/pipeline/common_functions.sh

# Define function for cleaning up temp directory created during script runtime
function cleanup {
    if [[ ! -z ${TMP} ]]; then
        rm -rf ${TMP}
        echo "Deleted temp working directory ${TMP}"
        echo ${helper_text}
    fi
}

# Create a temp dir to place files in
if [[ -z "$TMP" ]]; then
    TMP=$(mktemp -d) || exit 1
    trap cleanup EXIT
fi

# Set ssh key path for potential use with authenticated github
if [[ ! -z $2 ]]; then
    ssh_key=$2
else
    ssh_key=${HOME}/.ssh/id_rsa
fi
set -e

#<<<<<<< upstreamscoping
extra_params="-p clusterScoped=${cluster_scoped}"

# Figure out if we're using air-gapped machines that should pull images from somewhere other than dockerhub
is_external="false"
is_internal="false"
helm_image=
build_image=
if [[ "${github}" == "github.com" ]]; then
    is_external="true"
=======
# Figure out if we're using air-gapped machines that should pull images from somewhere other than dockerhub
extra_params=''
is_public_repo="false"
is_custom_repo="false"
helm_image=
build_image=
if [[ "${github}" == "github.com" ]]; then
    is_public_repo="true"
#>>>>>>> upstreamscoping-master
    build_image="docker.io/yakinikku/suede_compile"
    helm_image="docker.io/lachlanevenson/k8s-helm:latest"
    extra_params="${extra_params} -p build_image=${build_image} -p helm_image=${helm_image}"
    cp ${repo_root}/pipeline/statefulset.yaml ${TMP}/
else
#<<<<<<< upstreamscoping
    is_internal="true"
=======
    is_custom_repo="true"
#>>>>>>> upstreamscoping-master
    build_image="${dockerhub_hostname}/suede_compile:latest"
    helm_image="${dockerhub_hostname}/k8s-helm"
    extra_params="${extra_params} -p build_image=${build_image} -p helm_image=${helm_image}"
    cp ${repo_root}/pipeline/statefulset.yaml ${TMP}/
    sed -i.bak "s|image: docker.io/yakinikku/suede:latest|image: ${dockerhub_hostname}/suede|g" ${TMP}/statefulset.yaml
fi

# Figure out if we're running on OpenShift or Kubernetes (kind)
is_openshift="false"
is_kubernetes="false"
client=kubectl
pipeline_sa=pipeline
set +e
kubectl get ns | grep openshift-apiserver
rc=$?
if [[ $rc -eq 0 ]]; then
    is_openshift=true
    client=oc
    pipeline_sa=pipeline
else
    is_kubernetes=true
    client=kubectl
    pipeline_sa=default
fi
#<<<<<<< upstreamscoping
if [[ ${is_kubernetes} == "true" ]]; then
    # Assume this is a kind cluster, and install nfs client pvc
    set -e
    kubectl apply -f ${repo_root}/pipeline/nfs.yaml
    helm repo add stable https://charts.helm.sh/stable
    ip=$(kubectl get svc -n default nfs-service -o jsonpath='{.spec.clusterIP}')
    helm upgrade --install nfs-provisioner stable/nfs-client-provisioner --values ${repo_root}/pipeline/nfs-values.yaml --set nfs.server=${ip} --namespace nfs-provisioner --create-namespace
=======
if [[ ${is_kubernetes} == "true" && ${is_kind} == "true" ]]; then
    # If we're running kind and we don't have a default storage class, install nfs client for RWX volume support
    set +e
    kubectl get sc | grep -v "standard (default)" | grep "(default)"
    rc=$?
    set -e
    if [[ $rc -ne 0 ]]; then
        kubectl apply -f ${repo_root}/pipeline/nfs.yaml -n default
        helm repo add stable https://charts.helm.sh/stable
        ip=$(kubectl get svc -n default nfs-service -o jsonpath='{.spec.clusterIP}')
        helm upgrade --install nfs-provisioner stable/nfs-client-provisioner --values ${repo_root}/pipeline/nfs-values.yaml --set nfs.server=${ip} --namespace nfs-provisioner --create-namespace
    fi
#>>>>>>> upstreamscoping-master
fi

# See if an install namespace will need to be created
set +e
rc=1
if [[ ! -z $1 ]]; then
    kubectl get ns $1
    rc=$?
else
    kubectl get ns fybrik-system
    rc=$?
fi

# Create new project if necessary
if [[ $rc -ne 0 ]]; then
    if [[ ${is_openshift} == "true" ]]; then
        oc new-project ${1:-fybrik-system}
    else
        kubectl create ns ${1:-fybrik-system}
        kubectl config set-context --current --namespace=${1:-fybrik-system}
    fi
else
    if [[ ${is_openshift} == "true" ]]; then
        oc project ${1:-fybrik-system}
    else
        kubectl config set-context --current --namespace=${1:-fybrik-system}
    fi
fi
unique_prefix=$(kubectl config view --minify --output 'jsonpath={..namespace}'; echo)

#<<<<<<< upstreamscoping
if [[ "${unique_prefix}" == "fybrik-system" ]]; then
  blueprint_namespace="fybrik-blueprints"
else
  blueprint_namespace="${unique_prefix}-blueprints"
fi
extra_params="${extra_params} -p blueprintNamespace=${blueprint_namespace}"

fybrik_values="cluster.name=AmsterdamCluster,cluster.zone=Netherlands,cluster.region=Netherlands,cluster.vaultAuthPath=kubernetes,coordinator.catalog=WKC,coordinator.catalogConnectorURL=wkc-connector:50090"
if [[ ${cluster_scoped} == "false" ]]; then
    if [[ ${use_application_namespace} == "false" ]]; then
      fybrik_values="${fybrik_values},applicationNamespace=${unique_prefix}"
    else 
      fybrik_values="${fybrik_values},applicationNamespace=${unique_prefix}-app"
    fi
fi

if [[ ${cluster_scoped} == "false" && ${use_application_namespace} == "true"  ]]; then
  set +e
  rc=1
  kubectl get ns ${unique_prefix}-app
  rc=$?
  set -e
  # Create new project if necessary
  if [[ $rc -ne 0 ]]; then
    if [[ ${is_openshift} == "true" ]]; then
      oc new-project ${unique_prefix}-app
      oc project ${unique_prefix} 
    else
      kubectl create ns ${unique_prefix}-app 
    fi
  fi

  cat > ${TMP}/approle.yaml <<EOH
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: ${unique_prefix}-app-role
  namespace: ${unique_prefix}-app 
rules:
- apiGroups:
  - '*'
  resources:
  - '*'
  verbs:
  - '*'
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: ${unique_prefix}-app-rb
  namespace: ${unique_prefix}-app
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: ${unique_prefix}-app-role
subjects:
- kind: ServiceAccount
  name: manager
  namespace: ${unique_prefix}
EOH
  set +e
  oc delete -f ${TMP}/approle.yaml
  set -e
  oc apply -f ${TMP}/approle.yaml
fi

set +e
# Be smarter about this - just a quick hack for typical default OpenShift & Kind installs so we can control the default storage class
oc patch storageclass managed-nfs-storage -p '{"metadata": {"annotations": {"storageclass.kubernetes.io/is-default-class": "true"}}}'
oc patch storageclass standard -p '{"metadata": {"annotations": {"storageclass.kubernetes.io/is-default-class": "false"}}}'
=======
set +e
# Be smarter about this - just a quick hack for typical default OpenShift & Kind installs so we can control the default storage class
kubectl patch storageclass managed-nfs-storage -p '{"metadata": {"annotations": {"storageclass.kubernetes.io/is-default-class": "true"}}}'
kubectl patch storageclass standard -p '{"metadata": {"annotations": {"storageclass.kubernetes.io/is-default-class": "false"}}}'
#>>>>>>> upstreamscoping-master
set -e

# Install Tekton & Knative eventing
if [[ ${is_openshift} == "true" ]]; then
    oc apply -f ${repo_root}/pipeline/subscription.yaml
    oc apply -f ${repo_root}/pipeline/serverless-subscription.yaml

    cat > ${TMP}/streams_csv_check_script.sh <<EOH
#!/bin/bash
set -x
oc get -n openshift-pipelines csv | grep redhat-openshift-pipelines-operator 
oc get -n openshift-pipelines csv | grep redhat-openshift-pipelines-operator | grep Succeeded
EOH
#<<<<<<< upstreamscoping
    chmod u+x ${TMP}/streams_csv_check_script.sh
=======
chmod u+x ${TMP}/streams_csv_check_script.sh
#>>>>>>> upstreamscoping-master
    try_command "${TMP}/streams_csv_check_script.sh"  40 true 5

    cat > ${TMP}/streams_csv_check_script.sh <<EOH
#!/bin/bash
set -x
oc get -n openshift-operators csv | grep serverless-operator
oc get -n openshift-operators csv | grep serverless-operator | grep -e Succeeded -e Replacing
EOH
#<<<<<<< upstreamscoping
    chmod u+x ${TMP}/streams_csv_check_script.sh
    try_command "${TMP}/streams_csv_check_script.sh"  40 false 5
    oc apply -f ${repo_root}/pipeline/knative-eventing.yaml
else
    kubectl apply --filename https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml
    kubectl apply -f https://github.com/knative/operator/releases/download/v0.15.4/operator.yaml
    kubectl apply -f https://github.com/knative/eventing/releases/download/v0.21.0/eventing-crds.yaml
    kubectl apply -f https://github.com/knative/eventing/releases/download/v0.21.0/eventing-core.yaml
    kubectl wait pod -n tekton-pipelines --all --for=condition=Ready --timeout=3m
=======
chmod u+x ${TMP}/streams_csv_check_script.sh
    try_command "${TMP}/streams_csv_check_script.sh"  40 false 5
    oc apply -f ${repo_root}/pipeline/knative-eventing.yaml
else
    kubectl apply -f https://storage.googleapis.com/tekton-releases/pipeline/previous/v0.26.0/release.yaml
    kubectl apply -f https://github.com/knative/operator/releases/download/v0.15.4/operator.yaml
    kubectl apply -f https://github.com/knative/eventing/releases/download/v0.21.0/eventing-crds.yaml
    kubectl apply -f https://github.com/knative/eventing/releases/download/v0.21.0/eventing-core.yaml
    try_command "kubectl wait pod -n tekton-pipelines --all --for=condition=Ready --timeout=3m" 2 true 1
#>>>>>>> upstreamscoping-master
    set +e
    kubectl create ns knative-eventing
    set -e
    cat > ${TMP}/knative-eventing.yaml <<EOH
apiVersion: operator.knative.dev/v1alpha1
kind: KnativeEventing
metadata:
  name: knative-eventing
  namespace: knative-eventing
EOH
    ls -alrt ${TMP}/
    kubectl apply -f ${TMP}/knative-eventing.yaml
fi

if [[ ${is_openshift} == "true" ]]; then
    # Give extended privileges to the pipeline service account, make sure images can be pulled from the integrated openshift registry
    oc adm policy add-cluster-role-to-user cluster-admin system:serviceaccount:${unique_prefix}:pipeline
    oc adm policy add-cluster-role-to-user cluster-admin system:serviceaccount:${unique_prefix}:root-sa
    oc adm policy add-role-to-group system:image-puller system:serviceaccounts:${unique_prefix} --namespace ${unique_prefix}
#<<<<<<< upstreamscoping
    oc adm policy add-role-to-group system:image-puller system:serviceaccounts:${blueprint_namespace} --namespace ${unique_prefix}
    oc adm policy add-role-to-group system:image-puller system:serviceaccounts:${unique_prefix}-app --namespace ${unique_prefix}
=======
    oc adm policy add-role-to-group system:image-puller system:serviceaccounts:fybrik-blueprints --namespace ${unique_prefix}
#>>>>>>> upstreamscoping-master
    oc adm policy add-role-to-user system:image-puller system:serviceaccount:${unique_prefix}:wkc-connector --namespace ${unique_prefix}
    
    # Temporary hack pending a better solution
    oc adm policy add-scc-to-user anyuid system:serviceaccount:${unique_prefix}:opa-connector
    oc adm policy add-scc-to-user anyuid system:serviceaccount:${unique_prefix}:manager
else
    set +e
    # Tekton pipeline will run under the default service account in Kind/Kubernetes, so give that admin privileges as well
    kubectl create clusterrolebinding ${unique_prefix}-default-cluster-admin --clusterrole=cluster-admin --serviceaccount=${unique_prefix}:default
fi

set -e
set +x
helper_text="If this step fails, tekton related pods may be restarting or initializing:

1. Please rerun in a minute or so
"
set -x
#<<<<<<< upstreamscoping
oc apply -f ${repo_root}/pipeline/shell.yaml
oc apply -f ${repo_root}/pipeline/tasks/make.yaml
oc apply -f ${repo_root}/pipeline/tasks/git-clone.yaml
oc apply -f ${repo_root}/pipeline/tasks/buildah.yaml
oc apply -f ${repo_root}/pipeline/tasks/skopeo-copy.yaml
oc apply -f ${repo_root}/pipeline/tasks/openshift-client.yaml
oc apply -f ${repo_root}/pipeline/tasks/helm-upgrade-from-source.yaml 
oc apply -f ${repo_root}/pipeline/tasks/helm-upgrade-from-repo.yaml 
=======
kubectl apply -f ${repo_root}/pipeline/tasks/make.yaml
kubectl apply -f ${repo_root}/pipeline/tasks/git-clone.yaml
kubectl apply -f ${repo_root}/pipeline/tasks/buildah.yaml
kubectl apply -f ${repo_root}/pipeline/tasks/skopeo-copy.yaml
kubectl apply -f ${repo_root}/pipeline/tasks/openshift-client.yaml
kubectl apply -f ${repo_root}/pipeline/tasks/helm-upgrade-from-source.yaml 
kubectl apply -f ${repo_root}/pipeline/tasks/helm-upgrade-from-repo.yaml 
#>>>>>>> upstreamscoping-master
helper_text=""

# Wipe old pipeline definitions in case of merge conflicts
set +e
#<<<<<<< upstreamscoping
oc delete -f ${repo_root}/pipeline/pipeline.yaml
oc delete -f ${repo_root}/pipeline/wkc-pipeline.yaml

# If running open source, exclude WKC from the pipeline
set -e
if [[ "${github}" == "github.com" ]]; then 
    oc apply -f ${repo_root}/pipeline/pipeline.yaml
else
    oc apply -f ${repo_root}/pipeline/wkc-pipeline.yaml
=======
kubectl delete -f ${repo_root}/pipeline/pipeline.yaml
if [[ -f ${repo_root}/pipeline/custom_pipeline_cleanup.sh ]]; then
    ${repo_root}/pipeline/custom_pipeline_cleanup.sh
fi

set -e
if [[ "${is_public_repo}" == "true" ]]; then
    kubectl apply -f ${repo_root}/pipeline/pipeline.yaml
else 
     kubectl apply -f ${repo_root}/pipeline/wkc-pipeline.yaml
     if [[ -f ${repo_root}/pipeline/custom_pipeline_create.sh ]]; then
         ${repo_root}/pipeline/custom_pipeline_create.sh
     else
         set +x
         echo "You are running with a non public repo, but no custom pipeline creation script exists at ${repo_root}/pipeline/custom_pipeline_create.sh"
         exit 1
     fi
#>>>>>>> upstreamscoping-master
fi

# Delete old registry credentials
set +e
#<<<<<<< upstreamscoping
oc delete secret -n ${unique_prefix} regcred --wait
oc delete secret -n ${unique_prefix} regcred-test --wait
oc delete secret -n ${unique_prefix} sourceregcred --wait
=======
kubectl delete secret -n ${unique_prefix} regcred --wait
kubectl delete secret -n ${unique_prefix} regcred-test --wait
kubectl delete secret -n ${unique_prefix} sourceregcred --wait
#>>>>>>> upstreamscoping-master
set -e
set -x

# See if we have a pull secret available on cluster that has access to authenticated registries we need
set +e
#<<<<<<< upstreamscoping
oc get secret -n openshift-config pull-secret -o yaml > ${TMP}/secret.yaml
rc=$?
if [[ ${rc} -eq 0 ]]; then
    helper_text=""
    oc get secret -n openshift-config pull-secret -o=go-template='{{index .data ".dockerconfigjson"}}' | base64 --decode | grep "${image_source_repo}"
=======
kubectl get secret -n openshift-config pull-secret -o yaml > ${TMP}/secret.yaml
rc=$?
if [[ ${rc} -eq 0 ]]; then
    helper_text=""
    kubectl get secret -n openshift-config pull-secret -o=go-template='{{index .data ".dockerconfigjson"}}' | base64 --decode | grep "${image_source_repo}"
#>>>>>>> upstreamscoping-master
    rc=$?
    if [[ ${rc} -eq 0 ]]; then
        set -e
        cp ${TMP}/secret.yaml ${TMP}/secret.yaml.orig
        sed -i.bak "s|namespace: openshift-config|namespace: ${unique_prefix}|g" ${TMP}/secret.yaml
        sed -i.bak "s|name: pull-secret|name: regcred|g" ${TMP}/secret.yaml
        cat ${TMP}/secret.yaml
#<<<<<<< upstreamscoping
        oc apply -f ${TMP}/secret.yaml
=======
        kubectl apply -f ${TMP}/secret.yaml
#>>>>>>> upstreamscoping-master
    else
        if [[ ! -z ${image_source_repo_password} ]]; then
            set -e
            auth=$(echo -n "${image_source_repo_username:-$git_username}:${image_source_repo_password}" | base64 ${base64_arg})
            cat > ${TMP}/secret.yaml <<EOH
{"auths":{"${image_source_repo}":{"username":"${image_source_repo_username:-$git_username}","password":"${image_source_repo_password}","auth":"${auth}"}}}
EOH
            kubectl create secret -n ${unique_prefix} generic regcred --from-file=.dockerconfigjson=${TMP}/secret.yaml --type=kubernetes.io/dockerconfigjson
        else
            helper_text="Run the following commands to set up credentials for ${image_source_repo}:

            export image_source_repo_password=xxx
            export image_source_repo_username=user@email.com
            "
            exit 1
        fi
    fi
else
    helper_text=""
    if [[ ! -z ${image_source_repo_password} ]]; then
        set -e
        auth=$(echo -n "${image_source_repo_username:-$git_username}:${image_source_repo_password}" | base64 ${base64_arg})
        cat > ${TMP}/secret.yaml <<EOH
{"auths":{"${image_source_repo}":{"username":"${image_source_repo_username:-$git_username}","password":"${image_source_repo_password}","auth":"${auth}"}}}
EOH
        kubectl create secret -n ${unique_prefix} generic regcred --from-file=.dockerconfigjson=${TMP}/secret.yaml --type=kubernetes.io/dockerconfigjson
    else
        helper_text="Run the following commands to set up credentials for ${image_source_repo}:
    
        export image_source_repo_password=xxx
        export image_source_repo_username=user@email.com
        "
        exit 1
    fi
fi

# Patch service accounts with necessary secrets for pulling images from authenticated registries
if [[ ${is_openshift} == "true" ]]; then
    oc secrets link pipeline regcred --for=mount
    oc secrets link builder regcred --for=mount
    oc secrets link pipeline regcred --for=pull
else
    kubectl patch serviceaccount default -p '{"imagePullSecrets": [{"name": "regcred"}]}'
    kubectl patch serviceaccount default -p '{"secrets": [{"name": "regcred"}]}'
fi

#<<<<<<< upstreamscoping
extra_params="${extra_params} -p deployVault='true'"
deploy_vault="true"
set +e
oc get crd | grep "fybrikapplications.app.fybrik.io"
=======
# Install resources that are cluster scoped only if installing to fybrik-system
cluster_scoped="false"
deploy_vault="false"
if [[ "${unique_prefix}" == "fybrik-system" ]]; then
    extra_params="${extra_params} -p clusterScoped='true' -p deployVault='true'"
    cluster_scoped="true"
    deploy_vault="true"
fi
set +e
kubectl get crd | grep "fybrikapplications.app.fybrik.ibm.com"
#>>>>>>> upstreamscoping-master
rc=$?
deploy_crd="false"
if [[ $rc -ne 0 ]]; then
    extra_params="${extra_params} -p deployCRD='true'"
    deploy_crd="true"
fi

# Don't attempt to reinstall certmanager if some form of it is already installed
#<<<<<<< upstreamscoping
oc get crd | grep "certmanager"
=======
kubectl get crd | grep "certmanager"
#>>>>>>> upstreamscoping-master
rc=$?
deploy_cert_manager="false"
if [[ $rc -ne 0 ]]; then
    extra_params="${extra_params} -p deployCertManager='true'"
    deploy_cert_manager="true"
fi

#<<<<<<< upstreamscoping
# Create a workspace to allow users to exec in and run arbitrary commands
oc apply -f ${repo_root}/pipeline/rootsa.yaml
oc apply -f ${TMP}/statefulset.yaml
oc apply -f ${repo_root}/pipeline/pvc.yaml
=======
set +e
kubectl get ns fybrik-system
rc=$?
set -e
if [[ $rc -ne 0 ]]; then
    set +x
    helper_text="please install into fybrik-system first - currently vault can only be installed in one namespace, and needs to go in fybrik-system"
    exit 1
fi

# Create a workspace to allow users to exec in and run arbitrary commands
kubectl apply -f ${repo_root}/pipeline/rootsa.yaml
kubectl apply -f ${TMP}/statefulset.yaml
kubectl apply -f ${repo_root}/pipeline/pvc.yaml
#>>>>>>> upstreamscoping-master
if [[ ${is_openshift} == "true" ]]; then
    oc adm policy add-scc-to-user privileged system:serviceaccount:${unique_prefix}:root-sa
fi

# Install tekton triggers
pushd ${TMP}
wget https://storage.googleapis.com/tekton-releases/triggers/latest/release.yaml
if [[ ${is_openshift} == "true" ]]; then
    sed -i.bak 's|namespace: tekton-pipelines|namespace: openshift-pipelines|g' ${TMP}/release.yaml
fi
cat ${TMP}/release.yaml
wget https://storage.googleapis.com/tekton-releases/triggers/latest/interceptors.yaml
if [[ ${is_openshift} == "true" ]]; then
    sed -i.bak 's|namespace: tekton-pipelines|namespace: openshift-pipelines|g' ${TMP}/interceptors.yaml
fi
cat ${TMP}/interceptors.yaml
popd
#<<<<<<< upstreamscoping
oc apply -f ${TMP}/release.yaml
oc apply -f ${TMP}/interceptors.yaml

# Delete old apiserversource
set +e
oc delete apiserversource generic-watcher
set -e

# Install triggers for rebuilds of specific tasks
oc apply -f ${repo_root}/pipeline/eventlistener/triggerbinding.yaml
oc apply -f ${repo_root}/pipeline/eventlistener/triggertemplate.yaml
oc apply -f ${repo_root}/pipeline/eventlistener/apiserversource.yaml
oc apply -f ${repo_root}/pipeline/eventlistener/role.yaml
oc apply -f ${repo_root}/pipeline/eventlistener/serviceaccount.yaml
=======
kubectl apply -f ${TMP}/release.yaml
kubectl apply -f ${TMP}/interceptors.yaml

# Delete old apiserversource
set +e
kubectl delete apiserversource generic-watcher
set -e

# Install triggers for rebuilds of specific tasks
kubectl apply -f ${repo_root}/pipeline/eventlistener/triggerbinding.yaml
kubectl apply -f ${repo_root}/pipeline/eventlistener/triggertemplate.yaml
kubectl apply -f ${repo_root}/pipeline/eventlistener/apiserversource.yaml
kubectl apply -f ${repo_root}/pipeline/eventlistener/role.yaml
kubectl apply -f ${repo_root}/pipeline/eventlistener/serviceaccount.yaml
#>>>>>>> upstreamscoping-master

set +x
helper_text="If this step fails, run again - knative related pods may be restarting and unable to process the webhook
"
set -x
#<<<<<<< upstreamscoping
oc apply -f ${repo_root}/pipeline/eventlistener/eventlistener.yaml
helper_text=""
set +e
oc delete rolebinding generic-watcher
oc delete rolebinding tekton-task-watcher
set -e
oc create rolebinding tekton-task-watcher --role=tekton-task-watcher --serviceaccount=${unique_prefix}:tekton-task-watcher

set +e
oc delete secret git-ssh-key
oc delete secret git-token
set -e

# Determine which set of vault values to used, based on whether or not WKC components will be installed
vault_values="/workspace/source/vault-plugin-secrets-wkc-reader/helm-deployment/vault-single-cluster/values.yaml"
if [[ "${github}" == "github.com" ]]; then
    vault_values="/workspace/source/fybrik/third_party/vault/vault-single-cluster/values.yaml"
=======
kubectl apply -f ${repo_root}/pipeline/eventlistener/eventlistener.yaml
helper_text=""
set +e
kubectl delete rolebinding generic-watcher
kubectl delete rolebinding tekton-task-watcher
set -e
kubectl create rolebinding tekton-task-watcher --role=tekton-task-watcher --serviceaccount=${unique_prefix}:tekton-task-watcher

set +e
kubectl delete secret git-ssh-key
kubectl delete secret git-token
set -e

# Determine which set of vault values to used, based on whether or not custom components will be installed
vault_values=
if [[ "${is_public_repo}" == "true" ]]; then
    vault_values="/workspace/source/fybrik/third_party/vault/vault-single-cluster/values.yaml"
else
    if [[ -f ${repo_root}/pipeline/custom_vault_values_reference.sh ]]; then
        source ${repo_root}/pipeline/custom_vault_values_reference.sh
    fi
#>>>>>>> upstreamscoping-master
fi
extra_params="${extra_params} -p vaultValues=\"${vault_values}\""

# Determine which set of repositories to use, based on whether or not we're dealing with open source
#<<<<<<< upstreamscoping
if [[ -z ${GH_TOKEN} && "${github}" != "github.com" ]]; then
=======
if [[ -z ${GH_TOKEN} && "${is_public_repo}" != "true" ]]; then
#>>>>>>> upstreamscoping-master
    cat ~/.ssh/known_hosts | base64 ${base64_arg} > ${TMP}/known_hosts
    set +x
    helper_text="If this step fails, make the second positional arg the path to an ssh key authenticated with Github Enterprise
    
    ex: bash -x bootstrap.sh fybrik-system /path/to/private/ssh/key
    "
    set -x
#<<<<<<< upstreamscoping
    oc create secret generic git-ssh-key --from-file=ssh-privatekey=${ssh_key} --type=kubernetes.io/ssh-auth
    helper_text=""
    oc annotate secret git-ssh-key --overwrite 'tekton.dev/git-0'="${github}"
=======
    kubectl create secret generic git-ssh-key --from-file=ssh-privatekey=${ssh_key} --type=kubernetes.io/ssh-auth
    helper_text=""
    kubectl annotate secret git-ssh-key --overwrite 'tekton.dev/git-0'="${github}"
#>>>>>>> upstreamscoping-master
    if [[ ${is_openshift} == "true" ]]; then
        oc secrets link pipeline git-ssh-key --for=mount
        set +e
        oc secrets unlink pipeline git-token
    else
        kubectl patch serviceaccount default -p '{"secrets": [{"name": "git-ssh-key"}]}'
        set +e
#<<<<<<< upstreamscoping
#        kubectl patch serviceaccount default --type=json -p='[{"op": "remove", "path": "/data/mykey"}]'
#        kubectl patch deploy/some-deployment --type=json -p='[{"op": "remove", "path": "/spec/template/spec/containers/0/ports/0"},{"op": "remove", "path": "/spec/template/spec/containers/0/ports/2"}]
#        oc get sa default -o yaml | grep -A3 "secrets:" | awk '/git-token/ { print NR }' 
    fi
    set -e
elif [[ ! -z ${GH_TOKEN} && "${github}" != "github.com" ]]; then
=======
    fi
    set -e
elif [[ ! -z ${GH_TOKEN} && "${is_public_repo}" != "true" ]]; then
#>>>>>>> upstreamscoping-master
    cat > ${TMP}/git-token.yaml <<EOH
apiVersion: v1
kind: Secret
metadata:
  name: git-token
  annotations:
    tekton.dev/git-0: https://${github} # Described below
type: kubernetes.io/basic-auth
stringData:
  username: ${git_user}
  password: ${GH_TOKEN}
EOH
#<<<<<<< upstreamscoping
    oc apply -f ${TMP}/git-token.yaml
=======
    kubectl apply -f ${TMP}/git-token.yaml
#>>>>>>> upstreamscoping-master
    if [[ ${is_openshift} == "true" ]]; then
        oc secrets link pipeline git-token --for=mount
        set +e
        oc secrets unlink pipeline git-ssh-key
    else
        kubectl patch serviceaccount default -p '{"secrets": [{"name": "git-token"}]}'
        set +e
    fi
    set -e
fi
extra_params="${extra_params} -p git-url=${git_url}"
#<<<<<<< upstreamscoping
if [[ "${github}" != "github.com" ]]; then
    extra_params="${extra_params} -p wkc-connector-git-url=${wkc_connector_git_url} -p vault-plugin-secrets-wkc-reader-url=${vault_plugin_secrets_wkc_reader_url}"
fi

# Set up credentials for WKC
if [[ "${github}" != "github.com" ]]; then
    cat > ${TMP}/wkc-credentials.yaml <<EOH
apiVersion: v1
kind: Secret
metadata:
  name: wkc-credentials
  namespace: ${unique_prefix}
type: kubernetes.io/Opaque
stringData:
  CP4D_USERNAME: ${cpd_username}
  CP4D_PASSWORD: ${cpd_password}
  WKC_username: ${cpd_username}
  WKC_password: ${cpd_password}
  WKC_USERNAME: ${cpd_username}
  WKC_PASSWORD: ${cpd_password}
  CP4D_SERVER_URL: ${cpd_url}
EOH
    cat ${TMP}/wkc-credentials.yaml
    oc apply -f ${TMP}/wkc-credentials.yaml

  extra_params="${extra_params} -p wkcConnectorServerUrl=https://cpd-cpd4.apps.cpstreamsx4.cp.fyre.ibm.com"

  if [[ ${cluster_scoped} == "false" && ${use_application_namespace} == "true" ]]; then 
    cat > ${TMP}/wkc-credentials.yaml <<EOH
apiVersion: v1
kind: Secret
metadata:
  name: wkc-credentials
  namespace: ${unique_prefix}-app
type: kubernetes.io/Opaque
stringData:
  CP4D_USERNAME: ${cpd_username}
  CP4D_PASSWORD: ${cpd_password}
  WKC_username: ${cpd_username}
  WKC_password: ${cpd_password}
  WKC_USERNAME: ${cpd_username}
  WKC_PASSWORD: ${cpd_password}
  CP4D_SERVER_URL: ${cpd_url}
EOH
    cat ${TMP}/wkc-credentials.yaml
    oc apply -f ${TMP}/wkc-credentials.yaml
    extra_params="${extra_params} -p wkcConnectorServerUrl=${cpd_url}"
  fi
fi


# Determine whether images should be sent to ICR for security scanning if creds exist
set +e
oc get secret us-south-creds
=======

if [[ "${is_public_repo}" != "true" ]]; then
    if [[ -f ${repo_root}/pipeline/custom_repo_references.sh ]]; then
        source ${repo_root}/pipeline/custom_repo_references.sh
    fi
fi

# Determine whether images should be sent to ICR for security scanning if creds exist
set +e
kubectl get secret us-south-creds
#>>>>>>> upstreamscoping-master
rc=$?
transfer_images_to_icr=false
if [[ $rc -eq 0 ]]; then
    transfer_images_to_icr=true
fi
extra_params="${extra_params} -p transfer-images-to-icr=${transfer_images_to_icr}"

# If a github_workspace was specified, don't clone the code, copy it to volume from the local host
set -e
if [[ ! -z "${github_workspace}" ]]; then
    kubectl describe pvc
    try_command "kubectl wait pod workspace-0 --for=condition=Ready --timeout=1m" 15 false 5
    ls ${github_workspace}
    ls ${github_workspace}/..
    if [[ ${is_kubernetes} == "true" ]]; then
        kubectl cp $github_workspace workspace-0:/workspace/source/
    else 
        oc rsync $github_workspace workspace-0:/workspace/source/
    fi
    git_url=""
    extra_params="${extra_params} -p git-url="
fi
set +x

echo "
# for a pre-existing PVC that will be deleted when the namespace is deleted
#<<<<<<< upstreamscoping
tkn pipeline start build-and-deploy -w name=images-url,emptyDir=\"\" -w name=artifacts,claimName=artifacts-pvc -w name=shared-workspace,claimName=source-pvc -p docker-hostname=${image_repo} -p dockerhub-hostname=${dockerhub_hostname} -p docker-namespace=${unique_prefix} -p NAMESPACE=${unique_prefix} -p skipTests=${skip_tests} -p fybrik-values=${fybrik_values} ${extra_params} -p git-revision=pipeline"

if [[ ${run_tkn} -eq 1 ]]; then
    set -x

    cat > ${TMP}/pipelinerun.yaml <<EOH
=======
tkn pipeline start build-and-deploy -w name=images-url,emptyDir=\"\" -w name=artifacts,claimName=artifacts-pvc -w name=shared-workspace,claimName=source-pvc -p docker-hostname=${image_repo} -p dockerhub-hostname=${dockerhub_hostname} -p docker-namespace=${unique_prefix} -p NAMESPACE=${unique_prefix} -p skipTests=${skip_tests} ${extra_params} -p git-revision=pipeline"

if [[ ${run_tkn} -eq 1 ]]; then
    set -x
    if [[ ${is_public_repo} == "true" ]]; then
        cat > ${TMP}/pipelinerun.yaml <<EOH
#>>>>>>> upstreamscoping-master
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  labels:
    tekton.dev/pipeline: build-and-deploy
  name: build-and-deploy-run
  namespace: ${unique_prefix} 
spec:
  params:
  - name: NAMESPACE
    value: ${unique_prefix} 
  - name: docker-hostname
    value: ${image_repo}
  - name: dockerhub-hostname
    value: ${dockerhub_hostname}
#<<<<<<< upstreamscoping
  - name: blueprintNamespace
    value: ${blueprint_namespace}
=======
#>>>>>>> upstreamscoping-master
  - name: docker-namespace
    value: ${unique_prefix} 
  - name: git-revision
    value: pipeline
  - name: wkcConnectorServerUrl
    value: ${cpd_url}
  - name: git-url
    value: "${git_url}"
#<<<<<<< upstreamscoping
  - name: wkc-connector-git-url
    value: "${wkc_connector_git_url}" 
  - name: vault-plugin-secrets-wkc-reader-url 
    value: "${vault_plugin_secrets_wkc_reader_url}"
=======
#>>>>>>> upstreamscoping-master
  - name: skipTests
    value: "${skip_tests}"
  - name: transfer-images-to-icr
    value: "${transfer_images_to_icr}"
  - name: clusterScoped
    value: "${cluster_scoped}"
  - name: deployVault
    value: "${deploy_vault}"
  - name: deployCRD
    value: "${deploy_crd}"
  - name: build_image
    value: "${build_image}"
  - name: helm_image
    value: "${helm_image}"
  - name: deployCertManager
    value: "${deploy_cert_manager}"
  - name: vaultValues
    value: "${vault_values}"
#<<<<<<< upstreamscoping
  - name: fybrik-values
    value: "${fybrik_values}"
=======
#>>>>>>> upstreamscoping-master
  pipelineRef:
    name: build-and-deploy
  serviceAccountName: ${pipeline_sa}
  timeout: 1h0m0s
  workspaces:
  - emptyDir: {}
    name: images-url
  - name: artifacts
    persistentVolumeClaim:
      claimName: artifacts-pvc
  - name: shared-workspace
    persistentVolumeClaim:
      claimName: source-pvc
EOH
#<<<<<<< upstreamscoping
    cat ${TMP}/pipelinerun.yaml
    oc apply -f ${TMP}/pipelinerun.yaml
 
    cat > ${TMP}/streams_csv_check_script.sh <<EOH
#!/bin/bash
set -x
oc get taskrun,pvc,po
for i in $(oc get taskrun --no-headers | grep "False" | cut -d' ' -f1); do oc logs -l tekton.dev/taskRun=$i --all-containers; done
oc get pipelinerun --no-headers
oc get pipelinerun --no-headers | grep -e "Failed" -e "Completed"
EOH
    chmod u+x ${TMP}/streams_csv_check_script.sh
    try_command "${TMP}/streams_csv_check_script.sh"  40 false 30
    echo "debug: pods"
    oc describe pods
    echo "debug: events"
    oc get events
    echo "debug: taskruns"
    for i in $(oc get taskrun --no-headers | grep "False" | cut -d' ' -f1); do oc logs -l tekton.dev/taskRun=$i --all-containers; done
    set +e
    oc get pipelinerun -o yaml | grep "Completed"
=======
        cat ${TMP}/pipelinerun.yaml
        kubectl apply -f ${TMP}/pipelinerun.yaml
    else
         if [[ -f ${repo_root}/pipeline/custom_run_tkn.sh ]]; then
             ${repo_root}/pipeline/custom_run_tkn.sh
         else
             set +x
             echo "If run_tkn is on, please put a script in ${repo_root}/pipeline/custom_run_tkn.sh to define the custom pipelinerun"
             exit 1
         fi
    fi

    cat > ${TMP}/streams_csv_check_script.sh <<EOH
#!/bin/bash
set -x
kubectl get taskrun,pvc,po
for i in $(kubectl get taskrun --no-headers | grep "False" | cut -d' ' -f1); do kubectl logs -l tekton.dev/taskRun=$i --all-containers; done
kubectl get pipelinerun --no-headers
kubectl get pipelinerun --no-headers | grep -e "Failed" -e "Completed"
EOH
    chmod u+x ${TMP}/streams_csv_check_script.sh
    try_command "${TMP}/streams_csv_check_script.sh"  60 false 30
    echo "debug: pods"
    kubectl describe pods
    echo "debug: events"
    kubectl get events
    echo "debug: taskruns"
    for i in $(kubectl get taskrun --no-headers | grep "False" | cut -d' ' -f1); do kubectl logs $(kubectl get po -l tekton.dev/taskRun=$i --no-headers | cut -d' ' -f1) --all-containers --since=0s; done
    set +e
    kubectl get pipelinerun --no-headers | grep "True"
#>>>>>>> upstreamscoping-master
    rc=$?
    if [[ $rc -ne 0 ]]; then
        exit $rc
    fi
    set -e
fi
