#!/bin/sh

set -e

wait_simd() {
  timeout 30 sh -c 'until nc -z $0 $1; do sleep 1; done' localhost 26657
}

echo clearing old configs
rm -rf /root/.simapp
rm -rf /root/.simcli
# init config files
simd init simd --chain-id testing
# configure cli
simcli config chain-id testing
simcli config output json
simcli config trust-node true

# use keyring backend

# create accounts
echo 12345678 | simcli keys add fd

# give the accounts some money
simd add-genesis-account $(simcli keys show fd -a) 1000000000000stake

# save configs for the daemon
echo "12345678" | simd gentx --name fd --amount 10000000000stake

# input genTx to the genesis file
simd collect-gentxs
# verify genesis file is fine
simd validate-genesis
sed -i 's/127.0.0.1/0.0.0.0/g' /root/.simapp/config/config.toml

# start simd
echo starting simd...
simd start --pruning=nothing &
pid=$!
echo simd started with PID $pid

echo awaiting for simd to be ready
wait_simd
echo simd is ready
sleep 10


# send transaction to deterministic address
addr=$(simcli keys show fd -a)
echo sending transaction with addr $addr
echo "12345678" | simcli tx send "$addr" cosmos1dee3p7gl4emuzyaqw7us5flaq5ajq3de8vrw3l 100stake --yes --broadcast-mode=block --chain-id="testing"

sleep 10

echo stopping simd...
kill -9 $pid

echo zipping data dir and saving to /tmp/data.tar.gz

tar -czvf /tmp/data.tar.gz /root/

echo new address for bootstrap.json "$addr"