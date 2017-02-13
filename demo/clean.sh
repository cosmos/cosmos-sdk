#! /bin/bash

killall -9 basecoin tendermint
TMROOT=./data/chain1/tendermint tendermint unsafe_reset_all
TMROOT=./data/chain2/tendermint tendermint unsafe_reset_all

rm -rf ./data/chain1/basecoin/merkleeyes.db
rm -rf ./data/chain2/basecoin/merkleeyes.db

rm ./*.log

rm ./data/chain1/tendermint/*.bak
rm ./data/chain2/tendermint/*.bak
