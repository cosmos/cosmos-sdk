#!/usr/bin/env bash

set -o errexit
set -o nounset
set -x

ROOT=$PWD

SIMD="$ROOT/build/simdv2"
CONFIG="${CONFIG:-$HOME/.simappv2/config}"

COSMOS_BUILD_OPTIONS=v2 make build     

if [ -d "$($SIMD config home)" ]; then rm -rv $($SIMD config home); fi

$SIMD init simapp-v2-node --chain-id simapp-v2-chain

cd "$CONFIG"

# to change the voting_period
jq '.app_state.gov.voting_params.voting_period = "600s"' genesis.json > temp.json && mv temp.json genesis.json

# to change the inflation
jq '.app_state.mint.minter.inflation = "0.300000000000000000"' genesis.json > temp.json && mv temp.json genesis.json

$SIMD config set client chain-id simapp-v2-chain
$SIMD keys add test_validator --indiscreet
VALIDATOR_ADDRESS=$($SIMD keys show test_validator -a --keyring-backend test)

$SIMD genesis add-genesis-account "$VALIDATOR_ADDRESS" 1000000000stake
$SIMD genesis gentx test_validator 1000000000stake --keyring-backend test
$SIMD genesis collect-gentxs

$SIMD start &
SIMD_PID=$!

# wait 10s then export snapshot at height 10
sleep 10

kill -9 "$SIMD_PID"

$SIMD store export --height 5

# clear sc & ss data
rm -rf "$HOME/.simappv2/data/application.db"
rm -rf "$HOME/.simappv2/data/ss"

# restore

$SIMD restore 5 3
