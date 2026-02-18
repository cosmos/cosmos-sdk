#!/bin/bash
set -exuo pipefail

make localnet-clean
make build
make build-docker

# Number of nodes to create
NUM_NODES=5
CHAIN_ID="poa-localnet-1"
TOKEN_AMOUNT="1000000000000token"
DOCKER_IMAGE="poa-node"

echo "Creating $NUM_NODES node home directories..."
for i in $(seq 1 "$NUM_NODES"); do
    NODE_DIR="./build/node$i"
    docker run --rm -v "$PWD/$NODE_DIR:/root/.simapp" --name "node$i" $DOCKER_IMAGE \
        simd init "node$i" --chain-id "$CHAIN_ID"
done

echo "Creating keys for each node..."
for i in $(seq 1 "$NUM_NODES"); do
    NODE_DIR="./build/node$i"
    docker run --rm -v "$PWD/$NODE_DIR:/root/.simapp" --name "node$i" $DOCKER_IMAGE \
        simd keys add account --keyring-backend test
done

echo "Fetching node addresses..."
declare -a ACCOUNTS
for i in $(seq 1 "$NUM_NODES"); do
    NODE_DIR="./build/node$i"
    ACCOUNTS[$i]=$(docker run --rm -v "$PWD/$NODE_DIR:/root/.simapp" --name "node$i" $DOCKER_IMAGE \
        simd keys show account -a --keyring-backend test | tr -d '\r')
done

echo "Setting persistent_peers for each node..."
declare -a NODE_IDS
for i in $(seq 1 "$NUM_NODES"); do
    NODE_DIR="./build/node$i"
    NODE_IDS[$i]=$(./build/simd tendermint show-node-id --home "$NODE_DIR")
done

for i in $(seq 1 "$NUM_NODES"); do
    PEERS=""
    for j in $(seq 1 "$NUM_NODES"); do
        if [ "$i" -ne "$j" ]; then
            PEERS+="${NODE_IDS[$j]}@node$j:26656,"
        fi
    done
    PEERS=${PEERS%,}

    CONFIG_FILE="./build/node$i/config/config.toml"
    echo "Node $i persistent_peers: $PEERS"
    sed -i.bak "s|^persistent_peers *=.*|persistent_peers = \"$PEERS\"|" "$CONFIG_FILE"
done

echo "Persistent peers configured for all nodes."

sed -i.bak 's|^laddr = "tcp://127.0.0.1:26657"|laddr = "tcp://0.0.0.0:26657"|' ./build/node1/config/config.toml
sed -i.bak 's|^address = "localhost:9090"|address = "0.0.0.0:9090"|' ./build/node1/config/app.toml

echo "Adding genesis accounts..."
NODE1_DIR="./build/node1"
for i in $(seq 1 "$NUM_NODES"); do
    docker run --rm -v "$PWD/$NODE1_DIR:/root/.simapp" --name "node1" $DOCKER_IMAGE \
        simd genesis add-genesis-account "${ACCOUNTS[$i]}" "$TOKEN_AMOUNT"
done

echo "Copying node1 genesis to shared genesis.json..."
cp "$NODE1_DIR/config/genesis.json" ./build/genesis.json

echo "Setting POA admin account..."
jq --arg acc "${ACCOUNTS[1]}" \
    '.app_state.poa.params.admin = $acc' \
    ./build/genesis.json > ./build/genesis.tmp.json \
    && mv ./build/genesis.tmp.json ./build/genesis.json

echo "Adding validators to genesis..."
for i in $(seq 1 "$NUM_NODES"); do
    PRIV_KEY="./build/node$i/config/priv_validator_key.json"
    jq --slurpfile pk <(jq '{pub_key: { "@type": "/cosmos.crypto.ed25519.PubKey", key: .pub_key.value }, power: 10000000, metadata: {moniker: "node'${i}'", operator_address: "'${ACCOUNTS[$i]}'"}}' "$PRIV_KEY") \
        '.app_state.poa.validators |= (if . == null then [] else . end + $pk)' \
        ./build/genesis.json > ./build/genesis.tmp.json \
        && mv ./build/genesis.tmp.json ./build/genesis.json
done

echo "Updating gov params to use 'token' denom..."
jq '.app_state.gov.params.min_deposit[0].denom = "token"' ./build/genesis.json > ./build/genesis.tmp.json \
    && mv ./build/genesis.tmp.json ./build/genesis.json
jq '.app_state.gov.params.expedited_min_deposit[0].denom = "token"' ./build/genesis.json > ./build/genesis.tmp.json \
    && mv ./build/genesis.tmp.json ./build/genesis.json

echo "Enabling secp256k1 pubkey type for consensus..."
jq '.consensus.params.validator.pub_key_types = ["ed25519", "secp256k1"]' ./build/genesis.json > ./build/genesis.tmp.json \
    && mv ./build/genesis.tmp.json ./build/genesis.json

echo "Distributing genesis.json to all nodes..."
for i in $(seq 1 "$NUM_NODES"); do
    cp ./build/genesis.json "./build/node$i/config/genesis.json"
done

echo "Setup complete for $NUM_NODES nodes."
