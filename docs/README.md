---
layout: index
title: Cosmos SDK Documentation
description: The Cosmos-SDK is a framework for building blockchain applications in Golang. It is being used to build Gaia, the first implementation of the Cosmos Hub.
features:
  - cta: Use
    title: HelloChain
    desc: Quickly create your first “hello chain” application using the SDK.
    label: 15-20 min
    special: dark
    url: https://github.com/cosmos/hellochain
  - cta: Read
    title: Introduction to Cosmos SDK
    desc: Learn about all the parts of the Cosmos SDK.
    label: 5 min
    url: /intro/
  - cta: Learn
    title: SDK Tutorial
    desc: Build a complete blockchain application from scratch.
    label: 30-40 min
    url: https://github.com/cosmos/sdk-application-tutorial
sections:
  - title: Introduction
    desc: Short sentence describing this section.
  - title: Basics
    desc: Short sentence describing this section.
  - title: SDK Core
    desc: Short sentence describing this section.
  - title: Building Modules
    desc: Short sentence describing this section.
  - title: Using SDK
    desc: Short sentence describing this section.
  - title: Interfaces
    desc: Short sentence describing this section.
---

# Cosmos SDK Documentation

## Get Started

- **[SDK Intro](./intro/intro.md)**: High-level overview of the Cosmos SDK.
- **[SDK application tutorial](https://github.com/cosmos/sdk-application-tutorial)**: A tutorial to learn the SDK. It showcases how to build an SDK-based blockchain from scratch, and explains the basic principles of the SDK in the process.

## Resources

- [Specifications](./spec/README.md): Specifications of modules and other parts of the Cosmos SDK.
- [SDK API Reference](https://godoc.org/github.com/cosmos/cosmos-sdk): Godocs of the Cosmos SDK.
- [REST API spec](https://cosmos.network/rpc/): List of endpoints to interact with a `gaia` full-node through REST.

## Creating a new SDK project

To create a new project, you can either:

- Fork [this repo](https://github.com/cosmos/sdk-application-tutorial/). Do not forget to remove the `nameservice` module from the various files if you don't need it.
- Use community tools like [chainkit](https://github.com/blocklayerhq/chainkit).

## Cosmos Hub

The Cosmos Hub (`gaia`) docs have moved [here](https://github.com/cosmos/gaia/tree/master/docs).

## Languages

The Cosmos-SDK is currently written in [Golang](https://golang.org/), though the
framework could be implemented similarly in other languages.
Contact us for information about funding an implementation in another language.

## Contribute

See [this file](https://github.com/cosmos/cosmos-sdk/blob/master/docs/DOCS_README.md) for details of the build process and
considerations when making changes.

## Version

This documentation is built from the following commit:
