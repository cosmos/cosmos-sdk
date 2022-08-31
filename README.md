<!--
parent:
  order: false
-->

<div align="center">
  <h1> Cosmos SDK </h1>
</div>

![banner](docs/cosmos-sdk-image.jpg)

<div align="center">
  <a href="https://github.com/cosmos/cosmos-sdk/releases/latest">
    <img alt="Version" src="https://img.shields.io/github/tag/cosmos/cosmos-sdk.svg" />
  </a>
  <a href="https://github.com/cosmos/cosmos-sdk/blob/master/LICENSE">
    <img alt="License: Apache-2.0" src="https://img.shields.io/github/license/cosmos/cosmos-sdk.svg" />
  </a>
  <a href="https://pkg.go.dev/github.com/cosmos/cosmos-sdk?tab=doc">
    <img alt="GoDoc" src="https://godoc.org/github.com/cosmos/cosmos-sdk?status.svg" />
  </a>
  <a href="https://goreportcard.com/report/github.com/cosmos/cosmos-sdk">
    <img alt="Go report card" src="https://goreportcard.com/badge/github.com/cosmos/cosmos-sdk" />
  </a>
  <a href="https://codecov.io/gh/cosmos/cosmos-sdk">
    <img alt="Code Coverage" src="https://codecov.io/gh/cosmos/cosmos-sdk/branch/master/graph/badge.svg" />
  </a>
</div>
<div align="center">
  <a href="https://github.com/cosmos/cosmos-sdk">
    <img alt="Lines Of Code" src="https://tokei.rs/b1/github/cosmos/cosmos-sdk" />
  </a>
  <a href="https://discord.gg/AzefAFd">
    <img alt="Discord" src="https://img.shields.io/discord/669268347736686612.svg" />
  </a>
  <a href="https://sourcegraph.com/github.com/cosmos/cosmos-sdk?badge">
    <img alt="Imported by" src="https://sourcegraph.com/github.com/cosmos/cosmos-sdk/-/badge.svg" />
  </a>
    <img alt="Sims" src="https://github.com/cosmos/cosmos-sdk/workflows/Sims/badge.svg" />
    <img alt="Lint Satus" src="https://github.com/cosmos/cosmos-sdk/workflows/Lint/badge.svg" />
</div>

The Cosmos SDK is a framework for building blockchain applications. [Tendermint Core (BFT Consensus)](https://github.com/tendermint/tendermint) and the Cosmos SDK are written in the Golang programming language. Cosmos SDK is used to build [Gaia](https://github.com/cosmos/gaia), the first implementation of the Cosmos Hub.

**WARNING**: The Cosmos SDK has mostly stabilized, but we are still making some
breaking changes.

**Note**: Requires [Go 1.18+](https://golang.org/dl/)

## Quick Start

To learn how the Cosmos SDK works from a high-level perspective, see the Cosmos SDK [High-Level Intro](./docs/intro/overview.md).

If you want to get started quickly and learn how to build on top of Cosmos SDK, visit [Cosmos SDK Tutorials](https://tutorials.cosmos.network). You can also fork the tutorial's repository to get started building your own Cosmos SDK application.

For more information, see the [Cosmos SDK Documentation](./docs/).

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for details how to contribute and participate in our [dev calls](./CONTRIBUTING.md#teams-dev-calls).
If you want to follow the updates or learn more about the latest design then join our [Discord](https://discord.com/invite/cosmosnetwork).

## Tools and Frameworks

The Cosmos ecosystem is vast. We will only make a few notable mentions here.

+ [Tools](https://v1.cosmos.network/tools): notable frameworks and modules.
+ [CosmJS](https://github.com/cosmos/cosmjs): the Swiss Army knife to power JavaScript based client solutions.

### Cosmos Hub Mainnet

The Cosmos Hub application, `gaia`, has moved to its own [cosmos/gaia repository](https://github.com/cosmos/gaia). Go there to join the Cosmos Hub mainnet and more.

### Inter-Blockchain Communication (IBC)

The IBC module for the Cosmos SDK has moved to its own [cosmos/ibc-go repository](https://github.com/cosmos/ibc-go). Go there to build and integrate with the IBC module.

### Ignite CLI

Ignite CLI is the all-in-one platform to build, launch, and maintain any crypto application on a sovereign and secured blockchain. If you are building a new app or a new module, use [Ignite CLI](https://github.com/ignite-hq/cli) to get started and speed up development.

## Disambiguation

This Cosmos SDK project is not related to the [React-Cosmos](https://github.com/react-cosmos/react-cosmos) project (yet). Many thanks to Evan Coury and Ovidiu (@skidding) for this Github organization name. As per our agreement, this disambiguation notice will stay here.
