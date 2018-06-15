#!/usr/bin/tcsh

# Just run tcsh install_sdk_bsd.sh
# XXX: this script is intended to be run from
# a fresh Digital Ocean droplet with FreeBSD

# upon its completion, you must either reset
# your terminal or run `source ~/.tcshrc`

# This assumes your installing it through tcsh as root.
# Change the relevant lines from tcsh to csh if your
# installing as a different user, along with changing the
# gopath.

# change this to a specific release or branch
set BRANCH=master

sudo pkg update

sudo pkg upgrade -y
sudo pkg install -y gmake
sudo pkg install -y git

# get and unpack golang
curl -O https://storage.googleapis.com/golang/go1.10.freebsd-amd64.tar.gz
tar -xvf go1.10.freebsd-amd64.tar.gz

# move go binary and add to path
mv go /usr/local
set path=($path /usr/local/go/bin)


# create the go directory, set GOPATH, and put it on PATH
mkdir go
echo "setenv GOPATH /root/go" >> ~/.tcshrc
setenv GOPATH /root/go
echo "set path=($path $GOPATH/bin)" >> ~/.tcshrc

source ~/.tcshrc

# get the code and move into repo
set REPO=github.com/cosmos/cosmos-sdk
go get $REPO
cd $GOPATH/src/$REPO

# build & install master
git checkout $BRANCH
gmake get_tools
gmake get_vendor_deps
gmake install
gmake install_examples
