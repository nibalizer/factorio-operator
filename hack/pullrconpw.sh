#!/bin/bash

if [ -z $1 ]; then
    echo "$0 Error, provide factorioserver name"
    echo "usage: $0 <factorioserver>"
    exit 1
fi

fsid=$1

# thanks stackoverflow
pod=$(kubectl get pod  -l fsid=${fsid} -o=jsonpath='{.items[*].metadata.name}')
echo Found pod $pod

echo "RCON PW:"
kubectl exec ${pod} -- cat /factorio/config/rconpw
