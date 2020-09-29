#!/bin/bash

rm -rf $HOME/.simd/

cd $HOME

simd init --chain-id=testing testing --home=$HOME/.simd
simd keys add validator --keyring-backend=test --home=$HOME/.simd
simd add-genesis-account $(simd keys show validator -a --keyring-backend=test --home=$HOME/.simd) 1000000000validatortoken,1000000000stake --home=$HOME/.simd
simd gentx validator --keyring-backend=test --home=$HOME/.simd --chain-id testing
simd collect-gentxs --home=$HOME/.simd

simd start --home=$HOME/.simd
