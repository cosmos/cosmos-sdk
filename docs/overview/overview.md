# Overview

The Cosmos-SDK is a framework for building Tendermint ABCI applications in
Golang. It is designed to allow developers to easily create custom interoperable 
blockchain applications within the Cosmos Network.

We envision the SDK as the `npm`-like framework to build secure blockchain applications on top of Tendermint.

To achieve its goals of flexibility and security, the SDK makes extensive use of
the [object-capability
model](https://en.wikipedia.org/wiki/Object-capability_model) 
and the [principle of least
privelege](https://en.wikipedia.org/wiki/Principle_of_least_privilege).

For an introduction to object-capabilities, see this [article](http://habitatchronicles.com/2017/05/what-are-capabilities/).

## Languages

The Cosmos-SDK is currently writen in [Golang](https://golang.org/), though the
framework could be implemented similarly in other languages.
Contact us for information about funding an implementation in another language.

## Directory Structure

The SDK is laid out in the following directories:

- `baseapp`: Defines the template for a basic [ABCI](https://cosmos.network/whitepaper#abci) application so that your Cosmos-SDK application can communicate with the underlying Tendermint node.
- `client`: CLI and REST server tooling for interacting with SDK application.
- `examples`: Examples of how to build working standalone applications.
- `server`: The full node server for running an SDK application on top of
  Tendermint.
- `store`: The database of the SDK - a Merkle multistore supporting multiple types of underling Merkle key-value stores.
- `types`: Common types in SDK applications.
- `x`: Extensions to the core, where all messages and handlers are defined.
