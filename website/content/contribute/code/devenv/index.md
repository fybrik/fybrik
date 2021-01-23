---
title: Development Environment
summary: This page contains instructions for setting up a development environment 
weight: 2
---

This page describes what you need to install as a developer and contributor to this project, for setting up a development environment.

# Operating system

Currently, only the Linux and Mac OS operating systems are officially supported. 
If you have a Windows machine consider using Windows Subsystem for Linux (WSL 2), 
a remote Linux machine, or any other solution such as a virtual machine.

# Dependencies

Install the following on your machine:

1. [go 1.13.x](https://golang.org/dl/)
1. [Docker](https://docs.docker.com/get-docker/)
1. `make`
1. `jq`
1. `unzip`
1. Maven (`mvn`) 
1. Java Development Kit version 8 or above
1. **Mac only**: `brew install coreutils` (installs the timeout command)

The project has other dependencies as well, but you don't need to manually install them. Running `make install-tools` installs all other dependencies to `hack/tools/bin`. The `make` targets (e.g., `make test`) are configured to use the binaries from `hack/tools/bin`. However, you may want to add some of these tools to your system PATH for direct usage from your terminal (e.g., for using `kubectl`).

# Editors

The project is predominantly written in Go so we recommend [Visual Studio Code](https://marketplace.visualstudio.com/items?itemName=ms-vscode.Go) for its good Go support. Alternatively you can select from [Editors](https://golang.org/doc/editors.html)


