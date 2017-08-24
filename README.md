# Cosmos SDK

![banner](cosmos-sdk-image.png)

[![version](https://img.shields.io/github/tag/cosmos/cosmos-sdk.svg)](https://github.com/cosmos/cosmos-sdk/releases/latest)
[![API Reference](https://godoc.org/github.com/cosmos/cosmos-sdk?status.svg
)](https://godoc.org/github.com/cosmos/cosmos-sdk)
[![Rocket.Chat](https://demo.rocket.chat/images/join-chat.svg)](https://cosmos.rocket.chat/)
[![license](https://img.shields.io/github/license/cosmos/cosmos-sdk.svg)](https://github.com/cosmos/cosmos-sdk/blob/master/LICENSE)

Branch    | Tests | Coverage | Report Card
----------|-------|----------|-------------
develop   | [![CircleCI](https://circleci.com/gh/cosmos/cosmos-sdk/tree/develop.svg?style=shield)](https://circleci.com/gh/cosmos/cosmos-sdk/tree/develop) | [![codecov](https://codecov.io/gh/cosmos/cosmos-sdk/branch/develop/graph/badge.svg)](https://codecov.io/gh/cosmos/cosmos-sdk) | [![Go Report Card](https://goreportcard.com/badge/github.com/cosmos/cosmos-sdk/tree/develop)](https://goreportcard.com/report/github.com/cosmos/cosmos-sdk/tree/develop)
master    | [![CircleCI](https://circleci.com/gh/cosmos/cosmos-sdk/tree/master.svg?style=shield)](https://circleci.com/gh/cosmos/cosmos-sdk/tree/master) | [![codecov](https://codecov.io/gh/cosmos/cosmos-sdk/branch/master/graph/badge.svg)](https://codecov.io/gh/cosmos/cosmos-sdk) | [![Go Report Card](https://goreportcard.com/badge/github.com/cosmos/cosmos-sdk/tree/master)](https://goreportcard.com/report/github.com/cosmos/cosmos-sdk/tree/master)

The Cosmos SDK is an [ABCI application](https://github.com/tendermint/abci) designed to be used with the [Tendermint consensus engine](https://tendermint.com/) to form a Proof-of-Stake cryptocurrency. It also provides a general purpose framework
for extending the feature-set of the cryptocurrency by implementing plugins.

Within this repository, the `basecoin` app serves as a reference implementation for how we build ABCI applications in Go, and is the framework in which we implement the [Cosmos Hub](https://cosmos.network). **It's easy to use, and doesn't require any forking** - just implement your plugin, import the libraries, and away
you go with a full-stack blockchain and command line tool for transacting.

## Prerequisites

* [golang](https://golang.org/doc/install)

## Installation

```
go get -u github.com/cosmos/cosmos-sdk/cmd/basecoin
```

See the [install guide](/docs/guide/install.md) for more details.

## Guides

* Getting started with the [Basecoin basics](/docs/guide/basecoin-basics.md)
* Learn to [use the plugin system](/docs/guide/basecoin-plugins.md)
* More features of the [Basecoin tool](/docs/guide/basecoin-tool.md)
* Learn how to use [Inter-Blockchain Communication (IBC)](/docs/guide/ibc.md)
* See [more examples](https://github.com/cosmos/cosmos-academy)

To deploy a testnet, see our [repository of deployment tools](https://github.com/tendermint/tools).
