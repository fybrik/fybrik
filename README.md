[![GitHub Actions Build](https://github.com/fybrik/fybrik/actions/workflows/build.yml/badge.svg)](https://github.com/fybrik/fybrik/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/fybrik/fybrik)](https://goreportcard.com/report/github.com/fybrik/fybrik)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

# Fybrik

[Fybrik](https://fybrik.io/) is a cloud native platform to control the usage of data within an organization and thus enable business agility while securing enterprise data. Fybrik encapsulates a containerized workload and intermediates all data flows both into the workload and out of the workload, enabling enforcement of policies defining what can be done with the data.

For more information please visit [our website](https://fybrik.io/).

## Repositories

Fybrik is composed of the following repositories:

- [fybrik/fybrik](https://github.com/fybrik/fybrik). This is the main code repository. It hosts the core components, install artifacts, and sample programs. It includes:

  - [manager](manager) This directory contains the code for the main operator that is responsible for the control plane of Fybrik.

- [fybrik/charts](https://github.com/fybrik/charts) — Helm charts for deploying Fybrik.

- [fybrik/arrow-flight-module](https://github.com/fybrik/arrow-flight-module). This is the code repository for read/write data access module based on Arrow/Flight.

- [fybrik/mover](https://github.com/fybrik/mover). This is the code respository for an implicit copy module based on Apache SparkSQL.

Other modules and connectors are maintained under the Fybrik organization (for example: `openmetadata-connector`, `airbyte-module`, `trino-module`, `dremio-module`).  

See the full list at: https://github.com/fybrik

## Issue management

We use GitHub issues to track all of our bugs and feature requests.
