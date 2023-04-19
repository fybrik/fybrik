#!/bin/bash

# Copyright 2023 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

# error and exit configuration
set -eu

FLAG_DURATION="d"
FLAG_NAMESPACES="n"
FLAG_PATH="p"
FLAG_APP="a"
FLAG_UUID="u"
FLAG_HELP="h"
ARGS_SEPARATOR=","
APP_CONFLICT_MESSAGE="Only one of -$FLAG_APP/-$FLAG_UUID may be used"

# set initial values of the flags
duration=1h
namespaces_list=()
output_path=$(readlink -f .)/
fybrik_application_ns=""
fybrik_application_name=""
fybrik_application_uuid=""

function print_description {
    echo -e "A script for collecting relevant fybrik logs and information into a local directory, using kubectl.\n"
}

function print_usage {
    echo -e "Usage:\t\t$(basename $0) [-$FLAG_DURATION <duration>] [-$FLAG_NAMESPACES <ns1,ns2,...>] [-$FLAG_PATH <path>] [-$FLAG_APP <ns,app-name> OR -$FLAG_UUID <app-uuid>]";
    echo -e "Default:\t$(basename $0) -$FLAG_DURATION 1h -$FLAG_NAMESPACES <any namespace with the word 'fybrik'> -$FLAG_PATH ./";
}

function print_flags {
    echo -e "Flags:"
    echo -e "\t-$FLAG_DURATION\t the relative duration of time (e.g. 5s, 2m, 3h), only logs newer than the duration will be saved in the output. Default: 1h"
    echo -e "\t-$FLAG_NAMESPACES\t comma-separated namespaces, only resources from those namespaces will be saved in the output. Default: all the namespaces that contains the word 'fybrik'"
    echo -e "\t-$FLAG_PATH\t a path of an existant directory in which the output directory will be created. Default: current working directory"
    echo -e "\t-$FLAG_APP\t comma-separated namespace and name of fybrikapplication, for saving its yaml if the resource exists, and creating an additional file of the logs containing its uuid. Default: none. $APP_CONFLICT_MESSAGE"
    echo -e "\t-$FLAG_UUID\t an app.fybrik.io/app-uuid for saving its yaml if the resource exists in the namespaces, and creating an additional file of the logs containing the uuid. Default: none. $APP_CONFLICT_MESSAGE"
    echo -e "\t-$FLAG_HELP\t print this message and exit"
}

while getopts "$FLAG_DURATION:$FLAG_NAMESPACES:$FLAG_PATH:$FLAG_UUID:$FLAG_APP:$FLAG_HELP" opt; do
    case "$opt" in
        $FLAG_DURATION)
            duration="$OPTARG"
            # check that the argument is valid
            re='^[0-9]+[smh]$'
            if [[ ! $duration =~ $re ]]; then
                echo "error: invalid argument '$duration' for $FLAG_DURATION flag: should contain a number and a time unit (s/m/h), e.g. 1h or 30m"
                print_usage
                exit 1
            fi
            ;;
        $FLAG_NAMESPACES)
            namespaces_list=($(echo "$OPTARG" | awk -v RS="$ARGS_SEPARATOR" '{print}'))
            # check that all the given namespaces exist
            existant_ns=$(kubectl get ns -o=jsonpath='{.items[*].metadata.name}')
            for ns in ${namespaces_list[@]}; do
                if [[ ! " ${existant_ns[@]} " =~ " ${ns} " ]]; then
                    echo "error: given namespace '$ns' doesn't exist"
                    print_usage
                    exit 1
                fi
            done
            ;;
        $FLAG_PATH)
            output_path=$(realpath "$OPTARG")/
            if [[ ! -d $output_path ]]; then
                echo "error: given path '$output_path' doesn't exist"
                print_usage
                exit 1
            fi
            ;;
        $FLAG_APP)
            application_args_list=()
            application_args_list=($(echo "$OPTARG" | awk -v RS="$ARGS_SEPARATOR" '{print}'))
            if [ ${#application_args_list[@]} -ne 2 ] || [ -z ${application_args_list[0]} ] || [ -z ${application_args_list[1]} ]; then
                echo "error: given application namespace and name should be separated by a single comma"
                print_usage
                exit 1
            fi
            existant_ns=$(kubectl get ns -o=jsonpath='{.items[*].metadata.name}')
            fybrik_application_ns=${application_args_list[0]}
            fybrik_application_name=${application_args_list[1]}
            if [[ ! " ${existant_ns[@]} " =~ " ${fybrik_application_ns} " ]]; then
                echo "error: given namespace '$fybrik_application_ns' in flag '$FLAG_APP' doesn't exist"
                print_usage
                exit 1
            fi
            fybrik_applications_in_ns=$(kubectl get fybrikapplications -n $fybrik_application_ns -o=jsonpath='{.items[*].metadata.name}')
            if [[ ! " ${fybrik_applications_in_ns[@]} " =~ " ${fybrik_application_name} " ]]; then
                echo -e "error: given fybrikapplication '$fybrik_application_ns:$fybrik_application_name' doesn't exist"
                print_usage
                exit 1
            fi
            ;;
        $FLAG_UUID)
            fybrik_application_uuid="$OPTARG"
            ;;
        $FLAG_HELP)
            print_description
            print_usage
            print_flags
            exit 0
            ;;
        : | ?)
            print_usage
            exit 1
            ;;
    esac
