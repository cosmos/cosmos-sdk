#!/usr/bin/env bash
set -Eeuo pipefail

if [ -z "$ROCKSDB_VERSION" ]; then
  echo "ROCKSDB_VERSION is not set."
  exit 1
fi

# Update and install dependencies
sudo apt update && sudo apt-get install -y libsnappy-dev zlib1g-dev libbz2-dev liblz4-dev libzstd-dev build-essential

# Clone RocksDB repository
git clone https://github.com/facebook/rocksdb.git /home/runner/rocksdb
cd /home/runner/rocksdb || exit 1
git checkout "v${ROCKSDB_VERSION}"

# Build shared library
sudo make -j "$(nproc --all)" shared_lib
sudo cp --preserve=links ./librocksdb.* /usr/local/lib/
sudo cp -r ./include/rocksdb/ /usr/local/include/
