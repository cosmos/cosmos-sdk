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

**Note**: Requires [Go 1.9+](https://golang.org/dl/)


## Overview

The Cosmos SDK is a platform for building multi-asset Proof-of-Stake
cryptocurrencies, like the [Cosmos Hub](https://cosmos.network). It is both a library for building applications and a set of tools for securely interacting with them.

The goal of the Cosmos-SDK is to allow developers to easily create their custom interoperable blockchain applications within the Cosmos Network without having to code every single functionality of their application, thus abstracting away the complexities of building a Tendermint ABCI application. We envision the SDK as the `npm`-like framework to build secure blockchain applications on top of Tendermint.

In terms of its design, the SDK optimizes flexibility and security. The framework is designed around a modular execution stack which allows applications to mix and match elements as desired. In addition, all modules are sandboxed for greater application security.

It is based on two major principles:

- **Composability**: Anyone can create a module for the Cosmos-SDK, and using the already-built modules in your blockchain is as simple as importing them into your application.  As a developer, you only have to create the modules required by your application that do not already exist.

- **Capabilities**: The SDK is inspired by capabilities-based security, and informed by years of wrestling with blockchain state machines. Most developers will need to access other modules when building their own modules. Given that the Cosmos-SDK is an open framework, we designed it using object-capabilities (_ocaps_) based principles. This is because we assume that some modules are malicious. In practice, this means that instead of having each module keep an access control list to give access to other modules, each module implements `keepers` that can be passed to other modules to grant a pre-defined set of capabilities. For example, if an instance of module A's `keepers` is passed to module B, the latter will be able to call a restricted set of module A's functions. The capabilities of each `keeper` are defined by the module's developer, and it's the job of the application developer to instantiate and pass a `keeper` from module to module properly. For a deeper look at capabilities, you can read this [article](http://habitatchronicles.com/2017/05/what-are-capabilities/).

_Note: For now the Cosmos-SDK only exists in [Golang](https://golang.org/), which means that developers can only develop SDK modules in Golang. In the future, we expect that the SDK will be implemented in other programming languages. The Cosmos team will also financially support their development._

## Application architecture

#### Modules

A `module` is a fundamental unit in the Cosmos-SDK. Each module is an extension to the BaseApp functionalities that defines its own transactions, handles its own state as well as the state transition logic. The Cosmos-SDK has all the necessary pre-built modules to add functionality on top of a `BaseApp`, which is the template to build a blockchain dApp in Cosmos. Some of the most important ones are:

- `Auth` - Defines a `BaseAccount` structure and how transaction signers are authenticated.
- `Bank` - Defines how coins (i.e cryptocurrencies) are transferred.
- `Governance` -  Governance related specifications including proposals and voting.
- `Staking` - Proof of Stake related specifications including bonding and delegation transactions, inflation, fees, unbonding, etc.
- `IBC` - Defines the intereoperability of blockchain zones according to the specifications of the [IBC Protocol](https://cosmos.network/whitepaper#inter-blockchain-communication-ibc).
- Handlers for messages and transactions
- REST and CLI for secure user interactions

#### Directories

The key directories of the SDK are:

- `baseapp`: Defines the template for a basic application. It implements the [ABCI protocol](https://cosmos.network/whitepaper#abci) so that your Cosmos-SDK application can communicate with the underlying Tendermint node.
- `client`: Command-Line to interface with the application.
- `server`: Rest server to communicate with the node.
- `examples`: Contains example on how to build a working application based on baseapp and module.
- `store`: Contains code for the multistore (i.e state). Each module can create any number of KVStores from the multistore. Be careful to properly handle access rights to each store with appropriate keepers.
- `types`: Common types required in any SDK-based application.
- `x`(for e**X**tensions): Folder for storing the BaseApp module. You will find all the already-built modules in this directory. To use any of these modules, you just need to properly import them in your application.

## Prerequisites

- [Golang](https://golang.org/doc/install)

## Getting Started

See the [documentation](https://cosmos-sdk.readthedocs.io).
