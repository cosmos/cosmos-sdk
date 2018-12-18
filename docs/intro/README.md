# SDK Intro

The [Cosmos-SDK](https://github.com/cosmos/cosmos-sdk) is a framework for building multi-asset Proof-of-Stake (PoS) blockchains, like the Cosmos Hub, as well as Proof-Of-Authority (PoA) blockchains.

The goal of the Cosmos SDK is to allow developers to easily create custom  blockchains from scratch that can natively interoperate with other blockchains. We envision the SDK as the npm-like framework to build secure blockchain applications on top of [Tendermint](https://github.com/tendermint/tendermint).

It is based on two major principles:

- **Composability:** Anyone can create a module for the Cosmos-SDK, and integrating the already-built modules is as simple as importing them into your blockchain application.

- **Capabilities:** The SDK is inspired by capabilities-based security, and informed by years of wrestling with blockchain state-machines. Most developers will need to access other 3rd party modules when building their own modules. Given that the Cosmos-SDK is an open framework, some of the modules may be malicious, which means there is a need for security principles to reason about inter-module interactions. These principles are based on object-cababilities. In practice, this means that instead of having each module keep an access control list for other modules, each module implements special objects called keepers that can be passed to other modules to grant a pre-defined set of capabilities. For example, if an instance of module A's keepers is passed to module B, the latter will be able to call a restricted set of module A's functions. The capabilities of each keeper are defined by the module's developer, and it's the developer's job to understand and audit the safety of foreign code from 3rd party modules based on the capabilities they are passing into each third party module. For a deeper look at capabilities, jump to [this section](./ocap.md).

## Learn more about the SDK

- [SDK application architecture](./sdk-app-architecture.md)
- [SDK security paradigm: ocap](./ocap.md)

## Creating a new SDK project

To create a new project, you can either:

- Fork [this repo](https://github.com/cosmos/sdk-application-tutorial/). Do not forget to remove the `nameservice` module from the various files if you don't need it. 
- Copy the `docs/examples/basecoin` directory. 
- Use community tools! More info to come.

## SDK Directory Structure

The SDK is laid out in the following directories:

- `baseapp`: Defines the template for a basic [ABCI](https://github.com/tendermint/tendermint/tree/master/abci) application so that your Cosmos-SDK application can communicate with the underlying Tendermint node.
- `client`: CLI and REST server tooling for interacting with SDK application.
- `docs/examples`: Examples of how to build working standalone applications.
- `server`: The full node server for running an SDK application on top of
  Tendermint.
- `store`: The database of the SDK - a Merkle multistore supporting multiple types of underling Merkle key-value stores.
- `types`: Common types in SDK applications.
- `x`: Extensions to the core, where all messages and handlers are defined.

## Languages

The Cosmos-SDK is currently written in [Golang](https://golang.org/), though the
framework could be implemented similarly in other languages.
Contact us for information about funding an implementation in another language.
