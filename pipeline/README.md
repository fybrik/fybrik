# Tekton pipeline 

[vscode tekton pipelines extension doc](https://github.com/redhat-developer/vscode-tekton)

## Bootstrapping

Initial install for a cluster must happen in fybrik-system (currently).  It won't hurt anything if you aren't sure, and reinstall in fybrik-system.
```
. source-external.sh
bash -x bootstrap-pipeline.sh fybrik-system
# follow on screen instructions
```

Subsequent installs can go in any namespace
```
. source-external.sh
bash -x bootstrap-pipeline.sh fybrik-myname
# follow on screen instructions
```

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
