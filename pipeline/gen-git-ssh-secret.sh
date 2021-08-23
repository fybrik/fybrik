#!/bin/bash

ssh_key=${1:-$HOME/.ssh/id_rsa}
github_host=${2:-github.com}

kubectl delete secret git-ssh-key
kubectl create secret generic git-ssh-key --from-file=ssh-privatekey=${ssh_key} --type=kubernetes.io/ssh-auth
kubectl annotate secret git-ssh-key --overwrite 'tekton.dev/git-0'="${github_host}"
kubectl secrets link pipeline git-ssh-key --for=mount
