---
sidebar_position: 1
---

# Overview of `app.go`

:::note Synopsis

Since `v0.47.0`, the Cosmos SDK allows much easier wiring an `app.go`, thanks to [`depinject`](../tooling/02-depinject.md).

:::

:::note

### Pre-requisite Readings

* [Cosmos SDK Dependency Injection Framework](../tooling/02-depinject.md)

:::

This section is intended to provide an overview of the `SimApp` `app.go` file with App Wiring.

## `app_config.go`

The `app_config.go` file is the single place to configure all modules parameters.

1. Create the `AppConfig` variable:

    ```go reference
    https://github.com/cosmos/cosmos-sdk/blob/0d8787c/simapp/app_config.go#L77-L78
    ```

2. Configure the `runtime` module:

    ```go reference
    https://github.com/cosmos/cosmos-sdk/blob/0d8787c/simapp/app_config.go#L79-L137
    ```

3. Configure the modules defined in the `BeginBlocker` and `EndBlocker` and the `tx` module:

    ```go reference
    https://github.com/cosmos/cosmos-sdk/blob/0d8787c/simapp/app_config.go#L138-L156
    ```

    ```go reference
    https://github.com/cosmos/cosmos-sdk/blob/0d8787c/simapp/app_config.go#L170-L173
    ```

### Full `app_config.go`

```go reference
https://github.com/cosmos/cosmos-sdk/blob/0d8787c/simapp/app_config.go#L52-L233
```

## `app.go`

`app.go` is the place where `SimApp` is constructed. `depinject.Inject` facilitates that by automatically wiring the app modules and keepers, provided an application configuration, `AppConfig`. `SimApp` is constructed, when calling the injected `*runtime.AppBuilder`, with `appBuilder.Build(...)`.    
In short `depinject` and the [`runtime` package](https://pkg.go.dev/github.com/cosmos/cosmos-sdk/runtime) abstract the wiring of the app, and the `AppBuilder` is the place where the app is constructed. [`runtime`](https://pkg.go.dev/github.com/cosmos/cosmos-sdk/runtime) takes care of registering the codecs, KV store, subspaces and instantiating `baseapp`.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/0d8787c/simapp/app.go#L227-L254
```

### Advanced Configuration

In advanced cases, it is possible to inject extra (module) configuration in a way that is not (yet) supported by `AppConfig`.  
In this case, use `depinject.Configs` for combining the extra configuration and `AppConfig`, and `depinject.Supply` to providing that extra configuration.
More information on how work `depinject.Configs` and `depinject.Supply` can be found in the [`depinject` documentation](https://pkg.go.dev/cosmossdk.io/depinject).

```go reference
https://github.com/cosmos/cosmos-sdk/blob/0d8787c/simapp/app.go#L193-L224
```

### Full `app.go`

```go reference
https://github.com/cosmos/cosmos-sdk/blob/0d8787c/simapp/app.go#L94-L427
```
