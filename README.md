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

## Prerequisites

[Install and setup Golang](https://tendermint.com/docs/guides/install-go).

## Installation

```
go get -u github.com/tendermint/basecoin/cmd/basecoin
```

See the [install guide](/docs/guide/install.md) for more details.


## Guide

1. Getting started with the [Basecoin basics](/docs/guide/basecoin-basics.md)
1. Learning to [use the plugin system](/docs/guide/basecoin-plugins.md)
1. More features of the [Basecoin tool](/docs/guide/basecoin-tool.md)
1. Learn how to use [InterBlockchain Communication (IBC)](/docs/guide/ibc.md)


To deploy a testnet, see our [repository of deployment tools](https://github.com/tendermint/tools).
