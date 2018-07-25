#!/bin/sh
# new-testnet - example make call to create a new set of validator nodes in AWS
# WARNING: Run it from the current directory - it uses relative paths to ship the binary

if [ $# -ne 4 ]; then
  echo "Usage: ./new-testnet.sh <testnetname> <clustername> <regionlimit> <numberofnodesperavailabilityzone>"
  exit 1
fi
set -eux

if [ -z "`file ../build/gaiad | grep 'ELF 64-bit'`" ]; then
  # Build the linux binary we're going to ship to the nodes
  make -C .. build-linux
fi

# The testnet name is the same on all nodes
export TESTNET_NAME=$1
export CLUSTER_NAME=$2
export REGION_LIMIT=$3
export SERVERS=$4

# Build the AWS validator nodes and extract the genesis.json and config.toml from one of them
rm -rf remote/ansible/keys
make validators-start extract-config

# Save the private key seed words from the validators
SEEDFOLDER="${TESTNET_NAME}-${CLUSTER_NAME}-seedwords"
mkdir -p "${SEEDFOLDER}"
test ! -f "${SEEDFOLDER}/node0" && mv remote/ansible/keys/* "${SEEDFOLDER}"

