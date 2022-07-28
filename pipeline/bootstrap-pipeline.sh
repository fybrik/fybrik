#!/bin/bash
set -x
set +e

export cluster_scoped=${cluster_scoped:-false}
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
export use_application_namespace=${use_application_namespace:-false}
export git_url="${git_url:-https://github.com/fybrik/fybrik.git}"
export is_kind="${is_kind:-false}"

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

# Figure out if we're using air-gapped machines that should pull images from somewhere other than dockerhub
extra_params="-p clusterScoped=${cluster_scoped}"
is_public_repo="false"
is_custom_repo="false"
helm_image=
build_image=
if [[ "${github}" == "github.com" ]]; then
    is_public_repo="true"
    build_image="docker.io/yakinikku/suede_compile:latest"
    helm_image="docker.io/lachlanevenson/k8s-helm:latest"
    extra_params="${extra_params} -p build_image=${build_image} -p helm_image=${helm_image}"
    cp ${repo_root}/pipeline/statefulset.yaml ${TMP}/
else
    is_custom_repo="true"
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
        helm repo add nfs-subdir-external-provisioner https://kubernetes-sigs.github.io/nfs-subdir-external-provisioner/
        helm upgrade --install nfs-provisioner nfs-subdir-external-provisioner/nfs-subdir-external-provisioner --values ${repo_root}/pipeline/nfs-values.yaml --set nfs.server=${ip} --namespace nfs-provisioner --create-namespace
    fi
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

if [[ "${unique_prefix}" == "fybrik-system" ]]; then
  modules_namespace="fybrik-blueprints"
else
  modules_namespace="${unique_prefix}-blueprints"
fi
extra_params="${extra_params} -p modulesNamespace=${modules_namespace}"

if [[ ${cluster_scoped} == "false" ]]; then
  set +e
  rc=1
  kubectl get ns ${modules_namespace}
  rc=$?
  set -e
  # Create new project if necessary
  if [[ $rc -ne 0 ]]; then
    if [[ ${is_openshift} == "true" ]]; then
      oc new-project ${modules_namespace}
      oc project ${unique_prefix} 
    else
      kubectl create ns ${modules_namespace} 
    fi
  fi
fi

if [[ -f ${repo_root}/pipeline/custom_fybrik_values.sh ]]; then
    source ${repo_root}/pipeline/custom_fybrik_values.sh
else
    fybrik_values=""
fi

if [[ ${cluster_scoped} == "false" ]]; then
    if [[ ${use_application_namespace} == "false" ]]; then
        if [[ ! -z ${fybrik_values} ]]; then
            fybrik_values="${fybrik_values},applicationNamespace=${unique_prefix}"
        else
            fybrik_values="applicationNamespace=${unique_prefix}"
        fi
    else
        if [[ ! -z ${fybrik_values} ]]; then
            fybrik_values="${fybrik_values},applicationNamespace=${unique_prefix}-app"
        else
            fybrik_values="applicationNamespace=${unique_prefix}-app"
        fi
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
fi

set +e
# Be smarter about this - just a quick hack for typical default OpenShift & Kind installs so we can control the default storage class
kubectl patch storageclass managed-nfs-storage -p '{"metadata": {"annotations": {"storageclass.kubernetes.io/is-default-class": "true"}}}'
kubectl patch storageclass standard -p '{"metadata": {"annotations": {"storageclass.kubernetes.io/is-default-class": "false"}}}'
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
    chmod u+x ${TMP}/streams_csv_check_script.sh
    try_command "${TMP}/streams_csv_check_script.sh"  40 true 5

    cat > ${TMP}/streams_csv_check_script.sh <<EOH
#!/bin/bash
set -x
oc get -n openshift-operators csv | grep serverless-operator
oc get -n openshift-operators csv | grep serverless-operator | grep -e Succeeded -e Replacing
EOH
    chmod u+x ${TMP}/streams_csv_check_script.sh
    try_command "${TMP}/streams_csv_check_script.sh"  40 false 5
    oc apply -f ${repo_root}/pipeline/knative-eventing.yaml
