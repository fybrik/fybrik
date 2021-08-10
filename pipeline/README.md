# Tekton pipeline 

[vscode tekton pipelines extension doc](https://github.com/redhat-developer/vscode-tekton)



## Bootstrapping

Run the following command to bootstrap a Fybrik instance. By default the instance of Fybrik will be deployed as namespaced scoped, meaning the controller will only watch and process FybrikApplications from a specific namespace. Therefore limiting you to create FybrikApplications from one namespace. By that namespace is the one specific on the bootstrap script (a.k.a *system* namespace). You can change this behavior by setting environment variables. For details see [Scoping the deployment](#scoping-the-deployment).

The parameter specified is the *system* namespace. 
```
. source-external.sh
bash -x bootstrap-pipeline.sh fybrik-myname
# follow on screen instructions
```

After running the "tkn" command that was returned from the bootstrap script, the following namespaces will be created. Illistrated by using the example *system* namespace of fybrik-myname:
1. fybrik-myname - This will be the *system* namespace where Fybrik is deployed. By default the controller will only process FybrikApplications from this namespace.
2. fybrik-myname-blueprints - this is the designated namespace for blueprints custom resources. This is also where the data access module is deployed to. 
3. fybrik-myname-app - This namespace is only created if running namespace scoped AND the following environment variable is set: use_application_namespace=true.  

## Restarting individual tasks

Tasks can be restarted in vscode by right-clicking on the task in the tekton pipelines extension.

From command line, it can be done with a series of commands like this:
```
# kubectl get taskrun build-and-deploy-run-run-integration-tests-full-deploy-zfwgc -o yaml > /tmp/taskrun.yaml
# vi /tmp/taskrun.yaml
# # delete metadata.name & metadata.namespace
# # add metadata.generateName: restarted-task-
# kubectl create -f /tmp/taskrun.yaml
```

knative eventing is used to restart all downstream tasks.  Code rebuilds will trigger image rebuilds.  Image rebuilds will trigger helm upgrades.

## Scoping the deployment
In the following section the *system* namespace is the namespace specified when running the bootstrap script.

By default the bootstrap script will deploy Fybrik namespaced scoped. Deploying as namespaced scoped allows multiple instances of Fybrik to be installed in unique namespaces in the same cluster.  You can change this behavior but you must do so with care.

### Environment variables
| Name  | Default  | Valid Values | Description |
|-------|----------|--------------|-------------|
| cluster_scoped | false | true, false | Indicates if deploy as cluster scoped meaning the controller will watch and process FybrikApplications from all namespaces. For caveats see [tips](#tips) |
| use_application_namespace | false | true, false | Specifies if you want the bootstrap script to create a namespace for FybrikApplications. This will be ignore if cluster_scoped=true. If set to false the *system* namespace will be used. |

### Tips
1. When deploying a Fybrik instance as cluster scoped - **Disclaimer: Only use cluster scope if you know what you’re doing.**
   1. There can only be one instance of Fybrik deployed in cluster. 
   2. The controller will be setup so it watches all namespaces for FybrikApplications.
   3. You cannot run namespaced scoped and cluster scoped in same cluster.

2. When deploying a Fybrik instance as namespaced scoped:
   1. You can only create FybrikApplication custom resources from one namespace. The default will be to allow only creating from the *system* namespace. You can override this behavior by setting the use_application_namespace environment variable to true, the bootstrap code will create the *system*-app namespace and necessary objects to allow creating FybrikApplications. Hint: You will see the application namespace in the -p mesh-for-data-values= in the tkn command that is returned from running the bootstrap script.

3. When using wkc-connector, the wkc-credentials secret must be in the same namespace you are using when creating FybrikApplication custom resources. The bootstrap script will unconditionally create a wkc-credentials secret in *system* namespace. 
   1. If running cluster scoped (i.e cluster_scoped=true) and you use a different namespace than *system* for your FybrikApplications, you will need to create a wkc-credentials secret in that namespace. 
   2. If running namespaced scoped (i.e cluster_scoped=false) and you set use_application_namespace=true, bootstrap will create the secret in *system*-app.
