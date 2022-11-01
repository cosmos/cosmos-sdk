---
sidebar_position: 1
---

# Overview of `app.go`

Since `v0.47.0` the Cosmos SDK have made much easier wiring an `app.go` thanks to `depinject`.
This section is intended to provide an overview of the `SimApp` `app.go` file with App Wiring.

## `app_config.go`

The `app_config.go` file is intended to provide a single place to configure all modules parameters.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/0d8787c/simapp/app_config.go#L52-L233
```

## `app.go`

```go reference
https://github.com/cosmos/cosmos-sdk/blob/0d8787c/simapp/app.go#L94-L427
```
