# Transactions

## TxWithdrawDelegation

When a delegator wishes to withdraw their transaction fees it must send
`TxWithdrawDelegation`. Note that parts of this transaction logic is also
triggered each with any change in individual delegations, such as an unbond,
redelegation, or delegation of additional tokens to a specific validator.  

```golang
type TxWithdrawDelegation struct {
    delegation sdk.AccAddress
}
```

```
if GetValidator(delegation) == found ; Exit 

```

## TxWithdrawValidator

When a validator wishes to withdraw their transaction fees it must send
`TxWithdrawDelegation`. Note that parts of this transaction logic is also
triggered each with any change in individual delegations, such as an unbond,
redelegation, or delegation of additional tokens to a specific validator.  

```golang
type TxWithdrawDelegation struct {
    delegation sdk.AccAddress
}
```

```
TODO: pseudo-code
```

## Common Calculations

Collected fees are pooled globally and divided out passively to validators and
delegators. Each validator has the opportunity to charge commission to the
delegators on the fees collected on behalf of the delegators by the validators.
Fees are paid directly into a global fee pool. Due to the nature of of passive
accounting whenever changes to parameters which affect the rate of fee
distribution occurs, withdrawal of fees must also occur. 
 
 - when withdrawing one must withdrawal the maximum amount they are entitled
   too, leaving nothing in the pool, 
 - when bonding, unbonding, or re-delegating tokens to an existing account a
   full withdrawal of the fees must occur (as the rules for lazy accounting
   change), 
 - when a validator chooses to change the commission on fees, all accumulated 
   commission fees must be simultaneously withdrawn.

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
     validator.ProposerRewardPool += proposerReward
     
     communityFunding = feesCollectedDec * communityTax
     gs.CommunityFund += communityFunding
     
     poolReceived = feesCollectedDec - proposerReward - communityFunding
     g.Pool += poolReceived
     g.EverReceivedPool += poolReceived
     g.LastReceivedPool = poolReceived
```

The entitlement to the fee pool held by the each validator can be accounted for
lazily.  To begin this calculation we must determine the  validator's `Count`
and `Adjustment`. The `Count` represents a lazy accounting of what a
validator's entitlement to the fee pool would be if all validators have always
had static voting power, and no validator had ever withdrawn their entitled
rewards. 

``` 
func (v ValidatorDistribution) Count() int64
    return v.BondedTokens() * BlockHeight

func (p Pool) Count() int64
    return p.TotalBondedTokens() * BlockHeight
``` 

The `Adjustment` term accounts for changes in voting power and withdrawals of
fees. The adjustment factor must be persisted with the validator and modified
whenever fees are withdrawn from the validator or the voting power of the
validator changes. When the voting power of the validator changes the
`Adjustment` factor is increased/decreased by the cumulative difference in the
voting power if the voting power has been the new voting power as opposed to
the old voting power for the entire duration of the blockchain up the previous
block. Each time there is an adjustment change the `Global.Adjustment` must
also be updated.

```
func (v ValidatorDistibution) SimplePool(g Global) DecCoins
    return v.Count() / g.Count() * g.SumFeesReceived

func (v ValidatorDistibution) ProjectedPool(height int64, g Global, 
                                            pool stake.Pool, val stake.Validator) DecCoins
    return v.PrevPower * (height-1) 
           / (g.PrevPower * (height-1)) 
           * g.EverReceivedPool
           + val.Power() / Pool.TotalPower() 
           * g.LastReceivedPool

func UpdateAdjustment(g Global, v ValidatorDistibution, 
                      simplePool, projectedPool DecCoins) (Global, ValidatorDistibution)
                                            
    AdjustmentChange = simplePool - projectedPool
    v.AdjustmentRewardPool += AdjustmentChange
    g.Adjustment += AdjustmentChange
    return g, v
```

Every instance that the voting power changes, information about the state of
the validator set during the change must be recorded as a `powerChange` for
other validators to run through. Before any validator modifies its voting power
it must first run through the above calculation to determine the change in
their `candidate.AdjustmentRewardPool` for all historical changes in the set
of `powerChange` which they have not yet synced to. 

Note that the adjustment factor may result as negative if the voting power of a
different validator has decreased.  

``` 
validator.AdjustmentRewardPool += withdrawn
gs.Adjustment += withdrawn
``` 

Now the entitled fee pool of each validator can be lazily accounted for at 
any given block:

```
func (v ValidatorDistibution) Pool(g Global) DecCoins
    return v.simplePool() - v.Adjustment
```

So far we have covered two sources fees which can be withdrawn from: Fees from
proposer rewards (`validator.ProposerRewardPool`), and fees from the fee pool
(`validator.feePool`). However we should note that all fees from fee pool are
subject to commission rate from the owner of the validator. These next
calculations outline the math behind withdrawing fee rewards as either a
delegator to a validator providing commission, or as the owner of a validator
who is receiving commission.

### Calculations For Delegators and Validators

The same mechanism described to calculate the fees which an entire validator is
entitled to is be applied to delegator level to determine the entitled fees for
each delegator and the validators entitled commission from `gs.FeesPool` and
`validator.ProposerRewardPool`. 

The calculations are identical with a few modifications to the parameters:
 - Delegator's entitlement to `rewardPool`:
   - entitled party voting power should be taken as the effective voting power
     after commission is retrieved, 
     `bond.Shares/validator.TotalDelegatorShares * validator.VotingPower * (1 - validator.Commission)`
 - Delegator's entitlement to `validatorDistribution.ProposerFeePool` 
   - global power in this context is actually shares
     `validator.TotalDelegatorShares`
   - entitled party voting power should be taken as the effective shares after
     commission is retrieved, `bond.Shares * (1 - validator.Commission)`
 - Validator's commission entitlement to `rewardPool` 
   - entitled party voting power should be taken as the effective voting power
     of commission portion of total voting power, 
     `validator.VotingPower * validator.Commission`
 - Validator's commission entitlement to `validatorDistribution.ProposerFeePool` 
   - global power in this context is actually shares
     `validator.TotalDelegatorShares`
   - entitled party voting power should be taken as the of commission portion
     of total delegators shares, 
     `validator.TotalDelegatorShares * validator.Commission`

For more implementation ideas see spreadsheet `spec/AbsoluteFeeDistrModel.xlsx`

As mentioned earlier, every time the voting power of a delegator bond is
changing either by unbonding or further bonding, all fees must be
simultaneously withdrawn. Similarly if the validator changes the commission
rate, all commission on fees must be simultaneously withdrawn.  

### Other general notes on fees accounting

- When a delegator chooses to re-delegate shares, fees continue to accumulate
  until the re-delegation queue reaches maturity. At the block which the queue
  reaches maturity and shares are re-delegated all available fees are
  simultaneously withdrawn. 
- Whenever a totally new validator is added to the validator set, the `accum`
  of the entire validator must be 0, meaning that the initial value for
  `validator.Adjustment` must be set to the value of `canidate.Count` for the
  height which the validator is added on the validator set.
- The feePool of a new delegator bond will be 0 for the height at which the bond
  was added. This is achieved by setting `DelegatorBond.FeeWithdrawalHeight` to
  the height which the bond was added. 

