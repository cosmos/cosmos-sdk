#!/usr/bin/env bash

# XXX: this script is intended to be run from
# a fresh Digital Ocean droplet with Ubuntu

# upon its completion, you must either reset
# your terminal or run `source ~/.profile`

# change this to a specific release or branch
BRANCH=master

sudo apt-get update -y
sudo apt-get upgrade -y
sudo apt-get install -y make

# get and unpack golang
curl -O https://storage.googleapis.com/golang/go1.10.linux-amd64.tar.gz
tar -xvf go1.10.linux-amd64.tar.gz

# move go binary and add to path
mv go /usr/local
echo "export PATH=\$PATH:/usr/local/go/bin" >> ~/.profile

# create the goApps directory, set GOPATH, and put it on PATH
mkdir goApps
echo "export GOPATH=/root/goApps" >> ~/.profile
echo "export PATH=\$PATH:\$GOPATH/bin" >> ~/.profile

source ~/.profile

# get the code and move into repo
REPO=github.com/cosmos/cosmos-sdk
go get $REPO
cd $GOPATH/src/$REPO

# build & install master
git checkout $BRANCH
make get_tools
make get_vendor_deps
make install
make install_examples
