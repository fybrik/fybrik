# Build and Test

## Build the project images

```bash
make docker-build
```

## Run unit tests

```bash
make test
```

Some tests for controllers are written in a fashion that they can be run on a simulated environment using 
[envtest](https://pkg.go.dev/github.com/kubernetes-sigs/controller-runtime/pkg/envtest) or on an already existing
Kubernetes cluster (or local kind cluster). The default is to use envtest. In order to run the tests in a local cluster
the following environment variables can be set:
```bash
NO_SIMULATED_PROGRESS=true USE_EXISTING_CLUSTER=true make -C manager test
```

Please be aware that the controller is running locally in this case! If a controller is already deployed onto the
cluster then the tests can be run with the command below. This will ensure that the tests are only creating CRDs on 
the cluster and checking their status:
```bash
USE_EXISTING_CONTROLLER=true NO_SIMULATED_PROGRESS=true USE_EXISTING_CLUSTER=true make -C manager test
```

### Environment variables description

| Environment variable    | Default | Description
| -                       | -       | - 
| USE_EXISTING_CLUSTER    | false   | This variable controls if an existing K8s cluster should be used or not. If not envtest will spin up an artificial environment that includes a local etcd setup.
| NO_SIMULATED_PROGRESS   | false   | This variable can be used by tests that can manually simulate progress of e.g. jobs or pods. e.g. the simulated test environment from testEnv does not progress pods etc while when testing against an external Kubernetes cluster this will actually run pods.
| USE_EXISTING_CONTROLLER | false   | This variable controls if a controller should be set up and run by this test suite or if an external one should be used. E.g. in integration tests running against an existing setup a controller is already existing in the Kubernetes cluster and should not be started by the test as two controllers competing may influence the test.


## Running integration tests

### Running in one step

With the following you will then setup a kind cluster with the local registry,
build and push current docker images and finally run the integration
tests on it:

```bash
make run-integration-tests
```

### Running step by step

It is also possible to call the commands step by step, which sometimes is
useful if you want to only repeat a specific step which failed without having
to rerun  the entire sequence

```bash
# use the local kind registry
export DOCKER_HOSTNAME=kind-registry:5000
export DOCKER_NAMESPACE=m4d-system

# build a local kind cluser
make kind

# deploy the the cluster 3rd party such as cert-manager and vault
make cluster-prepare

# build all docker images and push them to the local registry
make docker

# build the mock/test docker images and push them to local registry
make -C test/services docker-build docker-push

# wait until cluster-prepare setup really completed
make cluster-prepare-wait

# configure Vault
make configure-vault

# deploy the m4d CRDs to the kind cluster
make -C manager deploy-crd

# deploy m4d manager to the kind cluster
make -C manager deploy_it

# wait until manager is ready
make -C manager wait_for_manager

# build and push helm charts to the local registry
make helm

# actually run the integration tests
make -C manager run-integration-tests
```

## Building in a multi cluster environment

As Mesh for Data can run in a multi-cluster environment there is also a test environment
that can be used that simulates this scenario. Using kind one can spin up two separate kubernetes
clusters with differnt contexts and develop and test in these. 

Two kind clusters that share the same kind-registry can be set up using:
```bash
make kind-setup-multi
``` 
