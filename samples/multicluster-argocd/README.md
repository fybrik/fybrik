# Fybrik multi-cluster management using ArgoCD


[Argo CD](https://argo-cd.readthedocs.io/en/stable/) is a declarative, GitOps continuous delivery tool for Kubernetes.
This README contains information about how to deploy Fybrik with ArgoCD as its multi-cluster manager using and to apply FybrikApplication to test it.

## Argo CD in high level

Argo CD is an open-source continuous delivery (CD) tool designed to automate the deployment and lifecycle management of applications in Kubernetes clusters. It follows the GitOps methodology, which means it uses a Git repository as the source of truth for defining and managing the desired state of applications and infrastructure.

With Argo CD, you can declaratively define your application deployment specifications in a Kubernetes manifest file or a Helm chart and store them in a Git repository. Argo CD continuously monitors this repository for changes and automatically reconciles the desired state with the current state of your Kubernetes cluster. It ensures that your applications are always running as intended and provides a robust mechanism for managing updates and rollbacks.

Argo CD is built as a Kubernetes controller and Custom Resource Definition (CRD). Its fundamental CRD resource is referred to as an "Application." The Application resource encapsulates essential details, such as the Git repository information stored in the `source` field, and the cluster to which the resources are synchronized, identified by the `destination` field.

For example:

```bash
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: guestbook
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/argoproj/argocd-example-apps.git
    targetRevision: HEAD
    path: guestbook
  destination:
    server: https://kubernetes.default.svc
    namespace: guestbook
  ```

## Fybrik Multi cluster management in high level

The high-level flow of Fybrik's multi-cluster management can be summarized as follows:

Within the coordinator cluster, the Plotter controller generates a blueprint and transmits it to the remote cluster using the multi-cluster manager (currently implemented for local, razee).
In the remote cluster, the Blueprint controller handles the received blueprints and updates their status.
Back in the coordinator cluster, the Plotter controller retrieves the updated blueprint status.

In order to facilitate this flow with Argo CD, it is necessary to supply a GitHub repository during Fybrik deployment. This repository will store the Fybrik Blueprint, which needs to be synchronized with the clusters using Argo CD.


## Before you begin

- Fybrik pre-requisite: https://fybrik.io/v1.3/get-started/quickstart/
- Argo CD CLI v2.7 and above
- Github account and Github repository to store the Blueprints


## Setup multi-cluster environment

This README demonstrates the setup of Fybrik using Argo CD as cluster manager using two clusters.

For local testing `hack/tools/create_kind.sh` script can be used to setup two kind clusters:

- `kind-control`: the coordinator cluster.
- `kind-kind`: the remote cluster.

To do so please execute the command:

```bash
make kind-setup-multi
```

Most of the commands in this readme should be executed in the coordinator cluster thus we first set kubectl context by executing the following commands:
```bash
kubectl config set-context coordinator --cluster=kind-control --namespace=argocd --user=kind-control
kubectl config set-context remote --cluster=kind-kind --namespace=default --user=kind-kind
```

If you are using OpenShift cluster you should change `--cluster` and `--user` arguments above according
to the values in `~/.kube/config` configuration file.

Next, if you are using OpenShift clusters you should run the following command:
```bash
export WITH_OPENSHIFT=true
```

Then, run the following commands to deploy Argo CD on the coordinator cluster:

```bash
kubectl config use-context coordinator
make -C third_party/argocd deploy
make -C third_party/argocd deploy-wait
```
    ---
    > _NOTE:_ If you are using OpenShift cluster you will see that the deployment fails because OpenShift doesn't allow `privileged: true` value in `securityContext` field by default. Thus, you should add the service account of the module's deployment to the `privileged SCC` using the following command:
    ```bash
    oc adm policy add-scc-to-user privileged system:serviceaccount:fybrik-blueprints:<SERVICE_ACCOUNT_NAME>
    ```
    > Then, the deployment will restart the failed pods and the pods in `fybrik-blueprints` namespace should start successfully.
    ---


## Logging to Argo CD

To execute Argo CD CLI commands in this tutorial, run the following commands:

First, retrieve the admin password from the command line:
```bash
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
```

Then, port-forward the argocd service:

```bash
kubectl config use-context coordinator
kubectl port-forward service/argo-argocd-server -n argocd 8080:443 &
```

The API server can now be accessed using https://localhost:8080 using the username `admin` and the password extracted above.

Then logging using Argo CD CLI with the admin password retrieved above.

```bash
kubectl config use-context coordinator
argocd login localhost:8080 --insecure
```

For more information please refer to the Argo CD [getting started page](https://argo-cd.readthedocs.io/en/stable/getting_started/).

## Configure Argo CD deployment

The following steps needs to be executed to configure the argo CD server:

1. Add clusters:

The coordinator cluster is automatically registered in argo CD server. However its default name is `in-cluster`. In this tutorial we change it to `coordinator`. This can be done manually via the Argo CD GUI or by executing the following command:

```bash
kubectl config use-context coordinator
argocd cluster --insecure add kind-control --in-cluster --name coordinator -y
```

For the remote cluster, please execute the following command.

```bash
kubectl config use-context coordinator
argocd cluster add kind-kind --name remote --cluster-endpoint kube-public -y
```

Note that the `--cluster-endpoint` is a new option added in Argo CD v2.7. For older versions
please refer this [workaround](https://github.com/argoproj/argo-cd/issues/4204).

If you are using OpenShift cluster you should change the clusters names (`kind-control` and `kind-kind` in the commands above).
These values can be extracted from the `~/.kube/config` configuration file.

2. Add private repositories:

A github repository needs to be provided for the purpose of hosting the Fybrik Blueprints.
Run the following command to register a private repository in Argo CD:

```bash
kubectl config use-context coordinator
argocd repo add https://github.com/xxx/argocd-fybrik-blueprints --name my-blueprints --username username --password xxx
```

Alternatively, the repository can be added using the Argo CD GUI.

3. Reduce sync interval

The automatic sync interval is determined by the timeout.reconciliation value in the [argocd-cm ConfigMap](https://argo-cd.readthedocs.io/en/stable/user-guide/auto_sync/), which defaults to 180s (3 minutes).
We recommend to reduce it to 10s.

To do so please change the value of `timeout.reconciliation` in argocd-cm config map:

```bash
k edit cm argocd-cm -n argocd
```


## Deploy Cert-manager and Vault

Run the following commands to deploy cert-manager and Vault on the clusters.
Vault is deployed only on the coordinator cluster.

```bash
kubectl config use-context coordinator
kubectl apply -f samples/multicluster-argocd/vault-app.yaml
kubectl apply -f samples/multicluster-argocd/cert-manager-appset.yaml
```

If you are running on OpenShift cluster you should execute the followng command:

```bash
kubectl patch application vault -n argocd --type=json -p='[{"op": "add", "path": "/spec/source/helm/valueFiles/-", "value": "https://raw.githubusercontent.com/fybrik/fybrik/v1.3.2/charts/vault/vault-openshift-values.yaml"}]'
```

Please note that the deployments are automatically synced as defined in the applications.
To view the status of the deployments in Argo CD GUI please press the `Applications` bottom on the left bar in the GUI.
The deployments should be in `Synced` state.

## Fybrik deployment using Argo CD

This section describe Fybrik deployment on the clusters using Argo CD. In addition, Argo CD serves as the multi-cluster manager for Fybrik and thus relevant information needs to be provided upon Fybrik deployment.

The folder `samples/multicluster-argocd/fybrik-applications/` contains Argo CD applications that install the fybrik-crd and fybrik helm chart on the clusters.
It's important to note that the default prefix for ArgoCD application names related to the Fybrik helm chart deployment is 'fybrik'. The complete application names for fybrik deployment are expected to be in the format: fybrik-<cluster-name>. For example, if the cluster name is "remote" and the prefix is "fybrik", then the ArgoCD application name should be "fybrik-remote".
The application name prefix ('fybrik') is customized and can be changed upon fybrik deployment by changing the relavent helm parameter.
It is crucial that the ArgoCD application names for Fybrik deployments follow the specified syntax, as Fybrik relies on it when retrieving cluster information from the Argo CD server.

**Note** Fybrik is deployed with `clusterScope=false` option due to an [open issue](https://github.com/kubernetes-sigs/aws-load-balancer-controller/issues/3188) in Argo CD which causes cert-manager related resources in Fybrik deployment to be generated with a wrong cert-manager version. With `clusterScope=false` option Fybrik application namespace needs to be manually created before Fybrik deployment by running the following commands:

```bash
kubectl config use-context coordinator
kubectl create ns fybrik-notebook-sample
kubectl config use-context remote
kubectl create ns fybrik-notebook-sample
```

Before installing the applications, details about Argo CD local deployment and the git repository needs to be updated in the `samples/multicluster-argocd/fybrik-applications/fybrik-coordinator.yaml` file:

### Coordinator cluster

File `samples/multicluster-argocd/fybrik-applications/fybrik-coordinator.yaml` contains Fybrik deployment on the coordinator cluster.

To do so the following fields needs to be updated:

```bash
        - name: coordinator.argocd.password
          value: "password"
        - name: coordinator.argocd.appsGitRepo.user
          value: "gitUsername"
        - name: coordinator.argocd.appsGitRepo.url
          value: "https://github.com/fybrik/argocd-test"
        - name: coordinator.argocd.appsGitRepo.password
          value: "gitPassowrd"
```

For example, for Argo CD password the following command can be used to retrieve the password:

```bash
kubectl config use-context coordinator

kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
```

Then, replace the value of `coordinator.argocd.password` in `samples/multicluster-argocd/fybrik-applications/fybrik-coordinator.yaml` with the password value above.

### Remote cluster

The remote clusters only need the watch keeper and cluster subscription agents installed. The remote clusters do not need the coordinator component of Fybrik. Thus the coordinator configuration in `samples/multicluster-argocd/fybrik-applications/fybrik-remote.yaml` file looks like the following:

```bash
    helm:
      parameters:
        - name: coordinator.enabled
          value: "false"
```


**Note** that the cluster information such as region and zone is different in the two applications.

Finally, apply the applications to deploy Fybrik chart on the clusters:

```bash
kubectl config use-context coordinator
kubectl apply -f samples/multicluster-argocd/fybrik-applications/fybrik-crd-coordinator.yaml
kubectl apply -f samples/multicluster-argocd/fybrik-applications/fybrik-crd-remote.yaml
kubectl apply -f samples/multicluster-argocd/fybrik-applications/fybrik-coordinator.yaml
kubectl apply -f samples/multicluster-argocd/fybrik-applications/fybrik-remote.yaml
```

**Note** that the applications except `fybrik-coordinator` are automatically synced as defined in the applications.
To view the status of the deployments in Argo CD GUI please press the `Applications` bottom on the left bar in the GUI.
The deployments should be in `Synced` state.

### Manually sync `fybrik-coordinator` application

Note that the auto sync option is disabled for the `fybrik-coordinator` application. This is done to allow changes in the `fybrik-adminconfig` ConfigMap in fybrik deployment that are shown in the upcoming section. To manually sync the `fybrik-coordinator` application please go to the Argo CD GUI and press `Applications`  bottom on the left bar. Then enter the `fybrik-coordinator` application and press the `sync` bottom.

TODO: add support in Fybrik to add policies to the adminConfig via the helm values.yaml.

## Deploy Fybrik modules

Next, deploy the [arrow flight module](https://github.com/fybrik/arrow-flight-module)
which enables reading data through Apache Arrow Flight API.

```bash
kubectl config use-context coordinator
kubectl apply -f https://raw.githubusercontent.com/fybrik/arrow-flight-module/master/module.yaml -n fybrik-system
```

## Add Adminconfig policy

Add an [extended policy](https://fybrik.io/dev/concepts/config-policies/#extended-policies) to meet advanced deployment requirements. In this sample a policy which specify where the transform modules should run is deployed. As the katalog Asset region is `theshire` then the blueprint is expected to be created on the remote cluster `remote`.

```bash
kubectl config use-context coordinator
kubectl edit cm fybrik-adminconfig -n fybrik-system
```

Add the following policy:

```rego
    config[{"capability": "transform", "decision": decision}] {
        policy := {"ID": "transform-geo", "description":"Governance based transformations must take place in the geography where the data is stored", "version": "0.1"}
        cluster_restrict := {"property": "metadata.region", "values": [input.request.dataset.geography]}
        decision := {"policy": policy, "restrictions": {"clusters": [cluster_restrict]}}
    }
```

## Apply Argo CD application for the Blueprints

Upon Fybrik deployment, a new directory named `blueprints` is automatically created (if not exists) on the github repository with sub-directories for each of the clusters to hold the blueprints of that cluster.

File  `samples/multicluster-argocd/blueprints-appset.yaml` contains Argo CD applicationSet to sync the Fybrik blueprints from the git repo described above with the clusters. The Argo CD applications are generated with name prefix "blueprints" while the full applications names for the blueprints deployment are expected to be of the form: blueprints-<cluster-name>. For example, when the cluster name is "remote" the application name is `blueprints-remote`.

Next execute the following command to apply the applicationSet:

```bash
kubectl config use-context coordinator
kubectl apply -f samples/multicluster-argocd/blueprints-appset.yaml
```

## Run the notebook read flow sample

Execute the [`before we begin`](https://fybrik.io/v1.3/samples/pre-steps/) section and [notebook-read](https://fybrik.io/v1.3/samples/notebook-read/) section, using the katalog as data catalog. Stop before `Create a FybrikApplication resource for the notebook` section.

## Apply Fybrik application

Execute the following command to create fybrikapplication resource.

```bash
cat <<EOF | kubectl apply -f -
apiVersion: app.fybrik.io/v1beta1
kind: FybrikApplication
metadata:
  name: my-notebook
  namespace: fybrik-notebook-sample
  labels:
    app: my-notebook
spec:
  selector:
    clusterName: remote
    workloadSelector:
      matchLabels:
        app: my-notebook
  appInfo:
    intent: Fraud Detection
  data:
    - dataSetID: 'fybrik-notebook-sample/paysim-csv'
      requirements:
        interface: 
          protocol: fybrik-arrow-flight
EOF
```

### Manual refresh `blueprints-remote` application

Due to an [open issue](https://github.com/argoproj/argo-cd/issues/10329) in Argo CD a manual refresh needs to be done for the blueprint application to fetch the latest changes from Git repo. It can be done by pressing the `Applications` bottom on the left bar in the Argo CD GUI. Then, enter the `blueprints-remote` and press the Refresh bottom on the top of the page.

Then Run the following command to wait until the FybrikApplication is ready

```bash
kubectl config use-context coordinator
while [[ $(kubectl get fybrikapplication my-notebook -o 'jsonpath={.status.ready}') != "true" ]]; do echo "waiting for FybrikApplication" && sleep 5; done
```


## Cleanup

Follow the [Fybrik cleanup](https://fybrik.io/v1.3/samples/cleanup/) section to cleanup the resources used in this sample.
Due to an [open issue](https://github.com/argoproj/argo-cd/issues/10329) in Argo CD a manual refresh needs to be done for the blueprint application to fetch the latest changes from Git repo. It can be done by pressing the `Applications` bottom on the left bar in the Argo CD GUI. Then, enter the `blueprints-remote` and press the Refresh bottom on the top of the page.



