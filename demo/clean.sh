#! /bin/bash

killall -9 basecoin tendermint
TMROOT=./data/chain1 tendermint unsafe_reset_all
TMROOT=./data/chain2 tendermint unsafe_reset_all

rm ./*.log

rm ./data/chain1/*.bak
rm ./data/chain2/*.bak
