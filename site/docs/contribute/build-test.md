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

With the following you will then setup a kind cluster with the local registry,
build and push current docker images and finally run the integration
tests on it:

```bash
make run-integration-tests
```

You can run `make kind-cleanup` to delete the created clusters when you're done.

## Building in a multi cluster environment

As Fybrik can run in a multi-cluster environment there is also a test environment
that can be used that simulates this scenario. Using kind one can spin up two separate kubernetes
clusters with differnt contexts and develop and test in these. 

Two kind clusters that share the same kind-registry can be set up using:
```bash
make kind-setup-multi
``` 

You can run `make kind-cleanup` to delete the created clusters when you're done.
