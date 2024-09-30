#!/usr/bin/env bash

set -o errexit
set -o nounset
set -x

ROOT=$PWD

SIMD="$ROOT/build/simdv2"

COSMOS_BUILD_OPTIONS=v2 make build     

$SIMD testnet init-files --chain-id=testing --output-dir="$HOME/.testnet" --validator-count=3 --keyring-backend=test --minimum-gas-prices=0.000001stake --commit-timeout=900ms --single-host

$SIMD start --log_level=info --home "$HOME/.testnet/node0/simdv2" &
$SIMD start --log_level=info --home "$HOME/.testnet/node1/simdv2" &
$SIMD start --log_level=info --home "$HOME/.testnet/node2/simdv2" 