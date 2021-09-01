#!/bin/bash

github_user=${1:?Need to specify the github (Enterprise) username}
pat=${2:?Need to specify the personal access token}
github_host=${3:-github.com}

kubectl delete secret git-pat
kubectl create secret generic git-pat --from-literal=username=${github_user} --from-literal=password=${pat} --type=kubernetes.io/basic-auth
kubectl annotate secret git-pat --overwrite 'tekton.dev/git-0'="https://${github_host}"
kubectl secrets link pipeline git-pat --for=mount
