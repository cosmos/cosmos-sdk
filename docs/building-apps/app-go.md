<!--
order: 0
-->

# Overview of `app.go` and how to wire it up

This section is intended to provide an overview of the `app.go` file and is still a work in progress.
For now we invite you to read the [tutorials](https://tutorials.cosmos.network) for a deep dive on how to build a chain.

## `app.go`

Since `v0.47.0` the Cosmos SDK have made easier wiring an `app.go` thanks to dependency injection:

+++ https://github.com/cosmos/cosmos-sdk/blob/main/simapp/app_config.go

+++ https://github.com/cosmos/cosmos-sdk/blob/main/simapp/app.go

## `app_legacy.go`

+++ https://github.com/cosmos/cosmos-sdk/blob/main/simapp/app_legacy.go