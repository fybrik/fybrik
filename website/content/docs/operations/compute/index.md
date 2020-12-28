---
title: "Registering Compute Resources"
date: 2020-05-03T21:49:19+03:00
draft: true
weight: 100
---

<!-- Not implemented nor designed yet -->

Currently {{< name >}} does not support multicluster operation. The administrator needs to register the _OpenShift Projects_ (namespaces) that the {{< name >}} control plane operates on. This is done by creating `M4DMemberRoll` resources in the `m4d-system` project. For example, the following adds the fraudanalysis project to {{< name >}}:

```yaml
apiVersion: admin.m4d.ibm.com/v1alpha1
kind: M4DMemberRoll
metadata:
    name: default
spec:
    members:
    # a list of projects joined into the control plane
    - fraudanalysis
```

