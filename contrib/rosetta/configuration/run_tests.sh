#!/bin/sh

set -e

addr="abcd"

send_tx() {
  echo '12345678' | simd tx bank send $addr "$1" "$2"
}

detect_account() {
  line=$1
}

wait_for_rosetta() {
  timeout 30 sh -c 'until nc -z $0 $1; do sleep 1; done' rosetta 8080
}

echo "waiting for rosetta instance to be up"
wait_for_rosetta

echo "checking data API"
rosetta-cli check:data --configuration-file ./config/rosetta.json

echo "checking construction API"
rosetta-cli check:construction --configuration-file ./config/rosetta.json

