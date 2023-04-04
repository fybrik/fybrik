#!/usr/bin/env bash

set -e

# Tools versions, they are updated automatically from requirements.env by running 'make reconcile-requirements' from main Makefile
HELM_VERSION=v3.10.3
YQ_VERSION=4.6.0
KUBE_VERSION=1.24.1
KIND_VERSION=0.17.0
CERT_MANAGER_VERSION=v1.6.2
AWSCLI_VERSION=2.7.18

if [[ "$FYBRIK_VERSION" == "" ]]; then
  # Get Fybrik lateset realease from github
  FYBRIK_VERSION=$(git -c 'versionsort.suffix=-' ls-remote --tags --sort='v:refname' https://github.com/fybrik/fybrik.git | tail --lines=1 | cut --delimiter='/' --fields=3)
fi
FYBRIK_VERSION_VAULT="$FYBRIK_VERSION"
if [[ "$FYBRIK_VERSION" == "master" ]]; then
  FYBRIK_VERSION=""
else
  FYBRIK_VERSION="--version ${FYBRIK_VERSION:1}"
fi

# Arguments handeling:
usage() {
    echo "Usage: $0 This is how to use the installer:"
    echo "Flags:
          l - (optional) tools location - default: current dir.
          m - (optional) module name - default: afm. options: afm (arrow-flight-module), abm (airbyte-module).
          c - use kind or current cluster - default - use kind."
}

exit_abnormal() {
    usage
    exit 1
}

# Default arguments:
TOOLS=.
MODULE=arrow-flight-module
USE_KIND=1

# Flags:
# l - (optional) tools location - default: current dir
# m - (optional) module name - default: afm. options: afm (arrow-flight-module), abm (airbyte-module).
# c - use kind or current cluster - default: use kind

while getopts ":l:m:c" arg; do
    case "${arg}" in
        l)
            TOOLS=${OPTARG}
            if ! [[ -d ${TOOLS} ]]; then
                echo "Couldn't find directory ${TOOLS}. creating a new directory"
                if ! mkdir ${TOOLS}; then
                echo "Error: Couldn't create a new directory ${TOOLS}"
                exit_abnormal
                fi
            elif ! [[ -r ${TOOLS} && -w ${TOOLS} ]]; then
                echo "Error: Don't have read/write premissions to ${TOOLS}"
                exit_abnormal
            fi
            ;;
        m)
            MODULE=${OPTARG}
            if [[ "${MODULE}" = "afm" || "${MODULE}" = "arrow-flight-module" ]]; then
                MODULE=arrow-flight-module
            elif [[ "${MODULE}" = "abm" || "${MODULE}" = "airbyte-module" ]]; then
                MODULE=airbyte-module
            else
                echo "Error: Module \"${MODULE}\" is not recognizable"
                exit_abnormal
            fi
            ;;
        c)
            USE_KIND=0
            ;;
        :)
            echo "Error: -${OPTARG} requires an argument."
            exit_abnormal   
            ;;
        ?)
            exit_abnormal
            ;;
    esac
done


# OS Check:
arch=amd64
os="unknown"

if [[ "$OSTYPE" == "linux-gnu" ]]; then
  os="linux"
else 
  echo "OS '$OSTYPE' is not yet supported. MacOS support coming soon. Aborting." >&2
  exit 1
fi

# Helper functions:
header_color=$'\e[1;32m'
sub_header_color=$'\e[1;35m'
reset_color=$'\033[0m'

function header {
  echo -e "$header_color$*$reset_color"
}
function sub_header {
  echo -e "$sub_header_color$*$reset_color"
}


header "Installing Tools:"
cd ${TOOLS}
PATH=$(pwd)/bin:${PATH}
mkdir -p bin

header "Checking for bin/helm ${HELM_VERSION}"
if [[ -f bin/helm &&  `bin/helm version --template='{{.Version}}'` == ${HELM_VERSION} ]];
then 
    header "  bin/helm ${HELM_VERSION} already exists"
else
    header "Installing bin/helm ${HELM_VERSION}"
    curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3
    chmod 700 get_helm.sh
    HELM_INSTALL_DIR=bin ./get_helm.sh -v ${HELM_VERSION} --no-sudo
    rm -rf get_helm.sh
fi

header "Checking for bin/yq ${YQ_VERSION}"
if [[ -f bin/yq && `bin/yq --version | awk '{print $3}'` == ${YQ_VERSION} ]]
then 
    header "  bin/yq ${YQ_VERSION} already exists"
