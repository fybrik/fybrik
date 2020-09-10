---
title: "Storage"
date: 2020-05-03T21:51:02+03:00
draft: true
weight: 100
---

<!-- Not implemented nor designed yet -->

{{< name >}} uses storage resources for two main reasons:
1. Storing internal implicit data copies
2. Providing storage for an application to use

In both cases the storage is owned by the {{< name >}} control plane. This enables full control of data lifecycle management, access credentials, etc.

To create a storage resource {{< name >}} currently uses [Red Hat OpenShift Container Storage 4](https://www.openshift.com/products/container-storage/) for file storage, block storage and object storage.

