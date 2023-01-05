---
sidebar_position: 1
---

# Confix

`Confix` is a configuration management tool that allows you to manage your configuration via CLI.

It is based on the [Tendermint RFC 019](https://github.com/tendermint/tendermint/blob/5013bc3f4a6d64dcc2bf02ccc002ebc9881c62e4/docs/rfc/rfc-019-config-version.md).

## Installation

## Usage

Use standalone:

```shell
confix --help
```

Use in simd:

```shell
simd config fix --help
```

### Get

Get a configuration value, e.g.:

```shell
simd config get app pruning # gets the value pruning from app.toml
simd config get client chain-id # gets the value chain-id from client.toml
```

```shell
confix get ~/.simapp/config/app.toml pruning # gets the value pruning from app.toml
confix get ~/.simapp/config/client.toml chain-id # gets the value chain-id from client.toml
```

### Set

Set a configuration value, e.g.:

```shell
simd config set app pruning "enabled" # sets the value pruning from app.toml
simd config set client chain-id "foo-1" # sets the value chain-id from client.toml
```

```shell
confix set ~/.simapp/config/app.toml pruning "enabled" # sets the value pruning from app.toml
confix set ~/.simapp/config/client.toml chain-id "foo-1" # sets the value chain-id from client.toml
```

## Credits

This project is based on the [Tendermint RFC 019](https://github.com/tendermint/tendermint/blob/5013bc3f4a6d64dcc2bf02ccc002ebc9881c62e4/docs/rfc/rfc-019-config-version.md) and their own implementation of [confix](https://github.com/tendermint/tendermint/blob/v0.36.x/scripts/confix/confix.go).
