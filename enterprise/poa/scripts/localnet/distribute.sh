#!/bin/bash
set -exuo pipefail

NUM_NODES=5

echo "Distributing genesis.json to all nodes..."
for i in $(seq 1 "$NUM_NODES"); do
    cp ./build/genesis.json "./build/node$i/config/genesis.json"
done
