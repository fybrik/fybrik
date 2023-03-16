#!/bin/bash

# Copyright 2023 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

# error and exit configuration
set -eu

FLAG_DURATION="-d|--duration"
FLAG_NAMESPACES="-n|--namespaces"
FLAG_PATH="-p|--path"
FLAG_UUID="-u|--uuid"
LONG_POS=3

# define the possible flag options
VALID_ARGS=$(getopt -o d:n:p:u:h --long duration:,namespaces:,path:,uuid:,help -n $(basename $0) -- "$@")

# set the default values of the flags
duration=1h
namespaces_list=()
for ns in $(kubectl get ns -o=jsonpath='{.items[*].metadata.name}'); do
    if [[ $ns == *"fybrik"* ]]; then
        namespaces_list+=("$ns")
    fi
done
output_path=$(readlink -f .)/
fybrik_application_uuid=""

function print_description {
    echo -e "A script for collecting relevant fybrik logs and information into a local directory, using kubectl.\n"
}

function print_usage {
    echo -e "Usage:\t\t$(basename $0) [$FLAG_DURATION <arg>] [$FLAG_NAMESPACES <ns1,ns2,...>] [$FLAG_PATH <arg>] [$FLAG_UUID <app-uuid>]";
    echo -e "Default:\t$(basename $0) ${FLAG_DURATION:$LONG_POS} 1h ${FLAG_NAMESPACES:$LONG_POS} <any namespace with the word 'fybrik'> ${FLAG_PATH:$LONG_POS} ./";
}

function print_flags {
    echo -e "Flags:"
    echo -e "\t$FLAG_DURATION\t the relative duration of time (e.g. 5s, 2m, 3h), only logs newer than the duration will be saved in the output. Default: 1h"
    echo -e "\t$FLAG_NAMESPACES\t comma-separated namespaces, only resources from those namespaces will be saved in the output. Default: all the namespaces that contains the word 'fybrik'"
    echo -e "\t$FLAG_PATH\t a path of an existant directory in which the output directory will be created. Default: current working directory"
    echo -e "\t$FLAG_UUID\t an app.fybrik.io/app-uuid for saving its yaml if the resource exists in the namespaces, and creating an additional file of the logs containing the uuid. Default: none"
}

eval set -- "$VALID_ARGS"
while [ : ]; do
    case "$1" in
        -d | --duration)
            duration=$2
            # check that the argument is valid
            re='^[0-9]+[smh]$'
            if [[ ! $duration =~ $re ]]; then
                echo "error: invalid argument '$duration' for -d|--duration flag: should contain a number and a time unit (s/m/h), e.g. 1h or 30m"
                print_usage
                exit 1
            fi
            shift 2
            ;;
        -n | --namespaces)
            readarray -d , -t namespaces_list <<< $2
            # check that all the given namespaces exist
            existant_ns=$(kubectl get ns -o=jsonpath='{.items[*].metadata.name}')
            for ns in ${namespaces_list[@]}; do
                if [[ ! " ${existant_ns[@]} " =~ " ${ns} " ]]; then
                    echo "error: given namespace '$ns' doesn't exist"
                    print_usage
                    exit 1
                fi
            done
            shift 2
            ;;
        -p | --path)
            output_path=$(readlink -f $2)/
            if [[ ! -d $output_path ]]; then
                echo "error: given path '$output_path' doesn't exist"
                print_usage
                exit 1
            fi
            shift 2
            ;;
        -u | --uuid)
            fybrik_application_uuid=$2
            shift 2
            ;;
        -h | --help)
            print_description
            print_usage
            print_flags
            exit 0
            ;;
        --) shift; 
            break 
            ;;
    esac
done

# create a directory for saving output
dirname_logs=$output_path"fybrik_logs_$(date +%Y-%m-%d_%H-%M)_last_$duration"
echo "Output will be saved in $dirname_logs"
mkdir $dirname_logs
dirname_logs_by_container=$dirname_logs/logs_by_container
mkdir $dirname_logs_by_container

relevant_configmaps=("cluster-metadata" "fybrik-config")
is_found_fybrik_application_yaml=0

