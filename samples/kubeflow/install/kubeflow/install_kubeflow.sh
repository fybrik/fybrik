#!/usr/bin/env bash

set -e

if [[ ! -f ./kfctl ]]; then
  wget https://github.com/kubeflow/kfctl/releases/download/v1.0.2/kfctl_v1.0.2-0-ga476281_linux.tar.gz
  tar -zxvf kfctl_v1.0.2-0-ga476281_linux.tar.gz
  rm kfctl_v1.0.2-0-ga476281_linux.tar.gz
fi

cp kubeflow.yaml kfctl_k8s_istio.v1.0.2.yaml
./kfctl apply -V -f kfctl_k8s_istio.v1.0.2.yaml

echo "Done"
