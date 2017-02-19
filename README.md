# Basecoin

_DISCLAIMER: Basecoin is not associated with Coinbase.com, an excellent Bitcoin/Ethereum service._

Basecoin is an [ABCI application](https://github.com/tendermint/abci) designed to be used with the [Tendermint consensus engine](https://tendermint.com/) to form a Proof-of-Stake cryptocurrency.
It also provides a general purpose framework for extending the feature-set of the cryptocurrency
by implementing plugins.

Basecoin serves as a reference implementation for how we build ABCI applications in Go,
and is the framework in which we implement the [Cosmos Hub](https://cosmos.network).
It's easy to use, and doesn't require any forking - just implement your plugin, import the basecoin libraries,
and away you go with a full-stack blockchain and command line tool for transacting.

WARNING: Currently uses plain-text private keys for transactions and is otherwise not production ready.

## Installation

On a good day, basecoin can be installed like a normal Go program:

```
go get -u github.com/tendermint/basecoin/cmd/basecoin
```

In some cases, if that fails, or if another branch is required,
we use `glide` for dependency management.

The guaranteed correct way of compiling from source, assuming you've already 
run `go get` or otherwise cloned the repo, is:

```
cd $GOPATH/src/github.com/tendermint/basecoin
git checkout develop # (until we release tendermint v0.9)
make get_vendor_deps
make install
```

This will create the `basecoin` binary in `$GOPATH/bin`.


## Command Line Interface

The basecoin CLI can be used to start a stand-alone basecoin instance (`basecoin start`),
or to start basecoin with Tendermint in the same process (`basecoin start --in-proc`).
It can also be used to send transactions, eg. `basecoin tx send --to 0x4793A333846E5104C46DD9AB9A00E31821B2F301 --amount 100btc,10gold`
See `basecoin --help` and `basecoin [cmd] --help` for more details`.

## Learn more

1. Getting started with the [Basecoin tool](/docs/guide/basecoin-basics.md)
1. Learn more about [Basecoin's design](/docs/guide/basecoin-design.md)
1. Extend Basecoin [using the plugin system](/docs/guide/example-plugin.md)
1. Learn more about [plugin design](/docs/guide/plugin-design.md)
1. See some [more example applications](/docs/guide/more-examples.md)
1. Learn how to use [InterBlockchain Communication (IBC)](/docs/guide/ibc.md)
1. [Deploy testnets](deployment.md) running your basecoin application.