else
    header "Installing bin/yq ${YQ_VERSION}"
    curl -L https://github.com/mikefarah/yq/releases/download/v${YQ_VERSION}/yq_${os}_${arch} -o bin/yq
    chmod +x bin/yq
fi

header "Checking for bin/kubectl ${KUBE_VERSION}"
if [[ -f bin/kubectl && `bin/kubectl version -o=yaml 2> /dev/null | bin/yq e '.clientVersion.gitVersion' -` == "v${KUBE_VERSION}" ]];
then
    header "  bin/kubectl ${KUBE_VERSION} already exists"
else
    curl -LO https://dl.k8s.io/release/v${KUBE_VERSION}/bin/${os}/${arch}/kubectl
    chmod +x ./kubectl
    mv ./kubectl bin/kubectl
fi

if [ ${USE_KIND} -eq 1 ]; then
    header "Checking for bin/kind ${KIND_VERSION}"
    if [[ -f bin/kind && `bin/kind --version | awk '{print $3}'` == ${KIND_VERSION} ]]
    then
        header "  bin/kind ${KIND_VERSION} already exists"
    else 
        header "Installing bin/kind ${KIND_VERSION}"
        curl -Lo ./kind https://kind.sigs.k8s.io/dl/v${KIND_VERSION}/kind-${os}-${arch}
        chmod +x ./kind
        mv ./kind bin/kind
    fi
fi

header "Checking for bin/7zzs"
if [[ -f bin/7zzs ]]
then
    header "  7z already exists"
else 
    header "Installing bin/7zzs"
    mkdir -p 7z-install
    curl -L https://www.7-zip.org/a/7z2201-linux-x64.tar.xz -o 7z-install/7z.tar.xz
    tar -xf 7z-install/7z.tar.xz -C 7z-install
    chmod u+x 7z-install/7zzs
    mv 7z-install/7zzs ./bin
    rm -r 7z-install
fi

header "Checking for aws-cli v2"
if [[ -f bin/aws && -d bin/aws-source/v2 ]]
then
    header "  bin/aws v2 already exists"
else
    header "Installing bin/aws ${AWSCLI_VERSION}"
    # Installed this way due to a known open bug: https://github.com/aws/aws-cli/issues/6852
    mkdir -p awscli-install
    curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64-${AWSCLI_VERSION}.zip" -o awscli-install/awscliv2.zip
    ./bin/7zzs x awscli-install/awscliv2.zip -oawscli-install
    ./awscli-install/aws/install -i bin/aws-source -b bin
    rm bin/aws bin/aws_completer
    ln -s aws-source/v2/${AWSCLI_VERSION}/bin/aws ./bin/aws
    rm -r ./awscli-install
fi


if [ ${USE_KIND} -eq 1 ]; then
    header "\nCreate kind cluster"
    cluster_name=kind-fybrik-installation-sample
    kubernetesVersion=$(bin/kubectl version -o=yaml | bin/yq e '.clientVersion.minor' -)
    if ((${kubernetesVersion} <= 21)); then
      echo "Error: kubernetes version ${kubernetesVersion} may not be supported. please choose a version greater than 21"
      exit_abnormal
    fi
    bin/kind delete clusters ${cluster_name}
    bin/kind create cluster --name=${cluster_name} --image=kindest/node:v1.${kubernetesVersion}.0
fi

header "\nUpdating helm repositories"
bin/helm repo add jetstack https://charts.jetstack.io
bin/helm repo add hashicorp https://helm.releases.hashicorp.com
bin/helm repo add fybrik-charts https://fybrik.github.io/charts
bin/helm repo update

header "\nInstall Cert-manager"
bin/helm install cert-manager jetstack/cert-manager \
    --namespace cert-manager \
    --version ${CERT_MANAGER_VERSION} \
    --create-namespace \
    --set installCRDs=true \
    --wait --timeout 600s

header "\nInstall Vault"
bin/helm install vault fybrik-charts/vault --create-namespace -n fybrik-system \
    --set "vault.injector.enabled=false" \
    --set "vault.server.dev.enabled=true" \
    --values https://raw.githubusercontent.com/fybrik/fybrik/${FYBRIK_VERSION_VAULT}/charts/vault/env/dev/vault-single-cluster-values.yaml
bin/kubectl wait --for=condition=ready --all pod -n fybrik-system --timeout=120s

