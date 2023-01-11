#!/bin/bash

# Make globs that don't match anything return an empty list instead of the raw
# pattern.
shopt -s nullglob

# Source the files under /entrypoint.d
########################################
# Source any readable environment files
read_env_files(){
    local file
    for file in /entrypoint.d/*.env ; do
        if [[ -r "$file" ]] ; then
            source "$file"
        fi
    done
}
read_env_files

# Run the main
exec /main.sh "$@"
