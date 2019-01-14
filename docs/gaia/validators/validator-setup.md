# Run a Validator on the public testnet

::: tip
Information on how to join the current testnet (`genesis.json` file and seeds) is held [in our `testnet` repo](https://github.com/cosmos/testnets/tree/master/latest). Please check there if you are looking to join our latest testnet. 
:::

__Note__: This documentation is only intended for validators of the **public testnet**

Before setting up your validator node, make sure you've already gone through the [Full Node Setup](../join-testnet.md) guide.

## What is a Validator?

[Validators](./overview.md) are responsible for committing new blocks to the blockchain through voting. A validator's stake is slashed if they become unavailable or sign blocks at the same height. Please read about [Sentry Node Architecture](./validator-faq.md#how-can-validators-protect-themselves-from-denial-of-service-attacks) to protect your node from DDOS attacks and to ensure high-availability.

::: danger Warning
If you want to become a validator for the Hub's `mainnet`, you should [research security](./security.md).
:::

You may want to skip the next section if you have already [set up a full-node](../join-testnet.md).

## Create Your Validator

Your `cosmosvalconspub` can be used to create a new validator by staking tokens. You can find your validator pubkey by running:

```bash
gaiad tendermint show-validator
```

Next, craft your `gaiad gentx` command:

::: warning Note
Don't use more `STAKE` thank you have! You can always get more by using the [Faucet](https://faucet.cosmos.network/)!
:::

```bash
gaiacli tx staking create-validator \
  --amount=5STAKE \
  --pubkey=$(gaiad tendermint show-validator) \
  --moniker="choose a moniker" \
  --chain-id=<chain_id> \
  --from=<key_name> \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" 
```

__Note__: When specifying commission parameters, the `commission-max-change-rate`
is used to measure % _point_ change over the `commission-rate`. E.g. 1% to 2% is
a 100% rate increase, but only 1 percentage point.

__Note__: If unspecified, `consensus_pubkey` will default to the output of `gaiad tendermint show-validator`.
`key_name` is the name of the private key that will be used to sign the transaction.

## Participate in genesis as a validator

__Note__: This section only concerns validators that want to be in the genesis file. If the chain you want to validate is already live, skip this section.

__Note__: `Gaia-9002` and `Game of stakes` will not use this process. They will be bootsrapped using Tendermint seed validators. You will just need to use the [create-validator](#create-your-validator) command in order to join as a validator for these networks.

If you want to participate in genesis as a validator, you need to justify that you (or a delegator) have some stake at genesis, create one (or multiple) transaction to bond this stake to your validator address, and include this transaction in the genesis file. 

We thus need to distinguish two cases:

- Case 1: You want to bond the initial stake from your validator's address.
- Case 2: You want to bond the initial stake from a delegator's address.

### Case 1: The initial stake comes from your validator's address

In this case, you will create a `gentx`:

```bash
gaiad gentx \
  --amount <amount_of_delegation> \
  --commission-rate <commission_rate> \
  --commission-max-rate <commission_max_rate> \
  --commission-max-change-rate <commission_max_change_rate> \
  --pubkey <consensus_pubkey> \
  --name <key_name>
```

__Note__: This command automatically store your `gentx` in `~/.gaiad/config/gentx` for it to be processed at genesis.

::: tip
Consult `gaiad gentx --help` for more information on the flags defaults.
:::

A `gentx` is a JSON file carrying a self-delegation. All genesis transactions are collected by a `genesis coordinator` and validated against an initial `genesis.json`. Such initial `genesis.json` contains only a list of accounts and their coins. Once the transactions are processed, they are merged in the `genesis.json`'s `gentxs` field.

### Case 2: The initial stake comes from a delegator's address

In this case, you need both the signature of the validator and the delegator. Start by creating an unsigned `create-validator` transaction, and save it in a file called `unsignedValTx`: 

```bash
gaiacli tx staking create-validator \
  --amount=5STAKE \
  --pubkey=$(gaiad tendermint show-validator) \
  --moniker="choose a moniker" \
  --chain-id=<chain_id> \
  --from=<key_name> \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --address-delegator="address of the delegator" \
  --generate-only \
  > unsignedValTx.json
```

Then, sign this `unsignedValTx` with your validator's private key, and save the output in a new file `signedValTx.json`:

```bash
gaiacli tx sign unsignedValTx.json --from=<validator_key_name> > signedValTx.json
```

Then, pass this file to the delegator, who needs to run the following command:

```bash
gaiacli tx sign signedValTx.json --from=<delegator_key_name> > gentx.json
```

This `gentx.json` needs to be included in the `~/.gaiad/config/gentx` folder on the validator's machine to be processed at genesis, just like in case 1 (except here it needs to be copied manually into the folder).

### Copy the Initial Genesis File and Process Genesis Transactions

Fetch the `genesis.json` file into `gaiad`'s config directory.

```bash
mkdir -p $HOME/.gaiad/config
curl https://raw.githubusercontent.com/cosmos/testnets/master/latest/genesis.json > $HOME/.gaiad/config/genesis.json
```

__Note:__ We use the `latest` directory in the [testnets repo](https://github.com/cosmos/testnets)
which contains details for the latest testnet. If you are connecting to a different testnet, ensure you get the right files.


You also need to fetch the genesis transactions of all the other genesis validators. For now there is no repository where genesis transactions can be submitted by validators, but this will as soon as we try out this feature in a testnet.

Once you've collected all genesis transactions in `~/.gaiad/config/gentx`, you can run:

```bash
gaiad collect-gentxs
```

__Note:__ The accounts from which you delegate in the `gentx` transactions need to possess stake tokens in the genesis file, otherwise `collect-gentx` will fail.

The previous command will collect all genesis transactions and finalise `genesis.json`. To verify the correctness of the configuration and start the node run:

```bash
gaiad start
```

## Edit Validator Description

You can edit your validator's public description. This info is to identify your validator, and will be relied on by delegators to decide which validators to stake to. Make sure to provide input for every flag below, otherwise the field will default to empty (`--moniker` defaults to the machine name).

The `--identity` can be used as to verify identity with systems like Keybase or UPort. When using with Keybase `--identity` should be populated with a 16-digit string that is generated with a [keybase.io](https://keybase.io) account. It's a cryptographically secure method of verifying your identity across multiple online networks. The Keybase API allows us to retrieve your Keybase avatar. This is how you can add a logo to your validator profile.

```bash
gaiacli tx staking edit-validator
  --moniker="choose a moniker" \
  --website="https://cosmos.network" \
  --identity=6A0D65E29A4CBC8E \
  --details="To infinity and beyond!" \
  --chain-id=<chain_id> \
  --from=<key_name> \
  --commission-rate="0.10"
```

__Note__: The `commission-rate` value must adhere to the following invariants:

- Must be between 0 and the validator's `commission-max-rate`
- Must not exceed the validator's `commission-max-change-rate` which is maximum
  % point change rate **per day**. In other words, a validator can only change
  its commission once per day and within `commission-max-change-rate` bounds.

## View Validator Description

View the validator's information with this command:

```bash
gaiacli query staking validator <account_cosmos>
```

## Track Validator Signing Information

In order to keep track of a validator's signatures in the past you can do so by using the `signing-info` command:

```bash
gaiacli query slashing signing-info <validator-pubkey>\
  --chain-id=<chain_id>
```

## Unjail Validator

When a validator is "jailed" for downtime, you must submit an `Unjail` transaction from the operator account in order to be able to get block proposer rewards again (depends on the zone fee distribution).

```bash
gaiacli tx slashing unjail \
	--from=<key_name> \
	--chain-id=<chain_id>
```

## Confirm Your Validator is Running

Your validator is active if the following command returns anything:

```bash
gaiacli query tendermint-validator-set | grep "$(gaiad tendermint show-validator)"
```

You should also be able to see your validator on the [Explorer](https://explorecosmos.network/validators). You are looking for the `bech32` encoded `address` in the `~/.gaiad/config/priv_validator.json` file.

::: warning Note
To be in the validator set, you need to have more total voting power than the 100th validator.
:::

## Common Problems

### Problem #1: My validator has `voting_power: 0`

Your validator has become auto-unbonded. In `gaia-8000`, we unbond validators if they do not vote on `50` of the last `100` blocks. Since blocks are proposed every ~2 seconds, a validator unresponsive for ~100 seconds will become unbonded. This usually happens when your `gaiad` process crashes.

Here's how you can return the voting power back to your validator. First, if `gaiad` is not running, start it up again:

```bash
gaiad start
```

Wait for your full node to catch up to the latest block. Next, run the following command. Note that `<cosmos>` is the address of your validator account, and `<name>` is the name of the validator account. You can find this info by running `gaiacli keys list`.

```bash
gaiacli tx slashing unjail <cosmos> --chain-id=<chain_id> --from=<from>
```

::: danger Warning
If you don't wait for `gaiad` to sync before running `unjail`, you will receive an error message telling you your validator is still jailed.
:::

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
