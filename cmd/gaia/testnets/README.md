# Connect to the `gaia-6002` Testnet

Note: We are aware this documentation is sub-par. We are working to
improve the tooling and the documentation to make this process as painless as
possible. In the meantime, join the
[Validator Chat](https://riot.im/app/#/room/#cosmos_validators:matrix.org)
for technical support. Thanks very much for your patience. :)

## Setting Up a New Node

These instructions are for setting up a brand new full node from scratch. If you ran a full node on a previous testnet, please skip to [Upgrading From Previous Testnet](#upgrading-from-previous-testnet).

### Install Go

Install `go` by following the [official docs](https://golang.org/doc/install).
**Go 1.10+** is required for the Cosmos SDK.

### Install Cosmos SDK

Next, let's install the testnet's version of the Cosmos SDK. 

```
mkdir -p $GOPATH/src/github.com/cosmos
cd $GOPATH/src/github.com/cosmos
git clone https://github.com/cosmos/cosmos-sdk
cd cosmos-sdk && git checkout v0.19.0
make get_tools && make get_vendor_deps && make install
```

That will install the `gaiad` and `gaiacli` binaries. Verify that everything is OK:

```
gaiad version
0.19.0-<commit>
```

### Node Setup

Create the required configuration files:

```
gaiad init
```

Name your node by editing the `moniker` in `$HOME/.gaiad/config/config.toml`. Note that only ASCII characters are supported. Using Unicode renders your node unconnectable.

```
# A custom human readable name for this node
moniker = "<your_custom_name>"
```

Your full node has been initialized! Please skip to [Genesis & Seeds](#genesis--seeds).

## Upgrading From Previous Testnet

These instructions are for full nodes that have ran on previous testnets and would like to upgrade to the latest testnet.

### Reset Data

First, remove the outdated files and reset the data.

```
rm $HOME/.gaiad/config/addrbook.json $HOME/.gaiad/config/genesis.json
gaiad unsafe_reset_all
```

Your node is now in a pristine state while keeping the original `priv_validator.json` and `config.toml`. If you had any sentry nodes or full nodes setup before, 
your node will still try to connect to them, but may fail if they haven't also
been upgraded.

**WARNING:** Make sure that every node has a unique `priv_validator.json`. Do not copy the `priv_validator.json` from an old node to multiple new nodes. Running two nodes with the same `priv_validator.json` will cause you to double sign.

### Software Upgrade

Now it is time to upgrade the software:

```
cd $GOPATH/src/github.com/cosmos/cosmos-sdk
git fetch --all && git checkout v0.19.0
make update_tools && make get_vendor_deps && make install
```

Your full node has been cleanly upgraded!

## Genesis & Seeds

### Copy the Genesis File

Copy the testnet's `genesis.json` file and place it in `gaiad`'s config directory.

```
mkdir -p $HOME/.gaiad/config
cp -a $GOPATH/src/github.com/cosmos/cosmos-sdk/cmd/gaia/testnets/gaia-6002/genesis.json $HOME/.gaiad/config/genesis.json
```

### Add Seed Nodes

Your node needs to know how to find peers. You'll need to add healthy seed nodes to `$HOME/.gaiad/config/config.toml`. Here are some seed nodes you can use: 

```
# Comma separated list of seed nodes to connect to
seeds = "38aa9bec3998f12ae9088b21a2d910d19d565c27@gaia-6002.coinculture.net:46656,80a35a46ce09cfb31ee220c8141a25e73e0b239b@seed.cosmos.cryptium.ch:46656,80a35a46ce09cfb31ee220c8141a25e73e0b239b@35.198.166.171:46656,032fa56301de335d835057fb6ad9f7ce2242a66d@165.227.236.213:46656"
```

You can also [ask other validators](https://riot.im/app/#/room/#cosmos_validators:matrix.org) for a persistent peer and add it under the `persistent_peers` key. For more information on seeds and peers, [read this](https://github.com/tendermint/tendermint/blob/develop/docs/using-tendermint.md#peers).

## Run a Full Node

Start the full node with this command:

```
gaiad start
```

Check that everything is running smoothly:

```
gaiacli status
```

View the status of the network with the [Cosmos Explorer](https://explorecosmos.network). Once your full node syncs up to the current block height, you should see it appear on the [list of full nodes](https://explorecosmos.network/validators). If it doesn't show up, that's ok--the Explorer does not connect to every node.

## Generate Keys

You'll need a private and public key pair \(a.k.a. `sk, pk` respectively\) to be able to receive funds, send txs, bond tx, etc.

To generate a new key \(default _ed25519_ elliptic curve\):

```
gaiacli keys add <your_key_name>
```

Next, you will have to create a passphrase. Save the _seed_ _phrase_ in a safe place in case you forget the password.

If you check your private keys, you'll now see `<your_key_name>`:

```
gaiacli keys show <your_key_name>
```

You can see all your available keys by typing:

```
gaiacli keys list
```

View the validator pubkey for your node by typing:

```
gaiad tendermint show_validator
```

Save your address and pubkey to environment variables for later use:

```
MYADDR=<your_newly_generated_address>
MYPUBKEY=<your_newly_generated_public_key>
```

**WARNING:** We strongly recommend NOT using the same passphrase for multiple keys. The Tendermint team and the Interchain Foundation will not be responsible for the loss of funds.

## Get Tokens

The best way to get tokens is from the [Cosmos Testnet Faucet](https://faucetcosmos.network). If the faucet is not working for you, try asking [#cosmos-validators](https://riot.im/app/#/room/#cosmos-validators:matrix.org).

After receiving tokens to your address, you can view your account's balance by typing: 

```
gaiacli account <your_newly_generated_address>
```

Note: When you query an account balance with zero tokens, you will get this error: `No account with address <your_newly_generated_address> was found in the state.` This is expected! We're working on improving our error messages.

## Send Tokens

```
gaiacli send --amount=10faucetToken --chain-id=<name_of_testnet_chain> --name=<key_name> --to=<destination_address>
```

Note: The `--amount` flag accepts the format `--amount=<value|coin_name>`.

Now, view the updated balances of the origin and destination accounts:

```
gaiacli account <origin_address>
gaiacli account <destination_address>
```

You can also check your balance at a given block by using the `--block` flag:

```
gaiacli account <your_address> --block=<block_height>
```

## Run a Validator Node

[Validators](https://cosmos.network/validators) are responsible for committing new blocks to the blockchain through voting. A validator's stake is slashed if they become unavailable, double sign a transaction, or don't cast their votes. If you only want to run a full node, a VM in the cloud is fine. However, if you are want to become a validator for the Hub's `mainnet`, you should research hardened setups. Please read [Sentry Node Architecture](https://github.com/cosmos/cosmos/blob/master/VALIDATORS_FAQ.md#how-can-validators-protect-themselves-from-denial-of-service-attacks) to protect your node from DDOS and ensure high-availability. Also see the [technical requirements](https://github.com/cosmos/cosmos/blob/master/VALIDATORS_FAQ.md#technical-requirements)). There's also more info on our [website](https://cosmos.network/validators).

Your `pubkey` can be used to create a new validator by staking tokens. You can find your validator pubkey by running:

```
gaiad tendermint show_validator
```

Next, craft your `gaiacli stake create-validator` command:

```
gaiacli stake create-validator --amount=5steak --pubkey=<your_node_pubkey> --address-validator=<your_address> --moniker=satoshi --chain-id=<name_of_the_testnet_chain> --name=<key_name>
```

You can add more information to the validator, such as`--website`, `--keybase-sig`, or `--details`. Here's how:

```
gaiacli stake edit-validator --details="To the cosmos !" --website="https://cosmos.network"
```

View the validator's information with this command:

```
gaiacli stake validator --address-validator=<your_address> --chain-id=<name_of_the_testnet_chain>
```

To check that the validator is active, look for it here:

```
gaiacli advanced tendermint validator-set
```

**Note:** To be in the validator set, you need to have more total voting power than the 100th validator.

## Delegating to a Validator

On the upcoming mainnet, you can delegate `atom` to a validator. These [delegators](https://cosmos.network/resources/delegators) can receive part of the validator's fee revenue. Read more about the [Cosmos Token Model](https://github.com/cosmos/cosmos/raw/master/Cosmos_Token_Model.pdf).

### Bond Tokens

On the testnet, we delegate `steak` instead of `atom`. Here's how you can bond tokens to a testnet validator:

```
gaiacli stake delegate --amount=10steak --address-delegator=<your_address> --address-validator=<bonded_validator_address> --name=<key_name> --chain-id=<name_of_testnet_chain>
```

While tokens are bonded, they are pooled with all the other bonded tokens in the network. Validators and delegators obtain a percentage of shares that equal their stake in this pool. 

### Unbond Tokens

If for any reason the validator misbehaves, or you want to unbond a certain amount of tokens, use this following command. You can unbond a specific amount of`shares`\(eg:`12.1`\) or all of them \(`MAX`\).

```
gaiacli stake unbond --address-delegator=<your_address> --address-validator=<bonded_validator_address> --shares=MAX --name=<key_name> --chain-id=<name_of_testnet_chain>
```

You can check your balance and your stake delegation to see that the unbonding went through successfully.

```
gaiacli account <your_address>
gaiacli stake delegation --address-delegator=<your_address> --address-validator=<bonded_validator_address> --chain-id=<name_of_testnet_chain>
```
