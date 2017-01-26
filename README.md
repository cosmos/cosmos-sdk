# Basecoin

DISCLAIMER: Basecoin is not associated with Coinbase.com, an excellent Bitcoin/Ethereum service.

Basecoin is a sample [ABCI application](https://github.com/tendermint/abci) designed to be used with the [tendermint consensus engine](https://tendermint.com/) to form a Proof-of-Stake cryptocurrency. This project has two main purposes:

  1. As an example for anyone wishing to build a custom application using tendermint.
  2. As a framework for anyone wishing to build a tendermint-based currency, extensible using the plugin system.

## Contents

  1. [Installation](#installation)
  1. [(Advice for go novices)](./GoBasics.md)
  1. [Using the plugin system](#plugins)
  1. [Forking the codebase](#forking)
  1. [Tutorials and other reading](#tutorials)

## Installation

We use glide for dependency management.  The prefered way of compiling from source is the following:

```
go get github.com/tendermint/basecoin
cd $GOPATH/src/github.com/tendermint/basecoin
make get_vendor_deps
make install
```

This will create the `basecoin` binary.

## Plugins

Basecoin handles public-key authentication of transaction, maintaining the balance of arbitrary types of currency (BTC, ATOM, ETH, MYCOIN, ...), sending currency (one-to-one or n-to-n multisig), and providing merkle-proofs of the state. These are common factors that many people wish to have in a crypto-currency system, so instead of trying to start from scratch, you can take advantage of the basecoin plugin system.

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

`RunTx` is where you can handle any special transactions directed to your application. To see a very simple implementation, look at the demo [counter plugin](./plugins/counter/counter.go). If you want to create your own currency using a plugin, you don't have to fork basecoin at all.  Just make your own repo, add the implementation of your custom plugin, and then build your own main script that instatiates BaseCoin and registers your plugin.

An example is worth a 1000 words, so please take a look [at this example](https://github.com/tendermint/basecoin/blob/abci_proof/cmd/paytovote/main.go#L25-L31), in a dev branch for now.  You can use the same technique in your own repo.

There are a lot of changes on the dev branch, which should be merged in my early February, so experiment, but things will change soon....

## Forking

If you do want to fork basecoin, we would be happy if this was done in a public repo and any enhancements made as PRs on github.  However, this is under the Apache license and you are free to keep the code private if you wish.

If you don't have much experience forking in go, there are a few tricks you want to keep in mind to avoid headaches. Basically, all imports in go are absolute from GOPATH, so if you fork a repo with more than one directory, and you put it under github.com/MYNAME/repo, all the code will start caling github.com/ORIGINAL/repo, which is very confusing.  My prefered solution to this is as follows:

  * Create your own fork on github, using the fork button.
  * Go to the original repo checked out locally (from `go get`)
  * `git remote rename origin upstream`
  * `git remote add origin git@github.com:YOUR-NAME/basecoin.git`
  * `git push -u origin master`
  * You can now push all changes to your fork and all code compiles, all other code referencing the original repo, now references your fork.
  * If you want to pull in updates from the original repo:
    * `git fetch upstream`
    * `git rebase upstream/master` (or whatever branch you want)

## Tutorials

We are working on some tutorials that will show you how to set up the genesis block, build a plugin to add custom logic, deploy to a tendermint testnet, and connect a UI to your blockchain.  They should be published during the course of February 2017, so stay tuned....


