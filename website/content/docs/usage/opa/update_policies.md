---
title: Using OPA
linktitle: Using OPA
weight: 30
---

There are several ways how to add/update policies and data of the OPA service. 
They all are documented in the opa documentation: https://www.openpolicyagent.org/docs/latest/management/

This tutorial shows how to update policies using Configmap update, which is the recomended approach for new users due to its simplicity and the fact that the update will not be lost if pods are restarted. The downside of this approach is that there is a limitation of 1 MB for the size of the mapconfig, so this is also the limitation of the size on all the OPA data and policies together.

There are other ways to update the policies, e.g. REST API, bundles, etc., that are not covered by this tutorial.

### Configmap update

This approach is based on the fact that OPA is deployed on kuberenetes and stores both data and policies in configmap.
To add new policies/data or update existing ones using this approach we should combine all the data and policies (the new and the old ones together in case of update) into one configmap and upload this configmap into the cluster.

Steps to update data and policies:
1. Put all policies and OPA data files into a single folder, for example opa_files

2. Create single configmap and upload it to kubernetes:
    ```bash
    kubectl create configmap opa-policy --from-file=opa_files/ --dry-run -o yaml | kubectl replace  -f -
    ```

3.  Restart OPA pod:
    ```bash
    kubectl rollout restart deployment.apps/opa
    ```

