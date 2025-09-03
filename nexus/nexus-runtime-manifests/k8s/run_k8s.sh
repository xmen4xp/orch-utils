#!/bin/bash

set -x

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# cleanup_cluster handles deletion of all resources related to kind setup
function cleanup_cluster {
	docker compose -f ${SCRIPT_DIR}/compose.yaml down
}

while getopts "n:ds:m:h:k:p:" flag; do
    case "${flag}" in
        d) delete_cluster=true;;
        p) start_port=${OPTARG};;
    esac
done

if [ -z "$start_port" ]
then
      echo "start port number is mandatory. Provide a valid start port number with -s argument"
      exit 1
fi

# kubectl proxy port
export KUBECTL_PROXY_PORT=$start_port

if [ "$delete_cluster" == true ] ; then
    cleanup_cluster
    exit 0
fi

# Cleanup stale cluster, if any
cleanup_cluster

# Install a k8s cluster
docker compose -f ${SCRIPT_DIR}/compose.yaml up -d
sleep 5

exit 0

