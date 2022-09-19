#!/bin/bash

OM_VERSION=0.12.0
OPENMETADATA_AIRFLOW_VERSION=0.12.0
OPENMETADATA_HELM_CHART_VERSION=0.0.36

if [ $# -gt 1 ]; then
    echo "Usage: "$0" [open-metadata-version]"
    exit -1
fi

if [ $# -eq 1 ]; then
    OM_VERSION=$1
fi

supported_om_version=(0.12.0 0.11.4)
if [[ ! ${supported_om_version[*]} =~ ${OM_VERSION} ]]; then
    echo supported OM versions are: ${supported_om_version[*]}
    exit -1
fi

if [ "$OM_VERSION" = "0.11.4" ]; then
    OPENMETADATA_AIRFLOW_VERSION=0.11.4
    OPENMETADATA_HELM_CHART_VERSION=0.0.34
fi

# create temp directory
tmp_dir=$(mktemp -d)

# download files to temp directory
cd $tmp_dir

files_to_download=(Makefile Makefile.env pv1.yaml pv2.yaml values-deps.yaml go.mod go.sum prepare_OM_for_fybrik.go)
for file in "${files_to_download[@]}"
    do curl https://raw.githubusercontent.com/fybrik/fybrik/master/third_party/openmetadata/$file -o $file
done

# install OM
make

kubectl port-forward svc/openmetadata -n open-metadata 8585:8585 &
sleep 2
JOB=$(jobs | grep "kubectl port-forward svc/openmetadata -n open-metadata 8585:8585" | cut -d"[" -f2 | cut -d"]" -f1)

go run .

kill %$JOB

# cleanup
cd -
rm -Rf $tmp_dir
