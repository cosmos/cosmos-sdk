<!--
order: 8
-->

# Client

## CLI

A user can query and interact with the `distribution` module using the CLI.

### Query

The `query` commands allow users to query `distribution` state.

```
simd query distribution --help
```

#### commission

The `commission` command allows users to query validator commission rewards by address.

```
simd query distribution commission [address] [flags]
```

Example:

```
simd query distribution commission cosmosvaloper1..
```

Example Output:

```
commission:
- amount: "1000000.000000000000000000"
  denom: stake
```

#### community-pool

The `community-pool` command allows users to query all coin balances within the community pool.

```
simd query distribution community-pool [flags]
```

Example:

```
simd query distribution community-pool
```

Example Output:

```
pool:
- amount: "1000000.000000000000000000"
  denom: stake
```

#### params

The `params` command allows users to query the parameters of the `distribution` module.

```
simd query distribution params [flags]
```

Example:

```
simd query distribution params
```

Example Output:

```
base_proposer_reward: "0.010000000000000000"
bonus_proposer_reward: "0.040000000000000000"
community_tax: "0.020000000000000000"
withdraw_addr_enabled: true
```

#### rewards

The `rewards` command allows users to query delegator rewards. Users can optionally include the validator address to query rewards earned from a specific validator.

```
simd query distribution rewards [delegator-addr] [validator-addr] [flags]
```

Example:

```
simd query distribution rewards cosmos1..
```

Example Output:

```
rewards:
- reward:
  - amount: "1000000.000000000000000000"
    denom: stake
  validator_address: cosmosvaloper1..
total:
- amount: "1000000.000000000000000000"
  denom: stake
```

#### slashes

The `slashes` command allows users to query all slashes for a given block range.

```
simd query distribution slashes [validator] [start-height] [end-height] [flags]
```

Example:

```
simd query distribution slashes cosmosvaloper1.. 1 1000
```

Example Output:

```
pagination:
  next_key: null
  total: "0"
slashes:
- validator_period: 20,
  fraction: "0.009999999999999999"
```

#### validator-outstanding-rewards

The `validator-outstanding-rewards` command allows users to query all outstanding (un-withdrawn) rewards for a validator and all their delegations.

```
simd query distribution validator-outstanding-rewards [validator] [flags]
```

Example:

```
simd query distribution validator-outstanding-rewards cosmosvaloper1..
```

Example Output:

```
rewards:
- amount: "1000000.000000000000000000"
  denom: stake
```

### Transactions

The `tx` commands allow users to interact with the `distribution` module.

```
simd tx distribution --help
```

#### fund-community-pool

The `fund-community-pool` command allows users to send funds to the community pool.

```
simd tx distribution fund-community-pool [amount] [flags]
```

Example:

```
simd tx distribution fund-community-pool 100stake --from cosmos1..
```

#### set-withdraw-addr

The `set-withdraw-addr` command allows users to set the withdraw address for rewards associated with a delegator address.

```
simd tx distribution set-withdraw-addr [withdraw-addr] [flags]
```

Example:

```
simd tx distribution set-withdraw-addr cosmos1.. --from cosmos1..
```

#### withdraw-all-rewards

The `withdraw-all-rewards` command allows users to withdraw all rewards for a delegator.

```
simd tx distribution withdraw-all-rewards [flags]
```

Example:

```
simd tx distribution withdraw-all-rewards --from cosmos1..
```

#### withdraw-rewards

The `withdraw-rewards` command allows users to withdraw all rewards from a given delegation address,
and optionally withdraw validator commission if the delegation address given is a validator operator and the user proves the `--commision` flag.

```
simd tx distribution withdraw-rewards [validator-addr] [flags]
```

Example:

```
simd tx distribution withdraw-rewards cosmosvaloper1.. --from cosmos1.. --commision
```

## gRPC

A user can query the `distribution` module using gRPC endpoints.

### Params

The `Params` endpoint allows users to query parameters of the `distribution` module.

Example:

```
grpcurl -plaintext \
    localhost:9090 \
    cosmos.distribution.v1beta1.Query/Params
```

Example Output:

