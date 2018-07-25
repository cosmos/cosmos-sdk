# Transactions

TODO XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
Include this next text in the transactions

Each time there is an recipient adjustment change the distributor adjustment
must also be updated, this modification is defined by each particular
distribution scenario. 

Each time a withdrawal is made by a recipient the adjustment term must be
modified for each block with a change in distributors shares since the time of
last withdrawal.  This is accomplished by iterating over all relevant
`PowerChange`'s stored in distribution state.

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

### Distribution scenario

A common form of abstracted calculations exists between validators and
delegations attempting to withdrawal their rewards, either from `Global.Pool` 
or from `ValidatorDistribution.ProposerPool`. With the following interface 
fulfilled the entitled fees for the various scenarios can be calculated.

```golang
type DistributionScenario interface {
    DistributorTokens()             DecCoins // current tokens from distributor
    DistributorCumulativeTokens()   DecCoins // total tokens ever received 
    DistributorPrevReceivedTokens() DecCoins // last value of tokens received 
    DistributorShares()             sdk.Dec  // current shares 
    DistributorPrevShares()         sdk.Dec  // shares last block

    RecipientAdjustment()           sdk.Dec  
    RecipientShares()               sdk.Dec  // current shares 
    RecipientPrevShares()           sdk.Dec  // shares last block

    ModifyAdjustments(withdrawal sdk.Dec) // proceedure to modify adjustment factors
}
```

#### Entitled reward from distribution scenario

The entitlement to the distributor's tokens held can be accounted for lazily.
To begin this calculation we must determine the recipient's _simple pool_ and
_projected pool_. The simple pool represents a lazy accounting of what a
recipient's entitlement to the distributor's tokens would be if all recipients
for that distributor had static shares (equal to the current shares), and no
recipients had ever withdrawn their entitled rewards. The projected pool
represents the anticipated recipient's entitlement to the distributors tokens
base on the current blocks token input to the distributor, and the
distributor's tokens and shares of the previous block assuming that neither had
changed in the current block. Using the simple and projected pools we can determine 
all cumulative changes which have taken place outside of the recipient and adjust 
the recipient's _adjustment factor_ to account for these changes and ultimately 
keep track of the correct entitlement to the distributors tokens. 

```
func (d DistributionScenario) RecipientCount(height int64) sdk.Dec
    return v.RecipientShares() * height

func (d DistributionScenario) GlobalCount(height int64) sdk.Dec
    return d.DistributorShares() * height

func (d DistributionScenario) SimplePool() DecCoins
    return d.RecipientCount() / d.GlobalCount() * d.DistributorCumulativeTokens

func (d DistributionScenario) ProjectedPool(height int64, g Global, 
                                            pool stake.Pool, val stake.Validator) DecCoins
    return d.RecipientPrevShares() * (height-1) 
           / (d.DistributorPrevShares() * (height-1)) 
           * d.DistributorCumulativeTokens
           + d.RecipientShares() / d.DistributorShares() 
           * d.DistributorPrevReceivedTokens() 
```

The `DistributionScenario` _adjustment_ terms account for changes in
recipient/distributor shares and recipient withdrawals. The adjustment factor
must be modified whenever the recipient withdraws from the distributor or the
distributor's/recipient's shares are changed. 
 - When the shares of the recipient is changed the adjustment factor is
   increased/decreased by the difference between the _simple_ and _projected_
   pools.  In other words, the cumulative difference in the shares if the shares
   has been the new shares as opposed to the old shares for the entire duration of
   the blockchain up the previous block. 
 - When a recipient makes a withdrawal the adjustment factor is increased by the
   withdrawal amount. 

```
func (d DistributionScenario) UpdateAdjustmentForPowerChange(simplePool, projectedPool DecCoins) 
    AdjustmentChange = simplePool - projectedPool
    if AdjustmentChange > 0 
        d.ModifyAdjustments(AdjustmentChange) 

func (d DistributionScenario) WithdrawalEntitlement() DecCoins
    entitlement = d.SimplePool() - d.RecipientAdjustment()
    d.ModifyAdjustments(entitlement)
    return entitlement
```

### Distribution scenarios

#### Delegation's entitlement to Global.Pool

