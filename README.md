[![Build Status](https://travis-ci.com/IBM/the-mesh-for-data.svg?branch=master)](https://travis-ci.com/IBM/the-mesh-for-data "Travis")
[![Go Report Card](https://goreportcard.com/badge/github.com/IBM/the-mesh-for-data)](https://goreportcard.com/report/github.com/IBM/the-mesh-for-data)

# Mesh for Data

[Mesh for Data](https://ibm.github.io/the-mesh-for-data/) is a cloud native platform to control the usage of data within an organization and thus enable business agility while securing enterprise data. Mesh for Data encapsulates a containerized workload and intermediates all data flows both into the workload and out of the workload, enabling enforcement of policies defining what can be done with the data.

For more information please visit [our website](https://ibm.github.io/the-mesh-for-data/).

## Repositories

Mesh for Data is composed of the following repositiores

- [IBM/the-mesh-for-data](https://github.com/IBM/the-mesh-for-data). This is the main code repository. It hosts the core components, install artifacts, and sample programs. It includes:

  - [manager](manager) This directory contains the code for the main operator that is responsible for the control plane of Mesh for Data.

- [IBM/the-mesh-for-data-flight-module](https://github.com/IBM/the-mesh-for-data-flight-module). This is the code repository for read/write data access module based on Arrow/Flight.

- [IBM/the-mesh-for-data-mover](https://github.com/IBM/the-mesh-for-data-mover). This is the code respository for an implicit copy module based on Apache SparkSQL.

## Issue management

We use GitHub issues to track all of our bugs and feature requests.
