# Install the SDK

This guide will explain how to install the [Cosmos SDK](/sdk/overview.md) onto your system. With the SDK installed on a server, you can participate in the latest testnet as either a [Full Node](full-node.md) or a [Validator](/validators/validator-setup.md).

## Install Go

Install `go` by following the [official docs](https://golang.org/doc/install). Remember to set your `$GOPATH`, `$GOBIN`, and `$PATH` environment variables, for example:

```bash
mkdir -p $HOME/go/bin
echo "export GOPATH=$HOME/go" >> ~/.bash_profile
echo "export GOBIN=$GOPATH/bin" >> ~/.bash_profile
echo "export PATH=$PATH:$GOBIN" >> ~/.bash_profile
```

::: tip
**Go 1.10+** is required for the Cosmos SDK.
:::

## Install Cosmos SDK

Next, let's install the testnet's version of the Cosmos SDK.

```bash
mkdir -p $GOPATH/src/github.com/cosmos
cd $GOPATH/src/github.com/cosmos
git clone https://github.com/cosmos/cosmos-sdk
cd cosmos-sdk && git checkout v0.19.0
make get_tools && make get_vendor_deps && make install
```

That will install the `gaiad` and `gaiacli` binaries. Verify that everything is OK:

```bash
$ gaiad version
0.19.0-c6711810

$ gaiacli version
0.19.0-c6711810
```

## Run a Full Node

With Cosmos SDK installed, you can run [a full node on the latest testnet](full-node.md).
