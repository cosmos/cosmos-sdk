<div align="left">
  <h1> Cosmos SDK </h1>
</div>

![banner](docs/static/img/banner.svg)

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
</div>
<div align="center">
  <a href="https://discord.com/invite/interchain">
    <img alt="Discord" src="https://img.shields.io/discord/669268347736686612.svg" />
  </a>
  <a href="https://sourcegraph.com/github.com/cosmos/cosmos-sdk?badge">
    <img alt="Imported by" src="https://sourcegraph.com/github.com/cosmos/cosmos-sdk/-/badge.svg" />
  </a>
</div>

The Cosmos SDK is a modular, open-source blockchain SDK for building secure, high-performance Layer 1 chains with full customizability used by 200+ chains in production.   Developers can use the Cosmos SDK to easily and quickly spin up custom blockchains that can natively interoperate.

The Cosmos SDK is tailored for building secure, sovereign application-specific blockchains. Developers building with the Cosmos SDK can use predefined modules that cover standard blockchain functionality or create custom modules for their specific use case. This composable architecture enables robust customization. The SDK provides abstractions for permissioning, governance, state management, account abstraction, tokenization processes, application logic, and more.

Cosmos SDK blockchains get interoperability out-of-the-box via a native integration with the [Inter-Blockchain Communication Protocol (IBC)](https://github.com/cosmos/ibc-go). ibc-go is implemented as a Go module in the Cosmos SDK. 

While the Cosmos SDK is plug-and-play with any consensus engine, we recommend using [CometBFT](https://github.com/cometbft/cometbft) for a fast, battle-tested, high-throughput, configurable BFT state machine. CometBFT is developed as part of the Cosmos Stack and its releases are updated alongside the SDK.

**WARNING**: The Cosmos SDK has mostly stabilized, but we are still making some breaking changes.

## Quick Start

To learn how the Cosmos SDK works from a high-level perspective, see the Cosmos SDK [High-Level Intro](https://docs.cosmos.network/main/intro/overview).

If you want to get started quickly and learn how to build on top of Cosmos SDK, visit [Cosmos SDK Tutorials](https://tutorials.cosmos.network). You can also fork the tutorial's repository to get started building your own Cosmos SDK application.

Note: We advise to always use the latest maintained [Go version](https://go.dev/dl/) for building Cosmos SDK applications.

## Modules

The Cosmos SDK maintains a set of modules that can be included in your blockchain application.  For more information
on modules, see our [introduction doc](./x/README.md).

## Enterprise Modules

In addition to the core SDK modules, we maintain enterprise-grade modules designed for specialized use cases such as permissioned networks and consortium chains. These modules are located in the `enterprise/` directory and have different licensing terms than the core SDK.

## Maintainers
[Cosmos Labs](https://cosmoslabs.io/) maintains the core components of the stack: Cosmos SDK, CometBFT, IBC, Cosmos EVM, and various developer tools and frameworks. The detailed maintenance policy can be found [here](https://github.com/cosmos/security/blob/main/POLICY.md). In addition to developing and maintaining the Cosmos Stack, Cosmos Labs provides advisory and engineering services for blockchain solutions. [Get in touch with Cosmos Labs](https://www.cosmoslabs.io/contact).

Cosmos Labs is a wholly-owned subsidiary of the [Interchain Foundation](https://interchain.io/), the Swiss nonprofit responsible for treasury management, funding public goods, and supporting governance for Cosmos. 

The Cosmos Stack is supported by a robust community of open-source contributors. 

## History
The Cosmos SDK was first released in 2019, and the first blockchain to use the SDK in production was the [Cosmos Hub](https://hub.cosmos.network/main). Today, the Cosmos SDK is a popular, battle-tested, open-source framework used by hundreds of chains.

The Cosmos Hub still receives the most up-to-date Cosmos SDK versions. The Cosmos Hub application, `gaia`, has its own [cosmos/gaia repository](https://github.com/cosmos/gaia). 

## Developer Community and Support

The issue list of this repo is exclusively for bug reports and feature requests. We have active, helpful communities on Discord, Telegram, and Slack.

**| Need Help? | Support & Community: [Discord](https://discord.com/invite/interchain) - [Telegram](https://t.me/CosmosOG) - [Talk to an Expert](https://cosmos.network/interest-form) - [Join the #Cosmos-tech Slack Channel](https://forms.gle/A8jawLgB8zuL1FN36) |**

## Documentation and Resources
**View the Cosmos SDK documentation: https://docs.cosmos.network/**

### Cosmos Stack Libraries

- [CometBFT](https://github.com/cometbft/cometbft) - High-performance, 10k+ TPS configurable BFT consensus engine.
- [The Inter-Blockchain Communication Protocol (IBC)](https://github.com/cosmos/ibc-go/) - A blockchain interoperability protocol that allows blockchains to transfer any type of data encoded in bytes.
- [Cosmos EVM](https://github.com/cosmos/evm) - Native EVM layer for Cosmos SDK chains. 

## Disambiguation

This Cosmos SDK project is not related to the [React-Cosmos](https://github.com/react-cosmos/react-cosmos) project (yet). Many thanks to Evan Coury and Ovidiu (@skidding) for this Github organization name. As per our agreement, this disambiguation notice will stay here.