done
shift "$(($OPTIND -1))"

# if no namespaces were provided, set value to default namespaces
if [[ ${#namespaces_list[@]} -eq 0 ]]; then
    for ns in $(kubectl get ns -o=jsonpath='{.items[*].metadata.name}'); do
        if [[ $ns == *"fybrik"* ]]; then
            namespaces_list+=("$ns")
        fi
    done
fi

# create a directory and subdirectories for saving output
dirname_logs=$output_path"fybrik_logs_$(date +%Y-%m-%d_%H-%M)_last_$duration"
echo "Output will be saved in $dirname_logs"
mkdir $dirname_logs
dirname_logs_by_container=$dirname_logs/logs_by_container
mkdir $dirname_logs_by_container
dirname_not_ready_pods=$dirname_logs/describe_not_ready_pods
mkdir $dirname_not_ready_pods
dirname_modules=$dirname_logs/fybrikmodules
mkdir $dirname_modules

is_found_fybrik_application_yaml=0

# update fybrik application variables and check flags usage
if [[ ! -z $fybrik_application_uuid ]]; then
    # uuid flag is used
    if [[ ! -z $fybrik_application_name ]]; then
        # application flag is also used
        echo "error: $APP_CONFLICT_MESSAGE"
        print_usage
        exit 1
    fi
else
    if [[ ! -z $fybrik_application_name ]]; then
        # only application flag is used
        kubectl get fybrikapplication $fybrik_application_name -n $fybrik_application_ns -o=yaml &> $dirname_logs/fybrikapplication_$fybrik_application_ns--$fybrik_application_name.yaml
        is_found_fybrik_application_yaml=1
        fybrik_application_uuid=$(kubectl get fybrikapplication $fybrik_application_name -n $fybrik_application_ns -o=jsonpath='{.metadata.uid}')
    fi
fi

# iterate over the relevant namespaces
for ns in ${namespaces_list[@]+"${namespaces_list[@]}"}; do
    # save the logs of each container separately
    for pod in $(kubectl get pods -n $ns -o=jsonpath='{.items[*].metadata.name}'); do
        # if pod is ready then save its logs, otherwise save its describe output
        if [[ $(kubectl get pod $pod -n $ns -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}') == "True" ]]; then
            containers=$(kubectl get pods $pod -n $ns -o jsonpath='{.spec.containers[*].name}')
            for container_name in $containers; do
                kubectl logs $pod -n $ns -c $container_name --since=$duration --timestamps &> $dirname_logs_by_container/$ns--$pod--$container_name.txt
            done
        else
            kubectl describe pod $pod -n $ns &> $dirname_not_ready_pods/$ns--$pod.txt
        fi
    done
    # save the non-certificates configmaps
    for cm in $(kubectl get cm -n $ns -o=jsonpath='{.items[*].metadata.name}'); do
        if [[ ! "$cm" == *.crt ]]; then
            cm_data=$(kubectl get cm $cm -n $ns -o=jsonpath='{.data}')
            prefix=$ns--$cm
            echo -e "$prefix:\n$cm_data\n" >> $dirname_logs/configmaps.txt
        fi
    done
    # save output of 'get pods' command
    get_pods_output=$(kubectl get pods -n $ns 2>&1)
    echo -e "$ns:\n$get_pods_output\n" >> $dirname_logs/get_pods_output.txt
    # save yamls of deployed modules
    for fybrik_module in $(kubectl get fybrikmodules -n $ns -o=jsonpath='{.items[*].metadata.name}'); do
        kubectl get fybrikmodule $fybrik_module -n $ns -o yaml &> $dirname_modules/$ns--$fybrik_module.yaml
    done
    # if given fybrik application uuid, save its yaml and create another file which only contains logs with the uuid
    if [[ ! -z fybrik_application_uuid ]] && [[ is_found_fybrik_application_yaml -eq 0 ]]; then
        for fybrik_application in $(kubectl get fybrikapplications -n $ns -o=jsonpath='{.items[*].metadata.name}'); do
            if [[ $(kubectl get fybrikapplication $fybrik_application -n $ns -o=jsonpath='{.metadata.uid}') == $fybrik_application_uuid ]]; then
                kubectl get fybrikapplication $fybrik_application -n $ns -o=yaml &> $dirname_logs/fybrikapplication_$ns--$fybrik_application.yaml
                is_found_fybrik_application_yaml=1
            fi
        done
    fi
done

FILENAME_COMBINED=combined_logs
FILENAME_COMBINED_SORTED=${FILENAME_COMBINED}_sorted
filepath_combined_logs=$dirname_logs/$FILENAME_COMBINED.txt
filepath_sorted_logs=$dirname_logs/$FILENAME_COMBINED_SORTED.txt

if [[ $(find $dirname_logs_by_container -type f | wc -l) -eq 0 ]]; then
    echo "there are no ready pods in the explored namespaces"
else
    # get maximal prefix length for the combined logs file
    max_len=0
    for f in $dirname_logs_by_container/*.txt; do
        f_short=$(basename -s .txt $f)
        if [ ${#f_short} -gt $max_len ]; then
            max_len=${#f_short}
        fi
    done

    # combine all the logs, with prefix of their ns--pod--container
    for f in $dirname_logs_by_container/*.txt; do
        f_short=$(basename -s .txt $f)
        prefix=$(printf "%-${max_len}s\n" "$f_short")
        while read line; do
            echo -e "$prefix || $line" >> $filepath_combined_logs
        done < $f
    done

    # if any logs are found, sort them by their timestamps and remove combined file
    if [[ -e $filepath_combined_logs ]]; then
        sort -k 3 $filepath_combined_logs > $filepath_sorted_logs
        rm $filepath_combined_logs
    else
        echo "no log entries were found in the pods in the explored namespaces"
    fi
fi

# if given fybrik application uuid, create another file which only contains logs with the uuid
filepath_filtered_logs_by_uuid=$dirname_logs/${FILENAME_COMBINED_SORTED}_$fybrik_application_uuid.txt
if [[ ! -z $fybrik_application_uuid ]]; then
    # check if its yaml file was found
    if [[ $is_found_fybrik_application_yaml -eq 0 ]]; then
        echo "the fybrikapplication '$fybrik_application_uuid' was not found in the namespaces, its yaml file was not retrieved"
    fi
    if [[ -e $filepath_sorted_logs ]]; then
        # create a file with the logs containing the uuid
        touch $filepath_filtered_logs_by_uuid
        while read line; do
            if [[ $line == *\"$fybrik_application_uuid\"* ]]; then
                echo "$line" >> $filepath_filtered_logs_by_uuid
            fi
        done < $filepath_sorted_logs
    fi
fi

# create another logs file filtered by only warn+error level logs
filepath_filtered_logs_by_level=$dirname_logs/${FILENAME_COMBINED_SORTED}_warn_and_error.txt
if [[ -e $filepath_sorted_logs ]]; then
    touch $filepath_filtered_logs_by_level
    while read line; do
        if [[ $line == *"\"level\":\"warn\""* || $line == *"\"level\":\"error\""* ]]; then
            echo "$line" >> $filepath_filtered_logs_by_level
        fi
    done < $filepath_sorted_logs
fi
