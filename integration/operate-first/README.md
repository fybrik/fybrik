# Deploying Fybrik on Operate First
[Operate First](https://www.operate-first.cloud/) is a concept of bringing pre-release open source software to a production cloud environment. The Mass Open Cloud (MOC) is a production cloud resource where projects are run. Deploying Fybrik on Operate First is the first step to integrating Fybrik with Open Data Hub, making Fybrik more easily accessible to data scientists. For further questions about Operate First or contributing to [Operate First GitHub](https://github.com/operate-first) repositories, join the Slack channel [here](http://operatefirst.slack.com/)

## Accessing the [MOC](https://massopen.cloud/) Smaug cluster
The Smaug cluster is where all user workloads are deployed. A deployment of [Open Data Hub](https://opendatahub.io/) (ODH) is also managed on the Smaug cluster. This is valuable to Fybrik since this ODH Deployment includes JupyterHub.

### Getting Access
Your GitHub username must be in [this file](https://github.com/operate-first/apps/blob/master/cluster-scope/base/user.openshift.io/groups/fybrik/group.yaml) to get access to the Fybrik user group on the Smaug cluster and login successfully. Create a PR in the [operate-first/apps](https://github.com/operate-first/apps) repository with your GitHub username added to `group.yaml` if you would like to be added as a user. 

### Logging In 
You can access the Smaug cluster with this [OpenShift console login link](https://oauth-openshift.apps.smaug.na.operate-first.cloud/oauth/authorize?client_id=console&redirect_uri=https%3A%2F%2Fconsole-openshift-console.apps.smaug.na.operate-first.cloud%2Fauth%2Fcallback&response_type=code&scope=user%3Afull&state=98ae2ceb). Click on `operate-first` to login with GitHub authentication. Once logged into the OpenShift console, you can use [this link](https://oauth-openshift.apps.smaug.na.operate-first.cloud/oauth/token/display) to get an `oc login` command with a token that will let you login to the Smaug OpenShift cluster from your terminal. 

## Deploying Fybrik Cluster-Scoped Resources
In the Operate First environment, cluster-scoped resource manifests must be added to the [operate-first/apps](https://github.com/operate-first/apps) repository to be deployed on the Smaug cluster because of security reasons. This `integration/operate-first` directory contains the raw YAML files of the cluster scoped resources deployed for Fybrik, mainly the custom resource definitions (CRDs)used by Fybrik. These files are generated from the Helm charts in `charts/fybrik`.

If the Helm chart has been updated, follow the steps below to generate the new yaml files:
1. Install yq and Helm in this repo. Run these commands from the root directory of this repo:
```bash
cd hack/tools
./install_yq.sh
./install_helm.sh
```
2. Go back to the `integration/operate-first` folder and set up the Python environment there
```bash
cd integration/operate-first
pipenv install
pipenv shell
```
3. Run Makefile to generate new YAML files from the Helm charts:
```bash
make all
```

After the cluster-scoped YAML files are generated, create a PR to the [operate-first/apps](https://github.com/operate-first/apps) repository in the [cluster-scope/base](https://github.com/operate-first/apps/tree/master/cluster-scope/base) directory with the YAML files in the subdirectories organized by resource type. Any resource added to base has also been added to [kustomization.yaml](https://github.com/operate-first/apps/blob/master/cluster-scope/overlays/prod/moc/smaug/kustomization.yaml) in [cluster-scope/overlays](https://github.com/operate-first/apps/tree/master/cluster-scope/overlays). Resources will only be deployed to the Smaug cluster if they are included in this `kustomization.yaml` file. Namespaces must also be added to this file to be created on the Smaug cluster. We currently have 3 namespaces that anyone in the Fybrik user group can access: `fybrik-system`, `fybrik-blueprints`, and `fybrik-applications`. More documentation about contributing to the `operate-first/apps` repository can be found [here](https://github.com/operate-first/apps/tree/master/docs/content).

## Deploying Namespace-Scoped resources
Operate First has an [ArgoCD](https://argo-cd.readthedocs.io/en/stable/) instance deployed on MOC that can be used to deploy OpenShift resources located on a Git Repository. Only namespace-scoped resources can be deployed with ArgoCD. Any cluster-scoped resource, such as CRDs or cluster roles, will be blocked by ArgoCD. The namespace-scoped resources required for Fybrik have been onboarded to ArgoCD by following [these instructions](https://github.com/operate-first/apps/blob/master/docs/content/argocd-gitops/onboarding_to_argocd.md) and an ArgoCD project has been created for Fybrik. You can login to [the ArgoCD instance](https://argocd.operate-first.cloud/applications?proj=&sync=&health=&namespace=&cluster=&labels=) with the same login method as above. We have deployed 2 [ArgoCD applications](https://argo-cd.readthedocs.io/en/stable/operator-manual/declarative-setup/#applications) which are automatically synced with the latest release of Fybrik. The `fybrik` and `vault` ArgoCD applications deployed on the Smaug cluster are in sync with the [`fybrik/charts`](https://github.com/fybrik/charts) repository.

The following are the [ArgoCD application manifests](https://github.com/operate-first/apps/tree/master/argocd/overlays/moc-infra/applications/envs/moc/smaug/fybrik) which have been added to the `operate-first/apps` repository:

```bash
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: fybrik
spec:
  destination:
    name: smaug
    namespace: fybrik-system
  source:
    path: charts/fybrik
    repoURL: 'https://github.com/fybrik/charts'
    targetRevision: HEAD
    helm:
      parameters:
        # Disable deploying Fbrik cluster scoped resources
        - name: clusterScoped
          value: "false"
        # Only watch for FybrikApplication from fybrik-applications
        - name: applicationNamespace
          value: fybrik-applications
  project: fybrik
```

```bash
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: vault
spec:
  project: fybrik
  source:
    repoURL: 'https://github.com/fybrik/charts'
    path: charts/vault
    targetRevision: HEAD
    helm:
      valueFiles:
        - env/dev/vault-single-cluster-values.yaml
      parameters:
        # authDelegator enables a cluster role binding to be attached to the service account.
        # The cluster role binding is already deployed in the smaug cluster and thus authDelegator can be disabled.
        - name: vault.server.authDelegator.enabled
          value: 'false'
        - name: vault.global.openshift
          value: 'true'
        - name: vault.injector.enabled
          value: 'false'
        - name: vault.server.dev.enabled
          value: 'true'
      values: |
        plugins:
          vaultPluginSecretsKubernetesReader:
            enabled: true
            clusterScope: false
            namespaces:
              - fybrik-applications
              - fybrik-system
        modulesNamespace: "fybrik-blueprints"
  destination:
    namespace: fybrik-system
    name: smaug
```

## Running the Fybrik notebook sample on Operate First
1) Follow the steps in [Fybrik notebook sample](https://fybrik.io/v0.5/samples/notebook/) to [prepare a dataset to be accessed by the notebook](https://fybrik.io/v0.5/samples/notebook/#prepare-a-dataset-to-be-accessed-by-the-notebook), [register the dataset in a data catalog](https://fybrik.io/v0.5/samples/notebook/#register-the-dataset-in-a-data-catalog), and [define data access policies](https://fybrik.io/v0.5/samples/notebook/#define-data-access-policies). Make sure to use the `fybrik-applications` namespace instead of the `fybrik-notebook-sample` namespace since `fybrik-applications` has already been created on the Smaug cluster. 
2) Access the JupyterHub instance deployed on the Smaug OpenShift cluster [here](https://oauth-openshift.apps.smaug.na.operate-first.cloud/oauth/authorize?response_type=code&redirect_uri=https%3A%2F%2Fjupyterhub-opf-jupyterhub.apps.smaug.na.operate-first.cloud%2Fhub%2Foauth_callback&client_id=system%3Aserviceaccount%3Aopf-jupyterhub%3Ajupyterhub-hub&state=eyJzdGF0ZV9pZCI6ICIwY2ZkYzYwMjA4MjY0OGZiYWY5MDk3OWJkOGFhZjE4NyIsICJuZXh0X3VybCI6ICIvaHViLyJ9&scope=user%3Ainfo) and login with the above method. 
3) Start a notebook server using the Elyra Notebook Image or any image of your choosing
4) Create a notebook and insert a new notebook cell with the Python code in Step 2 of [Read the dataset from the notebook](https://fybrik.io/v0.5/samples/notebook/#read-the-dataset-from-the-notebook). Make sure to change the `asset` to `fybrik-applications/paysim-csv` instead of `fybrik-notebook-sample/paysim-csv`
