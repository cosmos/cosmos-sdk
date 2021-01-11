<!--
order: 2
-->

# State

## FeePool

All globally tracked parameters for distribution are stored within
`FeePool`. Rewards are collected and added to the reward pool and
distributed to validators/delegators from here.

Note that the reward pool holds decimal coins (`DecCoins`) to allow
for fractions of coins to be received from operations like inflation.
When coins are distributed from the pool they are truncated back to
`sdk.Coins` which are non-decimal.

- FeePool:  `0x00 -> ProtocolBuffer(FeePool)`

```go
// coins with decimal
type DecCoins []DecCoin

type DecCoin struct {
    Amount sdk.Dec
    Denom  string
}
```

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/proto/cosmos/distribution/v1beta1/distribution.proto#L94-L101

## Validator Distribution

Validator distribution information for the relevant validator is updated each time:

 1. delegation amount to a validator is updated,
 2. a validator successfully proposes a block and receives a reward,
 3. any delegator withdraws from a validator, or
 4. the validator withdraws it's commission.

- ValidatorDistInfo:  `0x02 | ValOperatorAddr -> ProtocolBuffer(validatorDistribution)`

```go
type ValidatorDistInfo struct {
    OperatorAddress     sdk.AccAddress
    SelfBondRewards     sdk.DecCoins
    ValidatorCommission types.ValidatorAccumulatedCommission
}
```

## Delegation Distribution

Each delegation distribution only needs to record the height at which it last
withdrew fees. Because a delegation must withdraw fees each time it's
properties change (aka bonded tokens etc.) its properties will remain constant
and the delegator's _accumulation_ factor can be calculated passively knowing
only the height of the last withdrawal and its current properties.

- DelegationDistInfo: `0x02 | DelegatorAddr | ValOperatorAddr -> ProtocolBuffer(delegatorDist)`

```go
type DelegationDistInfo struct {
    WithdrawalHeight int64    // last time this delegation withdrew rewards
}
```
