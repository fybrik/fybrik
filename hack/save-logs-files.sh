#!/bin/bash

# Copyright 2023 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

# if a uid is passed, another file will be created, containing only logs with the uid
fybrik_application_uid=$1

# variable for limiting to recent logs, e.g. 50s, 20m, 3h
duration=1h

# create a directory for saving recent logs
dirname_logs=fybrik_logs_$(date +%Y-%m-%d_%H-%M)_last_$duration
mkdir $dirname_logs
mkdir $dirname_logs/by_pod

# iterate over the pods in the fybrik namespaces
for ns in $(kubectl get ns -o=jsonpath='{.items[*].metadata.name}'); do
    # check if the namespace contains the word "fybrik"
    if [[ $ns == *"fybrik"* ]]; then
        # save the logs of each container separately
        for pod in $(kubectl get pods -n $ns -o=jsonpath='{.items[*].metadata.name}'); do
            containers=$(kubectl get pods $pod -n $ns -o jsonpath="{.spec.containers[*].name}")
            for container_name in $containers; do
                kubectl logs $pod -n $ns -c $container_name --since=$duration --timestamps > ./$dirname_logs/by_pod/$ns--$pod--$container_name.txt
            done
        done
        # save configmaps
        for cm in $(kubectl get cm -n $ns -o=jsonpath='{.items[*].metadata.name}'); do
            cm_data=$(kubectl get cm $cm -n $ns -o=jsonpath='{.data}')
            prefix=$ns--$cm
            echo -e "$prefix:\n$cm_data" >> ./$dirname_logs/configmaps.txt
        done
    fi
done

# get maximal prefix length for the combined logs file
max_len=0
for f in ./$dirname_logs/by_pod/*.txt; do
    f_short=$(basename -s .txt $f)
    if [ ${#f_short} -gt $max_len ]; then
        max_len=${#f_short}
    fi
done

# combine all the logs, with prefix of their ns--pod--container, and sort them by their timestamps
for f in ./$dirname_logs/by_pod/*.txt; do
    f_short=$(basename -s .txt $f)
    prefix=$(printf "%-${max_len}s\n" "$f_short")
    while read line; do
        echo -e "$prefix || $line" >> ./$dirname_logs/combined_logs.txt
    done < $f
done
sort -k 3 ./$dirname_logs/combined_logs.txt > ./$dirname_logs/combined_logs_sorted.txt
rm ./$dirname_logs/combined_logs.txt

# if given fybrik application uid, create another file which only contains logs with the uid
if [[ $fybrik_application_uid ]]; then
    while read line; do
        if [[ $line == *$fybrik_application_uid* ]]; then
            echo "$line" >> ./$dirname_logs/combined_logs_sorted_onlyuid.txt
        fi
    done < ./$dirname_logs/combined_logs_sorted.txt
fi
