---
sidebar_position: 1
---

# Dependency Injection

:::note

### Pre-requisite Readings

* [Cosmos SDK Dependency Injection Framework](../tooling/02-depinject.md)

:::

[`depinject`](../tooling/02-depinject.md) is used to wire any module in `app.go`.
All core modules are already configured to support dependency injection.

To work with `depinject` a module must define its configuration and requirements so that `depinject` can provide the right dependencies.

In brief, as a module developer, the following steps are required:

1. Define the module configuration using Protobuf
2. Define the module dependencies in `x/{moduleName}/module.go`

A chain developer can then use the module by following these two steps:

1. Configure the module in `app_config.go` or `app.yaml`
2. Inject the module in `app.go`

## Module Configuration

The module available configuration is defined in a Protobuf file, located at `{moduleName}/module/v1/module.proto`.

```protobuf reference
https://github.com/cosmos/cosmos-sdk/blob/main/proto/cosmos/group/module/v1/module.proto
```

* `go_import` must point to the Go package of the custom module.
* Message fields define the module configuration.
  That configuration can be set in the `app_config.go` / `app.yaml` file for a chain developer to configure the module.  
  Taking `group` as example, a chain developer is able to decide, thanks to `uint64 max_metadata_len`, what the maximum metatada length allowed for a group porposal is.

  ```go reference
    https://github.com/cosmos/cosmos-sdk/blob/0d8787c/simapp/app_config.go#L202-L206
  ```

That message is generated using [`pulsar`](https://github.com/cosmos/cosmos-sdk/blob/main/scripts/protocgen-pulsar.sh) (by running `make proto-gen`).
In the case of the `group` module, this file is generated here: https://github.com/cosmos/cosmos-sdk/blob/main/api/cosmos/group/module/v1/module.pulsar.go.

The part that is relevant for the module configuration is:

```go reference
https://github.com/cosmos/cosmos-sdk/blob/0d8787c/api/cosmos/group/module/v1/module.pulsar.go#L514-L526
```

:::note
Pulsar is totally optional. The official [`protoc-gen-go`](https://developers.google.com/protocol-buffers/docs/reference/go-generated) can be used as well.
:::

## Dependency Definition

Once the configuration proto is defined, the module's `module.go` must define what dependencies are required by the module.
The boilerplate is similar for all modules.

:::warning

All methods, structs and their fields must be public for `depinject`.

:::

1. Import the module configuration generated package:

  ```go reference
  https://github.com/cosmos/cosmos-sdk/blob/main/x/group/module/module.go#L14-L15
  ```

  Define an `init()` function for defining the `providers` of the module configuration:  
  This registers the module configuration message and the wiring of the module.

  ```go reference
  https://github.com/cosmos/cosmos-sdk/blob/main/x/group/module/module.go#L184-L192
  ```

2. `ProvideModuleBasic` is calls `WrapAppModuleBasic` for wrapping the module `AppModuleBasic`, so that it can be injected and used by the runtime module.

  ```go reference
  https://github.com/cosmos/cosmos-sdk/blob/main/x/group/module/module.go#L194-L196
  ```

3. Define a struct that inherits `depinject.In` and define the module inputs (i.e. module dependencies):
   * `depinject` provides the right dependencies to the module.
   * `depinject` also checks that all dependencies are provided.

  :::tip
  For making a dependency optional, add the `optional:"true"` struct tag.  
  :::

  ```go reference
  https://github.com/cosmos/cosmos-sdk/blob/main/x/group/module/module.go#L198-L208
  ```

4. Define the module outputs with a public struct that inherits `depinject.Out`:
   The module outputs are the dependencies that the module provides to other modules. It is usually the module itself and its keeper.

  ```go reference
  https://github.com/cosmos/cosmos-sdk/blob/main/x/group/module/module.go#L210-L215
  ```

5. Create a function named `ProvideModule` (as called in 1.) and use the inputs for instantitating the module outputs.

  ```go reference
  https://github.com/cosmos/cosmos-sdk/blob/main/x/group/module/module.go#L217-L227
  ```

Following is the complete app wiring configuration for `group`:

```go reference
https://github.com/cosmos/cosmos-sdk/blob/0d8787c/x/group/module/module.go#L180-L227
```

The module is now ready to be used with `depinject` by a chain developer.

## App Wiring

The App Wiring is done in `app_config.go` / `app.yaml` and `app.go` and is explained in detail in the [overview of `app.go`](../building-apps/00-app-go.md).
