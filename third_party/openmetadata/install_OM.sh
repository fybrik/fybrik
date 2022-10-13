#!/bin/bash

OM_VERSION=0.12.1
export OPENMETADATA_AIRFLOW_VERSION=0.12.1
export OPENMETADATA_HELM_CHART_VERSION=0.0.39

export FYBRIK_BRANCH="${FYBRIK_BRANCH:-master}"
export FYBRIK_GITHUB_ORGANIZATION="${FYBRIK_NAMESPACE:-fybrik}"

if [ $# -gt 1 ]; then
    echo "Usage: "$0" [open-metadata-version]"
    exit -1
fi

if [ $# -eq 1 ]; then
    OM_VERSION=$1
fi

supported_om_version=(0.12.1)
if [[ ! ${supported_om_version[*]} =~ ${OM_VERSION} ]]; then
    echo supported OM versions are: ${supported_om_version[*]}
    exit -1
fi

# create temp directory
tmp_dir=$(mktemp -d)

# download files to temp directory
cd $tmp_dir

files_to_download=(Makefile Makefile.env pv1.yaml pv2.yaml values-deps.yaml)
for file in "${files_to_download[@]}"
    do curl https://raw.githubusercontent.com/${FYBRIK_GITHUB_ORGANIZATION}/fybrik/${FYBRIK_BRANCH}/third_party/openmetadata/$file -o $file
done

# install OM + prepare OM for Fybrik
make

# cleanup
cd -
rm -Rf $tmp_dir
