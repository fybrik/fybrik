# Deploying Fybrik on Operate First
[Operate First](https://www.operate-first.cloud/) is a concept of bringing pre-release open source software to a production cloud environment. The Mass Open Cloud (MOC) is a production cloud resource where projects are run. Deploying Fybrik on Operate First is the first step to integrating Fybrik with Open Data Hub, making Fybrik more easily accessible to data scientists. For further questions about Operate First or contributing to [Operate First GitHub](https://github.com/operate-first) repositories, join the Slack channel [here](http://operatefirst.slack.com/)

## Accessing the [MOC](https://massopen.cloud/) Smaug cluster
The Smaug cluster is where all user workloads are deployed. A deployment of [Open Data Hub](https://opendatahub.io/) (ODH) is also managed on the Smaug cluster. This is valuable to Fybrik since this ODH Deployment includes JupyterHub.

### Getting Access
Your GitHub username must be in [this file](https://github.com/operate-first/apps/blob/master/cluster-scope/base/user.openshift.io/groups/fybrik/group.yaml) to get access to the Fybrik user group on the Smaug cluster and login successfully. Create a PR in the [operate-first/apps](https://github.com/operate-first/apps) repository with your GitHub username added to `group.yaml` if you would like to be added as a user. 

### Logging In 
You can access the Smaug cluster with this [OpenShift console login link](https://oauth-openshift.apps.smaug.na.operate-first.cloud/oauth/authorize?client_id=console&redirect_uri=https%3A%2F%2Fconsole-openshift-console.apps.smaug.na.operate-first.cloud%2Fauth%2Fcallback&response_type=code&scope=user%3Afull&state=98ae2ceb). Click on `operate-first` to login with GitHub authentication. Once logged into the OpenShift console, you can use [this link](https://oauth-openshift.apps.smaug.na.operate-first.cloud/oauth/token/display) to get an `oc login` command with a token that will let you login to the Smaug OpenShift cluster from your terminal. 

## Deploying Fybrik Cluster-Scoped Resources
This `integration/operate-first` directory contains the raw YAML files of the cluster scoped resources deployed for Fybrik. These files are generated from the Helm charts in `charts/fybrik`.

If the Helm chart has been updated, follow the steps below to generate the new yaml files:
1. Install yq and Helm in this repo. Run these commands from the root directory of this repo:
```bash
cd hack/tools
./install_yq.sh
./install_helm.sh
```
2. Go back to the `integration/operate-first` folder and set up the Python environment there
```bash
cd samples/operate-first
pipenv install
pipenv shell
```
3. Run Makefile to generate new YAML files from the Helm charts:
```bash
make all
```

After the cluster-scoped YAML files are generated, create a PR to the [operate-first/apps](https://github.com/operate-first/apps) repository in the [cluster-scope/base](https://github.com/operate-first/apps/tree/master/cluster-scope/base) directory with the YAML files in the subdirectories organized by resource type. More documentation about contributing to the `operate-first/apps` repository can be found [here](https://github.com/operate-first/apps/tree/master/docs/content)

## Deploying Namespace-Scoped resources
Operate First has an [ArgoCD](https://argo-cd.readthedocs.io/en/stable/) instance deployed on MOC that can be used to deploy OpenShift resources located on a Git Repository. Fybrik has been onboarded to ArgoCD by following [these instructions](https://github.com/operate-first/apps/blob/master/docs/content/argocd-gitops/onboarding_to_argocd.md) and an ArgoCD project has been created for Fybrik. You can login to [the ArgoCD instance](https://argocd.operate-first.cloud/applications?proj=&sync=&health=&namespace=&cluster=&labels=) with the same login method as above. We have deployed 3 [ArgoCD applications](https://argo-cd.readthedocs.io/en/stable/operator-manual/declarative-setup/#applications) which are automatically synced with the latest release of Fybrik. See the ArgoCD applications deployed below pointing to the corresponding Fybrik repositories:
* fybrik -> https://github.com/fybrik/charts
* vault -> https://github.com/fybrik/fybrik

The complete ArgoCD application manifests have been added to the `operate-first/apps` repository [here](https://github.com/operate-first/apps/tree/master/argocd/overlays/moc-infra/applications/envs/moc/smaug/fybrik)

## Running the Fybrik quickstart on Operate First
1) Follow the steps in [Fybrik notebook sample](https://fybrik.io/v0.5/samples/notebook/) to prepare a dataset to be accessed by the notebook, register the dataset in a data catalog, and define data access policies.
2) Access the JupyterHub instance deployed on the Smaug OpenShift cluster [here](https://oauth-openshift.apps.smaug.na.operate-first.cloud/oauth/authorize?response_type=code&redirect_uri=https%3A%2F%2Fjupyterhub-opf-jupyterhub.apps.smaug.na.operate-first.cloud%2Fhub%2Foauth_callback&client_id=system%3Aserviceaccount%3Aopf-jupyterhub%3Ajupyterhub-hub&state=eyJzdGF0ZV9pZCI6ICIwY2ZkYzYwMjA4MjY0OGZiYWY5MDk3OWJkOGFhZjE4NyIsICJuZXh0X3VybCI6ICIvaHViLyJ9&scope=user%3Ainfo) and login with the above method. 
3) Start a notebook server using the Elyra Notebook Image or any image of your choosing
4) Create a notebook with the following cells:
```python
%pip install pandas pyarrow
```
```python
import json
import pyarrow.flight as fl
import pandas as pd
```
```python
# Create a Flight client
client = fl.connect('grpc://my-notebook-fybrik-applications-arrow-flight-module.fybrik-blueprints:80')
```
```python
# Prepare the request
request = {
    "asset": "fybrik-applications/paysim-csv",
    # To request specific columns add to the request a "columns" key with a list of column names
    # "columns": [...]
}
```
```python
# Send request and fetch result as a pandas DataFrame
info = client.get_flight_info(fl.FlightDescriptor.for_command(json.dumps(request)))
reader: fl.FlightStreamReader = client.do_get(info.endpoints[0].ticket)
df: pd.DataFrame = reader.read_pandas()
```
```python
df
```
5. Run each cell at a time and you should see the following data after printing `df`:

| step | type | amount   | nameOrig  | oldbalanceOrg | newbalanceOrig | nameDest | oldbalanceDest | newbalanceDest | isFraud     | isFlaggedFraud |     |
|------|------|----------|-----------|---------------|----------------|----------|----------------|----------------|-------------|----------------|-----|
| 0    | 1    | PAYMENT  | 9839.64   | XXXXX         | XXXXX          | XXXXX    | M1979787155    | 0.00           | 0.00        | 0              | 0   |
| 1    | 1    | PAYMENT  | 1864.28   | XXXXX         | XXXXX          | XXXXX    | M2044282225    | 0.00           | 0.00        | 0              | 0   |
| 2    | 1    | TRANSFER | 181.00    | XXXXX         | XXXXX          | XXXXX    | C553264065     | 0.00           | 0.00        | 1              | 0   |
| 3    | 1    | CASH_OUT | 181.00    | XXXXX         | XXXXX          | XXXXX    | C38997010      | 21182.00       | 0.00        | 1              | 0   |
| 4    | 1    | PAYMENT  | 11668.14  | XXXXX         | XXXXX          | XXXXX    | M1230701703    | 0.00           | 0.00        | 0              | 0   |
| ...  | ...  | ...      | ...       | ...           | ...            | ...      | ...            | ...            | ...         | ...            | ... |
| 95   | 1    | TRANSFER | 710544.77 | XXXXX         | XXXXX          | XXXXX    | C1359044626    | 738531.50      | 16518.36    | 0              | 0   |
| 96   | 1    | TRANSFER | 581294.26 | XXXXX         | XXXXX          | XXXXX    | C1590550415    | 5195482.15     | 19169204.93 | 0              | 0   |
| 97   | 1    | TRANSFER | 11996.58  | XXXXX         | XXXXX          | XXXXX    | C1225616405    | 40255.00       | 0.00        | 0              | 0   |
| 98   | 1    | PAYMENT  | 2875.10   | XXXXX         | XXXXX          | XXXXX    | M1651262695    | 0.00           | 0.00        | 0              | 0   |
| 99   | 1    | PAYMENT  | 8586.98   | XXXXX         | XXXXX          | XXXXX    | M494077446     | 0.00           | 0.00        | 0              | 0   |
