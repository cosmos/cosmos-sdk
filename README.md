# Basecoin

DISCLAIMER: Basecoin is not associated with Coinbase.com, an excellent Bitcoin/Ethereum service.

Basecoin is a sample [ABCI application](https://github.com/tendermint/abci) designed to be used with the [tendermint consensus engine](https://tendermint.com/) to form a Proof-of-Stake cryptocurrency. This project has two main purposes:

  1. As an example for anyone wishing to build a custom application using tendermint.
  2. As a framework for anyone wishing to build a tendermint-based currency, extensible using the plugin system.

## Contents

  1. [Installation](#installation)
  1. [Using the plugin system](#using-the-plugin-system)
  1. [Tutorials and other reading](#tutorials-and-other-reading)

## Installation

We use glide for dependency management.  The prefered way of compiling from source is the following:

```
go get github.com/tendermint/basecoin
cd $GOPATH/src/github.com/tendermint/basecoin
make get_vendor_deps
make install
```

This will create the `basecoin` binary in `$GOPATH/bin`.

## Using the Plugin System

Basecoin is designed to serve as a common base layer for developers building cryptocurrency applications.
It handles public-key authentication of transactions, maintaining the balance of arbitrary types of currency (BTC, ATOM, ETH, MYCOIN, ...), 
sending currency (one-to-one or n-to-m multisig), and providing merkle-proofs of the state. 
These are common factors that many people wish to have in a crypto-currency system, 
so instead of trying to start from scratch, developers can extend the functionality of Basecoin using the plugin system!

The Plugin interface is defined in `types/plugin.go`:

```
type Plugin interface {
  Name() string
  SetOption(store KVStore, key string, value string) (log string)
  RunTx(store KVStore, ctx CallContext, txBytes []byte) (res abci.Result)
  InitChain(store KVStore, vals []*abci.Validator)
  BeginBlock(store KVStore, height uint64)
  EndBlock(store KVStore, height uint64) []*abci.Validator
}
```

`RunTx` is where you can handle any special transactions directed to your application. 
To see a very simple implementation, look at the demo [counter plugin](./plugins/counter/counter.go). 
If you want to create your own currency using a plugin, you don't have to fork basecoin at all.  
Just make your own repo, add the implementation of your custom plugin, and then build your own main script that instatiates Basecoin and registers your plugin.

An example is worth a 1000 words, so please take a look [at this example](https://github.com/tendermint/basecoin/blob/develop/cmd/paytovote/main.go#L25-L31). 
Note for now it is in a dev branch.
You can use the same technique in your own repo.

## Tutorials and Other Reading

See our [introductory blog post](https://cosmos.network/blog/cosmos-creating-interoperable-blockchains-part-1), which explains the motivation behind Basecoin.

We are working on some tutorials that will show you how to set up the genesis block, build a plugin to add custom logic, deploy to a tendermint testnet, and connect a UI to your blockchain.  They should be published during the course of February 2017, so stay tuned....
