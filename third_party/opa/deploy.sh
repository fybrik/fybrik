#!/usr/bin/env bash

set -e
set -x

NAMESPACE=fybrik-system
TIMEOUT=8m

check_valid_data_folder(){
    count=`ls -1 *.json 2>/dev/null | wc -l`
    count="$(echo -e "${count}" | tr -d '[:space:]')"
    local retVal=1
    if [ $count == 0 ]
    then
        retVal=0
    fi
    echo $retVal
}

check_valid_policy_folder(){
    count=`ls -1 *.rego 2>/dev/null | wc -l`
    count="$(echo -e "${count}" | tr -d '[:space:]')"
    local retVal=1
    if [ $count == 0 ]
    then
        retVal=0
    fi
    echo $retVal
}

unloadpolicy() {
    cd $1
    retVal=$(check_valid_policy_folder)
    if [ $retVal -eq 1 ];
    then
        policyfolder="${1##*/}"
        echo $policyfolder
        kubectl delete configmap $policyfolder --namespace=$NAMESPACE
    else
        echo "$1 is not a valid policy folder"
    fi
    cd -
}

unloaddata() {
    cd $1
    retVal=$(check_valid_data_folder)
    if [ $retVal -eq 1 ];
    then
        policydatafolder="${1##*/}"
        echo $policydatafolder
        kubectl delete configmap $policydatafolder --namespace=$NAMESPACE
    else
        echo "$1 is not a valid data folder"
    fi
    cd -
}

loadpolicy(){
    cd $1
    retVal=$(check_valid_policy_folder)
    if [ $retVal -eq 1 ];
    then
        policyfolder="${1##*/}"
        echo $policyfolder
        kubectl create configmap $policyfolder --from-file=./ --namespace=$NAMESPACE
        kubectl label configmap $policyfolder openpolicyagent.org/policy=rego --namespace=$NAMESPACE
    else
        echo "$1 is not a valid policy folder"
    fi
    cd -
}

loaddata(){
    cd $1
    retVal=$(check_valid_data_folder)
    if [ $retVal -eq 1 ];
    then
        policydatafolder="${1##*/}"
        echo $policydatafolder
        kubectl create configmap $policydatafolder --from-file=./ --namespace=$NAMESPACE
        kubectl label configmap $policydatafolder openpolicyagent.org/data=opa --namespace=$NAMESPACE
    else
        echo "$1 is not a valid data folder"
    fi
    cd -
}

case "$1" in
    loadpolicy)
        loadpolicy "$2"
        ;;
    loaddata)
        loaddata "$2"
        ;;
    unloadpolicy)
        unloadpolicy "$2"
        ;;
    unloaddata)
        unloaddata "$2"
        ;;
    *)
        echo "usage: $0 [deploy|undeploy|loadpolicy <policydir>|loaddata <datadir>|unloadpolicy <policydir>|unloaddata <datadir>]"
        exit 1
        ;;
esac
