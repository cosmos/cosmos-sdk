# Fee Distribution

## Overview

Fees are pooled separately and withdrawn lazily, at any time.  They are not
bonded, and can be paid in multiple tokens.  An adjustment factor is maintained
for each validator and delegator to determine the true proportion of fees in
the pool they are entitled too.  Adjustment factors are updated every time a
validator or delegator's voting power changes.  Validators and delegators must
withdraw all fees they are entitled too before they can bond or unbond Atoms.

## Affect on Staking

Because fees are optimized to note

Commission on Atom Provisions and having atoms autobonded are mutually exclusive (we can’t have both). The reason
for this is that 

if there are atoms commissions and autobonding, the portion of
atoms the fee distribution calculation would become very large as the atom
portion for each delegator would change each block making a withdrawal of fees
for a delegator require a calculation for every single block since the last
withdrawal. Conclusion we can only have atom commission and unbonded atoms
provisions, or bonded atom provisions and no atom commission

## Fee Calculations

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
 - when a candidate chooses to change the commission on fees, all accumulated 
   commission fees must be simultaneously withdrawn.

When the validator is the proposer of the round, that validator (and their
delegators) receives between 1% and 5% of fee rewards, the reserve tax is then
charged, then the remainder is distributed socially by voting power to all
validators including the proposer validator.  The amount of proposer reward is
calculated from pre-commits Tendermint messages. All provision rewards are
added to a provision reward pool which validator holds individually. Here note
that `BondedShares` represents the sum of all voting power saved in the
`GlobalState` (denoted `gs`).

```
proposerReward = feesCollected * (0.01 + 0.04 
                  * sumOfVotingPowerOfPrecommitValidators / gs.BondedShares)
candidate.ProposerRewardPool += proposerReward

reserveTaxed = feesCollected * params.ReserveTax
gs.ReservePool += reserveTaxed

distributedReward = feesCollected - proposerReward - reserveTaxed
gs.FeePool += distributedReward
gs.SumFeesReceived += distributedReward
gs.RecentFee = distributedReward
```

The entitlement to the fee pool held by the each validator can be accounted for
lazily.  First we must account for a candidate's `count` and `adjustment`. The
`count` represents a lazy accounting of what that candidates entitlement to the
fee pool would be if there `VotingPower` was to never change and they were to
never withdraw fees. 

``` 
candidate.count = candidate.VotingPower * BlockHeight
``` 

Similarly the GlobalState count can be passively calculated whenever needed,
where `BondedShares` is the updated sum of voting powers from all validators.

``` 
gs.count = gs.BondedShares * BlockHeight
``` 

The `adjustment` term accounts for changes in voting power and withdrawals of
fees. The adjustment factor must be persisted with the candidate and modified
whenever fees are withdrawn from the candidate or the voting power of the
candidate changes. When the voting power of the candidate changes the
`Adjustment` factor is increased/decreased by the cumulative difference in the
voting power if the voting power has been the new voting power as opposed to
the old voting power for the entire duration of the blockchain up the previous
block. Each time there is an adjustment change the GlobalState (denoted `gs`)
`Adjustment` must also be updated.

```
simplePool = candidate.count / gs.count * gs.SumFeesReceived
projectedPool = candidate.PrevPower * (height-1) 
                / (gs.PrevPower * (height-1)) * gs.PrevFeesReceived
                + candidate.Power / gs.Power * gs.RecentFee

AdjustmentChange = simplePool - projectedPool
candidate.AdjustmentRewardPool += AdjustmentChange
gs.Adjustment += AdjustmentChange
```

Every instance that the voting power changes, information about the state of
the validator set during the change must be recorded as a `powerChange` for
other validators to run through. Before any validator modifies its voting power
it must first run through the above calculation to determine the change in
their `caandidate.AdjustmentRewardPool` for all historical changes in the set
of `powerChange` which they have not yet synced to.  The set of all
`powerChange` may be trimmed from its oldest members once all validators have
synced past the height of the oldest `powerChange`.  This trim procedure will
occur on an epoch basis.  

```golang
type powerChange struct {
    height      int64        // block height at change
    power       rational.Rat // total power at change
    prevpower   rational.Rat // total power at previous height-1 
    feesin      coins.Coin   // fees in at block height
    prevFeePool coins.Coin   // total fees in at previous block height
}
```

Note that the adjustment factor may result as negative if the voting power of a
different candidate has decreased.  

``` 
candidate.AdjustmentRewardPool += withdrawn
gs.Adjustment += withdrawn
``` 

Now the entitled fee pool of each candidate can be lazily accounted for at 
any given block:

```
candidate.feePool = candidate.simplePool - candidate.Adjustment
```

So far we have covered two sources fees which can be withdrawn from: Fees from
proposer rewards (`candidate.ProposerRewardPool`), and fees from the fee pool
(`candidate.feePool`). However we should note that all fees from fee pool are
subject to commission rate from the owner of the candidate. These next
calculations outline the math behind withdrawing fee rewards as either a
delegator to a candidate providing commission, or as the owner of a candidate
who is receiving commission.

### Calculations For Delegators and Candidates

The same mechanism described to calculate the fees which an entire validator is
entitled to is be applied to delegator level to determine the entitled fees for
each delegator and the candidates entitled commission from `gs.FeesPool` and
`candidate.ProposerRewardPool`. 

The calculations are identical with a few modifications to the parameters:
 - Delegator's entitlement to `gs.FeePool`:
   - entitled party voting power should be taken as the effective voting power
     after commission is retrieved, 
     `bond.Shares/candidate.TotalDelegatorShares * candidate.VotingPower * (1 - candidate.Commission)`
 - Delegator's entitlement to `candidate.ProposerFeePool` 
   - global power in this context is actually shares
     `candidate.TotalDelegatorShares`
   - entitled party voting power should be taken as the effective shares after
     commission is retrieved, `bond.Shares * (1 - candidate.Commission)`
 - Candidate's commission entitlement to `gs.FeePool` 
   - entitled party voting power should be taken as the effective voting power
     of commission portion of total voting power, 
     `candidate.VotingPower * candidate.Commission`
 - Candidate's commission entitlement to `candidate.ProposerFeePool` 
   - global power in this context is actually shares
     `candidate.TotalDelegatorShares`
   - entitled party voting power should be taken as the of commission portion
     of total delegators shares, 
     `candidate.TotalDelegatorShares * candidate.Commission`

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
  of the entire candidate must be 0, meaning that the initial value for
  `candidate.Adjustment` must be set to the value of `canidate.Count` for the
  height which the candidate is added on the validator set.
- The feePool of a new delegator bond will be 0 for the height at which the bond
  was added. This is achieved by setting `DelegatorBond.FeeWithdrawalHeight` to
  the height which the bond was added. 
