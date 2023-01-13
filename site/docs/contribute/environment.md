# Development Environment

This page describes what you need to install as a developer and contributor to this project, for setting up a development environment.

## Operating system

Linux and Mac OS operating systems are officially supported.

Windows users should consider using Windows Subsystem for Linux 2 (WSL 2), 
a remote Linux machine, or any other solution such as a virtual machine.

## Dependencies

Install the following on your machine:

1. [go](https://golang.org/dl/) 1.19
1. [Docker](https://docs.docker.com/get-docker/)
1. `make`
1. `jq`
1. `unzip`
1. **Mac only**: `brew install coreutils` (installs the timeout command)


Then, run the following command to install additional dependencies:

```bash
make install-tools
```

This installs additional dependencies to `hack/tools/bin`. The `make` targets (e.g., `make test`) are configured to use the binaries from `hack/tools/bin`. However, you may want to add some of these tools to your system PATH for direct usage from your terminal (e.g., for using `kubectl`).

Please note: For fybrik version 0.5.x and lower, Helm version greater than 3.3 but less than 3.7 is required when contributing. 
On the other side, for fybrik v0.6.x, Helm v3.7 or above is required. 

## Editors

The project is predominantly written in Go, so we recommend [Visual Studio Code](https://marketplace.visualstudio.com/items?itemName=ms-vscode.Go) for its good Go support. Alternatively you can select from [Editors](https://golang.org/doc/editors.html)

## Docker hub rate limits

As docker hub introduced rate limits on docker image downloads this may affect development using the local kind setup.
One option to fix the limit is to use a docker hub login for downloading the images. The environment will run
a docker registry as a proxy for all public images. This registry runs in a docker container next to the kind clusters. 

```shell
export DOCKERHUB_USERNAME='your docker hub username'
export DOCKERHUB_PASSWORD='your password'
```

## Optional Code Improvement and Verification Tools

Fybrik repositories use different linters to validate and improve code.  
*golangci-lint* works as a Go linters aggregator, which includes a lot of linter such as `staticcheck`, `revive`, `goimports` and more.  
You can check its configuration in the `.golangci.yml` files.


### Integrating golangci-lint with VS Code
Change the default lint to `golanci-lint` in VS Code:

1. Install golangci-lint: [https://golangci-lint.run/usage/install/](https://golangci-lint.run/usage/install/)
2. Open VS Code `setting.json`:
    1. Open the Command Palette: `Ctrl+Shift+P` 
    2. In the dropdown search box, search for "Open Settings (JSON)"
    3. Open `setting.json`
3. Add to `setting.json` the following:
```
"go.lintTool":"golangci-lint",
"go.lintFlags": [
  "--fast",
  "--allow-parallel-runners"
]
```

Golangci-lint automatically discovers `.golangci.yml` in the working project, you don't need to configure it in VS Code settings.

To integrate with other IDEs: [https://golangci-lint.run/usage/integrations/](https://golangci-lint.run/usage/integrations/)  
If you wish to run golangci-lint on cmd, run in the desired directory:  
```
golangci-lint run --fast
```


### Pre-commit

`pre-commit` is an optional tool that inspect the snapshot that's about to be committed according to the configured hooks, in our case, `golangci-lint`.  
Pre-commit configuration is in `.pre-commit-config.yaml`

How to use:

1. Install pre-commit: [https://pre-commit.com/](https://pre-commit.com/)
2. In the repository, run:  
```
pre-commit install
```

Now, `pre-commit` will automatically validate all your commits.

To run commits without `pre-commit` validation add the `--no-verify` flag to `git commit`.