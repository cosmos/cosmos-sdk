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

The Cosmos-SDK is a developer toolkit which allows developers to easily create their custom blockchain applications within the Cosmos ecosystem without having to code every single functionality of their application. The SDK will be the `npm`-like framework to build secure blockchain applications on top of Tendermint.

The SDK design optimizes flexibility and security. The framework is designed around a modular execution stack which allows applications to mix and match elements as desired. In addition, all modules are sandboxed for greater application security. 

It is based on two major principles:

- **Composability**: Anyone can create a module for the Cosmos-SDK, and using already-built modules in your blockchain is as simple as importing them into your application.  As a developer, you only have to create the modules required by your application that do not already exist.

- **Capabilities**: Most developers will need to access other modules when building their own modules. Given that the Cosmos-SDK is an open framework, we designed it using object-capabilities (ocaps) based principles. This is because we assume the thread that some modules are malicious. In practice, this means that instead of having each module keep an access control list to give access to other modules, each module implements `keepers` that can be passed to other modules to grant a pre-defined set of capabilities. For example, if an instance of module A's `keepers` is passed to module B, module B will be able to call a restricted set of module A's functions. The capabilities of each `keeper` are defined by the module's developer, and it is the job of the application developer to instanciate and pass a `keeper` from module to module properly. For a deeper look at capabilities, you can read this [article](http://habitatchronicles.com/2017/05/what-are-capabilities/).

_Note: For now the Cosmos-SDK only exists in [Golang](https://golang.org/), which means that module developers can only develop SDK modules in Golang. In the future, we expect that Cosmos-SDK in other programming languages will pop up._

## Application architecture

A module is a fundamental unit in the Cosmos-SDK. A module defines its own transaction, handles its own state as well as its own state transition logic. The Cosmos-SDK has all the necessary pre-build modules to add functionality on top of a BaseApp, which is the template to build a blockchain dApp in Cosmos. Some of the most important ones are: 
- Auth
- Token
- Governance
- Staking
- Handlers for messages and transactions
- REST and CLI sor secure user interactions

Key directories of the SDK:

- `baseapp`: Defines the template for a basic application. Basically it implements the ABCI protocol so that your Cosmos-SDK application can communicate with the underlying Tendermint node.
- `client`: Command-Line to interface with the application.
- `server`: Rest server to communicate with the node.
- `examples`: Contains example on how to build a working application based on baseapp and modules
- `store`: Contains code for the multistore (i.e state). Each module can create any number of KVStores from the multistore. Be careful to properly handle access rights to each store with appropriate keepers.
- `types`: Common types required in any SDK-based application.
- `x`: Folder for storing the BaseApp modules. You will find all the already-built modules in this directory. To use any of these modules, you just need to properly import them in your application.

## Prerequisites

- [Golang](https://golang.org/doc/install)

## Getting Started

See the [documentation](https://cosmos-sdk.readthedocs.io).


