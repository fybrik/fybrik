[![Build Status](https://travis.ibm.com/data-mesh-research/datamesh.svg?token=SFs8yc7zrXxhyzzSs8R8&branch=master)](https://travis.ibm.com/data-mesh-research/datamesh "Travis")


# The Mesh for Data

The Mesh for Data is a cloud native platform to control the usage of data within an organization and thus enble business agility while securing enterprise data.  The Mesh for Data encapsulate a containerized workload and intermediate all data flows both into the workload and out of the workload, enabling enforcement of policies defining what can be done with the data.

- For more information and details please visit [our website](https://pages.github.com/IBM/the-mesh-for-data/)

## Repositories

The Mesh for Data is composed of the following repositiores

- [IBM/the-mesh-for-data](https://github.com/ibm/the-mesh-for-data). This is the main code repository. It hosts the core components, install artifacts, and sample programs. It includes:

  - [manager](manager) This directory contains the code for the main operator that is responsible for the control plane of The Mesh for Data.

- [IBM/the-mesh-for-data-flight-module](https://github.com/IBM/the-mesh-for-data-flight-module). This is the code repository for read/write data access module based on Arrow/Flight.

- [IBM/the-mesh-for-data-mover](https://github.com/IBM/the-mesh-for-data-mover). This is the code respository for an implicit copy module based on Apache SparkSQL.

## Issue management

We use GitHub issues to track all of our bugs and feature requests.
