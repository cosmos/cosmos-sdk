# SDK Intro

The [Cosmos-SDK](https://github.com/cosmos/cosmos-sdk) is a framework for building multi-asset Proof-of-Stake (PoS) blockchains, like the Cosmos Hub, as well as Proof-Of-Authority (PoA) blockchains.

The goal of the Cosmos SDK is to allow developers to easily create custom  blockchains from scratch that can natively interoperate with other blockchains. We envision the SDK as the npm-like framework to build secure blockchain applications on top of [Tendermint](https://github.com/tendermint/tendermint).

It is based on two major principles:

- **Composability:** Anyone can create a module for the Cosmos-SDK, and integrating the already-built modules is as simple as importing them into your blockchain application.

- **Capabilities:** The SDK is inspired by capabilities-based security, and informed by years of wrestling with blockchain state-machines. Most developers will need to access other 3rd party modules when building their own modules. Given that the Cosmos-SDK is an open framework, some of the modules may be malicious, which means there is a need for security principles to reason about inter-module interactions. These principles are based on object-cababilities. In practice, this means that instead of having each module keep an access control list for other modules, each module implements special objects called keepers that can be passed to other modules to grant a pre-defined set of capabilities. For example, if an instance of module A's keepers is passed to module B, the latter will be able to call a restricted set of module A's functions. The capabilities of each keeper are defined by the module's developer, and it's the developer's job to understand and audit the safety of foreign code from 3rd party modules based on the capabilities they are passing into each third party module. For a deeper look at capabilities, jump to [this section](./ocap.md).

### Next, learn more about the [SDK Application Architecture](./sdk-app-architecture.md)