---
sidebar_position: 0
---

# Packages

The Cosmos SDK is a collection of Go modules. This section provides documentation on various packages that can used when developing a Cosmos SDK chain.
It lists all standalone Go modules that are part of the Cosmos SDK.

:::tip
For more information on SDK modules, see the [SDK Modules](https://docs.cosmos.network/main/modules) section.
For more information on SDK tooling, see the [Tooling](https://docs.cosmos.network/main/tooling) section.
:::

## Core

* [Core](https://pkg.go.dev/cosmossdk.io/core) - Core library defining SDK interfaces ([ADR-063](https://docs.cosmos.network/main/architecture/adr-063-core-module-api))
* [API](https://pkg.go.dev/cosmossdk.io/api) - API library containing generated SDK Pulsar API
* [Store](https://pkg.go.dev/cosmossdk.io/store) - Implementation of the Cosmos SDK store

## State Management

* [Collections](./02-collections.md) - State management library
* [ORM](./03-orm.md) - State management library

## Automation

* [Depinject](./01-depinject.md) - Dependency injection framework
* [Client/v2](https://pkg.go.dev/cosmossdk.io/client/v2) - Library powering [AutoCLI](https://docs.cosmos.network/main/core/autocli)

## Utilities

* [Log](https://pkg.go.dev/cosmossdk.io/log) - Logging library
* [Errors](https://pkg.go.dev/cosmossdk.io/errors) - Error handling library
* [Math](https://pkg.go.dev/cosmossdk.io/math) - Math library for SDK arithmetic operations

## Example

* [SimApp](https://pkg.go.dev/cosmossdk.io/simapp) - SimApp is **the** sample Cosmos SDK chain. This package should not be imported in your application.
