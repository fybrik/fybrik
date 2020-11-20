---
title: Contribution Flow
summary: This page describes the GitHub workflow, build and test instructions.

weight: 1
---

This page describes the flow that contributors should follow, including the GitHub workflow and how to build and test the project after making changes.

# Issues and Pull Requests

Contributing to {{< name >}} is done following the GitHub workflow of Pull Requests.

You should usually open a pull request in the following situations:
- Start work on a contribution that was already asked for, or that you’ve already discussed, in an issue.
- Submit trivial fixes (for example, a typo, a broken link or an obvious error).

A pull request doesn’t have to represent finished work. It’s usually better to open a pull request early on, so others can watch or give feedback on your progress. Just mark it as a “WIP” (Work in Progress) in the subject line. You can always add more commits later.

Here’s how to submit a pull request:

- **[Fork](https://{{< github_base >}}/{{< github_repo >}}/fork)** the main repository
- **Clone the forked repository locally**. Connect your local to the original “upstream” repository by adding it as a remote.
    ```shell
    git clone git@github.com:$(git config user.name)/{{< github_repo >}}.git
    git remote add upstream https://{{< github_base >}}/{{< github_repo >}}/{{< github_repo >}}.git
    git remote set-url --push upstream no_push
    ```
- **[Pull in changes](https://help.github.com/articles/syncing-a-fork/)** from “upstream” often so that you stay up to date so that when you submit your pull request, merge conflicts will be less likely.
    ```shell
    git fetch upstream
    git checkout master
    git merge upstream/master
    git push origin master
    ```
- **[Create a branch](https://guides.github.com/introduction/flow/)** for your edits from master. Note that your should never add edits to the master branch itself.
    ```shell
    git checkout -b <branch name>
    ```
- **Make commits of logical units**, ensuring that commit messages are in the [proper format](#format-of-the-commit-message).
- **Push your changes** to the created branch in your fork of the repository.
- **Open a pull request** to the original repository.
- **Reference any relevant issues** or supporting documentation in your PR (for example, “Closes #37.”)

As always, you must [follow code style](#normalize-the-code), ensure that [all tests pass](#building-and-testing), and add any new tests as appropriate.

**Thanks for your contribution!**

# Building and testing

[Setup a development environment](../devenv) and make sure `make install-tools` finished successfully.

Build and run unit tests with:
```bash
make build
make test
```

Some tests for controllers are written in a fashion that they can be run on a simulated environment using 
[testEnv](https://godoc.org/github.com/kubernetes-sigs/controller-runtime/pkg/envtest) or on an already existing
Kubernetes cluster (or local kind cluster). The default is to use testEnv. In order to run the tests in a local cluster
the following environment variables can be set.:
```bash
NO_SIMULATED_PROGRESS=true USE_EXISTING_CLUSTER=true make -C manager test
```

Please be aware that the controller is running locally in this case! If a controller is already deployed onto the
cluster then the tests can be run with the command below. This will ensure that the tests are only creating CRDs on 
the cluster and checking their status.
```bash
USE_EXISTING_CONTROLLER=true NO_SIMULATED_PROGRESS=true USE_EXISTING_CLUSTER=true make -C manager test
```

- USE_EXISTING_CLUSTER: (true/false)
  This variable controls if an existing K8s cluster should be used or not.
  If not testEnv will spin up an artificial environment that includes a local etcd setup.
- NO_SIMULATED_PROGRESS: (true/false)
  This variable can be used by tests that can manually simulate progress of e.g. jobs or pods.
  e.g. the simulated test environment from testEnv does not progress pods etc while when testing against
  an external Kubernetes cluster this will actually run pods.
- USE_EXISTING_CONTROLLER: (true/false)
  This setting controls if a controller should be set up and run by this test suite or if an external one
  should be used. E.g. in integration tests running against an existing setup a controller is already existing
  in the Kubernetes cluster and should not be started by the test as two controllers competing may influence the test.

Setup default local kind cluster with a local image registry:
```bash
make kind
```

By default the docker setup points to the official docker.io registry. When developing
locally and building/pushing docker containers the docker host and namespace should be changed. When 
using a local kind setup with a local registry the environment can be pointed to it by setting the following 
environment variables. (Please note that you also have to add an entry to your /etc/hosts file):

```bash
export DOCKER_HOSTNAME=kind-registry:5000
export DOCKER_NAMESPACE=m4d-system
export HELM_EXPERIMENTAL_OCI=1
```

There are make commands for building and pushing docker images separately or in one go:
```bash
make docker-build  # Only build docker images
make docker-push   # Only push docker images to registry defined with env $DOCKER_HOSTNAME

make docker        # Build and push images to the registry defined with env $DOCKER_HOSTNAME
```

Deploy on the cluster. This will install the CRDs, dependencies such as the certificate manager and the controller
itself. The default will pull the images from docker.io. If a local development setup is used please make sure
that $DOCKER_HOSTNAME is set to the registry that should be used. 

```bash
make deploy
```

Running end to end tests:
```bash
make e2e
```

# Building in a multi cluster environment

As {{< name >}} can run in a multi-cluster environment there is also a test environment
that can be used that simulates this scenario. Using kind one can spin up two separate kubernetes
clusters with differnt contexts and develop and test in these. 

Two kind clusters that share the same kind-registry can be set up using:
```bash
make kind-setup-multi
``` 

# Normalize the code

To ensure the code is formatted uniformly we use various linters which are
invoked using

```bash
make verify
```

# Format of the Commit Message

The project follows a rough convention for commit messages that is designed to answer two questions: what changed and why.
The subject line should feature the what and the body of the commit should describe the why.

Every commit must also include a DCO Sign Off at the end of the commit message. By doing this you state that you certify the [Developer Certificate of Origin](https://developercertificate.org/). This can be automated by adding the `-s` flag to `git commit`. You can also mass sign-off a whole PR with `git rebase --signoff master`.

Example commit message:
```
scripts: add the test-cluster command

this uses tmux to setup a test cluster that you can easily kill and
start for debugging.

Fixes #38

Signed-off-by: Legal Name <your.email@example.com>
```

The format can be described more formally as follows:

```
<subsystem>: <what changed>
<BLANK LINE>
<why this change was made>
<BLANK LINE>
<footer>
<BLANK LINE>
<signoff>
```

The first line is the subject and should be no longer than 70 characters, the second line is always blank, and other lines should be wrapped at 80 characters.
This allows the message to be easier to read on GitHub as well as in various git tools.
