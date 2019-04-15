# Genesis File

This document explains how the genesis file of the Cosmmos Hub mainnet is structured. It also explains how you can build a genesis file for your own `gaia` testnet.

Note that you can generate a default genesis file for your own testnet by running the following command:

```bash
gaiad init <moniker> --chain-id <chain-id>
```

The genesis file is stored in `~/.gaiad/config/genesis.toml`.

## What is a Genesis File

A genesis file is a JSON file which defines the initial state of your blockchain. It can be seen as height `0` of your blockchain. The first block, at height `1`, will reference the genesis file as its parent. 

The state defined in the genesis file contains all the necessary information, like initial token allocation, genesis time, default parameters, and more. Let us break down these information.

## Genesis Time and Chain_id 

The `genesis_time` is defined at the top of the genesis file. It is a `UTC` timestamps which specifies when the blockchain is due to start. At this time, genesis validators are supposed to come online and start participating in the consensus process. The blockchain starts when more than 2/3rd of the genesis validators (weighted by voting power) are online. 

```json
"genesis_time": "2019-03-13T17:00:00.000000000Z",
```

The `chain_id` is a unique identifier for your chain. It helps differentiate between different chains using the same version of the software.

```json
"chain_id": "cosmoshub-1",
```

## Consensus Parameters

Next, the genesis file defines consensus parameters. Consensus parameters regroup all the parameters that are related to the consensus layer, which is `Tendermint` in the case of `gaia`. Let us look at these parameters:

- `block`
    + `max_bytes`: Maximum number of bytes per block. 
    + `max_gas`: Gas limit per block. Each transaction included in the block will consume some gas. The total gas used by transactions included in a block cannot exceed this limit.
- `evidence`
    + `max_age`: An evidence is a proof that a validator signed two different blocks at the same height (and round). This is an explicitely malicious behaviour that is punished at the state-machine level. The `max_age` defines the maximum number of **blocks** after which an evidence is not valid anymore. 
- `validator`
    + `pub_key_types`: The types of pubkey (`ed25519`, `secp256k1`, ...) that are accepted for validators. Currently only `ed25519` is accepted.

```json
"consensus_params": {
    "block_size": {
      "max_bytes": "150000",
      "max_gas": "1500000"
    },
    "evidence": {
      "max_age": "1000000"
    },
    "validator": {
      "pub_key_types": [
        "ed25519"
      ]
    }
  },
```

## Application State

The application state defines the initial state of the state-machine. 

### Genesis Accounts

In this section, initial allocation of tokens is defined. It is possible to add accounts manually by directly editing the genesis file, but it is also possible to use the following command:

```bash
// Example: gaiad add-genesis-account cosmos1qs8tnw2t8l6amtzvdemnnsq9dzk0ag0z37gh3h 10000000uatom

gaiad add-genesis-account <account-address> <amount><denom>
```

This command creates an item in the `accounts` list, under the `app_state` section.

```json
"accounts": [
      {
        "address": "cosmos1qs8tnw2t8l6amtzvdemnnsq9dzk0ag0z37gh3h",
        "coins": [
          {
            "denom": "uatom",
            "amount": "10000000"
          }
        ],
        "sequence_number": "0",
        "account_number": "0",
        "original_vesting": [
          {
            "denom": "uatom",
            "amount": "26306000000"
          }
        ],
        "delegated_free": null,
        "delegated_vesting": null,
        "start_time": "0",
        "end_time": "10000"
      }
]
```

Let us break down the parameters:

- `sequence_number`: This number is used to count the number of transactions sent by this account. It is incremented each time a transaction is included in a block, and used to prevent replay attacks. Initial value is `0`.
- `account_number`: Unique identifier for the account. It is generated the first time a transaction including this account is included in a block.
- `original_vesting`: Vesting is natively supported by `gaia`. You can define an amount of token owned by the account that needs to be vested for a period of time before they can be transferred. Vested tokens can be delegated. Default value is `null`.
- `delegated_free`: Amount of delegated tokens that can be transferred after they've been vested. Most of the time, will be `null` in genesis. 
- `delegated_vesting`: Amount of delegated tokens that are still vesting. Most of the time, will be `null` in genesis.
- `start_time`: Block at which the vesting period starts. `0` most of the time in genesis.
- `end_time`: Block at which the vesting period ends. `0` if no vesting for this account.

### Bank 

The `bank` module handles tokens. The only parameter that needs to be defined in this section is wether `transfers` are enabled at genesis or not.

```json
"bank": {
      "send_enabled": false
    }
```

### Staking

The `staking` module handles the bulk of the Proof-of-Stake logic of the state-machine. This section should look like the following:

```json
"staking": {
      "pool": {
        "not_bonded_tokens": "10000000",
        "bonded_tokens": "0"
      },
      "params": {
        "unbonding_time": "1814400000000000",
        "max_validators": 100,
        "max_entries": 7,
        "bond_denom": "uatom"
      },
      "last_total_power": "0",
      "last_validator_powers": null,
      "validators": null,
      "bonds": null,
      "unbonding_delegations": null,
      "redelegations": null,
      "exported": false
    }
```

