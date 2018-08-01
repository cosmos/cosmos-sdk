# Validator Setup

::: warning Current Testnet
The current testnet is `gaia-7005`.
:::

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
  --chain-id=gaia-7005 \
  --name=<key_name>
```

### Edit Validator Description

You can edit your validator's public description. This info is to identify your validator, and will be relied on by delegators to decide which validators to stake to. Make sure to provide input for every flag below, otherwise the field will default to empty (`--moniker` defaults to the machine name).

The `--identity` can be used as to verify identity with systems like Keybase or UPort. When using with Keybase `--identity` should be populated with a 16-digit string that is generated with a [keybase.io](https://keybase.io) account. It's a cryptographically secure method of verifying your identity across multiple online networks. The Keybase API allows us to retrieve your Keybase avatar. This is how you can add a logo to your validator profile.

```bash
gaiacli stake edit-validator
  --address-validator=<account_cosmosaccaddr>
  --moniker="choose a moniker" \
  --website="https://cosmos.network" \
  --identity=6A0D65E29A4CBC8E
  --details="To infinity and beyond!"
  --chain-id=gaia-7005 \
  --name=<key_name>
```

### View Validator Description

View the validator's information with this command:

```bash
gaiacli stake validator \
  --address-validator=<account_cosmosaccaddr> \
  --chain-id=gaia-7005
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

Your validator has become auto-unbonded. In `gaia-7005`, we unbond validators if they do not vote on `50` of the last `100` blocks. Since blocks are proposed every ~2 seconds, a validator unresponsive for ~100 seconds will become unbonded. This usually happens when your `gaiad` process crashes.

Here's how you can return the voting power back to your validator. First, if `gaiad` is not running, start it up again:

```bash
gaiad start
```

Wait for your full node to catch up to the latest block. Next, run the following command. Note that `<cosmosaccaddr>` is the address of your validator account, and `<name>` is the name of the validator account. You can find this info by running `gaiacli keys list`.

```bash
gaiacli stake unrevoke <cosmosaccaddr> --chain-id=gaia-7005 --name=<name>
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

## Delegating to a Validator

On the upcoming mainnet, you can delegate `Atom` to a validator. These [delegators](https://cosmos.network/resources/delegators) can receive part of the validator's fee revenue. Read more about the [Cosmos Token Model](https://github.com/cosmos/cosmos/raw/master/Cosmos_Token_Model.pdf).

### Bond Tokens

On the testnet, we delegate `steak` instead of `Atom`. Here's how you can bond tokens to a testnet validator:

```bash
gaiacli stake delegate \
  --amount=10steak \
  --address-delegator=<account_cosmosaccaddr> \
  --address-validator=<validator_cosmosaccaddr> \
  --from=<key_name> \
  --chain-id=gaia-7001
```

While tokens are bonded, they are pooled with all the other bonded tokens in the network. Validators and delegators obtain a percentage of shares that equal their stake in this pool.

> _*NOTE:*_  Don't use more `steak` thank you have! You can always get more by using the [Faucet](https://gaia.faucetcosmos.network/)!

### Unbond Tokens

If for any reason the validator misbehaves, or you want to unbond a certain amount of tokens, use this following command. You can unbond a specific amount of`shares`\(eg:`12.1`\) or all of them \(`MAX`\).

```bash
gaiacli stake unbond \
  --address-delegator=<account_cosmosaccaddr> \
  --address-validator=<validator_cosmosaccaddr> \
  --shares=MAX \
  --from=<key_name> \
  --chain-id=gaia-7001
```

You can check your balance and your stake delegation to see that the unbonding went through successfully.

```bash
gaiacli account <account_cosmosaccaddr>

gaiacli stake delegation \
  --address-delegator=<account_cosmosaccaddr> \
  --address-validator=<validator_cosmosaccaddr> \
  --chain-id=gaia-7001
```

## Governance

Governance is the process from which users in the Cosmos Hub can come to consensus on software upgrades, parameters of the mainnet or on custom text proposals. This is done through voting on proposals, which will be submitted by `Atom` holders on the mainnet.

Some considerations about the voting process:

- Voting is done by bonded `Atom` holders on a 1 bonded `Atom` 1 vote basis
- Delegators inherit the vote of their validator if they don't vote
- **Validators MUST vote on every proposal**. If a validator does not vote on a proposal, they will be **partially slashed**
- Votes are tallied at the end of the voting period (2 weeks on mainnet). Each address can vote multiple times to update its `Option` value (paying the transaction fee each time), only the last casted vote will count as valid
- Voters can choose between options `Yes`, `No`, `NoWithVeto` and `Abstain`
At the end of the voting period, a proposal is accepted if `(YesVotes/(YesVotes+NoVotes+NoWithVetoVotes))>1/2` and `(NoWithVetoVotes/(YesVotes+NoVotes+NoWithVetoVotes))<1/3`. It is rejected otherwise

For more information about the governance process and how it works, please check out the Governance module [specification](https://github.com/cosmos/cosmos-sdk/tree/develop/docs/spec/governance).

### Create a Governance proposal

In order to create a governance proposal, you must submit an initial deposit along with the proposal details:

- `title`: Title of the proposal
- `description`: Description of the proposal
- `type`: Type of proposal. Must be of value _Text_ (types _SoftwareUpgrade_ and _ParameterChange_ not supported yet).

```bash
gaiacli gov submit-proposal \
  --title=<title> \
  --description=<description> \
  --type=<Text/ParameterChange/SoftwareUpgrade> \
  --proposer=<account_cosmosaccaddr> \
  --deposit=<40steak> \
  --from=<name> \
  --chain-id=gaia-7001
```


### Increase deposit

In order for a proposal to be broadcasted to the network, the amount deposited must be above a `minDeposit` value (default: `10 steak`). If the proposal you previously created didn't meet this requirement, you can still increase the total amount deposited to activate it. Once the minimum deposit is reached, the proposal enters voting period:

```bash
gaiacli gov deposit \
  --proposalID=<proposal_id> \
  --depositer=<account_cosmosaccaddr> \
  --deposit=<200steak> \
  --from=<name> \
  --chain-id=gaia-7001
```

> _NOTE_: Proposals that don't meet this requirement will be deleted after `MaxDepositPeriod` is reached.

#### Query proposal

Once created, you can now query information of the proposal:

```bash
gaiacli gov query-proposal \
  --proposalID=<proposal_id> \
  --chain-id=gaia-7001
```

### Vote on a proposal

After a proposal's deposit reaches the `MinDeposit` value, the voting period opens. Bonded `Atom` holders can then cast vote on it:

```bash
gaiacli gov vote \
  --proposalID=<proposal_id> \
  --voter=<account_cosmosaccaddr> \
  --option=<Yes/No/NoWithVeto/Abstain> \
  --from=<name> \
  --chain-id=gaia-7001
```

#### Query vote

Check the vote with the option you just submitted:

```bash
gaiacli gov query-vote \
  --proposalID=<proposal_id> \
  --voter=<account_cosmosaccaddr> \
  --chain-id=gaia-7001
```
