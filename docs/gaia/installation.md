## Install Gaia

This guide will explain how to install the `gaiad` and `gaiacli` entrypoints onto your system. With these installed on a server, you can participate in the latest testnet as either a [Full Node](./join-testnet.md#run-a-full-node) or a [Validator](./validators/validator-setup.md).

### Install Go

Install `go` by following the [official docs](https://golang.org/doc/install). Remember to set your `$GOPATH`, `$GOBIN`, and `$PATH` environment variables, for example:

```bash
mkdir -p $HOME/go/bin
echo "export GOPATH=$HOME/go" >> ~/.bash_profile
echo "export GOBIN=$GOPATH/bin" >> ~/.bash_profile
echo "export PATH=$PATH:$GOBIN" >> ~/.bash_profile
```

::: tip
**Go 1.11.5+** is required for the Cosmos SDK.
:::

### Install the binaries

Next, let's install the latest version of Gaia. Here we'll use the `master` branch, which contains the latest stable release.
If necessary, make sure you `git checkout` the correct
[released version](https://github.com/cosmos/cosmos-sdk/releases).

```bash
mkdir -p $GOPATH/src/github.com/cosmos
cd $GOPATH/src/github.com/cosmos
git clone https://github.com/cosmos/cosmos-sdk
cd cosmos-sdk && git checkout master
make tools install
```

> *NOTE*: If you have issues at this step, please check that you have the latest stable version of GO installed.

That will install the `gaiad` and `gaiacli` binaries. Verify that everything is OK:

```bash
$ gaiad version --long
$ gaiacli version --long
```

`gaiacli` for instance should output something similar to:

```
cosmos-sdk: 0.31.2-10-g1fba7308
git commit: 1fba7308fa226e971964cd6baad9527d4b51d9fc
vendor hash: 1aec7edfad9888a967b3e9063e42f66b28f447e6
build tags: netgo ledger
go version go1.11.5 linux/amd64
```

##### Build Tags

Build tags indicate special features that have been enabled in the binary.

| Build Tag | Description                                     |
| --------- | ----------------------------------------------- |
| netgo     | Name resolution will use pure Go code           |
| ledger    | Ledger devices are supported (hardware wallets) |

### Next

Now you can [join the public testnet](./join-testnet.md) or [create you own  testnet](./deploy-testnet.md)
