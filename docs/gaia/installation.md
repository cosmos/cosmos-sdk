# Install Gaia

This guide will explain how to install the `gaiad` and `gaiacli` entrypoints onto your system. With these installed on a server, you can participate in the latest testnet as either a [Full Node](./join-testnet.md#run-a-full-node) or a [Validator](./validators/validator-setup.md).

## Install Go

Install `go` by following the [official docs](https://golang.org/doc/install). Remember to set your `$GOPATH`, `$GOBIN`, and `$PATH` environment variables, for example:

```bash
mkdir -p $HOME/go/bin
echo "export GOPATH=$HOME/go" >> ~/.bash_profile
echo "export GOBIN=$GOPATH/bin" >> ~/.bash_profile
echo "export PATH=$PATH:$GOBIN" >> ~/.bash_profile
```

::: tip
**Go 1.11+** is required for the Cosmos SDK.
:::

## Install Cosmos SDK

Next, let's install the latest version of Gaia. Here we'll use the `master` branch, which contains the latest stable release.
If necessary, make sure you `git checkout` the correct
[released version](https://github.com/cosmos/cosmos-sdk/releases).

```bash
mkdir -p $GOPATH/src/github.com/cosmos
cd $GOPATH/src/github.com/cosmos
git clone https://github.com/cosmos/cosmos-sdk
cd cosmos-sdk && git checkout master
make get_tools && make get_vendor_deps && make install
```

> *NOTE*: If you have issues at this step, please check that you have the latest stable version of GO installed.

That will install the `gaiad` and `gaiacli` binaries. Verify that everything is OK:

```bash
$ gaiad version
$ gaiacli version
```

## Run a Full Node

With the binaries installed, you can run [a full node on the latest testnet](./join-testnet.md).
