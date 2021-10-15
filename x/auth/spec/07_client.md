<!--
order: 7
-->

# Client

## CLI

A user can query and interact with the `vesting` module using the CLI.

### Transactions

The `tx` commands allow users to interact with the `vesting` module.

```bash
simd tx vesting --help
```

#### create-periodic-vesting-account

The `create-periodic-vesting-account` command creates a new vesting account funded with an allocation of tokens, where a sequence of coins and period length in seconds. Periods are sequential, in that the duration of of a period only starts at the end of the previous period. The duration of the first period starts upon account creation.

```bash
simd tx vesting create-periodic-vesting-account [to_address] [periods_json_file] [flags]
```

Example:

```bash
simd tx vesting create-periodic-vesting-account cosmos1.. periods.json
```

#### create-vesting-account

The `create-vesting-account` command creates a new vesting account funded with an allocation of tokens. The account can either be a delayed or continuous vesting account, which is determined by the '--delayed' flag. All vesting accouts created will have their start time set by the committed block's time. The end_time must be provided as a UNIX epoch timestamp.

```bash
simd tx vesting create-vesting-account [to_address] [amount] [end_time] [flags]
```

Example:

```bash
simd tx vesting create-vesting-account cosmos1.. 100stake 2592000
```