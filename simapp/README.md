---
sidebar_position: 1
---

# `SimApp`

`SimApp` is a CLI application built using the Cosmos SDK for testing and educational purposes.

## Running testnets with `simd`

Except stated otherwise, all participants in the testnet must follow through with each step.

### 1. Download and Setup

Download the Cosmos SDK and unzip it. You can do this manually (via the GitHub UI) or with the git clone command.

```sh
git clone github.com/cosmos/cosmos-sdk.git
```

Next, run this command to build the `simd` binary in the `build` directory.

```sh
make build
```

Use the following command and skip all the next steps to configure your SimApp node:

```sh
make init-simapp
```

If you’ve run `simd` in the past, you may need to reset your database before starting up a new testnet. You can do that with this command:

```sh
# you need to provide the moniker and chain ID
$ ./simd init [moniker] --chain-id [chain-id]
```

The command should initialize a new working directory at the `~simapp` location. 

The `moniker` and `chain-id` can be anything but you need to use the same `chain-id` subsequently.


### 2. Create a New Key

Execute this command to create a new key.

```sh
 ./simd keys add [key_name]
```

The command will create a new key with your chosen name. 

⚠️ Save the output somewhere safe; you’ll need the address later.

### 3. Add Genesis Account

Add a genesis account to your testnet blockchain.

```sh
$ ./simd genesis add-genesis-account [key_name] [amount]
```

Where `key_name` is the same key name as before, and the `amount` is something like `10000000000000000000000000stake`.

### 4. Add the Genesis Transaction

This creates the genesis transaction for your testnet chain.

```sh
$ ./simd genesis gentx [key_name] [amount] --chain-id [chain-id]
```

The amount should be at least `1000000000stake`. When you start your node, providing too much or too little may result in errors.

### 5. Create the Genesis File

A participant must create the genesis file `genesis.json` with every participant's transaction. 

You can do this by gathering all the Genesis transactions under `config/gentx` and then executing this command.

```sh
$ ./simd genesis collect-gentxs
```

The command will create a new `genesis.json` file that includes data from all the validators. The command will create a new `genesis.json` file, including data from all the validators 

Once you've received the super genesis file, overwrite your original `genesis.json` file with
the new super `genesis.json`.

Modify your `config/config.toml` (in the simapp working directory) to include the other participants as
persistent peers:

```
# Comma-separated list of nodes to keep persistent connections to
persistent_peers = "[validator_address]@[ip_address]:[port],[validator_address]@[ip_address]:[port]"
```

You can find `validator_address` by executing:

```sh
$ ./simd comet show-node-id
```

The output will be the hex-encoded `validator_address`. The default `port` is 26656.

### 6. Start the Nodes

Finally, execute this command to start your nodes.

```sh
$ ./simd start
```

Now you have a small testnet that you can use to try out changes to the Cosmos SDK or CometBFT!

> ⚠️ NOTE: Sometimes, creating the network through the `collect-gents` will fail, and validators will start in a funny state (and then panic).
> 

If this happens, you can try to create and start the network first
with a single validator and then add additional validators using a `create-validator` transaction.
