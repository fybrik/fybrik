# `manager`

Kubernetes [custom resources and controllers](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) of Mesh for Data.

The `manager` binary includes all of the controllers that this project defines but you need to select which of the controllers to run by passing one or more of the following command line arguments:
- `enable-all-controllers` to enable all controllers
- `enable-application-controller` to enable the controller for `M4DApplication`
- `enable-blueprint-controller` to enable the controller for `Blueprint`
- `enable-motion-controller` to enable the controllers for `BatchTransfer` and `StreamTransfer`

The rest of this README describes the directory structure.

## `apis`

Holds the Customer Resource Definitions (CRDs) of the project:
- [app.m4d.ibm.com/v1alpha1](https://ibm.github.io/the-mesh-for-data/docs/reference/api/generated/app/#k8s-api-app-m4d-ibm-com-v1alpha1): Includes `M4DApplication`, administrator APIs `M4DModule` and `M4DBucket`, and internal CRDs `Blueprint` and `Plotter`.
- [motion.m4d.ibm.com/v1alpha1](https://ibm.github.io/the-mesh-for-data/docs/reference/api/generated/motion/#k8s-api-motion-m4d-ibm-com-v1alpha1): Includes data movements APIs `BatchTransfer` and `StreamTransfer`. Usually not used directly but rather invoked as a module.

## `controllers`

Holds the customer controllers of the project:
- `controllers/app` holds the controllers for `app.m4d.ibm.com` APIs `M4DApplication`, `Blueprint` and `Plotter`.
- `controllers/motion` holds the controllers for `motion.m4d.ibm.com` APIs `BatchTransfer` and `StreamTransfer`.

## `config`

The Kubernetes configuration is based on [kustomize](https://github.com/kubernetes-sigs/kustomize) which is a templating
framework for Kubernetes yaml files and is integrated into `kubectl`. When 
executing e.g. `kubectl apply -k config/default` the kustomize engine will template the
files in the given folder. Kustomize defines `kustomization.yaml` files in each folder that
describe what resources it should work on (e.g. `manager.yaml` in config/manager) and what patches 
should be applied. As different overlays can be created using kustomize different environments
can also be manged with this mechanism.

Description of configuration folder structure:
- `config/certmanager`: Creates a certificate that can be used for webhooks
- `config/control-plane-security`: Applies authorization and authentication policies to secure the control plane. Some of the policies are based on [Istio](https://istio.io/) thus Istio installation is prerequisite for this configuration. It's used in the make target `make deploy_control-plane-security` (in order for the Istio sidecar automatic injection to take effect the pods in the control-plane should be restarted after running the make command).
- `config/crd`: Contains the generated CRD definition as well as some patches for the webhook of the crd. This gets installed into the cluster when executing `make install`
- `config/default`: This is the default deployment that deploys everything but CRDs. This includes, manager deployment, certificates, rbac rules and webhook. This gets installed into the cluster when executing `make deploy`.
- `config/manager`: Kustomization of the deployment of the controller of the operator. Just the K8s deployment object of the manager.
- `config/movement-controller`: This is used to install a controller on K8s that is just managing the motion CRDs and no other CRDs. 
  it's used in the make target `make deploy_mc` (the movement-controller image must already be available in an image registry)
- `config/network-policies`: Applies Kubernetes NetworkPolicy resource to secure the control plane. It's used in the default configuration.
- `config/prod`: This is meant as the production profile that is based on `config/default` and applies patches to be run
 in the production environment. (NOT DEFINED YET)
- `config/prometheus`: For monitoring using prometheus (not used at the moment)
- `config/rbac`: Contains the rbac rules for the operator. (TODO describe roles in more detail)
- `config/samples`: Sample object instances of transfers that can be used for development and demos
- `config/test`: This is meant as the test profile that is based on `config/default` and applies patches to be run
 in the test environment. (NOT DEFINED YET)
- `config/webhook`: Contains kustomize rules for installing the webhook service and mutating and validating configurations

## `testdata`

Includes resources that are used in unit tests and in integration tests. 