# If using openshift:
# bin/helm install vault fybrik-charts/vault --create-namespace -n fybrik-system \
#     --set "vault.global.openshift=true" \
#     --set "vault.injector.enabled=false" \
#     --set "vault.server.dev.enabled=true" \
#     --values https://raw.githubusercontent.com/fybrik/fybrik/v${FYBRIK_VERSION}/charts/vault/env/dev/vault-single-cluster-values.yaml
# bin/kubectl wait --for=condition=ready --all pod -n fybrik-system --timeout=120s

header "\nInstall OpenMetaData"
export FYBRIK_BRANCH= ${FYBRIK_VERSION}
curl https://raw.githubusercontent.com/fybrik/fybrik/{FYBRIK_BRANCH}/third_party/openmetadata/install_OM.sh | bash -



header "\nInstall control plane"

bin/helm install fybrik-crd fybrik-charts/fybrik-crd -n fybrik-system  ${FYBRIK_VERSION} --wait

bin/helm install fybrik fybrik-charts/fybrik --set coordinator.catalog=openmetadata --set openmetadataConnector.openmetadata_endpoint=http://openmetadata.open-metadata:8585/api -n fybrik-system ${FYBRIK_VERSION} --wait # --values=../values.yaml --wait
sleep 5

header "\nInstall module"
bin/kubectl apply -f https://github.com/fybrik/${MODULE}/releases/latest/download/module.yaml -n fybrik-system

sleep 5

# END OF PART 1 OF THE QUICKSTART
# STARTING READ SAMPLE:

header "\nCreate notebook sample"
bin/kubectl create namespace fybrik-notebook-sample
bin/kubectl config set-context --current --namespace=fybrik-notebook-sample

header "\nInstall localstack"
bin/helm repo add localstack-charts https://localstack.github.io/helm-charts
bin/helm install localstack localstack-charts/localstack --version 0.4.3 --set startServices="s3" --set service.type=ClusterIP --set image.tag="1.2.0" --set livenessProbe.initialDelaySeconds=25
bin/kubectl wait --for=condition=ready --all pod -n fybrik-notebook-sample --timeout=600s
bin/kubectl port-forward svc/localstack 4566:4566 &

header "\nUpload sample dataset to localstack"
curl -L https://raw.githubusercontent.com/fybrik/fybrik/master/samples/notebook/PS_20174392719_1491204439457_log.csv -o sample.csv

export ACCESS_KEY=1234
export SECRET_KEY=1234
export ENDPOINT="http://127.0.0.1:4566"
export BUCKET="demo"
export OBJECT_KEY="sample.csv"
export REGION=theshire
bin/aws configure set aws_access_key_id ${ACCESS_KEY}
bin/aws configure set aws_secret_access_key ${SECRET_KEY}
bin/aws configure set region ${REGION}
bin/aws --endpoint-url=${ENDPOINT} s3api create-bucket --bucket ${BUCKET} --region ${REGION} --create-bucket-configuration LocationConstraint=${REGION}
bin/aws --endpoint-url=${ENDPOINT} s3api put-object --bucket ${BUCKET} --key ${OBJECT_KEY} --body sample.csv
rm sample.csv

header "\nRegister the dataset in a data catalog"
cat << EOF | bin/kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: paysim-csv
type: Opaque
stringData:
  access_key: "${ACCESS_KEY}"
  secret_key: "${SECRET_KEY}"
EOF

bin/kubectl port-forward svc/openmetadata-connector -n fybrik-system 8081:8080 &

sleep 5

cat << EOF | curl -X POST localhost:8081/createAsset -d @-
{
  "destinationCatalogID": "openmetadata",
  "destinationAssetID": "paysim-csv",
  "credentials": "/v1/kubernetes-secrets/paysim-csv?namespace=fybrik-notebook-sample",
  "details": {
    "dataFormat": "csv",
    "connection": {
      "name": "s3",
      "s3": {
        "endpoint": "http://localstack.fybrik-notebook-sample.svc.cluster.local:4566",
        "bucket": "demo",
        "object_key": "sample.csv"
      }
    }
  },
  "resourceMetadata": {
    "name": "Synthetic Financial Datasets For Fraud Detection",
    "geography": "theshire ",
    "tags": {
      "Purpose.finance": "true"
    },
    "columns": [
      {
        "name": "nameOrig",
        "tags": {
          "PII.Sensitive": "true"
        }
      },
      {
        "name": "oldbalanceOrg",
        "tags": {
          "PII.Sensitive": "true"
        }
      },
      {
        "name": "newbalanceOrig",
        "tags": {
          "PII.Sensitive": "true"
        }
      }
    ]
  }
}
EOF

