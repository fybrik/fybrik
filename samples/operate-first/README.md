# Operate First Cluster Scoped Resources
This directory contains the raw yaml files of the cluster scoped resources deployed for Fybrik. These files are generated from the Helm chart.

If the Helm chart has been updated, follow the steps below to generate the new yaml files:
1. Install yq and Helm in this repo. Run these commands from the root directory of the repo:
```bash
cd hack/tools
./install_yq.sh
./install_helm.sh
```
2. Go back to `operate-first` folder and set up Python environment there
```bash
cd samples/operate-first
pipenv install
pipenv shell
```
3. Run Makefile to generate new yaml files from Helm charts:
```bash
make all
```