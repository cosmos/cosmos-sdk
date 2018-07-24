## State

### Reward Pool

All rewards are collected in the reward pool and distributed to validators/delegators 
from this pool. 

 - RewardPool:  `0x00 -> amino(rewardPool)`

```golang
type RewardPool sdk.Coins // reward pool for all validators
```

### Validator Distribution

Validator Distribution information for the relavent validator is updated: each
time delegations to a validator are updated, a validator successfully proposes
a block and receives (recieving a reward), any delegator withdraws from a
validator, the validator withdraws. 

 - ValidatorDistribution:  `0x02 | ValOwnerAddr -> amino(validatorDistribution)`

```golang
type ValidatorDistribution struct {
    Adjustment         sdk.Rat   // commission adjustment factor
    ProposerRewardPool sdk.Coins // reward pool collected from being the proposer
	LastBondedTokens   sdk.Rat   // last bonded token amount
}
```

### Delegation

Each delegation holds multiple adjustment factors to specify its entitlement to
the rewards from a validator. `AdjustmentFeePool` is  used to passively
calculate each bonds entitled fees from the `RewardPool`.
`AdjustmentRewardPool` is used to passively calculate each bonds entitled fees
from `ValidatorDistribution.ProposerRewardPool`
 
 - DelegatorDist: ` 0x02 | DelegatorAddr | ValOwnerAddr -> amino(delegatorDist)`

```golang
type DelegatorDist struct {
    HeightLastWithdrawal int64     // last time this delegation withdrew rewards
    AdjustmentFeePool    sdk.Rat   // commission adjustment factor
    AdjustmentRewardPool sdk.Coins // reward pool collected from being the proposer
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
    prevFeePool coins.Coin   // total fees in at previous block height
}
```

### Max Power Change Sequence

To track the latest `PowerChange` record, the maximum `PCSequence` is stored

 - MaxPCSequence:   `0x04 -> amino(PCSequence)`
