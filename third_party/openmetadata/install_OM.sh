#!/bin/bash

OPERATION=install
OM_VERSION=0.12.1
K8S_TYPE=kind
export OPENMETADATA_AIRFLOW_VERSION=0.12.1
export OPENMETADATA_HELM_CHART_VERSION=0.0.39

export FYBRIK_BRANCH="${FYBRIK_BRANCH:-master}"
export FYBRIK_GITHUB_ORGANIZATION="${FYBRIK_GITHUB_ORGANIZATION:-fybrik}"

if [ $# -gt 3 ]; then
    echo "Usage: . ./install_OM.sh [install/getFiles/cleanup] [ibm-openshift/kind] [open-metadata-version]"
    return
fi

if [ $# -ge 1 ]; then
    OPERATION=$1
fi

if [ $# -ge 2 ]; then
    K8S_TYPE=$2
fi

if [ $# -ge 3 ]; then
    OM_VERSION=$3
fi

supported_operations=(install getFiles cleanup)
if [[ ! ${supported_operations[*]} =~ ${OPERATION} ]]; then
    echo supported operations: ${supported_operations[*]}
    echo ${OPERATION} not supported
    return
fi

if [ $OPERATION == "cleanup" ]; then
    if [ -z ${TMP_OM_INSTALLATION_DIR+x} ]; then
        echo could not find temporary installation dir. exiting
        return
    fi
    rm -Rf ${TMP_OM_INSTALLATION_DIR}
    echo removed directory ${TMP_OM_INSTALLATION_DIR}
    return
fi

supported_om_version=(0.12.1)
if [[ ! ${supported_om_version[*]} =~ ${OM_VERSION} ]]; then
    echo supported OM versions are: ${supported_om_version[*]}
    return
fi

# create temp directory
tmp_dir=$(mktemp -d)
echo about to download installation files to $tmp_dir

# download files to temp directory
if [ $K8S_TYPE == "ibm-openshift" ]; then
    export OPENSHIFT_INSTALLATION=true
    files_to_download=(Makefile Makefile.env pv1.yaml pv2.yaml ibm-openshift/pvc1.yaml ibm-openshift/pvc2.yaml ibm-openshift/pvc3.yaml ibm-openshift/pvc4.yaml values-deps.yaml)
else
    files_to_download=(Makefile Makefile.env pv1.yaml pvc1.yaml pv2.yaml pvc2.yaml values-deps.yaml)
fi
for file in "${files_to_download[@]}"
    do base=$(basename $file)
    curl https://raw.githubusercontent.com/${FYBRIK_GITHUB_ORGANIZATION}/fybrik/${FYBRIK_BRANCH}/third_party/openmetadata/$file -o $tmp_dir/$base
done

if [ $OPERATION == "getFiles" ]; then
    echo downloaded installation files to $tmp_dir
    echo to compile, go to the $tmp_dir directory,
    echo edit Makefile.env , and then run 'make'
    echo when you are done, re-run this script with the 'cleanup' parameter:
    echo "    ". ./install_OM.sh cleanup
    TMP_OM_INSTALLATION_DIR=$tmp_dir
    return
fi

# install OM + prepare OM for Fybrik
cd $tmp_dir
make

# cleanup
cd -
rm -Rf $tmp_dir
