#!/bin/sh
# add-datadog - add datadog agent to a set of nodes

if [ $# -ne 2 ]; then
  echo "Usage: ./add-datadog.sh <testnetname> <clustername>"
  exit 1
fi
set -eux

export TESTNET_NAME=$1
export CLUSTER_NAME=$2

make install-datadog

