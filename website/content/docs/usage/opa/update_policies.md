---
title: Using OPA
linktitle: Using OPA
weight: 30
---

There are several ways how to add/update policies and data for OPA service. 
They all are documented in opa documentation: https://www.openpolicyagent.org/docs/latest/management/

This tutorial shows how to update policies using Configmap update, this way is the most recomended for new users for its simplicity and as it does not require state and the update will not be lost if pods are restarted. Downside of this approach is that there is a limitation of 1 MB for the size of the mapconfig, so this is also the limitation of the size on all data and policies together.

There are other ways to update the policies, e.g. REST API, bundles, etc., that are not covered by this tutorial.

### Configmap update

This approach based on the fact that OPA deployed on kuberenetes stores both data and policies in configmap.
To add new policies/data or update existing one using this approach we should combine all the data and policies (new and the old one together in case of update) into one configmap and upload this configmap into the cluster.

Steps to update data and policies:
1. Put all policies and data files into a single folder, for example opa_files

2. Create single configmap and upload it to kubernetes:
    ```bash
    kubectl create configmap opa-policy --from-file=opa_files/ --dry-run -o yaml | kubectl replace  -f -
    ```

3.  Restart OPA pod:
    ```bash
    kubectl rollout restart deployment.apps/opa
    ```

