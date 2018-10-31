## State

### FeePool

All globally tracked parameters for distribution are stored within
`FeePool`. Rewards are collected and added to the reward pool and
distributed to validators/delegators from here. 

Note that the reward pool holds decimal coins (`DecCoins`) to allow 
for fractions of coins to be received from operations like inflation. 
When coins are distributed from the pool they are truncated back to 
`sdk.Coins` which are non-decimal. 

 - FeePool:  `0x00 -> amino(FeePool)`

```golang
// coins with decimal 
type DecCoins []DecCoin

type DecCoin struct {
    Amount sdk.Dec
    Denom  string
}

type FeePool struct {
    TotalValAccumUpdateHeight  int64    // last height which the total validator accum was updated
    TotalValAccum              sdk.Dec  // total valdator accum held by validators
    Pool                       DecCoins // funds for all validators which have yet to be withdrawn
    CommunityPool              DecCoins // pool for community funds yet to be spent
}
```

### Validator Distribution

Validator distribution information for the relevant validator is updated each time:
 1. delegation amount to a validator is updated, 
 2. a validator successfully proposes a block and receives a reward,
 3. any delegator withdraws from a validator, or 
 4. the validator withdraws it's commission.

 - ValidatorDistInfo:  `0x02 | ValOperatorAddr -> amino(validatorDistribution)`

```golang
type ValidatorDistInfo struct {
    FeePoolWithdrawalHeight     int64    // last height this validator withdrew from the global fee pool
    Pool                       DecCoins // rewards owed to delegators, commission has already been charged (includes proposer reward)
    PoolCommission             DecCoins // commission collected by this validator (pending withdrawal) 

    TotalDelAccumUpdateHeight  int64    // last height which the total delegator accum was updated
    TotalDelAccum              sdk.Dec  // total proposer pool accumulation factor held by delegators
}
```

### Delegation Distribution 

Each delegation distribution only needs to record the height at which it last
withdrew fees. Because a delegation must withdraw fees each time it's
properties change (aka bonded tokens etc.) its properties will remain constant
and the delegator's _accumulation_ factor can be calculated passively knowing
only the height of the last withdrawal and its current properties. 
 
 - DelegationDistInfo: ` 0x02 | DelegatorAddr | ValOperatorAddr -> amino(delegatorDist)`

```golang
type DelegationDistInfo struct {
    WithdrawalHeight int64    // last time this delegation withdrew rewards
}
```
