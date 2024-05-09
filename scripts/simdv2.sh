#!/usr/bin/env bash

set -o errexit
set -o nounset
set -x

rm -rf $HOME/.simappv2

ROOT=$PWD
cd "$ROOT"/simapp/v2
SIMD="${SIMD:-go run simdv2/main.go}"
CONFIG="${CONFIG:-$HOME/.simappv2/config}"

$SIMD init aurn-node --chain-id aurn-chain

cd $CONFIG

# to enable the api server
# sed -i '.bak' '/\[api\]/,+3 s/enable = false/enable = true/' app.toml
sed -i -E 's|enable = false|enable = true|g' app.toml

# to change the voting_period
jq '.app_state.gov.voting_params.voting_period = "600s"' genesis.json > temp.json && mv temp.json genesis.json

# to change the inflation
jq '.app_state.mint.minter.inflation = "0.300000000000000000"' genesis.json > temp.json && mv temp.json genesis.json

cd "$ROOT"/simapp/v2/
$SIMD keys add test_validator --keyring-backend test
VALIDATOR_ADDRESS=$($SIMD keys show test_validator -a --keyring-backend test)

$SIMD genesis add-genesis-account "$VALIDATOR_ADDRESS" 1000000000stake
$SIMD genesis gentx test_validator 1000000000stake --keyring-backend test
$SIMD genesis collect-gentxs

$SIMD start

# Comment since its not working for now
# # Wait to chain start then query block
# sleep 15
# cd ../..

# $SIMD query block --type=height 5
# if [ $? -eq 0 ]; then
#     echo "Success"
#     exit 0
# else 
#     echo "Fail"
#     exit 1
# fi