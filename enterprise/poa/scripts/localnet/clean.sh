#!/bin/bash
set -euo pipefail

echo "Cleaning node directories..."
rm -rf ./build/node[0-9]*

echo "Removing Docker containers..."
for c in $(docker ps -a --format '{{.Names}}' | grep -E '^node[0-9]+$$'); do
    echo "Removing container $c"
    docker rm -f "$c" 2>/dev/null || true
done

echo "Removing shared genesis.json..."
rm -f ./build/genesis.json

echo "Cleanup complete."
