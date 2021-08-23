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

From command line, it can be done with this script:
```
# ./restart-task.sh run-integration-tests
# # Where run-integration-task is a grep match for the task you want to restart
```

knative eventing is used to restart all downstream tasks.  Code rebuilds will trigger image rebuilds.  Image rebuilds will trigger helm upgrades.

## Building code from your development environment

Set the following environment variable to point at your code repo, and it will be copied to the volume tekton tasks will mount rather than cloning code from github
```
github_workspace="/path/to/fybrik"
. source-external.sh
```
