<!--
parent:
  order: false
-->

# Cosmos SDK

![banner](docs/cosmos-sdk-image.jpg)

[![version](https://img.shields.io/github/tag/cosmos/cosmos-sdk.svg)](https://github.com/cosmos/cosmos-sdk/releases/latest)
[![CircleCI](https://circleci.com/gh/cosmos/cosmos-sdk/tree/master.svg?style=shield)](https://circleci.com/gh/cosmos/cosmos-sdk/tree/master)
![Sims](https://github.com/cosmos/cosmos-sdk/workflows/Sims/badge.svg)
[![codecov](https://codecov.io/gh/cosmos/cosmos-sdk/branch/master/graph/badge.svg)](https://codecov.io/gh/cosmos/cosmos-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/cosmos/cosmos-sdk)](https://goreportcard.com/report/github.com/cosmos/cosmos-sdk)
[![license](https://img.shields.io/github/license/cosmos/cosmos-sdk.svg)](https://github.com/cosmos/cosmos-sdk/blob/master/LICENSE)
[![LoC](https://tokei.rs/b1/github/cosmos/cosmos-sdk)](https://github.com/cosmos/cosmos-sdk)
[![API Reference](https://godoc.org/github.com/cosmos/cosmos-sdk?status.svg)](https://godoc.org/github.com/cosmos/cosmos-sdk)
[![Discord chat](https://img.shields.io/discord/669268347736686612.svg)](https://discord.gg/AzefAFd)

The Cosmos-SDK is a framework for building blockchain applications in Golang.
It is being used to build [`Gaia`](https://github.com/cosmos/gaia), the first implementation of the Cosmos Hub.

**WARNING**: The SDK has mostly stabilized, but we are still making some
breaking changes.

**Note**: Requires [Go 1.14+](https://golang.org/dl/)

## Quick Start

To learn how the SDK works from a high-level perspective, go to the [SDK Intro](./docs/intro/overview.md).

If you want to get started quickly and learn how to build on top of the SDK, please follow the [SDK Application Tutorial](https://tutorials.cosmos.network/nameservice/tutorial/00-intro.html). You can also fork the tutorial's repository to get started building your own Cosmos SDK application.

For more, please go to the [Cosmos SDK Docs](./docs/).

## Cosmos Hub Mainnet

The Cosmos Hub application, `gaia`, has moved to its [own repository](https://github.com/cosmos/gaia). Go there to join the Cosmos Hub mainnet and more.

## Scaffolding

If you are starting a new app or a new module we provide a [scaffolding tool](https://github.com/cosmos/scaffold) to help you get started and speed up development. If you have any questions or find a bug, feel free to open an issue in the repo.

## Disambiguation

This Cosmos-SDK project is not related to the [React-Cosmos](https://github.com/react-cosmos/react-cosmos) project (yet). Many thanks to Evan Coury and Ovidiu (@skidding) for this Github organization name. As per our agreement, this disambiguation notice will stay here.