```
{
  "params": {
    "communityTax": "20000000000000000",
    "baseProposerReward": "10000000000000000",
    "bonusProposerReward": "40000000000000000",
    "withdrawAddrEnabled": true
  }
}
```

### ValidatorOutstandingRewards

The `ValidatorOutstandingRewards` endpoint allows users to query rewards of a validator address.

Example:

```
grpcurl -plaintext \
    -d '{"validator_address":"cosmosvalop1.."}' \
    localhost:9090 \
    cosmos.distribution.v1beta1.Query/ValidatorOutstandingRewards
```

Example Output:

```
{
  "rewards": {
    "rewards": [
      {
        "denom": "stake",
        "amount": "1000000000000000"
      }
    ]
  }
}
```

### ValidatorCommission

The `ValidatorCommission` endpoint allows users to query accumulated commission for a validator.

Example:

```
grpcurl -plaintext \
    -d '{"validator_address":"cosmosvalop1.."}' \
    localhost:9090 \
    cosmos.distribution.v1beta1.Query/ValidatorCommission
```

Example Output:

```
{
  "commission": {
    "commission": [
      {
        "denom": "stake",
        "amount": "1000000000000000"
      }
    ]
  }
}
```

### ValidatorSlashes

The `ValidatorSlashes` endpoint allows users to query slash events of a validator.

Example:

```
grpcurl -plaintext \
    -d '{"validator_address":"cosmosvalop1.."}' \
    localhost:9090 \
    cosmos.distribution.v1beta1.Query/ValidatorSlashes
```

Example Output:

```
{
  "slashes": [
    {
      "validator_period": "20",
      "fraction": "0.009999999999999999"
    }
  ],
  "pagination": {
    "total": "1"
  }
}
```

### DelegationRewards

The `DelegationRewards` endpoint allows users to query the total rewards accrued by a delegation.

Example:

```
grpcurl -plaintext \
    -d '{"delegator_address":"cosmos1..","validator_address":"cosmosvalop1.."}' \
    localhost:9090 \
    cosmos.distribution.v1beta1.Query/DelegationRewards
```

Example Output:

```
{
  "rewards": [
    {
      "denom": "stake",
      "amount": "1000000000000000"
    }
  ]
}
```

### DelegationTotalRewards

The `DelegationTotalRewards` endpoint allows users to query the total rewards accrued by each validator.

Example:

```
grpcurl -plaintext \
    -d '{"delegator_address":"cosmos1.."}' \
    localhost:9090 \
    cosmos.distribution.v1beta1.Query/DelegationTotalRewards
```

Example Output:

```
{
  "rewards": [
    {
      "validatorAddress": "cosmosvaloper1..",
      "reward": [
        {
          "denom": "stake",
          "amount": "1000000000000000"
        }
      ]
    }
  ],
  "total": [
    {
      "denom": "stake",
      "amount": "1000000000000000"
    }
  ]
}
```

### DelegatorValidators

The `DelegatorValidators` endpoint allows users to query all validators for given delegator.

Example:

```
grpcurl -plaintext \
    -d '{"delegator_address":"cosmos1.."}' \
    localhost:9090 \
    cosmos.distribution.v1beta1.Query/DelegatorValidators
```

Example Output:

```
{
  "validators": [
    "cosmosvaloper1.."
  ]
}
```

### DelegatorWithdrawAddress

The `DelegatorWithdrawAddress` endpoint allows users to query the withdraw address of a delegator.

Example:

```
grpcurl -plaintext \
    -d '{"delegator_address":"cosmos1.."}' \
    localhost:9090 \
    cosmos.distribution.v1beta1.Query/DelegatorWithdrawAddress
```

Example Output:

```
{
  "withdrawAddress": "cosmos1.."
}
```

### CommunityPool

The `CommunityPool` endpoint allows users to query the community pool coins.

Example:

```
grpcurl -plaintext \
    localhost:9090 \
    cosmos.distribution.v1beta1.Query/CommunityPool
```

Example Output:

```
{
  "pool": [
    {
      "denom": "stake",
      "amount": "1000000000000000000"
    }
  ]
}
```
