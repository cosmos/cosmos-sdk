# Validator Setup

Before setting up your validator node, make sure you've already gone through the [Full Node Setup](/getting-started/full-node.md) guide.

## Running a Validator Node

[Validators](/validators/overview.md) are responsible for committing new blocks to the blockchain through voting. A validator's stake is slashed if they become unavailable, double sign a transaction, or don't cast their votes. Please read about [Sentry Node Architecture](/validators/validator-faq.md#how-can-validators-protect-themselves-from-denial-of-service-attacks) to protect your node from DDOS attacks and to ensure high-availability.

::: danger Warning
If you want to become a validator for the Hub's `mainnet`, you should [research security](/validators/security.md).
:::

### Create Your Validator

Your `cosmosvalpub` can be used to create a new validator by staking tokens. You can find your validator pubkey by running:

```bash
gaiad tendermint show_validator
```

Next, craft your `gaiacli stake create-validator` command:

::: warning Note
Don't use more `steak` thank you have! You can always get more by using the [Faucet](https://faucetcosmos.network/)!
:::

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

### Confirm Your Validator is Running

Your validator is active if the following command returns anything:

```bash
gaiacli advanced tendermint validator-set | grep "$(gaiad tendermint show_validator)"
```

You should also be able to see your validator on the [Explorer](https://explorecosmos.network/validators). You are looking for the `bech32` encoded `address` in the `~/.gaiad/config/priv_validator.json` file.


::: warning Note
To be in the validator set, you need to have more total voting power than the 100th validator.
:::

## Common Problems

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

::: danger Warning
If you don't wait for `gaiad` to sync before running `unrevoke`, you will receive an error message telling you your validator is still jailed.
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
