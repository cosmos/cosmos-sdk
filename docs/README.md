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
    desc: Read about the core concepts like baseapp, the store, or the server.
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

* **[Cosmos SDK Intro](./intro/overview.md)**: High-level overview of the Cosmos SDK.
* **[Ignite CLI](https://docs.ignite.com)**: A developer-friendly interface to the Cosmos SDK to scaffold, launch, and maintain any crypto application on a sovereign and secured blockchain.
* **[SDK Tutorials](https://tutorials.cosmos.network/)**: Tutorials that showcase how to build Cosmos SDK-based blockchains from scratch and explain the basic Cosmos SDK principles in the process.

## Reference Docs

* **[Basics](./basics/)**: Basic concepts of the Cosmos SDK, including the standard anatomy of an application, the transaction lifecycle, and accounts management.
* **[Core](./core/)**: Core concepts of the Cosmos SDK, including `baseapp`, the `store`, or the `server`.
* **[Building Modules](./building-modules/)**: Important concepts for module developers like `message`, `keeper`, and `querier`.
* **[IBC](https://ibc.cosmos.network/)**: IBC protocol integration and concepts.
* **[Running a Node, API, CLI](./run-node/)**: How to run a node and interact with the node using the CLI and the API.
* **[Migrations](./migrations/)**: Migration guides for updating to newer versions of Cosmos SDK.

## Other Resources

* **[Module Directory](../x/)**: Cosmos SDK module implementations and their respective documentation.
* **[Specifications](./spec/)**: Specifications of modules and other parts of the Cosmos SDK.
* **[Cosmos SDK API Reference](https://pkg.go.dev/github.com/cosmos/cosmos-sdk)**: Godocs of the Cosmos SDK.
* **[REST and RPC Endpoints](https://cosmos.network/rpc/)**: List of endpoints to interact with a `gaia` full-node.
* **[Rosetta API](./run-node/rosetta.md)**: Rosetta API integration.

## Cosmos Hub

The Cosmos Hub (`gaia`) docs have moved to [github.com/cosmos/gaia](https://github.com/cosmos/gaia/tree/main/docs).

## Languages

The Cosmos SDK is written in [Golang](https://golang.org/), though the framework could be implemented similarly in other languages. Contact us for information about funding an implementation in another language.

## Contribute

See the [DOCS_README.md](https://github.com/cosmos/cosmos-sdk/blob/main/docs/DOCS_README.md) for details of the build process and considerations when making changes.
