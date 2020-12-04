#!/bin/sh

wait_node() {
      timeout 30 sh -c 'until nc -z $0 $1; do sleep 1; done' cosmos 26657
      echo "tendermint rpc is up"
}
echo "waiting for tm node to be up..."
wait_node
simcli rest-server --laddr tcp://0.0.0.0:1317 --node tcp://cosmos:26657 --chain-id testing