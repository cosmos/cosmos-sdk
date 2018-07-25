# Triggers

The following triggers are anticipated to be activated under various scenarios
triggered by other modules

## Create validator distribution

 - triggered-by: `stake.TxCreateValidator`

Whenever a totally new validator is added to the Tendermint validator set they
are entitled to begin earning rewards of atom provisions and fess. At this
point `ValidatorDistribution.Pool()` must be zero (as the validator has not yet
earned any rewards) meaning that the initial value for `validator.Adjustment`
must be set to the value of `validator.SimplePool()` for the height which the
validator is added on the validator set. 

```
func CreateValidatorDistribution() 
TODO: pseudo-code
```

## Create or modify delegation distribution
 
 - triggered-by: `stake.TxDelegate`

The pool of a new delegator bond will be 0 for the height at which the bond was
added. This is achieved by setting `DelegationDistribution.WithdrawalHeight` to
the height which the bond was added. Additionally the `AdjustmentPool` and
`AdjustmentProposerPool` must be set to the equivalent values of
`DelegationDistribution.SimplePool()` and
`DelegationDistribution.SimpleProposerPool()` for the height of delegation. 

```
func CreateOrModDelegationDistribution() 
TODO: pseudo-code
```

## Withdraw commission
 
Before any validator modifies its voting power it must first run through the
above calculation to determine the change in their
`validatorDistribution.Adjustment` for all historical changes in the set of
`powerChange` which they have not yet synced to. 

 - triggered-by: `stake.TxEditValidator`

If a validator changes its commission rate, all commission on fees must be
simultaneously withdrawn.  

```
func WithdrawCommission() 
TODO: pseudo-code
```

## Fees Collected
 
 - triggered-by: `auth.CollectFees` //XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX?
 - triggered-by: `stake.ProcessProvisions`

When the validator is the proposer of the round, that validator (and their
delegators) receives between 1% and 5% of fee rewards, the reserve tax is then
charged, then the remainder is distributed socially by voting power to all
validators including the proposer validator.  The amount of proposer reward is
calculated from pre-commits Tendermint messages. All provision rewards are
added to a provision reward pool which validator holds individually
(`ValidatorDistribution.ProvisionsRewardPool`). 

```
func (g Global) Update(feesCollected sdk.Coins, 
     sumPowerPrecommitValidators, totalBondedTokens, communityTax sdk.Dec)

     feesCollectedDec = MakeDecCoins(feesCollected)
     proposerReward = feesCollectedDec * (0.01 + 0.04 
                       * sumPowerPrecommitValidators / totalBondedTokens)
     validator.ProposerPool += proposerReward
     
     communityFunding = feesCollectedDec * communityTax
     g.CommunityFund += communityFunding
     
     poolReceived = feesCollectedDec - proposerReward - communityFunding
     g.Pool += poolReceived
     g.EverReceivedPool += poolReceived
     g.LastReceivedPool = poolReceived
```
