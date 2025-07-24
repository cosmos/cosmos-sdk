---
sidebar_position: 1
---

# Hubl

`Hubl` is a tool that allows you to query any Cosmos SDK based blockchain.
It takes advantage of the new [AutoCLI](https://pkg.go.dev/github.com/cosmos/cosmos-sdk/client/v2@v2.0.0-20220916140313-c5245716b516/cli) feature <!-- TODO replace with AutoCLI docs --> of the Cosmos SDK.

## Installation

Hubl can be installed using `go install`:

```shell
go install cosmossdk.io/tools/hubl/cmd/hubl@latest
```

Or build from source:

```shell
git clone --depth=1 https://github.com/cosmos/cosmos-sdk
make hubl
```

The binary will be located in `tools/hubl`.

## Usage

```shell
hubl --help
```

### Add chain

To configure a new chain just run this command using the --init flag and the name of the chain as it's listed in the chain registry (<https://github.com/cosmos/chain-registry>).

If the chain is not listed in the chain registry, you can use any unique name.

```shell
hubl init [chain-name]
hubl init regen
```

The chain configuration is stored in `~/.hubl/config.toml`.

:::tip

When using an unsecure gRPC endpoint, change the `insecure` field to `true` in the config file.

```toml
[chains]
[chains.regen]
[[chains.regen.trusted-grpc-endpoints]]
endpoint = 'localhost:9090'
insecure = true
```

Or use the `--insecure` flag:

```shell
hubl init regen --insecure
```

:::

### Query

To query a chain, you can use the `query` command.
Then specify which module you want to query and the query itself.

```shell
hubl regen query auth module-accounts
```
