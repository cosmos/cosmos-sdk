# Basecoin

DISCLAIMER: Basecoin is not associated with Coinbase.com, an excellent Bitcoin/Ethereum service.

Basecoin is an [ABCI application](https://github.com/tendermint/abci) designed to be used with the [tendermint consensus engine](https://tendermint.com/) to form a Proof-of-Stake cryptocurrency. 
It also provides a general purpose framework for extending the feature-set of the cryptocurrency
by implementing plugins.

Basecoin serves as a reference implementation for how we build ABCI applications in Go,
and is the framework in which we implement the [Cosmos Hub](https://cosmos.network). 
It's easy to use, and doesn't require any forking - just implement your plugin, import the basecoin libraries,
and away you go with a full-stack blockchain and command line tool for transacting.

WARNING: Currently uses plain-text private keys for transactions and is otherwise not production ready.

## Installation

We use glide for dependency management.  The prefered way of compiling from source is the following:

```
go get -d github.com/tendermint/basecoin/cmd/basecoin
cd $GOPATH/src/github.com/tendermint/basecoin
make get_vendor_deps
make install
```

This will create the `basecoin` binary in `$GOPATH/bin`.

## Command Line Interface

The basecoin CLI can be used to start a stand-alone basecoin instance (`basecoin start`),
or to start basecoin with tendermint in the same process (`basecoin start --in-proc`).
It can also be used to send transactions, eg. `basecoin sendtx --to 0x4793A333846E5104C46DD9AB9A00E31821B2F301 --amount 100`
See `basecoin --help` and `basecoin [cmd] --help` for more details`.

## Learn more

1. Getting started with the [Basecoin tool](/docs/guide/basecoin-basics.md)
1. Learn more about [Basecoin's design](/docs/guide/basecoin-design.md)
1. Make your own [cryptocurrency using Basecoin plugins](/docs/guide/example-counter.md)
1. Learn more about [plugin design](/docs/guide/plugin-design.md)
1. See some [more example applications](/docs/guide/more-examples.md)
1. Learn how to use [InterBlockchain Communication (IBC)](ibc.md)
1. [Deploy testnets](deployment.md) running your basecoin application.


