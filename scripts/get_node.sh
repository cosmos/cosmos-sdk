#!/usr/bin/env bash

NODE=v11.15.0

mkdir -p ~/.local/bin
mkdir -p ~/.local/node
wget http://nodejs.org/dist/${NODE}/node-${NODE}-linux-x64.tar.gz -O ~/.local/node/${NODE}.tar.gz
tar -xzf ~/.local/node/${NODE}.tar.gz -C ~/.local/node/
ln -s ~/.local/node/node-${NODE}-linux-x64/bin/node ~/.local/bin/node
ln -s ~/.local/node/node-${NODE}-linux-x64/bin/npm ~/.local/bin/npm
export PATH=~/.local/bin:$PATH
npm i -g dredd@11.0.1
