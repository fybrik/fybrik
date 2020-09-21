---
title: Write documentation
description: Details how to contribute new documentation.
date: 2020-04-26T22:00:53+03:00
draft: false
weight: 2
---

## Prerequisites

The page assumes that you [installed Hugo](../build/) and have a copy of the {{< website >}} repository files, following the [contribution workflow](../workflow/) guidelines. 

## Create a document of the right type

The documentation of the project contains several types of documents, addressing different needs or audiances.


| Type                 | Description                                                            |
| -------------------- | ---------------------------------------------------------------------- |
| Architecture         | High level description of the architecture and features of the project | 
| Setup                | Installation instructions                                              |
| Operations           | Managing and monitoring the {{< name >}} configuration and runtime     | 
| Usage                | Usage guides targeting readers with a specific role (e.g., a governance officer) |
| Component            | Component documents describing the internals of a specific feature     |
| Reference            | Auto generated API documentation                                       |


Reference documents are genereted directly from the protobuf and Kubernetes CRD definitions that must be well documented as part of the code contribution process. 

You create non-reference documents with Hugo. To create a new document of a specific type use:
```shell 
hugo new -k <type> <path>
``` 

For example, to create a (cool) usage guide use:
  ```plain 
  hugo new -k usage docs/usage/cool-guide/index.md
  ```

## Add content to the document

The content of the document must be written with a specific target audiance in mind. The level of details and the structure of the document must match the document type and the expectations of a typical reader. Avoid being too abstract and write concrete useful text instead.

To help writing good documentation, creating a new document generetes a template with guidelines specific to the chosen document type. Use those guidelines to learn, but you typically don't need to keep any of the genereted text. 

#### Follow the following guidelines when writing documentation
