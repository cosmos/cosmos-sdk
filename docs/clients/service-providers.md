# Service Providers

We define 'service providers' as entities providing services for end-users that involve some form of interaction with a Cosmos-SDK based blockchain (this includes the Cosmos Hub). More specifically, this document will be focused around interactions with tokens.

This section does not concern wallet builders that want to provide [Light-Client](https://github.com/cosmos/cosmos-sdk/tree/develop/docs/light) functionalities. Service providers are expected to act as trusted point of contact to the blockchain for their end-users. 

## High-level description of the architecture

There are three main pieces to consider:

- Full-nodes: To interact with the blockchain. 
- Rest Server: This acts as a relayer for HTTP calls.
- Rest API: Define available endpoints for the Rest Server.

## Running a Full-Node

### Installation and configuration

We will describe the steps to run and interract with a full-node for the Cosmos Hub. For other SDK-based blockchain, the process should be similar. 

First, you need to [install the software](../getting-started/installation.md).

Then, you can start [running a full-node](../getting-started/join-testnet.md).

### Command-Line interface

Next you will find a few useful CLI commands to interact with the Full-Node.

#### Creating a key-pair

To generate a new key (default secp256k1 elliptic curve):

```bash
gaiacli keys add <your_key_name>
```

You will be asked to create a passwords (at least 8 characters) for this key-pair. The command returns 4 informations:

- `NAME`: Name of your key
- `ADDRESS`: Your address. Used to receive funds.
- `PUBKEY`: Your public key. Useful for validators.
- `Seed phrase`: 12-words phrase. **Save this seed phrase somewhere safe**. It is used to recover your private key in case you forget the password.

You can see all your available keys by typing:

```bash
gaiacli keys list
```

#### Checking your balance

After receiving tokens to your address, you can view your account's balance by typing:

```bash
gaiacli account <YOUR_ADDRESS>
```

*Note: When you query an account balance with zero tokens, you will get this error: No account with address <YOUR_ADDRESS> was found in the state. This is expected! We're working on improving our error messages.*

#### Sending coins via the CLI

Here is the command to send coins via the CLI:

```bash
gaiacli send --amount=10faucetToken --chain-id=<name_of_testnet_chain> --from=<key_name> --to=<destination_address>
```

Flags:
- `--amount`: This flag accepts the format `<value|coinName>`.
- `--chain-id`: This flag allows you to specify the id of the chain. There will be different ids for different testnet chains and main chain.
- `--from`: Name of the key of the sending account.
- `--to`: Address of the recipient.

#### Help

If you need to do something else, the best command you can run is:

```bash
gaiacli 
```

It will display all the available commands. For each command, you can use the `--help` flag to get further information. 

## Setting up the Rest Server

The Rest Server acts as an intermediary between the front-end and the full-node. You don't need to run the Rest Server on the same machine as the full-node. 

To start the Rest server: 

```bash
gaiacli advanced rest-server --node=<full_node_address:full_node_port>
```

Flags:
- `--trust-node`: A boolean. If `true`, light-client verification is disabled. If `false`, it is disabled. For service providers, this should be set to `true`. By default, it set to `true`. 
- `--node`: This is where you indicate the address and the port of your full-node. The format is <full_node_address:full_node_port>. If the full-node is on the same machine, the address should be `tcp://localhost:26657`.
- `--laddr`: This flag allows you to specify the address and port for the Rest Server (default `1317`). You will mostly use this flag only to specify the port, in which case just input "localhost" for the address. The format is <rest_server_address:port>.


### Listening for incoming transaction

The recommended way to listen for incoming transaction is to periodically query the blockchain through the following endpoint of the LCD:

[`/bank/balance/{account}`](https://cosmos.network/rpc/#/ICS20/get_bank_balances__address_)

## Rest API

The Rest API documents all the available endpoints that you can use to interact
with your full node. It can be found [here](https://cosmos.network/rpc/).

The API is divided into ICS standards for each category of endpoints. For
example, the [ICS20](https://cosmos.network/rpc/#/ICS20/) describes the API to
interact with tokens.

To give more flexibility to implementers, we have included the ability to
generate unsigned transactions, [sign](https://cosmos.network/rpc/#/ICS20/post_tx_sign)
and [broadcast](https://cosmos.network/rpc/#/ICS20/post_tx_broadcast) them with
different API endpoints. This allows service providers to use their own signing
mechanism for instance.

In order to generate an unsigned transaction (example with
[coin transfer](https://cosmos.network/rpc/#/ICS20/post_bank_accounts__address__transfers)),
you need to use the field `generate_only` in the body of `base_req`.
