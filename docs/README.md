<!--
layout: homepage
title: Cosmos SDK Documentation
description: Cosmos SDK is the world’s most popular framework for building application-specific blockchains.
sections:
  - title: Introduction
    desc: High-level overview of the Cosmos SDK.
    url: /intro/overview.html
    icon: introduction
  - title: Basics
    desc: Anatomy of a blockchain, transaction lifecycle, accounts and more.
    icon: basics
    url: /basics/app-anatomy.html
  - title: Core Concepts
    desc: Read about the core concepts like `baseapp`, the store, or the server.
    icon: core
    url: /core/baseapp.html
  - title: Building Modules
    desc: Discover how to build modules for the Cosmos SDK.
    icon: modules
    url: /building-modules/intro.html
  - title: Running a Node
    desc: Running and interacting with nodes using the CLI and API.
    icon: interfaces
    url: /run-node/
  - title: Modules
    desc: Explore existing modules to build your application with.
    icon: specifications
    url: /modules/
stack:
  - title: Cosmos Hub
    desc: The first of thousands of interconnected blockchains on the Cosmos Network.
    color: "#BA3FD9"
    label: hub
    url: http://hub.cosmos.network
  - title: Tendermint Core
    desc: The leading BFT engine for building blockchains, powering Cosmos SDK.
    color: "#00BB00"
    label: core
    url: http://docs.tendermint.com
footer:
  newsletter: false
aside: false
-->

# Cosmos SDK Documentation

## Get Started

- **[SDK Intro](./intro/overview.md)**: High-level overview of the Cosmos SDK.
- **[Quick Start Guide](./using-the-sdk/quick-start.md)**: Scaffold a standard Cosmos SDK app and run a node.
- **[SDK Application Tutorial](https://github.com/cosmos/sdk-application-tutorial)**: A tutorial that showcases how to build an SDK-based blockchain from scratch and explains the basic principles of the SDK in the process.

## Reference

- **[Basics](./basics/)**: Documentation on the basic concepts of the Cosmos SDK, like the standard anatomy of an application, the transaction lifecycle and accounts management.
- **[Core](./core/)**: Documentation on the core concepts of the Cosmos SDK, like `baseapp`, the `store` or the `server`.
- **[Building Modules](./building-modules/)**: Important concepts for module developers like `message`s, `keeper`s, `handler`s and `querier`s.
- **[IBC](./ibc/)**: Documentation for the IBC protocol integration and concepts.
- **[Running a Node, API, CLI](./run-node/)**: Documentation on how to run a node, and how to interact with it using the CLI and the API.
- **[Migrations](./migrations/)**: Migration guides for updating to Stargate.

## Other Resources

- **[Module Directory](../x/)**: Module implementations and their respective documentation.
- **[Specifications](./spec/)**: Specifications of modules and other parts of the Cosmos SDK.
- **[SDK API Reference](https://godoc.org/github.com/cosmos/cosmos-sdk)**: Godocs of the Cosmos SDK.
- **[REST API spec](https://cosmos.network/rpc/)**: List of endpoints to interact with a `gaia` full-node through REST.

## Cosmos Hub

The Cosmos Hub (`gaia`) docs have moved [here](https://github.com/cosmos/gaia/tree/master/docs).

## Languages

The Cosmos-SDK is currently written in [Golang](https://golang.org/), though the
framework could be implemented similarly in other languages.
Contact us for information about funding an implementation in another language.

## Contribute

See [this file](https://github.com/cosmos/cosmos-sdk/blob/master/docs/DOCS_README.md) for details of the build process and
considerations when making changes.
