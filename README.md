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


WARNING: the libraries are still undergoing breaking changes as we get better ideas and start building out the Apps.

**Note: Requires Go 1.9+**

The Cosmos SDK is a platform for building multi-asset Proof-of-Stake cryptocurrencies,
like the [Cosmos Hub](https://cosmos.network). It is both a library for building applications 
and a set of tools for securely interacting with them.

The goals of the SDK are to abstract away the complexities of building a Tendermint ABCI application in Golang
and to provide a framework for building interoperable blockchain applications in the Cosmos Network.

It is inspired by the routing and middleware model of many web application frameworks, and informed by years
of wrestling with blockchain state machines. 

The SDK is fast, safe, easy-to-use, and easy-to-reason-about.
It is generic enough to be used to implement the state machines of other existing blockchains,
like Bitcoin and Ethereum, allowing seamless integration with them and their data structures.
It comes with batteries included, is easily extensible, and does not require developers
to fork it to access any of its current or extended functionality.
It provides both REST and command line interfaces for secure user interactions.

Applications in the Cosmos-SDK are defined in terms of handlers that process messages
and read and write to a store. Handlers are given restricted capabilities that determine which 
parts of the store they can access. The SDK provides common data structures for Accounts,
multi-asset Coins, checking signatures, preventing replay, and so on.

For more details on the design goals, see the [Design Document](docs/design.md)

## Prerequisites

* [golang](https://golang.org/doc/install)

## Getting Started

- See the [SDK Basics](docs/basics.md)
