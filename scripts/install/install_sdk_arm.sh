#!/usr/bin/env bash

# change this to a specific release or branch
BRANCH=master
REPO=github.com/cosmos/cosmos-sdk

GO_VERSION=1.12.5

sudo apt-get update -y
sudo apt-get upgrade -y
sudo apt-get install -y make

# get and unpack golang
curl -O https://dl.google.com/go/go$GO_VERSION.linux-armv6l.tar.gz
tar -xvf go$GO_VERSION.linux-armv6l.tar.gz

# move go binary and add to path
sudo mv go /usr/local
echo "export PATH=\$PATH:/usr/local/go/bin" >> ~/.profile

# create the go directory, set GOPATH, and put it on PATH
mkdir go
echo "export GOPATH=$HOME/go" >> ~/.profile
echo "export PATH=\$PATH:\$GOPATH/bin" >> ~/.profile
echo "export GO111MODULE=on" >> ~/.profile
source ~/.profile

# get the code and move into repo
go get $REPO
cd $GOPATH/src/$REPO

# build & install master
git checkout $BRANCH
LEDGER_ENABLED=false make tools
LEDGER_ENABLED=false make install

source ~/.profile
