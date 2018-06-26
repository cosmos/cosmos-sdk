# Cosmos SDK
![banner](docs/graphics/cosmos-sdk-image.png)

[![version](https://img.shields.io/github/tag/cosmos/cosmos-sdk.svg)](https://github.com/cosmos/cosmos-sdk/releases/latest)
[![API Reference](https://godoc.org/github.com/cosmos/cosmos-sdk?status.svg
)](https://godoc.org/github.com/cosmos/cosmos-sdk)
[![Rocket.Chat](https://demo.rocket.chat/images/join-chat.svg)](https://cosmos.rocket.chat/)
[![license](https://img.shields.io/github/license/cosmos/cosmos-sdk.svg)](https://github.com/cosmos/cosmos-sdk/blob/master/LICENSE)
[![LoC](https://tokei.rs/b1/github/cosmos/cosmos-sdk)](https://github.com/cosmos/cosmos-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/cosmos/cosmos-sdk)](https://goreportcard.com/report/github.com/cosmos/cosmos-sdk)

Branch    | Tests | Coverage
----------|-------|---------
develop   | [![CircleCI](https://circleci.com/gh/cosmos/cosmos-sdk/tree/develop.svg?style=shield)](https://circleci.com/gh/cosmos/cosmos-sdk/tree/develop) | [![codecov](https://codecov.io/gh/cosmos/cosmos-sdk/branch/develop/graph/badge.svg)](https://codecov.io/gh/cosmos/cosmos-sdk)
master    | [![CircleCI](https://circleci.com/gh/cosmos/cosmos-sdk/tree/master.svg?style=shield)](https://circleci.com/gh/cosmos/cosmos-sdk/tree/master) | [![codecov](https://codecov.io/gh/cosmos/cosmos-sdk/branch/master/graph/badge.svg)](https://codecov.io/gh/cosmos/cosmos-sdk)

**WARNING**: the libraries are still undergoing breaking changes as we get better ideas and start building out the Apps.

**Note**: Requires [Go 1.10+](https://golang.org/dl/)

## Testnet

For more information on connecting to the testnet, see
[cmd/gaia/testnets](/cmd/gaia/testnets)

For the latest status of the testnet, see the [status
file](/cmd/gaia/testnets/STATUS.md).


## Overview

The Cosmos-SDK is a platform for building multi-asset Proof-of-Stake (PoS) blockchains, like the [Cosmos Hub](https://cosmos.network). It is both a library for building and securely interacting with blockchain applications.

The goal of the Cosmos-SDK is to allow developers to easily create custom interoperable blockchain applications within the Cosmos Network without having to recreate common blockchain functionality, thus abstracting away the complexities of building a Tendermint ABCI application. We envision the SDK as the `npm`-like framework to build secure blockchain applications on top of Tendermint.

In terms of its design, the SDK optimizes flexibility and security. The framework is designed around a modular execution stack which allows applications to mix and match elements as desired. In addition, all modules are sandboxed for greater application security.

It is based on two major principles:

- **Composability**: Anyone can create a module for the Cosmos-SDK and integrating the already-built modules is as simple as importing them into your blockchain application.

- **Capabilities**: The SDK is inspired by capabilities-based security, and informed by years of wrestling with blockchain state machines. Most developers will need to access other 3rd party modules when building their own modules. Given that the Cosmos-SDK is an open framework and that we assume that some those modules may be malicious, we designed the SDK using object-capabilities (_ocaps_) based principles. In practice, this means that instead of having each module keep an access control list for other modules, each module implements `keepers` that can be passed to other modules to grant a pre-defined set of capabilities. For example, if an instance of module A's `keepers` is passed to module B, the latter will be able to call a restricted set of module A's functions.

  The capabilities of each `keeper` are defined by the module's developer, and it's their job to understand and audit the safety of foreign code from 3rd party modules based on the capabilities they are passing into each 3rd party module. For a deeper look at capabilities, you can read this [article](http://habitatchronicles.com/2017/05/what-are-capabilities/).

_Note: For now the Cosmos-SDK only exists in [Golang](https://golang.org/), which means that developers can only develop SDK modules in Golang. In the future, we expect that the SDK will be implemented in other programming languages. Funding opportunities supported by the Tendermint team may be available eventually._

## Application architecture

#### Modules

The Cosmos-SDK has all the necessary pre-built modules to add functionality on top of a `BaseApp`, which is the template to build a blockchain dApp in Cosmos. In this context, a `module` is a fundamental unit in the Cosmos-SDK. Each module is an extension of the `BaseApp` functionalities that defines transactions, handles application state and the state transition logic. Each module also contains handlers for messages and transactions, as well as REST and CLI for secure user interactions.

Some of the most important modules already integrated in the SDK are:

- `Auth`: Defines a standard account structure (`BaseAccount`) and how transaction signers are authenticated.
- `Bank`: Defines how coins (i.e cryptocurrencies) are transferred.
- `Governance`: Governance related implementation including proposals and voting.
- `Staking`: Proof of Stake related implementation including bonding and delegation transactions, inflation, fees, unbonding, etc.
- `IBC`: Defines the intereoperability of blockchain zones according to the specifications of the [IBC Protocol](https://cosmos.network/whitepaper#inter-blockchain-communication-ibc).


#### Directories

The key directories of the SDK are:

- `baseapp`: Defines the template for a basic [ABCI ](https://cosmos.network/whitepaper#abci) application so that your Cosmos-SDK application can communicate with the underlying Tendermint node.
- `client`: CLI and REST server tooling.
- `server`: RPC server to communicate with the node.
- `examples`: Contains examples on how to build working standalone applications.
- `store`: Contains code for the multistore (_i.e._ state). Each module can create any number of `KVStores` (key-value stores) from the multistore.
- `types`: Common types required in any SDK-based application.
- `x`(for e**X**tensions): Folder for storing the `BaseApp` module and all the already-built modules described in the previous section.

## Prerequisites

- [Golang](https://golang.org/doc/install)

## Getting Started

See the [documentation](https://cosmos-sdk.readthedocs.io).

## Disambiguation

This Cosmos-SDK project is not related to the [React-Cosmos](https://github.com/react-cosmos/react-cosmos) project (yet). Many thanks to Evan Coury and Ovidiu (@skidding) for this Github organization name. As per our agreement, this disambiguation notice will stay here.
