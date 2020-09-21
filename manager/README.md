[![Build Status](https://travis.ibm.com/data-mesh-research/data-movement-operator.svg?token=1MDqPqDts2bujj49quzq&branch=master)](https://travis.ibm.com/data-mesh-research/data-movement-operator)

# Controller manager
This operator is created using the kubebuilder tooling. A good starting point is
the documentation at https://book.kubebuilder.io/ . It's a multigroup project that 
contains different categories of CRDs.

Kubebuilder generates multiple files in different steps of the development.
All the go files are generated when executing `kubebuilder create api` commands.
The actual CRD definition and manifest for the webhook are generated when executing `make manifests`.
So whenever a field is added in a `*_types.go` file, the webhook is changed or any kubebuilder
 specific annotations are changed please execute `make manifests` so that 
the correct yaml files are generated. 

The Kubernetes configuration is based on [kustomize](https://github.com/kubernetes-sigs/kustomize) which is a templating
framework for Kubernetes yaml files and is integrated into `kubectl`. When 
executing e.g. `kubectl apply -k config/default` the kustomize engine will template the
files in the given folder. Kustomize defines `kustomization.yaml` files in each folder that
describe what resources it should work on (e.g. `manager.yaml` in config/manager) and what patches 
should be applied. As different overlays can be created using kustomize different environments
can also be manged with this mechanism.

Description of configuration folder structure:
- **config/certmanager**: Creates a certificate that can be used for the webhook
- **config/crd**: Contains the generated CRD definition as well as some patches for the
 webhook of the crd. This gets installed into K8s when executing `make install`
- **config/default**: This is the default deployment that deploys everything. This includes the CRD, manager deployment, certificates, rbac rules and webhook. 
 This gets installed into the cluster when executing `make deploy`.
- **config/manager**: Kustomization of the deployment of the controller of the operator. Just the K8s deployment object of the manager.
- **config/minikube**: This is one profile that is based on `config/default` and applies patches to
 make it easier develop the operator in minikube. (Sets image names to local ones and sets imagePullPolicy: IfNotPresent)
- **config/movement-controller**: This is used to install a controller on K8s that is just managing the motion CRDs and no other CRDs. 
  its used in the make target `make deploy_mc` (When the movement-controller image was deployed before from the /build folder)
- **config/prod**: This is meant as the production profile that is based on `config/default` and applies patches to be run
 in the production environment. (NOT DEFINED YET)
- **config/prometheus**: For monitoring using prometheus (not used at the moment)
- **config/rbac**: Contains the rbac rules for the operator. (TODO describe roles in more detail)
- **config/samples**: Sample object instances of transfers that can be used for development and demos
- **config/test**: This is meant as the test profile that is based on `config/default` and applies patches to be run
 in the test environment. (NOT DEFINED YET)
- **config/webhook**: Contains kustomize rules for installing the webhook service and mutating and validating configurations

This data movement operator supports different forms of transformations that are described [here](doc/Transformations.md).

### Local testing with minikube

Minikube is the perfect environment for local testing. It comes with an included registry.

```minikube start --vm-driver=virtualbox --addons=registry --memory=4000mb```

Then to expose the docker registry of minikube locally run:

```eval $(minikube docker-env)```

The image can be build locally afterwards and put directly into the docker registry of minikube.

```make docker-build-local```

Deploying to minikube using the local image can be done using a special kustomized specialization called minikube:

```make deploy-mini```

### Troubleshooting
If the operator is being redeployed by the tekton pipeline and there are already BatchTransfer objects deployed a deployment
may hang and fail as the BatchTransfer objects cannot be deleted. This is due to the fact that
the CI pipeline already removed the deployment that is running the finalizer. Solving this issue can be
done by manually editing the still existing BatchTransfer objects and removing all finalizers. 
The preferred way though is that all BatchTransfer objects are removed before a redeployment is done:
```kubectl delete batchtransfer --all -A```