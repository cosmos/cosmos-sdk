#!/usr/bin/env bash

set -o errexit
set -o nounset
set -x

ROOT=$PWD
SIMAPP_DIR="$ROOT/simapp/v2"

SIMD="$ROOT/build/simdv2"
CONFIG="${CONFIG:-$HOME/.simappv2/config}"

cd "$SIMAPP_DIR"
go build -o "$ROOT/build/simdv2" simdv2/main.go

if [ -d "$($SIMD config home)" ]; then rm -r $($SIMD config home); fi

$SIMD init simapp-v2-node --chain-id simapp-v2-chain

cd "$CONFIG"

# to change the voting_period
jq '.app_state.gov.voting_params.voting_period = "600s"' genesis.json > temp.json && mv temp.json genesis.json

# to change the inflation
jq '.app_state.mint.minter.inflation = "0.300000000000000000"' genesis.json > temp.json && mv temp.json genesis.json

$SIMD keys add test_validator --indiscreet
VALIDATOR_ADDRESS=$($SIMD keys show test_validator -a --keyring-backend test)

$SIMD genesis add-genesis-account "$VALIDATOR_ADDRESS" 1000000000stake
$SIMD genesis gentx test_validator 1000000000stake --keyring-backend test
$SIMD genesis collect-gentxs

$SIMD start &
SIMD_PID=$!

cnt=0
while ! $SIMD query block --type=height 5; do
  cnt=$((cnt + 1))
  if [ $cnt -gt 30 ]; then
    kill -9 "$SIMD_PID"
    exit 1
  fi
  sleep 1
done

kill -9 "$SIMD_PID"