## Kubeflow on OpenShift 4+ with Istio 1.7 Installation 

Follow [Kubeflow Documentation](https://www.kubeflow.org/docs/openshift/install-kubeflow/) to install Kubeflow with a few changes detailed below:
```bash
git clone https://github.com/opendatahub-io/manifests.git
```

A modified `kfctl_openshift.yaml` is provided in this repo with the following changes already made:
* Since we already have Istio 1.7 installed, comment out `istio-crds` and `istio-install`
* Comment out `cert-manager-crds` and `cert-manager` 
* Comment out `seldon-core-operator` 

Make sure to update the manifest uri to your local file path
```bash
sed -i 's#uri: .*#uri: '$PWD'#' ./kfdef/kfctl_openshift.yaml
```

Changes to `kf-istio-resources.yaml`:
```bash
cd /opt/openshift-kfdef/kustomize/istio/base 
```
 *Replace sni_hosts with sniHosts on lines 63 and 96

 Deploy Kubeflow resources:
```bash
cd /opt/openshift-kfdef
sudo kfctl build --file=kfctl_openshift.yaml
sudo kfctl apply --file=kfctl_openshift.yaml
```
Verify all Kubeflow resources are deployed:
```bash
kubectl get pods -n kubeflow
