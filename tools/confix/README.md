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

## Credits

This project is based on the [Tendermint RFC 019](https://github.com/tendermint/tendermint/blob/5013bc3f4a6d64dcc2bf02ccc002ebc9881c62e4/docs/rfc/rfc-019-config-version.md) and their own implementation of [confix](https://github.com/tendermint/tendermint/blob/v0.36.x/scripts/confix/confix.go).