else
    kubectl apply -f https://storage.googleapis.com/tekton-releases/pipeline/previous/v0.32.1/release.yaml
    set +e
    kubectl apply -f https://github.com/knative/operator/releases/download/knative-v1.2.0/operator.yaml
    kubectl apply -f https://github.com/knative/eventing/releases/download/knative-v1.2.0/eventing-crds.yaml
    kubectl apply -f https://github.com/knative/eventing/releases/download/knative-v1.2.0/eventing-core.yaml
    set -e
    try_command "kubectl wait pod -n tekton-pipelines --all --for=condition=Ready --timeout=3m" 2 true 1
    set +e
    kubectl create ns knative-eventing
    cat > ${TMP}/knative-eventing.yaml <<EOH
apiVersion: operator.knative.dev/v12
kind: KnativeEventing
metadata:
  name: knative-eventing
  namespace: knative-eventing
EOH
    ls -alrt ${TMP}/
    kubectl apply -f ${TMP}/knative-eventing.yaml
    set -e
fi

if [[ ${is_openshift} == "true" ]]; then
    # Give extended privileges to the pipeline service account, make sure images can be pulled from the integrated openshift registry
    oc adm policy add-cluster-role-to-user cluster-admin system:serviceaccount:${unique_prefix}:pipeline
    oc adm policy add-cluster-role-to-user cluster-admin system:serviceaccount:${unique_prefix}:root-sa
    oc adm policy add-role-to-group system:image-puller system:serviceaccounts:${unique_prefix} --namespace ${unique_prefix}
    oc adm policy add-role-to-group system:image-puller system:serviceaccounts:${modules_namespace} --namespace ${unique_prefix}
    oc adm policy add-role-to-group system:image-puller system:serviceaccounts:${unique_prefix}-app --namespace ${unique_prefix}
    
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
try_command "kubectl apply -f ${repo_root}/pipeline/tasks/shell.yaml" 3 true 60
try_command "kubectl apply -f ${repo_root}/pipeline/tasks/make.yaml" 3 true 60
try_command "kubectl apply -f ${repo_root}/pipeline/tasks/git-clone.yaml" 3 true 60
try_command "kubectl apply -f ${repo_root}/pipeline/tasks/buildah.yaml" 3 true 60
try_command "kubectl apply -f ${repo_root}/pipeline/tasks/skopeo-copy.yaml" 3 true 60
try_command "kubectl apply -f ${repo_root}/pipeline/tasks/openshift-client.yaml" 3 true 60
try_command "kubectl apply -f ${repo_root}/pipeline/tasks/helm-upgrade-from-source.yaml" 3 true 60
try_command "kubectl apply -f ${repo_root}/pipeline/tasks/helm-upgrade-from-repo.yaml" 3 true 60
helper_text=""

# Wipe old pipeline definitions in case of merge conflicts
set +e
kubectl delete -f ${repo_root}/pipeline/pipeline.yaml
if [[ -f ${repo_root}/pipeline/custom_pipeline_cleanup.sh ]]; then
    source ${repo_root}/pipeline/custom_pipeline_cleanup.sh
fi

set -e
if [[ "${is_public_repo}" == "true" ]]; then
    try_command "kubectl apply -f ${repo_root}/pipeline/pipeline.yaml" 3 true 60
else 
     if [[ -f ${repo_root}/pipeline/custom_pipeline_create.sh ]]; then
         source ${repo_root}/pipeline/custom_pipeline_create.sh
     else
         set +x
         echo "You are running with a non public repo, but no custom pipeline creation script exists at ${repo_root}/pipeline/custom_pipeline_create.sh"
         exit 1
     fi
fi

# Delete old registry credentials
set +e
kubectl delete secret -n ${unique_prefix} regcred --wait
kubectl delete secret -n ${unique_prefix} regcred-test --wait
kubectl delete secret -n ${unique_prefix} sourceregcred --wait
set -e
set -x

