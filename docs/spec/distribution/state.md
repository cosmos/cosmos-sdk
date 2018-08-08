## State

### Global

All globally tracked parameters for distribution are stored within
`Global`. Rewards are collected and added to the reward pool and
distributed to validators/delegators from here. 

Note that the reward pool holds decimal coins (`DecCoins`) to allow 
for fractions of coins to be received from operations like inflation. 
When coins are distributed from the pool they are truncated back to 
`sdk.Coins` which are non-decimal. 

 - Global:  `0x00 -> amino(global)`

```golang
// coins with decimal 
type DecCoins []DecCoin

type DecCoin struct {
    Amount sdk.Dec
    Denom  string
}

type Global struct {
    Accum             sdk.Dec  // global accumulation factor for lazy calculations
    Pool              DecCoins // funds for all validators which have yet to be withdrawn
    CommunityFund     DecCoins // pool for community funds yet to be spent
}
```

### Validator Distribution

Validator distribution information for the relevant validator is updated each time:
 1. delegation amount to a validator are updated, 
 2. a validator successfully proposes a block and receives a reward,
 3. any delegator withdraws from a validator, or 
 4. the validator withdraws it's commission.

 - ValidatorDistribution:  `0x02 | ValOwnerAddr -> amino(validatorDistribution)`

```golang
type ValidatorDistribution struct {
    CommissionWithdrawalHeight int64    // last time this validator withdrew commission
    Accum                      sdk.Dec  // global pool accumulation factor
    ProposerAccum              sdk.Dec  // proposer pool accumulation factor
    ProposerPool               DecCoins // reward pool collected from being the proposer
}
```

### Delegation Distribution 

Each delegation holds multiple accumulation factors to specify its entitlement to
the rewards from a validator. `Accum` is used to passively calculate
each bonds entitled rewards from the `RewardPool`. `AccumProposer` is used to
passively calculate each bonds entitled rewards from
`ValidatorDistribution.ProposerRewardPool`
 
 - DelegatorDistribution: ` 0x02 | DelegatorAddr | ValOwnerAddr -> amino(delegatorDist)`

```golang
type DelegatorDist struct {
    WithdrawalHeight int64    // last time this delegation withdrew rewards
    Accum            sdk.Dec  // reward provisioning accumulation factor
    AccumProposer    sdk.Dec  // proposers pool accumulation factor
}
```

