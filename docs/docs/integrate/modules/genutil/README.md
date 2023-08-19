# `x/genutil`

## Concepts

The `genutil` package contains a variaety of genesis utility functionalities for usage within a blockchain application. Namely:

* Genesis transactions related (gentx)
* Commands for collection and creation of gentxs
* `InitChain` processing of gentxs
* Genesis file validation
* Genesis file migration
* CometBFT related initialization
    * Translation of an app genesis to a CometBFT genesis

## Client

### CLI

The genutil commands are available under the `genesis` subcommand.

#### add-genesis-account

Add a genesis account to `genesis.json`. Learn more [here](https://docs.cosmos.network/main/run-node/run-node#adding-genesis-accounts).

#### collect-gentxs

Collect genesis txs and output a `genesis.json` file.

```shell
simd genesis collect-gentxs
```

This will create a new `genesis.json` file that includes data from all the validators (we sometimes call it the "super genesis file" to distinguish it from single-validator genesis files).

#### gentx

Generate a genesis tx carrying a self delegation.

```shell
simd genesis gentx [key_name] [amount] --chain-id [chain-id]
```

This will create the genesis transaction for your new chain. Here `amount` should be at least `1000000000stake`.
If you provide too much or too little, you will encounter an error when starting a node.

#### migrate

Migrate genesis to a specified target (SDK) version.

```shell
simd genesis migrate [target-version]
```

:::tip
The `migrate` command is extensible and takes a `MigrationMap`. This map is a mapping of target versions to genesis migrations functions.
When not using the default `MigrationMap`, it is recommended to still call the default `MigrationMap` corresponding the SDK version of the chain and prepend/append your own genesis migrations.
:::

#### validate-genesis

Validates the genesis file at the default location or at the location passed as an argument.

```shell
simd genesis validate-genesis
```

:::warning
Validate genesis only validates if the genesis is valid at the **current application binary**. For validating a genesis from a previous version of the application, use the `migrate` command to migrate the genesis to the current version.
:::
