---
sidebar_position: 1
---

# Dependency Injection

:::note

### Pre-requisite Readings

* [Cosmos SDK Dependency Injection Framework](../building-apps/01-depinject.md)

:::

[`depinject`](../building-apps/01-depinject.md) is used to wire any module in `app.go`.
All core modules are already configured to support dependency injection.

Modules, to work with `depinject`, must defines its configuration and its requirements so that `depinject` can provide the right dependencies.

In brief, as a module developer, the following steps are required:

1. Define the module's configuration using Protobuf
2. Define the module's dependency in `x/{moduleName}/module.go`

A chain developer can then use the module by following these two steps:

1. Wire the module in `app_config.go`
2. Inject the keeper in `app.go`

## Module Configuration

The module available configuration is defined in a Protobuf file, located at `{moduleName}/module/v1/module.proto`.

```proto reference
https://github.com/cosmos/cosmos-sdk/blob/main/proto/cosmos/group/module/v1/module.proto
```

* `go_import` must point to the Go package of the custom module.
* Message fields are the module's configuration.
  That configuration can be set in the `app_config.go` / `app.yaml` file for a chain developer to configure the module.
  Taking `group` as example, a chain developer is able to decide, thanks to `uint64 max_metadata_len`, what is the maximum metatada length allowed for a group porposal.

  ```go reference
    https://github.com/cosmos/cosmos-sdk/blob/306a9a7/simapp/app_config.go#L202-L206
  ```

That message is generated using [`pulsar`](https://github.com/cosmos/cosmos-sdk/blob/main/scripts/protocgen-pulsar.sh) (by running make `proto-gen`).
In the case of the `group` module, this file is generated here: https://github.com/cosmos/cosmos-sdk/blob/main/api/cosmos/group/module/v1/module.pulsar.go.
The part that is relevant for the module configuration is:

```go reference
https://github.com/cosmos/cosmos-sdk/blob/306a9a7/api/cosmos/group/module/v1/module.pulsar.go#L514-L526
```

Once the message is defined, `module.go` must define what dependencies are required by the module.

## Dependency Definition

// provide
// appConfig module struct
// depinject.In
// depinject.Out

## Wiring

