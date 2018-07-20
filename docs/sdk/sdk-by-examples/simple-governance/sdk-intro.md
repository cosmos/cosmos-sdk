## Introduction to the Cosmos-SDK

Developing a Tendermint-based blockchain means that you only have to code the application (i.e. the state-machine). But that in itself can prove to be rather difficult. This is why the Cosmos-SDK exists.

The [Cosmos-SDK](https://github.com/cosmos/cosmos-sdk) is a platform for building multi-asset Proof-of-Stake (PoS) blockchains, like the Cosmos Hub, as well as Proof-Of-Authority (PoA) blockchains.

The goal of the Cosmos-SDK is to allow developers to easily create custom interoperable blockchain applications within the Cosmos Network without having to recreate common blockchain functionality, thus removing the complexity of building a Tendermint ABCI application. We envision the SDK as the npm-like framework to build secure blockchain applications on top of Tendermint.

In terms of its design, the SDK optimizes flexibility and security. The framework is designed around a modular execution stack which allows applications to mix and match elements as desired. In addition, all modules are sandboxed for greater application security.

It is based on two major principles:

- **Composability:** Anyone can create a module for the Cosmos-SDK and integrating the already-built modules is as simple as importing them into your blockchain application.

- **Capabilities:** The SDK is inspired by capabilities-based security, and informed by years of wrestling with blockchain state-machines. Most developers will need to access other 3rd party modules when building their own modules. Given that the Cosmos-SDK is an open framework and that we assume that some of those modules may be malicious, we designed the SDK using object-capabilities (ocaps) based principles. In practice, this means that instead of having each module keep an access control list for other modules, each module implements special objects called keepers that can be passed to other modules to grant a pre-defined set of capabilities. For example, if an instance of module A's keepers is passed to module B, the latter will be able to call a restricted set of module A's functions. The capabilities of each keeper are defined by the module's developer, and it's the developer's job to understand and audit the safety of foreign code from 3rd party modules based on the capabilities they are passing into each 3rd party module. For a deeper look at capabilities, you can read this [article](http://habitatchronicles.com/2017/05/what-are-capabilities/).

*Note: For now the Cosmos-SDK only exists in Golang, which means that developers can only develop SDK modules in Golang. In the future, we expect that the SDK will be implemented in other programming languages. Funding opportunities supported by the Tendermint team may be available eventually.*