# See if we have a pull secret available on cluster that has access to authenticated registries we need
set +e
kubectl get secret -n openshift-config pull-secret -o yaml > ${TMP}/secret.yaml
rc=$?
if [[ ${rc} -eq 0 ]]; then
    helper_text=""
    kubectl get secret -n openshift-config pull-secret -o=go-template='{{index .data ".dockerconfigjson"}}' | base64 --decode | grep "${image_source_repo}"
    rc=$?
    if [[ ${rc} -eq 0 ]]; then
        set -e
        cp ${TMP}/secret.yaml ${TMP}/secret.yaml.orig
        sed -i.bak "s|namespace: openshift-config|namespace: ${unique_prefix}|g" ${TMP}/secret.yaml
        sed -i.bak "s|name: pull-secret|name: regcred|g" ${TMP}/secret.yaml
        cat ${TMP}/secret.yaml
        kubectl apply -f ${TMP}/secret.yaml
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

    if [[ ${cluster_scoped} == "false" ]]; then
      kubectl patch serviceaccount default -p '{"imagePullSecrets": [{"name": "regcred"}]}' -n ${modules_namespace}
      kubectl patch serviceaccount default -p '{"secrets": [{"name": "regcred"}]}' -n ${modules_namespace}
    fi
fi

extra_params="${extra_params} -p deployVault='true'"
deploy_vault="true"
set +e
kubectl get crd | grep "fybrikapplications.app.fybrik.io"
rc=$?
deploy_crd="false"
if [[ $rc -ne 0 ]]; then
    extra_params="${extra_params} -p deployCRD='true'"
    deploy_crd="true"
fi

# Don't attempt to reinstall certmanager if some form of it is already installed
kubectl get crd | grep "certmanager"
rc=$?
deploy_cert_manager="false"
if [[ $rc -ne 0 ]]; then
    extra_params="${extra_params} -p deployCertManager='true'"
    deploy_cert_manager="true"
fi

# Create a workspace to allow users to exec in and run arbitrary commands
kubectl apply -f ${repo_root}/pipeline/rootsa.yaml
kubectl apply -f ${TMP}/statefulset.yaml
kubectl apply -f ${repo_root}/pipeline/pvc.yaml
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

kubectl apply -f ${TMP}/release.yaml
kubectl apply -f ${TMP}/interceptors.yaml

# Delete old apiserversource
set +e
kubectl delete apiserversource generic-watcher
set -e

# Install triggers for rebuilds of specific tasks
try_command "kubectl apply -f ${repo_root}/pipeline/eventlistener/triggerbinding.yaml" 3 true 60
try_command "kubectl apply -f ${repo_root}/pipeline/eventlistener/triggertemplate.yaml" 3 true 60
try_command "kubectl apply -f ${repo_root}/pipeline/eventlistener/apiserversource.yaml" 3 true 60
kubectl apply -f ${repo_root}/pipeline/eventlistener/role.yaml
kubectl apply -f ${repo_root}/pipeline/eventlistener/serviceaccount.yaml

set +x
helper_text="If this step fails, run again - knative related pods may be restarting and unable to process the webhook
"
set -x
if [[ ${is_openshift} == "true" ]]; then
    try_command "kubectl apply -f ${repo_root}/pipeline/eventlistener/eventlistener.yaml" 3 true 60
else
    sed -i.bak "s|serviceAccountName: pipeline|serviceAccountName: default|g" ${repo_root}/pipeline/eventlistener/eventlistener.yaml
    kubectl apply -f ${repo_root}/pipeline/eventlistener/eventlistener.yaml
    mv ${repo_root}/pipeline/eventlistener/eventlistener.yaml.bak ${repo_root}/pipeline/eventlistener/eventlistener.yaml
fi

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
fi
extra_params="${extra_params} -p vaultValues=\"${vault_values}\""

