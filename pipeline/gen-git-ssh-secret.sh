#!/bin/bash

ssh_key=${1:-$HOME/.ssh/id_rsa}
github_host=${2:-github.ibm.com}

oc delete secret git-ssh-key
oc create secret generic git-ssh-key --from-file=ssh-privatekey=${ssh_key} --type=kubernetes.io/ssh-auth
oc annotate secret git-ssh-key --overwrite 'tekton.dev/git-0'="${github_host}"
oc secrets link pipeline git-ssh-key --for=mount
