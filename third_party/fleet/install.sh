#/usr/bin/env bash

set -x
set -e

FLEET_VERSION=0.3.0
MANAGER=k3d-manager
AGENT=k3d-agent
CLUSTER_LABELS="--set-string labels.env=dev"
API_SERVER_URL=$(echo $(hostname -I):8080 | sed -e 's/ //')

k3d_install() {
	which k3d 2>/dev/null && return
	curl -s https://raw.githubusercontent.com/rancher/k3d/main/install.sh | TAG=v3.0.0 bash
}

tools() {
	k3d_install
}

k3d_delete() {
	local name=$1
	k3d cluster delete $name 2>/dev/null || true
}

k3d_create() {
	local name=$1
	k3d cluster create --agents 2 $name
}

helm_uninstall() {
	local context=$1; shift
	local name=$1; shift
	helm --kube-context $context -n fleet-system uninstall $name 2>/dev/null || true
}

helm_install() {
	local context=$1; shift
	local name=$1; shift
	helm --kube-context $context -n fleet-system install --create-namespace --wait $* \
	    $name https://github.com/rancher/fleet/releases/download/v${FLEET_VERSION}/${name}-${FLEET_VERSION}.tgz 
}

cluster_status() {
	namespace=$1; shift
	kubectl --context $MANAGER get clusters -n $namespace
}

####################
##### manager ######
####################

manager_delete() {
	k3d_delete manager
}

manager_create() {
	k3d_create manager
}

manager_install() {
	helm_uninstall $MANAGER fleet
	helm_uninstall $MANAGER fleet-crd
	helm_install $MANAGER fleet-crd
	helm_install $MANAGER fleet ${CLUSTER_LABELS} \
		--set apiServerURL=${API_SERVER_URL}
	cluster_status fleet-local
}

manager_token() {
	local name=new-token
	local namespace=clusters

	cat <<EOF | kubectl apply --context $MANAGER --wait -f - 
apiVersion: v1
kind: Namespace
metadata:
  name: $namespace
---
kind: ClusterRegistrationToken
apiVersion: "fleet.cattle.io/v1alpha1"
metadata:
    name: $name
    namespace: $namespace
spec:
    ttl: 240h
EOF
	kubectl --context $MANAGER -n $namespace get secret $name -o 'jsonpath={.data.values}' | \
		base64 --decode > /tmp/values.yaml
	cat /tmp/values.yaml
}

manager_proxy_kill() {
	sudo killall -9 kubectl 2>/dev/null || true
}

manager_proxy() {
	kubectl proxy --accept-hosts '^.*' --address 0.0.0.0 --port 8080 --context $MANAGER&
}

manager_status() {
	kubectl --context $MANAGER -n fleet-system logs -l app=fleet-controller || true
	kubectl --context $MANAGER -n fleet-system get pods -l app=fleet-controller || true
}

manager() {
	manager_proxy_kill
	manager_delete
	manager_create
	manager_install
	manager_token
	manager_proxy
	manager_status
}

####################
#####  agent  ######
####################

agent_delete() {
	k3d_delete agent
}

agent_create() {
	k3d_create agent
}

agent_install() {
	helm_uninstall $AGENT fleet-agent
	helm_uninstall $AGENT fleet-crd
	helm_install $AGENT fleet-crd
	helm_install $AGENT fleet-agent ${CLUSTER_LABELS} --values /tmp/values.yaml
	cluster_status clusters
}

agent_status() {
	kubectl --context $AGENT -n fleet-system logs -l app=fleet-agent || true
	kubectl --context $AGENT -n fleet-system get pods -l app=fleet-agent || true
}

agent_cluster_status() {
	kubectl --context $MANAGER -n clusters get clusters.fleet.cattle.io
}

agent() {
	agent_delete
	agent_create
	agent_install
	agent_status
	agent_cluster_status
}

####################
#####   app   ######
####################

app_install() {
	local namespace=$1; shift
	cat <<EOF | kubectl apply --context $MANAGER -f -
kind: GitRepo
apiVersion: fleet.cattle.io/v1alpha1
metadata:
  name: simple
  namespace: $namespace
spec:
  repo: https://github.com/rancher/fleet-examples/
  paths:
  - multi-cluster/helm
  targets:
  - name: dev
    clusterSelector:
      matchLabels:
        env: dev
EOF
	sleep 5
	kubectl --context $MANAGER -n $namespace get fleet
}

app_status() {
	local context=$1; shift
	sleep 5
	local name=frontend
	local namespace=fleet-mc-helm-example
	kubectl --context $context rollout status deploy --namespace $namespace $name --timeout 30s || true
	kubectl --context $context get deploy --namespace $namespace $name
}

app_manager_status() {
	app_status $MANAGER
}

app_manager() {
	app_install fleet-local
}

app_agent_status() {
	app_status $AGENT
}

app_agent() {
	app_install clusters
}

app() {
	#app_manager
	#app_manager_status
	app_agent
	app_agent_status
}

####################
#####   app   ######
####################

cleanup() {
	manager_delete
	agent_delete
}

setup() {
	tools
	manager
	agent
	app
}

state() {
	#manager_status
	#app_manager_status
	agent_status
	app_agent_status
}

case "$1" in 
	status)
		state
		;;
	cleanup)
		cleanup
		;;
	*)
		setup
		;;
esac