export CATALOGED_ASSET="openmetadata-s3.default.demo.\"sample.csv\""


header "\nDefine data access policies"
cat << EOF > sample-policy.rego
package dataapi.authz
rule[{"action": {"name":"RedactAction", "columns": column_names}, "policy": description}] {
  description := "Redact columns tagged as PII.Sensitive in datasets tagged with Purpose.finance = true"
  input.action.actionType == "read"
  input.resource.metadata.tags["Purpose.finance"]
  column_names := [input.resource.metadata.columns[i].name | input.resource.metadata.columns[i].tags["PII.Sensitive"]]
  count(column_names) > 0
}

EOF

bin/kubectl -n fybrik-system create configmap sample-policy --from-file=sample-policy.rego
bin/kubectl -n fybrik-system label configmap sample-policy openpolicyagent.org/policy=rego
while [[ $(bin/kubectl get cm sample-policy -n fybrik-system -o 'jsonpath={.metadata.annotations.openpolicyagent\.org/policy-status}') != '{"status":"ok"}' ]]; do echo "waiting for policy to be applied" && sleep 5; done
rm sample-policy.rego

header "\nCreate a FybrikApplication resource for the notebook"
cat <<EOF | kubectl apply -f -
apiVersion: app.fybrik.io/v1beta1
kind: FybrikApplication
metadata:
  name: my-notebook
  labels:
    app: my-notebook
spec:
  selector:
    workloadSelector:
      matchLabels:
        app: my-notebook
  appInfo:
    intent: Fraud Detection
  data:
    - dataSetID: ${CATALOGED_ASSET}
      requirements:
        interface: 
          protocol: fybrik-arrow-flight
EOF


while [[ $(kubectl get fybrikapplication my-notebook -o 'jsonpath={.status.ready}') != "true" ]]; do echo "waiting for FybrikApplication" && sleep 5; done
export CATALOGED_ASSET_MODIFIED=$(echo $CATALOGED_ASSET | sed 's/\./\\\./g')
while [[ $(kubectl get fybrikapplication my-notebook -o "jsonpath={.status.assetStates.${CATALOGED_ASSET_MODIFIED}.conditions[?(@.type == 'Ready')].status}") != "True" ]]; do echo "waiting for ${CATALOGED_ASSET} asset" && sleep 5; done



header "\nRead the dataset from the notebook"

export ENDPOINT_SCHEME=$(kubectl get fybrikapplication my-notebook -o "jsonpath={.status.assetStates.${CATALOGED_ASSET_MODIFIED}.endpoint.fybrik-arrow-flight.scheme}")
export ENDPOINT_HOSTNAME=$(kubectl get fybrikapplication my-notebook -o "jsonpath={.status.assetStates.${CATALOGED_ASSET_MODIFIED}.endpoint.fybrik-arrow-flight.hostname}")
export ENDPOINT_PORT=$(kubectl get fybrikapplication my-notebook -o "jsonpath={.status.assetStates.${CATALOGED_ASSET_MODIFIED}.endpoint.fybrik-arrow-flight.port}")
export ENDPOINT_CONNECTION=${ENDPOINT_SCHEME}://${ENDPOINT_HOSTNAME}:${ENDPOINT_PORT}

cat << EOF > test.py
import json
import pyarrow.flight as fl
import pandas as pd

# Create a Flight client 
client = fl.connect('${ENDPOINT_CONNECTION}')

# Prepare the request
request = {
    "asset": "openmetadata-s3.default.demo.\"sample.csv\"",
    # To request specific columns add to the request a "columns" key with a list of column names
    "columns": ["type", "amount", "oldbalanceOrg", "isFraud"]
}

# Send request and fetch result as a pandas DataFrame
info = client.get_flight_info(fl.FlightDescriptor.for_command(json.dumps(request)))
reader: fl.FlightStreamReader = client.do_get(info.endpoints[0].ticket)
df: pd.DataFrame = reader.read_pandas()
print(df)
EOF


POD_NAME=$(bin/kubectl get pods -n fybrik-blueprints -o=name | sed "s/^.\{4\}//")
bin/kubectl cp ./test.py ${POD_NAME}:/tmp -n fybrik-blueprints
bin/kubectl exec -i ${POD_NAME} -n fybrik-blueprints -- python /tmp/test.py > res.out
rm test.py
cat res.out
header "\nFinished successfully"