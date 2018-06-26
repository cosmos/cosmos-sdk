# Connect to the `gaia-6002` Testnet

Note: We are aware this documentation is a work in progress. We are actively
working to improve the tooling and the documentation to make this process as painless as
possible. In the meantime, join the [Validator Chat](https://riot.im/app/#/room/#cosmos_validators:matrix.org)
for technical support, and [open issues](https://github.com/cosmos/cosmos-sdk) if you run into any! Thanks very much for your patience and support. :)

## Setting Up a New Node

These instructions are for setting up a brand new full node from scratch. If you ran a full node on a previous testnet, please skip to [Upgrading From Previous Testnet](#upgrading-from-previous-testnet).

### Install Go

Install `go` by following the [official docs](https://golang.org/doc/install).
**Go 1.10+** is required for the Cosmos SDK. Remember to properly setup your `$GOPATH`, `$GOBIN`, and `$PATH` variables, for example:

```bash
mkdir -p $HOME/go/bin
echo "export GOPATH=$HOME/go" >> ~/.bash_profile
echo "export GOBIN=$GOPATH/bin" >> ~/.bash_profile
echo "export PATH=$PATH:$GOBIN" >> ~/.bash_profile
```

### Install Cosmos SDK

Next, let's install the testnet's version of the Cosmos SDK.

```bash
mkdir -p $GOPATH/src/github.com/cosmos
cd $GOPATH/src/github.com/cosmos
git clone https://github.com/cosmos/cosmos-sdk
cd cosmos-sdk && git checkout v0.19.0
make get_tools && make get_vendor_deps && make install
```

That will install the `gaiad` and `gaiacli` binaries. Verify that everything is OK:

```bash
$ gaiad version
0.19.0-c6711810

$ gaiacli version
0.19.0-c6711810
```

### Node Setup

Create the required configuration files, and initialize the node:

```bash
gaiad init --name <your_custom_name>
```

> *NOTE:* Note that only ASCII characters are supported for the `--name`. Using Unicode renders your node unreachable.

You can also edit this `name` in the `~/.gaiad/config/config.toml` file:

```toml
# A custom human readable name for this node
moniker = "<your_custom_name>"
```

Your full node has been initialized! Please skip to [Genesis & Seeds](#genesis--seeds).

## Upgrading From Previous Testnet

These instructions are for full nodes that have ran on previous testnets and would like to upgrade to the latest testnet.

### Reset Data

First, remove the outdated files and reset the data.

```bash
rm $HOME/.gaiad/config/addrbook.json $HOME/.gaiad/config/genesis.json
gaiad unsafe_reset_all
```

Your node is now in a pristine state while keeping the original `priv_validator.json` and `config.toml`. If you had any sentry nodes or full nodes setup before,
your node will still try to connect to them, but may fail if they haven't also
been upgraded.

**WARNING:** Make sure that every node has a unique `priv_validator.json`. Do not copy the `priv_validator.json` from an old node to multiple new nodes. Running two nodes with the same `priv_validator.json` will cause you to double sign.

### Software Upgrade

Now it is time to upgrade the software:

```bash
cd $GOPATH/src/github.com/cosmos/cosmos-sdk
git fetch --all && git checkout v0.19.0
make update_tools && make get_vendor_deps && make install
```

Your full node has been cleanly upgraded!

## Genesis & Seeds

### Copy the Genesis File

Copy the testnet's `genesis.json` file and place it in `gaiad`'s config directory.

```bash
mkdir -p $HOME/.gaiad/config
cp -a $GOPATH/src/github.com/cosmos/cosmos-sdk/cmd/gaia/testnets/gaia-6002/genesis.json $HOME/.gaiad/config/genesis.json
```

### Add Seed Nodes

Your node needs to know how to find peers. You'll need to add healthy seed nodes to `$HOME/.gaiad/config/config.toml`. Here are some seed nodes you can use:

```toml
# Comma separated list of seed nodes to connect to
seeds = "38aa9bec3998f12ae9088b21a2d910d19d565c27@gaia-6002.coinculture.net:46656,1e124dd15bd9955a7ea844ab003b1b47f0998b70@seed.cosmos.cryptium.ch:46656"
```

If those seeds aren't working, you can find more seeds and persistent peers on the [Cosmos Explorer](https://explorecosmos.network/nodes). Open the the `Full Nodes` pane and select nodes that do not have private (`10.x.x.x`) or [local IP addresses](https://en.wikipedia.org/wiki/Private_network). The `Persistent Peer` field contains the connection string. For best results use 4-6.

For more information on seeds and peers, [read this](https://github.com/tendermint/tendermint/blob/develop/docs/using-tendermint.md#peers).

## Run a Full Node

Start the full node with this command:

```bash
gaiad start
```

Check that everything is running smoothly:

```bash
gaiacli status
```

View the status of the network with the [Cosmos Explorer](https://explorecosmos.network). Once your full node syncs up to the current block height, you should see it appear on the [list of full nodes](https://explorecosmos.network/validators). If it doesn't show up, that's ok--the Explorer does not connect to every node.

## Generating Keys

### A Note on Keys in Cosmos:

There are three types of key representations that are used in this tutorial:

- `cosmosaccaddr`
  * Derived from account keys generated by `gaiacli keys add`
  * Used to receive funds
  * e.g. `cosmosaccaddr15h6vd5f0wqps26zjlwrc6chah08ryu4hzzdwhc`

- `cosmosaccpub`
  * Derived from account keys generated by `gaiacli keys add`
  * e.g. `cosmosaccpub1zcjduc3q7fu03jnlu2xpl75s2nkt7krm6grh4cc5aqth73v0zwmea25wj2hsqhlqzm`

- `cosmosvalpub`
  * Generated when the node is created with `gaiad init`.
  * Get this value with `gaiad tendermint show_validator`
  * e.g. `cosmosvalpub1zcjduc3qcyj09qc03elte23zwshdx92jm6ce88fgc90rtqhjx8v0608qh5ssp0w94c`

### Key Generation

You'll need an account private and public key pair \(a.k.a. `sk, pk` respectively\) to be able to receive funds, send txs, bond tx, etc.

To generate a new key \(default _ed25519_ elliptic curve\):

```bash
gaiacli keys add <account_name>
```

Next, you will have to create a passphrase to protect the key on disk. The output of the above command will contain a _seed phrase_. Save the _seed phrase_ in a safe place in case you forget the password!

If you check your private keys, you'll now see `<account_name>`:

```bash
gaiacli keys show <account_name>
```

You can see all your available keys by typing:

```bash
gaiacli keys list
```

View the validator pubkey for your node by typing:

```bash
gaiad tendermint show_validator
```

**WARNING:** We strongly recommend NOT using the same passphrase for multiple keys. The Tendermint team and the Interchain Foundation will not be responsible for the loss of funds.

## Fund your account

The best way to get tokens is from the [Cosmos Testnet Faucet](https://faucetcosmos.network). If the faucet is not working for you, try asking [#cosmos-validators](https://riot.im/app/#/room/#cosmos-validators:matrix.org). The faucet needs the `cosmosaccaddr` from the account you wish to use for staking.

After receiving tokens to your address, you can view your account's balance by typing:

```bash
gaiacli account <account_cosmosaccaddr>
```

> _*Note:*_ When you query an account balance with zero tokens, you will get this error: `No account with address <account_cosmosaccaddr> was found in the state.` This can also happen if you fund the account before your node has fully synced with the chain. These are both normal. Also, we're working on improving our error messages!

## Run a Validator Node

[Validators](https://cosmos.network/validators) are responsible for committing new blocks to the blockchain through voting. A validator's stake is slashed if they become unavailable, double sign a transaction, or don't cast their votes. If you only want to run a full node, a VM in the cloud is fine. However, if you are want to become a validator for the Hub's `mainnet`, you should research hardened setups. Please read [Sentry Node Architecture](https://github.com/cosmos/cosmos/blob/master/VALIDATORS_FAQ.md#how-can-validators-protect-themselves-from-denial-of-service-attacks) to protect your node from DDOS and ensure high-availability. Also see the [technical requirements](https://github.com/cosmos/cosmos/blob/master/VALIDATORS_FAQ.md#technical-requirements)). There's also more info on our [website](https://cosmos.network/validators).

### Create Your Validator

Your `cosmosvalpub` can be used to create a new validator by staking tokens. You can find your validator pubkey by running:

```bash
gaiad tendermint show_validator
```

Next, craft your `gaiacli stake create-validator` command:

> _*NOTE:*_  Don't use more `steak` thank you have! You can always get more by using the [Faucet](https://faucetcosmos.network/)!

```bash
gaiacli stake create-validator \
  --amount=5steak \
  --pubkey=$(gaiad tendermint show_validator) \
  --address-validator=<account_cosmosaccaddr>
  --moniker="choose a moniker" \
  --chain-id=gaia-6002 \
  --name=<key_name>
```

### Edit Validator Description

You can edit your validator's public description. This info is to identify your validator, and will be relied on by delegators to decide which validators to stake to. Make sure to provide input for every flag below, otherwise the field will default to empty (`--moniker` defaults to the machine name).

The `--keybase-sig` is a 16-digit string that is generated with a [keybase.io](https://keybase.io) account. It's a cryptographically secure method of verifying your identity across multiple online networks. The Keybase API allows us to retrieve your Keybase avatar. This is how you can add a logo to your validator profile.

```bash
gaiacli stake edit-validator
  --address-validator=<account_cosmosaccaddr>
  --moniker="choose a moniker" \
  --website="https://cosmos.network" \
  --keybase-sig="6A0D65E29A4CBC8E"
  --details="To infinity and beyond!"
  --chain-id=gaia-6002 \
  --name=<key_name>
```

### View Validator Description
View the validator's information with this command:

```bash
gaiacli stake validator \
  --address-validator=<account_cosmosaccaddr> \
  --chain-id=gaia-6002
```

Your validator is active if the following command returns anything:

```bash
gaiacli advanced tendermint validator-set | grep "$(gaiad tendermint show_validator)"
```

You should also be able to see your validator on the [Explorer](https://explorecosmos.network/validators). You are looking for the `bech32` encoded `address` in the `~/.gaiad/config/priv_validator.json` file.

> _*Note:*_ To be in the validator set, you need to have more total voting power than the 100th validator. This is not normally an issue.

### Problem #1: My validator has `voting_power: 0`

Your validator has become auto-unbonded. In `gaia-6002`, we unbond validators if they do not vote on `50` of the last `100` blocks. Since blocks are proposed every ~2 seconds, a validator unresponsive for ~100 seconds will become unbonded. This usually happens when your `gaiad` process crashes.

Here's how you can return the voting power back to your validator. First, if `gaiad` is not running, start it up again:

```bash
gaiad start
```

Wait for your full node to catch up to the latest block. Next, run the following command. Note that `<cosmosaccaddr>` is the address of your validator account, and `<name>` is the name of the validator account. You can find this info by running `gaiacli keys list`.

```bash
gaiacli stake unrevoke <cosmosaccaddr> --chain-id=gaia-6002 --name=<name>
```

**WARNING:** If you don't wait for `gaiad` to sync before running `unrevoke`, you will receive an error message telling you your validator is still jailed.

Lastly, check your validator again to see if your voting power is back.

```bash
gaiacli status
```

You may notice that your voting power is less than it used to be. That's because you got slashed for downtime!

### Problem #2: My `gaiad` crashes because of `too many open files`

The default number of files Linux can open (per-process) is `1024`. `gaiad` is known to open more than `1024` files. This causes the process to crash. A quick fix is to run `ulimit -n 4096` (increase the number of open files allowed) and then restart the process with `gaiad start`. If you are using `systemd` or another process manager to launch `gaiad` this may require some configuration at that level. A sample `systemd` file to fix this issue is below:

```toml
# /etc/systemd/system/gaiad.service
[Unit]
Description=Cosmos Gaia Node
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu
ExecStart=/home/ubuntu/go/bin/gaiad start
Restart=on-failure
RestartSec=3
LimitNOFILE=4096

[Install]
WantedBy=multi-user.target
```

## Delegating to a Validator

On the upcoming mainnet, you can delegate `atom` to a validator. These [delegators](https://cosmos.network/resources/delegators) can receive part of the validator's fee revenue. Read more about the [Cosmos Token Model](https://github.com/cosmos/cosmos/raw/master/Cosmos_Token_Model.pdf).

### Bond Tokens

On the testnet, we delegate `steak` instead of `atom`. Here's how you can bond tokens to a testnet validator:

```bash
gaiacli stake delegate \
  --amount=10steak \
  --address-delegator=<account_cosmosaccaddr> \
  --address-validator=$(gaiad tendermint show_validator) \
  --name=<key_name> \
  --chain-id=gaia-6002
```

While tokens are bonded, they are pooled with all the other bonded tokens in the network. Validators and delegators obtain a percentage of shares that equal their stake in this pool.

> _*NOTE:*_  Don't use more `steak` thank you have! You can always get more by using the [Faucet](https://faucetcosmos.network/)!

### Unbond Tokens

If for any reason the validator misbehaves, or you want to unbond a certain amount of tokens, use this following command. You can unbond a specific amount of`shares`\(eg:`12.1`\) or all of them \(`MAX`\).

```bash
gaiacli stake unbond \
  --address-delegator=<account_cosmosaccaddr> \
  --address-validator=$(gaiad tendermint show_validator) \
  --shares=MAX \
  --name=<key_name> \
  --chain-id=gaia-6002
```

You can check your balance and your stake delegation to see that the unbonding went through successfully.

```bash
gaiacli account <account_cosmosaccaddr>

gaiacli stake delegation \
  --address-delegator=<account_cosmosaccaddr> \
  --address-validator=$(gaiad tendermint show_validator) \
  --chain-id=gaia-6002
```

## Other Operations

### Send Tokens

```bash
gaiacli send \
  --amount=10faucetToken \
  --chain-id=gaia-6002 \
  --name=<key_name> \
  --to=<destination_cosmosaccaddr>
```

> _*Note:*_ The `--amount` flag accepts the format `--amount=<value|coin_name>`.

Now, view the updated balances of the origin and destination accounts:

```bash
gaiacli account <account_cosmosaccaddr>
gaiacli account <destination_cosmosaccaddr>
```

You can also check your balance at a given block by using the `--block` flag:

```bash
gaiacli account <account_cosmosaccaddr> --block=<block_height>
```
