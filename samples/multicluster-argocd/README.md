# Fybrik multi-cluster management using ArgoCD


[Argo CD](https://argo-cd.readthedocs.io/en/stable/) is a declarative, GitOps continuous delivery tool for Kubernetes.
This README contains information about how to enable Fybrik to use ArgoCD as its multi-cluster manager. 

**Disclaimer**: This work is in progress.

# Argo CD in high level

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
- Argo CD CLI
- Github account and Github repository to store the Blueprints


## Setup multi-cluster environment

For local testing [`hack/setup-local-multi-cluster.sh`](../hack/setup-local-multi-cluster.sh) script can be used to setup two kind clusters:

- `kind-control`: the coordinator cluster.
- `kind-kind`: the remote cluster.

After running the script above the two clusters contains the following deployments:
- `kind-control`:

                  1. Argo CD in argocd namespace

                  2. Vault (*)

                  3. Cert Manager (*)

- `kind-kind`: 

                  1. Cert Manager (*)

(*) TODO: should be deployed with Argo CD

## Logging to Argo CD:

To execute Argo CD CLI commands in this tutorial, run the following login command:

First, retrieve the admin password from the command line:
```bash
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
```

Then, port-forward the argocd service:

```bash
kubectl port-forward service/argo-argocd-server -n argocd 8080:443 &
```

Lastly, logging using the admin password retrieved above.

```bash
kubectl config set-context kind-control --namespace=argocd
argocd login localhost:8080 --insecure
```

For more information please refer to the Argo CD [getting started page](https://argo-cd.readthedocs.io/en/stable/getting_started/).

## Logging to Argocd GUI

The API server can then be accessed using https://localhost:8080

For more information please refer to the Argo CD [getting started page](https://argo-cd.readthedocs.io/en/stable/getting_started/).

## Configure Argo CD deployment

The following steps needs to be executed to configure the argo CD server:

1. Add clusters:

The coordinator cluster is automatically registered in argo CD server. However its default name is `in-cluster`. In this tutorial we change it to `kind-control`. This could be done via the Argo CD GUI.
For the remote cluster, please follow the steps below that are taken from the [link](https://github.com/argoproj/argo-cd/issues/4204) which are require due to a bug in registering cluster in KinD. Note that for newer version of Argo CD a new flag was added to fix that (`cluster add --cluster-endpoint`) but it is not working properly.

```bash
kubectl config use-context kind-kind
kubectl get endpoints -A
```
Copy the endpoint for `kubernetes` to ~/.kube/config , replacing the existing `server:` field for 'kind-kind' cluster.

Then run the following command to register the cluster:
```bash
kubectl config set-context kind-control --namespace=argocd
argocd cluster --insecure add kind-kind
```

2. Add private repositories:

A github repository needs to be provided for the purpose of hosting the Fybrik Blueprints.
Run the following command to register a private repository in Argo CD:

```bash
argocd repo add https://github.com/xxx/argocd-fybrik-blueprints --name my-blueprints --username username --password xxx
```

Alternatively, the repository can be added using the Argo CD GUI.

### Fybrik deployment using Argo CD

This section describe Fybrik deployment on the clusters using Argo CD. In addition, Argo CD serves as the multi-cluster manager for Fybrik and thus relevant information needs to be provided upon Fybrik deployment.

The folder `samples/multicluster-argocd/fybrik-applications/` contains Argo CD applications that install the fybrik-crd and fybrik helm chart on the clusters.
It's important to note that the default prefix for ArgoCD application names related to the Fybrik helm chart deployment is 'fybrik'. The complete application names for fybrik deployment are expected to be in the format: fybrik-<cluster-name>. For example, if the cluster name is "kind-kind" and the prefix is "fybrik", then the ArgoCD application name should be "fybrik-kind-kind".
The application name prefix ('fybrik') can be customized and changed during fybrik deployment on the coordinator cluster via the argo CD application.
It is crucial that the ArgoCD application names for Fybrik deployments follow the specified syntax, as Fybrik relies on it when retrieving cluster information from the Argo CD server.

Before installing the applications, details about Argo CD local deployment and the git repository needs to be updated in the `samples/multicluster-argocd/fybrik-applications/fybrik-kind-control.yaml` file:

### Coordinator cluster

File `samples/multicluster-argocd/fybrik-applications/fybrik-kind-control.yaml` contains Fybrik deployment on the coordinator cluster.

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
kubectl config use-context kind-control

ARGO_PASSWORD=$(kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d)
```

Then, replace the value of `coordinator.argocd.password` in samples/multicluster-argocd/fybrik-applications/fybrik-kind-control.yaml with ARGO_PASSWORD value above.

#### Remote cluster

The remote clusters only need the watch keeper and cluster subscription agents installed. The remote clusters do not need the coordinator component of Fybrik. Thus the coordinator configuration in `samples/multicluster-argocd/fybrik-applications/fybrik-kind-kind.yaml` file looks like the following:

```bash
    helm:
      parameters:
        - name: coordinator.enabled
          value: "false"
```

Once all the Argo CD applications in `samples/multicluster-argocd/fybrik-applications/` folder are applied then they can be synced via the Argo CD GUI given that auto sync option is disabled.

### Apply Argo CD application for the Blueprints

Upon Fybrik deployment, a new directory named `blueprints` is automatically created on the github repository with sub-directories for each of the clusters to hold the blueprints of that cluster.

File  `samples/multicluster-argocd/blueprints-appset.yaml` contains Argo CD applicationSet to sync the Fybrik blueprints from the git repo described above with the clusters. The Argo CD applications are generated with name prefix "blueprints" and the full applications names for the blueprints deployment are expected to be of the form: blueprints-<cluster-name>. For example, when the cluster name is "kind-kind" the application name is `blueprints-kind-kind`.

Note: This stage should be done *After* Fybrik deployment as the later creates the sub directory for each cluster to hold the blueprint of that cluster.





