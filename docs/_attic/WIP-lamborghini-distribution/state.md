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
    PrevBondedTokens  sdk.Dec  // bonded token amount for the global pool on the previous block 
    Adjustment        sdk.Dec  // global adjustment factor for lazy calculations
    Pool              DecCoins // funds for all validators which have yet to be withdrawn
    PrevReceivedPool  DecCoins // funds added to the pool on the previous block
    EverReceivedPool  DecCoins // total funds ever added to the pool 
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
    CommissionWithdrawalHeight   int64   // last time this validator withdrew commission
    Adjustment                 sdk.Dec   // global pool adjustment factor
    ProposerAdjustment         DecCoins  // proposer pool adjustment factor
    ProposerPool               DecCoins  // reward pool collected from being the proposer
    EverReceivedProposerReward DecCoins  // all rewards ever collected from being the proposer
    PrevReceivedProposerReward DecCoins  // previous rewards collected from being the proposer
    PrevBondedTokens           sdk.Dec   // bonded token amount on the previous block 
    PrevDelegatorShares        sdk.Dec   // amount of delegator shares for the validator on the previous block 
}
```

### Delegation Distribution 

Each delegation holds multiple adjustment factors to specify its entitlement to
the rewards from a validator. `AdjustmentPool` is  used to passively calculate
each bonds entitled fees from the `RewardPool`.  `AdjustmentPool` is used to
passively calculate each bonds entitled fees from
`ValidatorDistribution.ProposerRewardPool`
 
 - DelegatorDistribution: ` 0x02 | DelegatorAddr | ValOwnerAddr -> amino(delegatorDist)`

```golang
type DelegatorDist struct {
    WithdrawalHeight   int64    // last time this delegation withdrew rewards
    Adjustment         sdk.Dec  // fee provisioning adjustment factor
    AdjustmentProposer DecCoins // proposers pool adjustment factor
    PrevTokens         sdk.Dec  // bonded tokens held by the delegation on the previous block
    PrevShares         sdk.Dec  // delegator shares held by the delegation on the previous block
}
```

### Power Change

Every instance that the voting power changes, information about the state of
the validator set during the change must be recorded as a `PowerChange` for
other validators to run through. Each power change is indexed by its block
height. 

 - PowerChange: `0x03 | amino(Height) -> amino(validatorDist)`

```golang
type PowerChange struct {
    Height                        int64     // block height at change
    ValidatorBondedTokens         sdk.Dec   // following used to create distribution scenarios
    ValidatorDelegatorShares      sdk.Dec
    ValidatorDelegatorShareExRate sdk.Dec
    ValidatorCommission           sdk.Dec
    PoolBondedTokens              sdk.Dec
    Global                        Global
    ValDistr                      ValidatorDistribution
    DelegationShares              sdk.Dec
    DelDistr                      DelegatorDistribution
}
```
