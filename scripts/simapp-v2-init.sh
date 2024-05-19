#!/usr/bin/env bash

set -o errexit
set -o nounset
set -x

ROOT=$PWD
SIMD="${SIMD:-go run ./simapp/v2/simdv2/main.go}"
CONFIG="${CONFIG:-$HOME/.simappv2/config}"

$SIMD init simapp-v2-node --chain-id simapp-v2-chain

cd $CONFIG

# to enable the api server
sed -i '.bak' '/\[api\]/,+3 s/enable = false/enable = true/' app.toml

# to change the voting_period
jq '.app_state.gov.voting_params.voting_period = "600s"' genesis.json > temp.json && mv temp.json genesis.json

# to change the inflation
jq '.app_state.mint.minter.inflation = "0.300000000000000000"' genesis.json > temp.json && mv temp.json genesis.json

# change the initial height to 2 to work around store/v2 and iavl limitations with a genesis block
jq '.initial_height = 2' genesis.json > temp.json && mv temp.json genesis.json

cd "$ROOT"
$SIMD keys add test_validator --keyring-backend test
VALIDATOR_ADDRESS=$($SIMD keys show test_validator -a --keyring-backend test)

$SIMD genesis add-genesis-account "$VALIDATOR_ADDRESS" 1000000000stake
$SIMD genesis gentx test_validator 1000000000stake --keyring-backend test
$SIMD genesis collect-gentxs
