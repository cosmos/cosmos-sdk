# Basecoin

DISCLAIMER: Basecoin is not associated with Coinbase.com, an excellent Bitcoin/Ethereum service.

Basecoin is a sample [ABCI application](https://github.com/tendermint/abci) designed to be used with the [tendermint consensus engine](https://tendermint.com/) to form a Proof-of-Stake cryptocurrency. This project has two main purposes:

  1. As an example for anyone wishing to build a custom application using tendermint.
  2. As a framework for anyone wishing to build a tendermint-based currency, extensible using the plugin system.

## Contents

  1. [Installation](#installation)
  1. [Using the plugin system](#using-the-plugin-system)
  1. [Using the cli](#using-the-cli)
  1. [Tutorials and other reading](#tutorials-and-other-reading)
  1. [Contributing](#contributing)

## Installation

We use glide for dependency management.  The prefered way of compiling from source is the following:

```
go get -d github.com/tendermint/basecoin/cmd/basecoin
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
so instead of trying to start from scratch, developers can extend the functionality of Basecoin using the plugin system, just writing the custom business logic they need, and leaving the rest to the basecoin system.

Interested in building a plugin?  Then [read more details here](./Plugins.md) and then you can follow a [simple tutorial](https://github.com/tendermint/basecoin-examples/blob/master/pluginDev/tutorial.md) to get your first plugin working.

## Using the CLI

The basecoin cli can be used to start a stand-alone basecoin instance (`basecoin start`),
or to start basecoin with tendermint in the same process (`basecoin start --in-proc`).
It can also be used to send transactions, eg. `basecoin sendtx --to 0x4793A333846E5104C46DD9AB9A00E31821B2F301 --amount 100`
See `basecoin --help` and `basecoin [cmd] --help` for more details`.

Or follow through a [step-by-step introduction](https://github.com/tendermint/basecoin-examples/blob/master/tutorial.md) to testing basecoin locally.

## Tutorials and Other Reading

See our [introductory blog post](https://cosmos.network/blog/cosmos-creating-interoperable-blockchains-part-1), which explains the motivation behind Basecoin.

We are working on some tutorials that will show you how to set up the genesis block, build a plugin to add custom logic, deploy to a tendermint testnet, and connect a UI to your blockchain.  They should be published during the course of February 2017, so stay tuned....

## Contributing

We will merge in interesting plugin implementations and improvements to Basecoin.

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
