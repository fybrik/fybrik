#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0


op=$1

source ./common.sh

case "$op" in
    cleanup)
        header_text "Uinstalling minikube cluster"
        bin/minikube version
        sudo -E $(which bin/minikube) delete
        rm -rf $HOME/.minikube
        ;;
    *)
        header_text "Installing minikube cluster"
        driver=$(sudo docker info --format '{{print .CgroupDriver}}')
        args="--vm-driver=none --extra-config=kubelet.cgroup-driver=$driver"
        bin/minikube version
        sudo -E $(which bin/minikube) start $args
        sudo -E chown -R $USER $HOME/.kube $HOME/.minikube
        bin/kubectl config use-context minikube
        ;;
esac
