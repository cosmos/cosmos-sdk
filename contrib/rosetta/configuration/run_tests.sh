#!/bin/sh

set -e

wait_for_node() {
    timeout 30 sh -c 'until nc -z $0 $1; do sleep 1; done' cosmos 26657
    echo "tendermint rpc is up"
    timeout 30 sh -c 'until nc -z $0 $1; do sleep 1; done' rest 1317
    echo "rest server is up"
}
wait_for_rosetta() {
  timeout 30 sh -c 'until nc -z $0 $1; do sleep 1; done' rosetta 8080
}

echo "waiting for node to be up"
wait_for_node

echo "waiting for rosetta instance to be up"
wait_for_rosetta

echo "checking data API"
rosetta-cli check:data --configuration-file rosetta.json

echo "checking construction API"
rosetta-cli check:construction --configuration-file rosetta.json