Let us break down the parameters:

- `pool`
    + `not_bonded_tokens`: Defines the amount of tokens not bonded (i.e. delegated) in genesis. Generally, it equals the total supply of the staking token (`uatom` in this example).
    + `bonded_tokens`: Amount of bonded tokens in genesis. Generally `0`.
- `params`
    + `unbonding_time`: Time in **nanosecond** it takes for tokens to complete unbonding. 
    + `max_validators`: Maximum number of active validators. 
    + `max_entries`: Maximum unbonding delegations and redelegations between a particular pair of delegator / validator.
    + `bond_denom`: Denomination of the staking token. 
- `last_total_power`: Total amount of voting power. Generally `0` in genesis (except if genesis was generated using a previous state).
- `last_validator_powers`: Power of each validator in last known state. Generally `null` in genesis (except if genesis was generated using a previous state).
- `validators`: List of last knoww validators. Generally `null` in genesis (except if genesis was generated using a previous state).
- `bonds`: List of last known delegation. Generally `null` in genesis (except if genesis was generated using a previous state).
- `unbonding_delegations`: List of last known unbonding delegations. Generally `null` in genesis (except if genesis was generated using a previous state).
- `redelegations`: List of last known redelegations. Generally `null` in genesis (except if genesis was generated using a previous state).
- `exported`: Wether this genesis was generated using the export of a previous state.

### Mint

The `mint` module governs the logic of inflating the supply of token. The `mint` section in the genesis file looks like the follwing:

```json
"mint": {
      "minter": {
        "inflation": "0.070000000000000000",
        "annual_provisions": "0.000000000000000000"
      },
      "params": {
        "mint_denom": "uatom",
        "inflation_rate_change": "0.130000000000000000",
        "inflation_max": "0.200000000000000000",
        "inflation_min": "0.070000000000000000",
        "goal_bonded": "0.670000000000000000",
        "blocks_per_year": "6311520"
      }
    }
```

Let us break down the parameters:

- `minter`
    + `inflation`: Initial yearly percentage of increase in the total supply of staking token, compounded weekly. A `0.070000000000000000` value means the target is `7%` yearly inflation, compounded weekly.
    + `annual_provisions`: Calculated each block. Initialize at `0.000000000000000000`.
- `params`
    + `mint_denom`: Denom of the staking token that is inflated.
    + `inflation_rate_change`: Max yearly change in inflation. 
    + `inflation_max`: Maximum level of inflation.
    + `inflation_min`: Minimum level of inflation.
    + `goal_bonded`: Percentage of the total supply that is targeted to be bonded. If the percentage of bonded staking tokens is below this target, the inflation increases (following `inflation_rate_change`) until it reaches `inflation_max`. If the percentage of bonded staking tokens is above this target, the inflation decreases (following `inflation_rate_change`) until it reaches `inflation_min`.
    + `blocks_per_year`: Estimation of the amount of blocks per year. Used to compute the block reward coming from inflated staking token (called block provisions). 

### Distribution

The `distr` module handles the logic of distribution block provisions and fees to validators and delegators. The `distr` section in the genesis file looks like the follwing:

```json
    "distr": {
      "fee_pool": {
        "community_pool": null
      },
      "community_tax": "0.020000000000000000",
      "base_proposer_reward": "0.010000000000000000",
      "bonus_proposer_reward": "0.040000000000000000",
      "withdraw_addr_enabled": false,
      "delegator_withdraw_infos": null,
      "previous_proposer": "",
      "outstanding_rewards": null,
      "validator_accumulated_commissions": null,
      "validator_historical_rewards": null,
      "validator_current_rewards": null,
      "delegator_starting_infos": null,
      "validator_slash_events": null
    }
```

Let us break down the parameters:

- `fee_pool`
    + `community_pool`: The community pool is a pool of tokens that can be used to pay for bounties. It is allocated via governance proposals. Generally `null` in genesis.
