#!/usr/bin/env bash

# XXX: this script is intended to be run from
# a fresh Digital Ocean droplet with Ubuntu

# change this to a specific release or branch
BRANCH=master

sudo apt-get update -y
sudo apt-get upgrade -y
sudo apt-get install -y make

# get and unpack golang
curl -O https://dl.google.com/go/go1.11.2.linux-amd64.tar.gz
tar -xvf go1.11.2.linux-amd64.tar.gz

# move go binary and add to path
mv go /usr/local
echo "export PATH=\$PATH:/usr/local/go/bin" >> ~/.profile

# create the go directory, set GOPATH, and put it on PATH
mkdir go
echo "export GOPATH=/root/go" >> ~/.profile
echo "export PATH=\$PATH:\$GOPATH/bin" >> ~/.profile

source ~/.profile

# get the code and move into repo
REPO=github.com/cosmos/cosmos-sdk
go get $REPO
cd $GOPATH/src/$REPO

# build & install master
git checkout $BRANCH
LEDGER_ENABLED=false make get_tools
LEDGER_ENABLED=false make get_vendor_deps
LEDGER_ENABLED=false make install

source ~/.profile