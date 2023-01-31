# dYdX Fork of CosmosSDK

This is a lightweight fork of CosmosSDK. The current version of the forked code resides on the [default branch](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/about-branches#about-the-default-branch).

## Making Changes to the Fork

1. Open a PR against the current default branch (i.e. `dydx-fork-v0.47.0-alpha2`).
2. Get approval, and merge.
3. After merging, update the `v4` repository's `go.mod`, and `go.sum` files with your merged `$COMMIT_HASH`.
4. (In `dydxprotocol/v4`) `go mod edit -replace github.com/cosmos/cosmos-sdk=github.com/dydxprotocol/cosmos-sdk@$COMMIT_HASH`
5. (In `dydxprotocol/v4`) `go mod tidy`
6. Open a PR in `dydxprotocol/v4` to bump the version of the fork.

## Fork maintenance

We'd like to keep the `main` branch up to date with `cosmos/cosmos-sdk`. You can utilize GitHub's [sync fork](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/working-with-forks/syncing-a-fork) button to accomplish this. ⚠️ Please only use this on the `main` branch, not on the fork branches as it will discard our commits.⚠️

Note that this doesn't pull in upstream tags, so in order to do this follow these steps:
1. `git fetch upstream`
2. `git push --tags`

## Updating CosmosSDK to new versions

When a new version of CosmosSDK is published, we may want to adopt the changes in our fork. This process can be somewhat tedious, but below are the recommended steps to accomplish this.

1. Ensure the `main` branch and all tags are up to date by following the steps above in "Fork maintenance".
2. Create a new branch off the desired CosmosSDK commit using tags. `git checkout -b dydx-fork-$VERSION <CosmosSDK repo's tag name>`. The new branch should be named something like `dydx-fork-$VERSION` where `$VERSION` is the version of CosmosSDK being forked (should match the CosmosSDK repo's tag name). i.e. `dydx-fork-v0.47.0-alpha2`.
3. Push the new branch.
4. Open a PR which cherry-picks each commit in the current default branch, in order, on to the new `dydx-fork-$VERSION` branch (note: you may want to consider creating multiple PRs for this process if there are difficulties or merge conflicts). For example, `git cherry-pick <commit hash>`.
5. Get approval, and merge.
6. Update `dydxprotocol/v4` by following the steps in "Making Changes to the fork" above.
7. Set `dydx-fork-$VERSION` as the [default branch](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-branches-in-your-repository/changing-the-default-branch) in this repository.


<div align="center">
  <h1> Cosmos SDK </h1>
</div>

![banner](docs/static/img/banner.jpg)

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
  <a href="https://codecov.io/gh/cosmos/cosmos-sdk">
    <img alt="Code Coverage" src="https://codecov.io/gh/cosmos/cosmos-sdk/branch/main/graph/badge.svg" />
  </a>
</div>
<div align="center">
  <a href="https://discord.gg/AzefAFd">
    <img alt="Discord" src="https://img.shields.io/discord/669268347736686612.svg" />
  </a>
  <a href="https://sourcegraph.com/github.com/cosmos/cosmos-sdk?badge">
    <img alt="Imported by" src="https://sourcegraph.com/github.com/cosmos/cosmos-sdk/-/badge.svg" />
  </a>
    <img alt="Sims" src="https://github.com/cosmos/cosmos-sdk/workflows/Sims/badge.svg" />
    <img alt="Lint Satus" src="https://github.com/cosmos/cosmos-sdk/workflows/Lint/badge.svg" />
</div>

The Cosmos SDK is a framework for building blockchain applications. [Tendermint Core (BFT Consensus)](https://github.com/tendermint/tendermint) and the Cosmos SDK are written in the Go programming language. Cosmos SDK is used to build [Gaia](https://github.com/cosmos/gaia), the implementation of the Cosmos Hub.

**WARNING**: The Cosmos SDK has mostly stabilized, but we are still making some breaking changes.

**Note**: Requires [Go 1.19+](https://go.dev/dl)

## Quick Start

To learn how the Cosmos SDK works from a high-level perspective, see the Cosmos SDK [High-Level Intro](https://docs.cosmos.network/main/intro/overview.html).

If you want to get started quickly and learn how to build on top of Cosmos SDK, visit [Cosmos SDK Tutorials](https://tutorials.cosmos.network). You can also fork the tutorial's repository to get started building your own Cosmos SDK application.

For more information, see the [Cosmos SDK Documentation](https://docs.cosmos.network).

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for details on how to contribute and participate in our [dev calls](./CONTRIBUTING.md#teams-dev-calls).
If you want to follow the updates or learn more about the latest design then join our [Discord](https://discord.com/invite/cosmosnetwork).

## Tools and Frameworks

The Cosmos ecosystem is vast.
[Awesome Cosmos](https://github.com/cosmos/awesome-cosmos) is a community-curated list of notable frameworks, modules and tools.

### Cosmos Hub Mainnet

The Cosmos Hub application, `gaia`, has its own [cosmos/gaia repository](https://github.com/cosmos/gaia). Go there to join the Cosmos Hub mainnet and more.

### Inter-Blockchain Communication (IBC)

The IBC module for the Cosmos SDK has its own [cosmos/ibc-go repository](https://github.com/cosmos/ibc-go). Go there to build and integrate with the IBC module.

### Ignite CLI

Ignite CLI is the all-in-one platform to build, launch, and maintain any crypto application on a sovereign and secured blockchain. If you are building a new app or a new module, use [Ignite CLI](https://github.com/ignite/cli) to get started and speed up development.

## Disambiguation

This Cosmos SDK project is not related to the [React-Cosmos](https://github.com/react-cosmos/react-cosmos) project (yet). Many thanks to Evan Coury and Ovidiu (@skidding) for this Github organization name. As per our agreement, this disambiguation notice will stay here.
