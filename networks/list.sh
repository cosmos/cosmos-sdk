#!/bin/sh
# list - list the IPs of a set of nodes

if [ $# -ne 1 ]; then
  echo "Usage: ./list.sh <clustername>"
  exit 1
fi
set -eux

export CLUSTER_NAME=$1

make list

