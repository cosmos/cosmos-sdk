<div align="center">
  <h1> Cosmos SDK </h1>
</div>

![banner](https://github.com/cosmos/cosmos-sdk-docs/blob/main/static/img/banner.jpg)

<div align="center">
  <a href="https://github.com/cosmos/cosmos-sdk/blob/main/LICENSE">
    <img alt="License: Apache-2.0" src="https://img.shields.io/github/license/cosmos/cosmos-sdk.svg" />
  </a>
  <a href="https://pkg.go.dev/github.com/cosmos/cosmos-sdk">
    <img src="https://pkg.go.dev/badge/github.com/cosmos/cosmos-sdk.svg" alt="Go Reference">
  </a>
  <a href="https://goreportcard.com/report/github.com/cosmos/cosmos-sdk">
    <img alt="Go report card" src="https://goreportcard.com/badge/github.com/cosmos/cosmos-sdk" />
  </a>
  <a href="https://sonarcloud.io/summary/overall?id=cosmos_cosmos-sdk">
    <img alt="Code Coverage" src="https://sonarcloud.io/api/project_badges/measure?project=cosmos_cosmos-sdk&metric=coverage" />
  </a>
  <a href="https://sonarcloud.io/summary/overall?id=cosmos_cosmos-sdk">
    <img alt="SonarCloud Analysis" src="https://sonarcloud.io/api/project_badges/measure?project=cosmos_cosmos-sdk&metric=alert_status">
  </a>
</div>
<div align="center">
  <a href="https://discord.gg/interchain">
    <img alt="Discord" src="https://img.shields.io/discord/669268347736686612.svg" />
  </a>
  <a href="https://sourcegraph.com/github.com/cosmos/cosmos-sdk?badge">
    <img alt="Imported by" src="https://sourcegraph.com/github.com/cosmos/cosmos-sdk/-/badge.svg" />
  </a>
    <img alt="Sims" src="https://github.com/cosmos/cosmos-sdk/workflows/Sims/badge.svg" />
    <img alt="Lint Status" src="https://github.com/cosmos/cosmos-sdk/workflows/Lint/badge.svg" />
</div>

The Cosmos SDK is a framework for building blockchain applications. [CometBFT (BFT Consensus)](https://github.com/cometbft/cometbft) and the Cosmos SDK are written in the Go programming language. Cosmos SDK is used to build [Gaia](https://github.com/cosmos/gaia), the implementation of the Cosmos Hub.

**WARNING**: The Cosmos SDK has mostly stabilized, but we are still making some breaking changes.

**Note**: Always use the latest maintained [Go](https://go.dev/dl) version for building Cosmos SDK applications.

## Quick Start

To learn how the Cosmos SDK works from a high-level perspective, see the Cosmos SDK [High-Level Intro](https://docs.cosmos.network/main/learn/intro/overview).

If you want to get started quickly and learn how to build on top of Cosmos SDK, visit [Cosmos SDK Tutorials](https://tutorials.cosmos.network). You can also fork the tutorial's repository to get started building your own Cosmos SDK application.

For more information, see the [Cosmos SDK Documentation](https://docs.cosmos.network).

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for details on how to contribute and participate in our [dev calls](./CONTRIBUTING.md#teams-dev-calls).
If you want to follow the updates or learn more about the latest design then join our [Discord](https://discord.gg/interchain).

## Tools and Frameworks

The Cosmos ecosystem is vast.
[Awesome Cosmos](https://github.com/cosmos/awesome-cosmos) is a community-curated list of notable frameworks, modules and tools.

### Inter-Blockchain Communication (IBC)

The IBC module for the Cosmos SDK has its own [cosmos/ibc-go repository](https://github.com/cosmos/ibc-go). Go there to build and integrate with the IBC module.

### Version Matrix

The version matrix below shows which versions of the Cosmos SDK, modules and libraries are compatible with each other.

#### Core Dependencies

Core Dependencies are the core libraries that a application may depend on. 

> Note: the ❌ signals that the version of the Cosmos SDK does not need to import the dependency.

| Cosmos SDK | cosmossdk.io/core | cosmossdk.io/api | cosmossdk.io/math | cosmossdk.io/errors | cosmossdk.io/depinject | cosmossdk.io/log | cosmossdk.io/store |
|------------|-------------------|------------------|-------------------|---------------------|------------------------|------------------|--------------------|
| 0.50.z     | 0.11.z            | 0.7.z            | 1.y.z             | 1.y.z               | 1.y.z                  | 1.y.z            | 1.y.z              |
| 0.47.z     | 0.5.z             | 0.3.z            | 1.y.z             | 1.y.z               | 1.y.z                  | 1.y.z            | ❌                  |
| 0.46.z     | ❌                 | ❌                | 1.y.z             | 1.y.z               | ❌                      | ❌                | ❌                  |

#### Module Dependencies

Module Dependencies are the modules that a application may depend on and which version of the Cosmos SDK they are compatible with.

> Note: The version table only goes back to 0.50.x, this is due to the reason that modules were not spun out into their own go.mods until 0.50.z. ❌ signals that the module was not spun out into its own go.mod file.


| Cosmos SDK                  | 0.50.z    | 0.y.z |
|-----------------------------|-----------|-------|
| cosmossdk.io/x/auth         | ❌         |       |
| cosmossdk.io/x/accounts     | ❌         |       |
| cosmossdk.io/x/bank         | ❌         |       |
| cosmossdk.io/x/circuit      | 0.1.z     |       |
| cosmossdk.io/x/consensus    | ❌         |       |
| cosmossdk.io/x/distribution | ❌         |       |
| cosmossdk.io/x/evidence     | 0.1.z     |       |
| cosmossdk.io/x/feegrant     | 0.1.z     |       |
| cosmossdk.io/x/gov          | ❌         |       |
| cosmossdk.io/x/group        | ❌         |       |
| cosmossdk.io/x/mint         | ❌         |       |
| cosmossdk.io/x/nft          | 0.1.z     |       |
| cosmossdk.io/x/protcolpool  | ❌         |       |
| cosmossdk.io/x/slashing     | ❌         |       |
| cosmossdk.io/x/staking      | ❌         |       |
| cosmossdk.io/x/tx           | =< 0.13.z |       |
| cosmossdk.io/x/upgrade      | 0.1.z     |       |




## Disambiguation

This Cosmos SDK project is not related to the [React-Cosmos](https://github.com/react-cosmos/react-cosmos) project (yet). Many thanks to Evan Coury and Ovidiu (@skidding) for this Github organization name. As per our agreement, this disambiguation notice will stay here.