For delegations (including validator's self-delegation) all fees from fee pool
are subject to commission rate from the owner of the validator. The global
shares should be taken as true number of global bonded shares. The recipients
shares should be taken as the bonded tokens less the validator's commission.

```
type DelegationFromGlobalPool struct {
    delegation sdk.Delegation
    validator  sdk.Validator 
    pool       sdk.StakePool
    global     Global
    valDistr   ValidatorDistribution
    delDistr   DelegatorDistribution
}

func (d DelegationFromGlobalPool) DistributorTokens() DecCoins
    return global.Pool

func (d DelegationFromGlobalPool) DistributorCumulativeTokens() DecCoins
    return global.EverReceivedPool

func (d DelegationFromGlobalPool) DistributorPrevReceivedTokens() DecCoins
    return global.PrevReceivedPool
    
func (d DelegationFromGlobalPool) DistributorShares() sdk.Dec
    return pool.BondedTokens

func (d DelegationFromGlobalPool) DistributorPrevShares() sdk.Dec
    return global.PrevBondedTokens

func (d DelegationFromGlobalPool) RecipientShares() sdk.Dec
    return d.delegation.Shares * d.validator.DelegatorShareExRate() * 
           d.validator.BondedTokens() * (1 - d.validator.Commission)

func (d DelegationFromGlobalPool) RecipientPrevShares() sdk.Dec
    return d.delDistr.PrevTokens

func (d DelegationFromGlobalPool) RecipientAdjustment() sdk.Dec
    return d.delDistr.Adjustment

func (d DelegationFromGlobalPool) ModifyAdjustments(withdrawal sdk.Dec)
    d.valDistr.Adjustment += withdrawal
    d.delDistr.Adjustment += withdrawal
    d.global.Adjustment += withdrawal
    SetValidatorDistribution(d.valDistr)
    SetDelegatorDistribution(d.delDistr)
    SetGlobal(global)
```

#### Delegation's entitlement to ValidatorDistribution.ProposerPool

Delegations (including validator's self-delegation) are still subject
commission on the rewards gained from the proposer pool.  Global shares in this
context is actually the validators total delegations shares.  The recipient's
shares is taken as the effective delegation shares less the validator's
commission. 

```
type DelegationFromProposerPool struct {
    delegation sdk.Delegation
    validator  sdk.Validator 
    pool       sdk.StakePool
    valDistr   ValidatorDistribution
    delDistr   DelegatorDistribution
}

func (d DelegationFromProposerPool) DistributorTokens() DecCoins
    return valDistr.ProposerPool

func (d DelegationFromProposerPool) DistributorCumulativeTokens() DecCoins
    return valDistr.EverReceivedProposerReward

func (d DelegationFromProposerPool) DistributorPrevReceivedTokens() DecCoins
    return valDistr.PrevReceivedProposerReward

func (d DelegationFromProposerPool) DistributorShares() sdk.Dec
    return validator.DelegatorShares

func (d DelegationFromProposerPool) DistributorPrevShares() sdk.Dec
    return validator.PrevDelegatorShares

func (d DelegationFromProposerPool) RecipientShares() sdk.Dec
    return d.delegation.Shares * (1 - d.validator.Commission)

func (d DelegationFromProposerPool) RecipientPrevShares() sdk.Dec
    return d.delDistr.PrevShares

func (d DelegationFromProposerPool) RecipientAdjustment() sdk.Dec
    return d.delDistr.AdjustmentProposer

func (d DelegationFromProposerPool) ModifyAdjustments(withdrawal sdk.Dec)
    d.valDistr.AdjustmentProposer += withdrawal
    d.delDistr.AdjustmentProposer += withdrawal
    SetValidatorDistribution(d.valDistr)
    SetDelegatorDistribution(d.delDistr)
```

#### Validators's commission entitlement to Global.Pool

Similar to a delegator's entitlement, but with recipient shares based on the
commission portion of bonded tokens.

```
type CommissionFromGlobalPool struct {
    delegation sdk.Delegation
    validator  sdk.Validator 
    pool       sdk.StakePool
    global     Global
    valDistr   ValidatorDistribution
}

func (c CommissionFromGlobalPool) DistributorTokens() DecCoins
    return global.Pool

func (c CommissionFromGlobalPool) DistributorCumulativeTokens() DecCoins
    return global.EverReceivedPool

func (c CommissionFromGlobalPool) DistributorPrevReceivedTokens() DecCoins
    return global.PrevReceivedPool
    
func (c CommissionFromGlobalPool) DistributorShares() sdk.Dec
    return pool.BondedTokens

func (c CommissionFromGlobalPool) DistributorPrevShares() sdk.Dec
    return global.PrevBondedTokens

func (c CommissionFromGlobalPool) RecipientShares() sdk.Dec
    return c.validator.BondedTokens() * c.validator.Commission

func (c CommissionFromGlobalPool) RecipientPrevShares() sdk.Dec
    return d.valDistr.PrevBondedTokens * c.validator.Commission

func (c CommissionFromGlobalPool) RecipientAdjustment() sdk.Dec
    return d.valDistr.Adjustment

func (c CommissionFromGlobalPool) ModifyAdjustments(withdrawal sdk.Dec)
    d.valDistr.Adjustment += withdrawal
    d.global.Adjustment += withdrawal
    SetValidatorDistribution(d.valDistr)
    SetGlobal(d.global)
```

#### Validators's commission entitlement to ValidatorDistribution.ProposerPool

Similar to a delegators entitlement to the proposer pool, but with recipient
shares based on the commission portion of the total delegator shares.

```
type CommissionFromProposerPool struct {
    validator  sdk.Validator 
    pool       sdk.StakePool
    valDistr   ValidatorDistribution
}

func (c CommissionFromProposerPool) DistributorTokens() DecCoins
    return valDistr.ProposerPool

func (c CommissionFromProposerPool) DistributorCumulativeTokens() DecCoins
    return valDistr.EverReceivedProposerReward

func (c CommissionFromProposerPool) DistributorPrevReceivedTokens() DecCoins
    return valDistr.PrevReceivedProposerReward

func (c CommissionFromProposerPool) DistributorShares() sdk.Dec
    return validator.DelegatorShares

func (c CommissionFromProposerPool) DistributorPrevShares() sdk.Dec
    return validator.PrevDelegatorShares

func (c CommissionFromProposerPool) RecipientShares() sdk.Dec
    return d.validator.DelegatorShares * (d.validator.Commission)

func (c CommissionFromProposerPool) RecipientPrevShares() sdk.Dec
    return d.validator.PrevDelegatorShares * (d.validator.Commission)

func (c CommissionFromGlobalPool) RecipientAdjustment() sdk.Dec
    return d.valDistr.AdjustmentProposer

func (c CommissionFromProposerPool) ModifyAdjustments(withdrawal sdk.Dec)
    d.valDistr.AdjustmentProposer += withdrawal
    d.delDistr.AdjustmentProposer += withdrawal
    SetValidatorDistribution(d.valDistr)
    SetDelegatorDistribution(d.delDistr)
```

