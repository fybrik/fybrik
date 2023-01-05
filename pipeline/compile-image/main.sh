#!/bin/bash

# The default main acts depending on
#   - whether a command is supplied
#   - whether there is a non-null stdin stream

# if <in-container command> : run it
# elif null input : run sleep infinity
# else : run bash

RUN(){
    if [[ $SLEEP_FOREVER ]] ; then
        "$@";
        exec sleep infinity
    fi
    exec "$@"
}

# If there is a command, run it
[[ $@ ]] && RUN "$@"

# If there is no input, sleep forever
[[ "$(readlink /proc/self/fd/0)" = "/dev/null" ]] && exec sleep infinity

# Others exec bash
exec /bin/bash
