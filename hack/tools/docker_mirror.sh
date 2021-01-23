#!/usr/bin/env bash

imagesfile=$1

SOURCE=${SOURCE:-docker.io}
MIRROR=${MIRROR:-quay.io/ibm}

readarray -t IMAGES < $imagesfile

do_pull_retag() {
	m=$1
	s=$2
	docker pull $m 2>/dev/null || return
	docker tag $m $s 2>/dev/null
	echo "$ docker pull $m"
	echo "$ docker tag $m $s"
	echo
}

mirror_pull_and_retag_to_source() {
	for s in ${IMAGES[@]}; do
		tag=${s#*:}
		[[ $s == '/' ]] && namespace=${s#\/*}
		name=${s#*\/}
		name=${name%:*} # remove suffix
		[ -n "$namespace" ] && do_pull_retag $MIRROR/$namespace-$name-x86_64:$tag $SOURCE/$s && continue
		[ -n "$namespace" ] && do_pull_retag $MIRROR/$namespace-$name-x86_64:latest $SOURCE/$s && continue
		[ -z "$namespace" ] && do_pull_retag $MIRROR/$name-x86_64:$tag $SOURCE/$s && continue
		[ -z "$namespace" ] && do_pull_retag $MIRROR/$name-x86_64:latest $SOURCE/$s && continue
	done
}

mirror_pull_and_retag_to_source
