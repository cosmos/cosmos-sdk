## State

### Global

All globally tracked parameters for distribution are stored within
`Global`.  Rewards are collected are added to the reward pool and
distributed to validators/delegators from here. 

Note that the reward pool is held decimal coins (`DecCoins`) to allow 
for fractions of coins to be received from operations like inflation. 
When coins are distributed from the pool they are truncated back to 
`DecCoins` which is non-decimal. 

 - Global:  `0x00 -> amino(global)`

```golang
// coins with decimal 
type DecCoins []DecCoin

type DecCoin struct {
    Amount sdk.Dec
    Denom  string
}

type Global struct {
    Adjustment        sdk.Dec  // global adjustment factor for lazy calculations
    Pool              DecCoins // funds pool for all validators
    LastReceivedPool  DecCoins // last funds added to the pool 
    EverReceivedPool  DecCoins // total funds ever added to the pool 
    CommunityFund     DecCoins // pool for community funds
}
```

### Validator Distribution

Validator Distribution information for the relevant validator is updated: each
time delegations to a validator are updated, a validator successfully proposes
a block and receives (receiving a reward), any delegator withdraws from a
validator, the validator withdraws. 

 - ValidatorDistribution:  `0x02 | ValOwnerAddr -> amino(validatorDistribution)`

```golang
type ValidatorDistribution struct {
    Adjustment       sdk.Dec   // commission adjustment factor
    ProposerPool     DecCoins  // reward pool collected from being the proposer
	LastBondedTokens sdk.Dec   // last bonded token amount
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
    WithdrawalHeight       int64    // last time this delegation withdrew rewards
    AdjustmentPool         sdk.Dec  // commission adjustment factor
    AdjustmentProposerPool DecCoins // reward pool collected from being the proposer
}
```

### Power Change

Every instance that the voting power changes, information about the state of
the validator set during the change must be recorded as a `PowerChange` for
other validators to run through. Each power change is stored under a sequence
number which increments by one for each new power change record

 - PowerChange: `0x03 | amino(PCSequence) -> amino(validatorDist)`

```golang
type PCSequence int64 

type PowerChange struct {
    height      int64        // block height at change
    power       rational.Rat // total power at change
    prevpower   rational.Rat // total power at previous height-1 
    feesIn      coins.Coin   // fees-in at block height
    prevPool    coins.Coin   // total fees in at previous block height
}
```

### Max Power Change Sequence

To track the latest `PowerChange` record, the maximum `PCSequence` is stored

 - MaxPCSequence:   `0x04 -> amino(PCSequence)`
