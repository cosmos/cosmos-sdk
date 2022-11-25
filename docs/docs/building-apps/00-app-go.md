---
sidebar_position: 1
---

# Overview of `app.go`

This section is intended to provide an overview of the `SimApp` `app.go` file and is still a work in progress.
For now please instead read the [tutorials](https://tutorials.cosmos.network) for a deep dive on how to build a chain.

<<<<<<< HEAD
The Cosmos SDK allows much easier wiring an `app.go` with App Wiring and the tool [`depinject`](../tooling/02-depinject.md).
Learn more about the rationale of App Wiring in [ADR-057](../architecture/adr-057-app-wiring.md).

Depinject, a tool used by this version of app.go is not stable and may still land breaking changes. 
:::

:::note

### Pre-requisite Readings

* [ADR 057: App Wiring](../architecture/adr-057-app-wiring.md)
* [Depinject Documentation](../tooling/02-depinject.md)

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

### Complete `app_config.go`
=======
## Complete `app.go`
>>>>>>> main

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-alpha1/simapp/app_legacy.go#L162-L503
```
