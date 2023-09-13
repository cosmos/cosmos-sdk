---
sidebar_position: 1
---

# `SimApp`

`SimApp` is an application built using the Cosmos SDK for testing and educational purposes.

## Running testnets with `simd`

If you want to spin up a quick testnet with your friends, you can follow these steps.
Unless otherwise noted, every step must be done by everyone who wants to participate
in this testnet.

1. From the root directory of the Cosmos SDK repository, run `$ make build`. This will build the
    `simd` binary inside a new `build` directory. The following instructions are run from inside
    the `build` directory.
2. If you've run `simd` before, you may need to reset your database before starting a new
    testnet. You can reset your database with the following command: `$ ./simd comet unsafe-reset-all`.
3. `$ ./simd init [moniker] --chain-id [chain-id]`. This will initialize a new working directory
    at the default location `~/.simapp`. You need to provide a "moniker" and a "chain id". These
    two names can be anything, but you will need to use the same "chain id" in the following steps.
4. `$ ./simd keys add [key_name]`. This will create a new key, with a name of your choosing.
    Save the output of this command somewhere; you'll need the address generated here later.
5. `$ ./simd genesis add-genesis-account [key_name] [amount]`, where `key_name` is the same key name as
    before; and `amount` is something like `10000000000000000000000000stake`.
6. `$ ./simd genesis gentx [key_name] [amount] --chain-id [chain-id]`. This will create the genesis
    transaction for your new chain. Here `amount` should be at least `1000000000stake`. If you
    provide too much or too little, you will encounter an error when starting your node.
7. Now, one person needs to create the genesis file `genesis.json` using the genesis transactions
   from every participant, by gathering all the genesis transactions under `config/gentx` and then
   calling `$ ./simd genesis collect-gentxs`. This will create a new `genesis.json` file that includes data
   from all the validators (we sometimes call it the "super genesis file" to distinguish it from
   single-validator genesis files).
8. Once you've received the super genesis file, overwrite your original `genesis.json` file with
    the new super `genesis.json`.
9. Modify your `config/config.toml` (in the simapp working directory) to include the other participants as
    persistent peers:

    ```text
    # Comma separated list of nodes to keep persistent connections to
    persistent_peers = "[validator_address]@[ip_address]:[port],[validator_address]@[ip_address]:[port]"
    ```

    You can find `validator_address` by running `$ ./simd comet show-node-id`. The output will
    be the hex-encoded `validator_address`. The default `port` is 26656.
10. Now you can start your nodes: `$ ./simd start`.

Now you have a small testnet that you can use to try out changes to the Cosmos SDK or CometBFT!

NOTE: Sometimes creating the network through the `collect-gentxs` will fail, and validators will start
in a funny state (and then panic). If this happens, you can try to create and start the network first
with a single validator and then add additional validators using a `create-validator` transaction.