# Determine which set of repositories to use, based on whether or not we're dealing with open source
if [[ -z ${GH_TOKEN} && "${is_public_repo}" != "true" ]]; then
    cat ~/.ssh/known_hosts | base64 ${base64_arg} > ${TMP}/known_hosts
    set +x
    helper_text="If this step fails, make the second positional arg the path to an ssh key authenticated with Github Enterprise
    
    ex: bash -x bootstrap.sh fybrik-system /path/to/private/ssh/key
    "
    set -x
    kubectl create secret generic git-ssh-key --from-file=ssh-privatekey=${ssh_key} --type=kubernetes.io/ssh-auth
    helper_text=""
    kubectl annotate secret git-ssh-key --overwrite 'tekton.dev/git-0'="${github}"
    if [[ ${is_openshift} == "true" ]]; then
        oc secrets link pipeline git-ssh-key --for=mount
        set +e
        oc secrets unlink pipeline git-token
    else
        kubectl patch serviceaccount default -p '{"secrets": [{"name": "git-ssh-key"}]}'
        set +e
    fi
    set -e
elif [[ ! -z ${GH_TOKEN} && "${is_public_repo}" != "true" ]]; then
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
    kubectl apply -f ${TMP}/git-token.yaml
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

if [[ "${is_public_repo}" != "true" ]]; then
    if [[ -f ${repo_root}/pipeline/custom_repo_references.sh ]]; then
        source ${repo_root}/pipeline/custom_repo_references.sh
    fi
fi

# Determine whether images should be sent to ICR for security scanning if creds exist
set +e
kubectl get secret us-south-creds
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

if [[ -f ${repo_root}/pipeline/custom_pre_tkn.sh ]]; then
    source ${repo_root}/pipeline/custom_pre_tkn.sh
fi
set +x

echo "
# for a pre-existing PVC that will be deleted when the namespace is deleted
tkn pipeline start build-and-deploy -w name=images-url,emptyDir=\"\" -w name=artifacts,claimName=artifacts-pvc -w name=shared-workspace,claimName=source-pvc -p docker-hostname=${image_repo} -p dockerhub-hostname=${dockerhub_hostname} -p docker-namespace=${unique_prefix} -p NAMESPACE=${unique_prefix} -p skipTests=${skip_tests} -p fybrik-values=${fybrik_values} ${extra_params} -p git-revision=pipeline"

if [[ ${run_tkn} -eq 1 ]]; then
    set -x
    if [[ ${is_public_repo} == "true" ]]; then
        cat > ${TMP}/pipelinerun.yaml <<EOH
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
  - name: modulesNamespace
    value: ${modules_namespace}
  - name: docker-namespace
    value: ${unique_prefix} 
  - name: git-revision
    value: pipeline
  - name: git-url
    value: "${git_url}"
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
  - name: fybrik-values
    value: "${fybrik_values}"
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
        cat ${TMP}/pipelinerun.yaml
        kubectl apply -f ${TMP}/pipelinerun.yaml
    else
         if [[ -f ${repo_root}/pipeline/custom_run_tkn.sh ]]; then
             source ${repo_root}/pipeline/custom_run_tkn.sh
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
    try_command "${TMP}/streams_csv_check_script.sh"  60 false 40
    echo "debug: pods"
    kubectl describe pods
    echo "debug: events"
    kubectl get events
    echo "debug: volumes"
    kubectl describe pvc
    for i in $(kubectl get po -n nfs-provisioner | cut -d' ' -f1); do kubectl logs -n nfs-provisioner $i --all-containers; done
    echo "debug: taskruns"
    for i in $(kubectl get taskrun --no-headers | grep -v "True" | cut -d' ' -f1); do kubectl logs $(kubectl get po -l tekton.dev/taskRun=$i --no-headers | cut -d' ' -f1) --all-containers --since=0s; done
    set +e
    kubectl get pipelinerun --no-headers | grep "True"
    rc=$?
    if [[ $rc -ne 0 ]]; then
        exit $rc
    fi
    set -e
fi
