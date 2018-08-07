#!/bin/sh
# add-cluster - example make call to add a set of nodes to an existing testnet in AWS
# WARNING: Run it from the current directory - it uses relative paths to ship the binary and the genesis.json,config.toml files

if [ $# -ne 4 ]; then
  echo "Usage: ./add-cluster.sh <testnetname> <clustername> <regionlimit> <numberofnodesperavailabilityzone>"
  exit 1
fi
set -eux

# The testnet name is the same on all nodes
export TESTNET_NAME=$1
export CLUSTER_NAME=$2
export REGION_LIMIT=$3
export SERVERS=$4

# Build the AWS full nodes
rm -rf remote/ansible/keys
make fullnodes-start

# Save the private key seed words from the nodes
SEEDFOLDER="${TESTNET_NAME}-${CLUSTER_NAME}-seedwords"
mkdir -p "${SEEDFOLDER}"
test ! -f "${SEEDFOLDER}/node0" && mv remote/ansible/keys/* "${SEEDFOLDER}"

