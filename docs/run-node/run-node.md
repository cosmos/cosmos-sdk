<!--
order: 2
-->

# Running a Node

Now that the application is ready and the keyring populated, it's time to see how to run the blockchain node. In this section, the application we are running is called [`simapp`](https://github.com/cosmos/cosmos-sdk/tree/v0.40.0-rc3/simapp), and its corresponding CLI binary `simd`. {synopsis}

## Pre-requisite Readings

- [Anatomy of an SDK Application](../basics/app-anatomy.md) {prereq}
- [Setting up the keyring](./keyring.md) {prereq}

## Initialize the Chain

::: warning
Make sure you can build your own binary, and replace `simd` with the name of your binary in the snippets.
:::

Before actually running the node, we need to initialize the chain, and most importantly its genesis file. This is done with the `init` subcommand:

```bash
# The argument <moniker> is the custom username of your node, it should be human-readable.
simd init <moniker> --chain-id my-test-chain
```

The command above creates all the configuration files needed for your node to run, as well as a default genesis file, which defines the initial state of the network. All these configuration files are in `~/.simapp` by default, but you can overwrite the location of this folder by passing the `--home` flag.

The `~/.simapp` folder has the following structure:

```bash
.                                   # ~/.simapp
  |- data                           # Contains the databases used by the node.
  |- config/
      |- app.toml                   # Application-related configuration file.
      |- config.toml                # Tendermint-related configuration file.
      |- genesis.json               # The genesis file.
      |- node_key.json              # Private key to use for node authentication in the p2p protocol.
      |- priv_validator_key.json    # Private key to use as a validator in the consensus protocol.
```

Before starting the chain, you need to populate the state with at least one account. To do so, first [create a new account in the keyring](./keyring.md#adding-keys-to-the-keyring) named `my_validator` under the `test` keyring backend (feel free to choose another name and another backend).

Now that you have created a local account, go ahead and grant it some `stake` tokens in your chain's genesis file. Doing so will also make sure your chain is aware of this account's existence:

```bash
simd add-genesis-account $MY_VALIDATOR_ADDRESS 100000000000stake
```

Recall that `$MY_VALIDATOR_ADDRESS` is a variable that holds the address of the `my_validator` key in the [keyring](./keyring.md#adding-keys-to-the-keyring). Also note that the tokens in the SDK have the `{amount}{denom}` format: `amount` is is a 18-digit-precision decimal number, and `denom` is the unique token identifier with its denomination key (e.g. `atom` or `uatom`). Here, we are granting `stake` tokens, as `stake` is the token identifier used for staking in [`simapp`](https://github.com/cosmos/cosmos-sdk/tree/v0.40.0-rc3/simapp). For your own chain with its own staking denom, that token identifier should be used instead.

Now that your account has some tokens, you need to add a validator to your chain. Validators are special full-nodes that participate in the consensus process (implemented in the [underlying consensus engine](../intro/sdk-app-architecture.md#tendermint)) in order to add new blocks to the chain. Any account can declare its intention to become a validator operator, but only those with sufficient delegation get to enter the active set (for example, only the top 125 validator candidates with the most delegation get to be validators in the Cosmos Hub). For this guide, you will add your local node (created via the `init` command above) as a validator of your chain. Validators can be declared before a chain is first started via a special transaction included in the genesis file called a `gentx`:

```bash
# Create a gentx.
simd gentx my_validator 100000000stake --chain-id my-test-chain --keyring-backend test

# Add the gentx to the genesis file.
simd collect-gentxs
```

A `gentx` does three things:

1. Registers the `validator` account you created as a validator operator account (i.e. the account that controls the validator).
2. Self-delegates the provided `amount` of staking tokens.
3. Link the operator account with a Tendermint node pubkey that will be used for signing blocks. If no `--pubkey` flag is provided, it defaults to the local node pubkey created via the `simd init` command above.

For more information on `gentx`, use the following command:

```bash
simd gentx --help
```

## Run a Localnet

Now that everything is set up, you can finally start your node:

```bash
simd start
```

You should see blocks come in.

The previous command allow you to run a single node. This is enough for the next section on interacting with this node, but you may wish to run multiple nodes at the same time, and see how consensus happens between them.

The naive way would be to run the same commands again in separate terminal windows. This is possible, however in the SDK, we leverage the power of [Docker Compose](https://docs.docker.com/compose/) to run a localnet. If you need inspiration on how to set up your own localnet with Docker Compose, you can have a look at the SDK's [`docker-compose.yml`](https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/docker-compose.yml).

## Configuring the Node Using `app.toml`

The Cosmos SDK automatically generates an `app.toml` file inside `~/.simapp/config`. This file is used to configure your app, such as state pruning strategies, telemetry, gRPC and REST servers configuration, state sync... The file itself is heavily commented, please refer to it directly to tweak your node.

Make sure to restart your node after modifying `app.toml`.

## Next {hide}

Read about the [Interacting with your Node](./interact-node.md) {hide}
