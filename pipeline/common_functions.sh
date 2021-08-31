#!/bin/bash

function try_command {
    MAX_TRIES=${2:-5}
    EXIT=${3:-true}
    DELAY=${4:-1}
    TRIES=0
    DELETED=0

    command="$1"

    repo_root=$(realpath $(dirname $(realpath $0))/..)
    test_logs=$repo_root/test_logs
    mkdir -p ${test_logs}

    set +e
    for (( TRIES=0; TRIES<=$MAX_TRIES; TRIES++ ))
    do
      ${command}
      if [ $? -ne 0 ];then
         sleep ${DELAY}
      else
         DELETED=1
         break
      fi
    done
    if [[ $DELETED -eq 0 ]]; then
      DELETED=0
      echo "Couldn't run $command successfully"
      echo "`date` Couldn't run $command successfully" >> $test_logs/4_failed_commands.log
      if [[ $EXIT == "true" ]]; then
          exit 1
      fi
    else
      DELETED=0
    fi
}

echo "hi"
