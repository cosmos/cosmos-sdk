#!/usr/bin/env bash

NODE=v11.15.0

cd ~
mkdir -p .local/bin
mkdir -p .local/node
cd .local/node
curl -O http://nodejs.org/dist/${NODE}/node-${NODE}-linux-x64.tar.gz
tar -xzf node-${NODE}-linux-x64.tar.gz
ln -s node-${NODE}-linux-x64 latest
cd ../bin
ln -s ../node/latest/bin/node
ln -s ../node/latest/bin/npm
echo 'export PATH=$HOME/.local/bin:$PATH' >> ~/.bashrc
source ~/.bashrc
