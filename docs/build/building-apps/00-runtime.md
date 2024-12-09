---
sidebar_position: 1
---

# What is `runtime`?

The `runtime` package is the Cosmos SDK package that combines the building blocks of your blockchain together. It wires together the modules, the applications, the codecs, and the stores.
It is a layer of abstraction between `baseapp` and the application modules that simplifies the process of building a Cosmos SDK application.

## Modules wiring

Runtime is responsible for wiring the modules together. It uses `depinject` to inject the dependencies of the modules.

## App wiring

Runtime is the base boilerplate of a Cosmos SDK application.
A user only needs to import `runtime` in their `app.go` and instantiate a `runtime.App`.

## Services

Modules have access to a multitude of services that are provided by the runtime.
These services include the `store`, the `event manager`, the `context`, and the `logger`.
As runtime is doing the wiring of modules, it can ensure that the services are scoped to their respective modules.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.52.0-beta.2/runtime/module.go#L250-L279
```
