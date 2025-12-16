---
sidebar_position: 1
---

# What is `runtime`?

The `runtime` package in the Cosmos SDK provides a flexible framework for configuring and managing blockchain applications. It serves as the foundation for creating modular blockchain applications using a declarative configuration approach.

## Overview

The runtime package acts as a wrapper around the `BaseApp` and `ModuleManager`, offering a hybrid approach where applications can be configured both declaratively through configuration files and programmatically through traditional methods.
It is a layer of abstraction between `baseapp` and the application modules that simplifies the process of building a Cosmos SDK application.

## Core Components

### App Structure

The runtime App struct contains several key components:

```go
type App struct {
    *baseapp.BaseApp
    ModuleManager    *module.Manager
    configurator     module.Configurator
    config           *runtimev1alpha1.Module
    storeKeys        []storetypes.StoreKey
    // ... other fields
}
```

Cosmos SDK applications should embed the `*runtime.App` struct to leverage the runtime module.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.53.0/simapp/app_di.go#L60-L61
```

### Configuration

The runtime module is configured using App Wiring. The main configuration object is the [`Module` message](https://github.com/cosmos/cosmos-sdk/blob/v0.53.0-rc.2/proto/cosmos/app/runtime/v1alpha1/module.proto), which supports the following key settings:

* `app_name`: The name of the application
* `begin_blockers`: List of module names to call during BeginBlock
* `end_blockers`: List of module names to call during EndBlock
* `init_genesis`: Order of module initialization during genesis
* `export_genesis`: Order for exporting module genesis data
* `pre_blockers`: Modules to execute before block processing

Learn more about wiring `runtime` in the [next section](./01-app-go-di.md).

#### Store Configuration

By default, the runtime module uses the module name as the store key.
However it provides a flexible store key configuration through:

* `override_store_keys`: Allows customizing module store keys
* `skip_store_keys`: Specifies store keys to skip during keeper construction

Example configuration:

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.53.0/simapp/app_config.go#L133-L138
```

## Key Features

### 1. BaseApp and other Core SDK components integration

The runtime module integrates with the `BaseApp` and other core SDK components to provide a seamless experience for developers.

The developer only needs to embed the `runtime.App` struct in their application to leverage the runtime module.
The configuration of the module manager and other core components is handled internally via the [`AppBuilder`](#4-application-building).

### 2. Module Registration

Runtime has built-in support for [`depinject`-enabled modules](../building-modules/15-depinject.md).
Such modules can be registered through the configuration file (often named `app_config.go`), with no additional code required.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.53.0/simapp/app_config.go#L210-L216
```

Additionally, the runtime package facilitates manual module registration through the `RegisterModules` method. This is the primary integration point for modules not registered via configuration.

:::warning
Even when using manual registration, the module should still be configured in the `Module` message in AppConfig.
:::

```go
func (a *App) RegisterModules(modules ...module.AppModule) error
```

The SDK recommends using the declarative approach with `depinject` for module registration whenever possible.

### 3. Service Registration

Runtime registers all [core services](https://pkg.go.dev/cosmossdk.io/core) required by modules.
These services include `store`, `event manager`, `context`, and `logger`.
Runtime ensures that services are scoped to their respective modules during the wiring process.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.53.0/runtime/module.go#L201-L235
```

Additionally, runtime provides automatic registration of other essential (i.e., gRPC routes) services available to the App:

* AutoCLI Query Service
* Reflection Service
* Custom module services

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.53.0/runtime/builder.go#L52-L54
```

### 4. Application Building

The `AppBuilder` type provides a structured way to build applications:

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.53.0/runtime/builder.go#L14-L19
```

Key building steps:

1. Configuration loading
2. Module registration
3. Service setup
4. Store mounting
5. Router configuration

An application only needs to call `AppBuilder.Build` to create a fully configured application (`runtime.App`).

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.53.0/runtime/builder.go#L26-L57
```

More information on building applications can be found in the [next section](./02-app-building.md).

## Best Practices

1. **Module Order**: Carefully consider the order of modules in begin_blockers, end_blockers, and pre_blockers.
2. **Store Keys**: Use override_store_keys only when necessary to maintain clarity
3. **Genesis Order**: Maintain correct initialization order in init_genesis
4. **Migration Management**: Use order_migrations to control upgrade paths

### Migration Considerations

When upgrading between versions:

1. Review the migration order specified in `order_migrations`
2. Ensure all required modules are included in the configuration
3. Validate store key configurations
4. Test the upgrade path thoroughly
