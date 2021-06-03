[![GitHub Actions Build](https://github.com/mesh-for-data/mesh-for-data/actions/workflows/build.yml/badge.svg)](https://github.com/mesh-for-data/mesh-for-data/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mesh-for-data/mesh-for-data)](https://goreportcard.com/report/github.com/mesh-for-data/mesh-for-data)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

# Mesh for Data

[Mesh for Data](https://mesh-for-data.github.io/mesh-for-data/) is a cloud native platform to control the usage of data within an organization and thus enable business agility while securing enterprise data. Mesh for Data encapsulates a containerized workload and intermediates all data flows both into the workload and out of the workload, enabling enforcement of policies defining what can be done with the data.

For more information please visit [our website](https://mesh-for-data.github.io/mesh-for-data/).

## Repositories

Mesh for Data is composed of the following repositiores

- [mesh-for-data/mesh-for-data](https://github.com/mesh-for-data/mesh-for-data). This is the main code repository. It hosts the core components, install artifacts, and sample programs. It includes:

  - [manager](manager) This directory contains the code for the main operator that is responsible for the control plane of Mesh for Data.

- [mesh-for-data/arrow-flight-module](https://github.com/mesh-for-data/arrow-flight-module). This is the code repository for read/write data access module based on Arrow/Flight.

- [mesh-for-data/mesh-for-data-mover](https://github.com/mesh-for-data/mesh-for-data-mover). This is the code respository for an implicit copy module based on Apache SparkSQL.

## Issue management

We use GitHub issues to track all of our bugs and feature requests.