# iterate over the pods in the relevant namespaces
for ns in ${namespaces_list[@]}; do
    # save the logs of each container separately
    for pod in $(kubectl get pods -n $ns -o=jsonpath='{.items[*].metadata.name}'); do
        containers=$(kubectl get pods $pod -n $ns -o jsonpath="{.spec.containers[*].name}")
        for container_name in $containers; do
            kubectl logs $pod -n $ns -c $container_name --since=$duration --timestamps &> $dirname_logs_by_container/$ns--$pod--$container_name.txt
        done
    done
    # save relevant configmaps from relevant namespaces
    for cm in $(kubectl get cm -n $ns -o=jsonpath='{.items[*].metadata.name}'); do
        if [[ " ${relevant_configmaps[@]} " =~ " ${cm} " ]]; then
            cm_data=$(kubectl get cm $cm -n $ns -o=jsonpath='{.data}')s
            prefix=$ns--$cm
            echo -e "$prefix:\n$cm_data\n" >> $dirname_logs/configmaps.txt
        fi
    done
    # save output of 'get pods' command
    get_pods_output=$(kubectl get pods -n $ns 2>&1)
    echo -e "$ns:\n$get_pods_output\n" >> $dirname_logs/get_pods_output.txt
    # save yamls of deployed modules
    for fybrik_module in $(kubectl get fybrikmodules -n $ns -o=jsonpath='{.items[*].metadata.name}'); do
        kubectl get fybrikmodule $fybrik_module -n $ns -o yaml &> $dirname_logs/fybrik_module_$ns--$fybrik_module.yaml
    done
    # if given fybrik application uuid, save its yaml and create another file which only contains logs with the uuid
    for fybrik_application in $(kubectl get fybrikapplications -n $ns -o=jsonpath='{.items[*].metadata.name}'); do
        if [[ $(kubectl get fybrikapplication $fybrik_application -n $ns -o=jsonpath='{.metadata.uid}') == $fybrik_application_uuid ]]; then
            kubectl get fybrikapplication $fybrik_application -n $ns -o=yaml &> $dirname_logs/fybrik_application_$ns--$fybrik_application.yaml
            is_found_fybrik_application_yaml=1
        fi
    done
done

# get maximal prefix length for the combined logs file
max_len=0
for f in $dirname_logs_by_container/*.txt; do
    f_short=$(basename -s .txt $f)
    if [ ${#f_short} -gt $max_len ]; then
        max_len=${#f_short}
    fi
done

FILENAME_COMBINED=combined_logs
FILENAME_COMBINED_SORTED=combined_logs_sorted

# combine all the logs, with prefix of their ns--pod--container
for f in $dirname_logs_by_container/*.txt; do
    f_short=$(basename -s .txt $f)
    prefix=$(printf "%-${max_len}s\n" "$f_short")
    while read line; do
        echo -e "$prefix || $line" >> $dirname_logs/$FILENAME_COMBINED.txt
    done < $f
done

# if any logs are found, sort them by their timestamps and remove combined file
if [[ -e $dirname_logs/$FILENAME_COMBINED.txt ]]; then
    sort -k 3 $dirname_logs/$FILENAME_COMBINED.txt > $dirname_logs/$FILENAME_COMBINED_SORTED.txt
    rm $dirname_logs/$FILENAME_COMBINED.txt
else
    echo "no log entries were found in the pods in the explored namespaces"
fi

# if given fybrik application uuid, create another file which only contains logs with the uuid
if [[ ! -z $fybrik_application_uuid ]]; then
    # check if its yaml file was found
    if [[ $is_found_fybrik_application_yaml -eq 0 ]]; then
        echo "the fybrikapplication '$fybrik_application_uuid' was not found in the namespaces, its yaml file was not retrieved"
    fi
    if [[ -e $dirname_logs/$FILENAME_COMBINED_SORTED.txt ]]; then
        # create a file with the logs containing the uuid
        touch $dirname_logs/${FILENAME_COMBINED_SORTED}_$fybrik_application_uuid.txt
        while read line; do
            if [[ $line == *\"$fybrik_application_uuid\"* ]]; then
                echo "$line" >> $dirname_logs/${FILENAME_COMBINED_SORTED}_$fybrik_application_uuid.txt
            fi
        done < $dirname_logs/$FILENAME_COMBINED_SORTED.txt
    fi
fi
