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

### Distribution scenarios 

A common form of abstracted calculations exists between validators and
delegations attempting to withdrawal their rewards, either from `Global.Pool` 
or from `ValidatorDistribution.ProposerPool`. With the following interface 
fulfilled the entitled fees for the various scenarios can be calculated 
as explained in later sections.

```golang
type DistributionScenario interface {
    GlobalShares()    sdk.Dec  
    RecipientShares() sdk.Dec
}
```

#### Delegator's entitlement to ValidatorDistribution.Pool()

For delegators all fees from fee pool are subject to commission rate from the
owner of the validator. The global shares should be taken as true number of
global bonded shares. The recipients shares should be taken as the bonded
tokens less the validator's commission.

```
type DelegationFromGlobalPool struct {
    delegation sdk.Delegation
    validator  sdk.Validator 
    pool  sdk.Validator 
}

func (d DelegationFromGlobalPool) GlobalShares() sdk.Dec
    return pool.BondedTokens

func (d DelegationFromGlobalPool) RecipientShares() sdk.Dec
    return d.delegation.Shares * d.validator.DelegatorShareExRate() * 
           validator.BondedTokens() * (1 - validator.Commission)
```

#### Delegator's entitlement to ValidatorDistribution.ProposerPool

Delegations are still subject commission on the rewards gained from the
proposer pool.  Global shares in this context is actually the validators total
delegations shares.  The reciepient shares is taken as the effective delegation
shares less the validator's commission. 


`bond.Shares * (1 - validator.Commission)`

```
type DelegationFromProposerPool struct {
    delegation sdk.Delegation
    validator  sdk.Validator 
    pool  sdk.Validator 
}

func (d DelegationFromProposerPool) GlobalShares() sdk.Dec
    return pool.BondedTokens

func (d DelegationFromProposerPool) RecipientShares() sdk.Dec
    return d.delegation.Shares * d.validator.DelegatorShareExRate() * 
           validator.BondedTokens() * (1 - validator.Commission)
```

#### Validator's commission entitlement to `rewardPool` 
   entitled party voting power should be taken as the effective voting power
     of commission portion of total voting power, 
     `validator.VotingPower * validator.Commission`

#### Validator's commission entitlement to `validatorDistribution.ProposerFeePool` 
   - global power in this context is actually shares
     `validator.TotalDelegatorShares`
   - entitled party voting power should be taken as the of commission portion
     of total delegators shares, 
     `validator.TotalDelegatorShares * validator.Commission`




### Entitled fees from distribution scenarios 

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
     validator.ProposerPool += proposerReward
     
     communityFunding = feesCollectedDec * communityTax
     g.CommunityFund += communityFunding
     
     poolReceived = feesCollectedDec - proposerReward - communityFunding
     g.Pool += poolReceived
     g.EverReceivedPool += poolReceived
     g.LastReceivedPool = poolReceived
```

The entitlement to the fee pool held by the each validator can be accounted for
lazily.  To begin this calculation we must determine the  validator's `Count`
and `Adjustment`. The `Count` represents a lazy accounting of what a
validator's entitlement to the fee pool would be if all validators have always
rewards. 
had static voting power, and no validator had ever withdrawn their entitled

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
    v.Adjustment += AdjustmentChange
    g.Adjustment += AdjustmentChange
    return g, v
```

Before any validator modifies its voting power it must first run through the
above calculation to determine the change in their
`validatorDistribution.Adjustment` for all historical changes in the set of
`powerChange` which they have not yet synced to. 

``` 
func (v ValidatorDistibution) Withdraw(g Global, withdrawal DecCoins) (Global, ValidatorDistribution)
    v.Adjustment += withdrawal
    g.Adjustment += withdrawal
    return g, v
``` 

The entitled pool of for validator can then be lazily accounted for at any
given block:

```
func (v ValidatorDistibution) Pool(g Global) DecCoins
    return v.simplePool() - v.Adjustment
```


