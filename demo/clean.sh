#! /bin/bash

killall -9 basecoin tendermint
TMHOME=./data/chain1 tendermint unsafe_reset_all
TMHOME=./data/chain2 tendermint unsafe_reset_all

rm ./*.log

rm ./data/chain1/*.bak
rm ./data/chain2/*.bak