- `community_tax`: The tax percentage on fees and block rewards that goes to the community pool.
- `base_proposer_reward`: Base bonus on transaction fees collected in a valid block that goes to the proposer of block. If value is `0.010000000000000000`, 1% of the fees go to the proposer. 
- `bonus_proposer_reward`: Max bonus on transaction fees collected in a valid block that goes to the proposer of block. The bonus depends on the number of `precommits` the proposer includes. If the proposer includes 2/3rd `precommits` weighted by voting power (minimum for the block to be valid), they get a bonus of `base_proposer_reward`. This bonus increases linearly up to `bonus_proposer_reward` if the proposer includes 100% of `precommits`.
- `withdraw_addr_enabled`: If `true`, delegators can set a different address to withdraw their rewards. Set to `false` if you want to disable transfers at genesis, as it can be used as a way to get around the restriction.
- `delegator_withdraw_infos`: List of delegators withdraw address. Generally `null` if genesis was not exported from previous state.
- `previous_proposer`: Proposer of the previous block. Set to `""` if genesis was not exported from previous state.
- `outstanding_rewards`: Outstanding (un-withdrawn) rewards. Set to `null` if genesis was not exported from previous state.
- `validator_accumulated_commission`: Outstanding (un-withdrawn) commission of validators. Set to `null` if genesis was not exported from previous state.
- `validator_historical_rewards`: Set of information related to the historical rewards of validators and used by the `distr` module for various computation. Set to `null` if genesis was not exported from previous state. 
- `validators_current_rewards`: Set of information related to the current rewards of validators and used by the `distr` module for various computation. Set to `null` if genesis was not exported from previous state.
- `delegator_starting_infos`: Tracks the previous validator period, the delegation's amount of staking token, and the creation height (to check later on if any slashes have occurred). Set to `null` if genesis was not exported from previous state.
- `validator_slash_events`: Set of information related to the past slashing of validators. Set to `null` if genesis was not exported from previous state.

### Governance

The `gov` module handles all governance-related transactions. The initial state of the `gov` section looks like the following:

```json
"gov": {
      "starting_proposal_id": "1",
      "deposits": null,
      "votes": null,
      "proposals": null,
      "deposit_params": {
        "min_deposit": [
          {
            "denom": "uatom",
            "amount": "512000000"
          }
        ],
        "max_deposit_period": "1209600000000000"
      },
      "voting_params": {
        "voting_period": "1209600000000000"
      },
      "tally_params": {
        "quorum": "0.4",
        "threshold": "0.5",
        "veto": "0.334",
        "governance_penalty": "0.0"
      }
    }
```

Let us break down the parameters:

- `starting_proposal_id`: This parameter defines the ID of the first proposal. Each proposal is identified by a unique ID.
- `deposits`: List of deposits for each proposal ID. Set to `null` if genesis was not exported from previous state.
- `votes`: List of votes for each proposal ID. Set to `null` if genesis was not exported from previous state.
- `proposals`: List of proposals for each proposal ID: Set to `null` if genesis was not exported from previous state.
- `deposit_params`
    + `min_deposit`: The minimum deposit required for the proposal to enter `Voting Period`. If multiple denoms are provided, the `OR`  operator applies.
    + `max_deposit_period`: The maximum period (in **nanoseconds**) after which it is not possible to deposit on the proposal anymore.
- `voting_params`
    + `voting_period`: Length of the voting period in **nanoseconds**. 
- `tally_params`
    + `quorum`: Minimum percentage of bonded staking tokens that needs to vote for the result to be valid.
    + `threshold`: Minimum percentage of votes that need to be `YES` for the result to be valid.
    + `veto`: Maximum percentage `NO_WITH_VETO` votes for the result to be valid.
    + `governance_penalty`: Penalty for validators that do not vote on a given proposal.

### Slashing 

The `slashing` module handles the logic to slash delegators if their validator misbehave. The `slashing` section in genesis looks as follows:

```json
"slashing": {
      "params": {
        "max_evidence_age": "1814400000000000",
        "signed_blocks_window": "10000",
        "min_signed_per_window": "0.050000000000000000",
        "downtime_jail_duration": "600000000000",
        "slash_fraction_double_sign": "0.050000000000000000",
        "slash_fraction_downtime": "0.000100000000000000"
      },
      "signing_infos": {},
      "missed_blocks": {}
    }
```

Let us break down the parameters:

- `params`
    + `max_evidence_age`: Maximum age of the evidence in **nanoseconds**.
    + `signed_blocks_window`: Moving window of blocks to figure out offline validators.
    + `min_signed_per_window`: Minimum percentage of `precommits`that must be present in the `block window` for the validator to be considered online.
    + `downtime_jail_duration`: Duration in **nanoseconds** for which a validator is jailed after they get slashed for downtime.
    + `slash_fraction_double_sign`: Percentage of delegators bonded stake slashed when their validator double signs. 
    + `slash_fraction_downtime`: Percentage of delegators bonded stake slashed when their validator is down.
- `signing_infos`: Various infos per validator needed by the `slashing` module. Set to `{}` if genesis was not exported from previous state.
- `missed_blocks`: Various infos related to missed blocks needed by the `slashing` module. Set to `{}` if genesis was not exported from previous state.

### Genesis Transactions

By default, the genesis file do not contain any `gentxs`. A `gentx` is a transaction that bonds staking token present in the genesis file under `accounts` to a validator, essentially creating a validator at genesis. The chain will start as soon as more than 2/3rds of the validators (weighted by voting power) that are the recipient of a valid `gentx` come online after `genesis_time`.

A `gentx` can be added manually to the genesis file, or via the following command:

```bash
gaiad collect-gentxs
```

This command will add all the `gentxs` stored in `~/.gaiad/config/gentx` to the genesis file. In order to create a genesis transaction, click [here](./validators/validator-setup.md#participate-in-genesis-as-a-validator).
