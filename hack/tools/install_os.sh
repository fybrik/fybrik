#!/bin/bash

set -x
set -e

ID=$(lsb_release -i)
ID=${ID##Distributor ID:[[:space:]]}
RELEASE=$(lsb_release -r)
RELEASE=${RELEASE##Release:[[:space:]]}

install_ubuntu() {
	sudo apt update
	sudo apt install -y python3
	sudo apt install -y python3-pip
	sudo apt install -y python3-setuptools
	sudo apt install -y asciidoctor
	sudo apt install -y conntrack
	sudo apt install -y ethtool
	sudo apt install -y apache2-utils
	sudo apt install -y maven
	sudo apt install -y docker.io
}

case $ID in
	Ubuntu)
		install_ubuntu
		;;
	*)
		echo "$id not supported!"
		exit 1
esac
