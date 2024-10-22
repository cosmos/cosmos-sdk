#!/usr/bin/env bash

set -o errexit -o nounset -o pipefail
# set -x

SIMD_BIN=${SIMD_BIN:=$(which simdv2 2>/dev/null)}
SIMD_HOME=$($SIMD_BIN config home)
export LC_ALL=C

init_config() {
    $SIMD_BIN config set client chain-id simapp-v2-chain
    $SIMD_BIN config set client keyring-backend test
    $SIMD_BIN config set client keyring-default-keyname alice
    sed -i '' "s/sc-type = 'iavl'/sc-type = 'iavl-v2'/" $SIMD_HOME/config/app.toml
}

gen_alias() {
    for i in $(seq 100); do
        alias=$(head /dev/urandom | tr -dc A-Za-z0-9 | head -c12)
        $SIMD_BIN keys add $alias --indiscreet
    done
    $SIMD_BIN keys add alice --indiscreet
    $SIMD_BIN keys add bob --indiscreet
    $SIMD_BIN init simapp-v2-node --chain-id simapp-v2-chain
    # to change the voting_period
    jq '.app_state.gov.params.voting_period = "600s"' $SIMD_HOME/config/genesis.json > temp.json && mv temp.json $SIMD_HOME/config/genesis.json
    jq '.app_state.gov.params.expedited_voting_period = "300s"' $SIMD_HOME/config/genesis.json > temp.json && mv temp.json $SIMD_HOME/config/genesis.json
    jq '.app_state.mint.minter.inflation = "0.300000000000000000"' $SIMD_HOME/config/genesis.json > temp.json && mv temp.json $SIMD_HOME/config/genesis.json # to change the inflation
}

add_genesis_accounts() {
    cd $SIMD_HOME/keyring-test
    ls *.info > ../aliases.txt
    cd ..
    for a in $(cat aliases.txt); do
        $SIMD_BIN genesis add-genesis-account $(basename $a .info) 100000000stake --keyring-backend test
    done
}

case $1 in
    init)
        init_config
    ;;

    gen-alias)
        gen_alias
    ;;

    gen-send-txs)
        $SIMD_BIN genesis generate-send-txs > $SIMD_HOME/load-txs.json
    ;;

    add-genesis-accounts)
        add_genesis_accounts
    ;;

    collect)
        $SIMD_BIN genesis gentx alice 1000000stake --chain-id simapp-v2-chain
        $SIMD_BIN genesis collect-gentxs
    ;;

    all)
        init_config
        gen_alias
        add_genesis_accounts
        $SIMD_BIN genesis gentx alice 1000000stake --chain-id simapp-v2-chain
        $SIMD_BIN genesis collect-gentxs
    ;;
esac