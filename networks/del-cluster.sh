#!/bin/sh
# del-cluster - example make call to delete a set of nodes on an existing testnet in AWS

if [ $# -ne 1 ]; then
  echo "Usage: ./add-cluster.sh <clustername>"
  exit 1
fi
set -eux

export CLUSTER_NAME=$1

# Delete the AWS nodes
make fullnodes-stop

