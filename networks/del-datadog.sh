#!/bin/sh
# del-datadog - aremove datadog agent from a set of nodes

if [ $# -ne 1 ]; then
  echo "Usage: ./del-datadog.sh <clustername>"
  exit 1
fi
set -eux

export CLUSTER_NAME=$1

make remove-datadog

