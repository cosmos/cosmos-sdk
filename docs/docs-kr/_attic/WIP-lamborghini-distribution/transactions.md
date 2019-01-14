# Transactions

## TxWithdrawDelegation

When a delegator wishes to withdraw their transaction fees it must send
`TxWithdrawDelegation`. Note that parts of this transaction logic are also
triggered each with any change in individual delegations, such as an unbond,
redelegation, or delegation of additional tokens to a specific validator.  

Each time a withdrawal is made by a recipient the adjustment term must be
modified for each block with a change in distributors shares since the time of
last withdrawal.  This is accomplished by iterating over all relevant
`PowerChange`'s stored in distribution state.


```golang
type TxWithdrawDelegation struct {
    delegatorAddr sdk.AccAddress
    withdrawAddr  sdk.AccAddress // address to make the withdrawal to
}

func WithdrawDelegator(delegatorAddr, withdrawAddr sdk.AccAddress) 
    entitlement = GetDelegatorEntitlement(delegatorAddr)
    AddCoins(withdrawAddr, totalEntitlment.TruncateDecimal())

func GetDelegatorEntitlement(delegatorAddr sdk.AccAddress) DecCoins
    
    // compile all the distribution scenarios
    delegations = GetDelegations(delegatorAddr)
    DelDistr = GetDelegationDistribution(delegation.DelegatorAddr,
                                         delegation.ValidatorAddr)
    pcs = GetPowerChanges(DelDistr.WithdrawalHeight)
        
    // update all adjustment factors for each delegation since last withdrawal
    for pc = range pcs 
        for delegation = range delegations 
            DelDistr = GetDelegationDistribution(delegation.DelegatorAddr,
                                                 delegation.ValidatorAddr)
            pc.ProcessPowerChangeDelegation(delegation, DelDistr) 
    
    // collect all entitled fees
    entitlement = 0
    for delegation = range delegations 
        global = GetGlobal() 
        pool = GetPool() 
        DelDistr = GetDelegationDistribution(delegation.DelegatorAddr,
                        delegation.ValidatorAddr)
        ValDistr = GetValidatorDistribution(delegation.ValidatorAddr)
        validator = GetValidator(delegation.ValidatorAddr)

        scenerio1 = NewDelegationFromGlobalPool(delegation, validator, 
                        pool, global, ValDistr, DelDistr)
        scenerio2 = NewDelegationFromProvisionPool(delegation, validator, 
                        ValDistr, DelDistr)
        entitlement += scenerio1.WithdrawalEntitlement()
        entitlement += scenerio2.WithdrawalEntitlement()
    
    return entitlement

func (pc PowerChange) ProcessPowerChangeDelegation(delegation sdk.Delegation, 
                      DelDistr DelegationDistribution) 

    // get the historical scenarios
    scenario1 = pc.DelegationFromGlobalPool(delegation, DelDistr) 
    scenario2 = pc.DelegationFromProvisionPool(delegation, DelDistr) 

    // process the adjustment factors 
    scenario1.UpdateAdjustmentForPowerChange(pc.Height) 
    scenario2.UpdateAdjustmentForPowerChange(pc.Height) 
```

## TxWithdrawValidator

When a validator wishes to withdraw their transaction fees it must send
`TxWithdrawDelegation`. Note that parts of this transaction logic is also
triggered each with any change in individual delegations, such as an unbond,
redelegation, or delegation of additional tokens to a specific validator. This
transaction withdraws the validators commission rewards, as well as any rewards
earning on their self-delegation. 

```golang
type TxWithdrawValidator struct {
    ownerAddr    sdk.AccAddress // validator address to withdraw from 
    withdrawAddr sdk.AccAddress // address to make the withdrawal to
}

func WithdrawalValidator(ownerAddr, withdrawAddr sdk.AccAddress)

    // update the delegator adjustment factors and also withdrawal delegation fees
    entitlement = GetDelegatorEntitlement(ownerAddr)
    
    // update the validator adjustment factors for commission 
    ValDistr = GetValidatorDistribution(ownerAddr.ValidatorAddr)
    pcs = GetPowerChanges(ValDistr.CommissionWithdrawalHeight)
    for pc = range pcs 
        pc.ProcessPowerChangeCommission()

    // withdrawal validator commission rewards
    global = GetGlobal() 
    pool = GetPool() 
    ValDistr = GetValidatorDistribution(delegation.ValidatorAddr)
    validator = GetValidator(delegation.ValidatorAddr)

    scenerio1 = NewCommissionFromGlobalPool(validator, 
                    pool, global, ValDistr)
    scenerio2 = CommissionFromProposerPool(validator, ValDistr)
    entitlement += scenerio1.WithdrawalEntitlement()
    entitlement += scenerio2.WithdrawalEntitlement()
    
    AddCoins(withdrawAddr, totalEntitlment.TruncateDecimal())

func (pc PowerChange) ProcessPowerChangeCommission() 

    // get the historical scenarios
    scenario1 = pc.CommissionFromGlobalPool()
    scenario2 = pc.CommissionFromProposerPool()

    // process the adjustment factors 
    scenario1.UpdateAdjustmentForPowerChange(pc.Height) 
    scenario2.UpdateAdjustmentForPowerChange(pc.Height) 
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

    ModifyAdjustments(withdrawal sdk.Dec)    // proceedure to modify adjustment factors
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
based on the current blocks token input (for example fees reward received)  to
the distributor, and the distributor's tokens and shares of the previous block
assuming that neither had changed in the current block. Using the simple and
projected pools we can determine all cumulative changes which have taken place
outside of the recipient and adjust the recipient's _adjustment factor_ to
account for these changes and ultimately keep track of the correct entitlement
to the distributors tokens. 

```
func (d DistributionScenario) RecipientCount(height int64) sdk.Dec
    return v.RecipientShares() * height

func (d DistributionScenario) GlobalCount(height int64) sdk.Dec
    return d.DistributorShares() * height

func (d DistributionScenario) SimplePool() DecCoins
    return d.RecipientCount() / d.GlobalCount() * d.DistributorCumulativeTokens

func (d DistributionScenario) ProjectedPool(height int64) DecCoins
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
func (d DistributionScenario) UpdateAdjustmentForPowerChange(height int64) 
    simplePool = d.SimplePool()
    projectedPool = d.ProjectedPool(height)
    AdjustmentChange = simplePool - projectedPool
    if AdjustmentChange > 0 
        d.ModifyAdjustments(AdjustmentChange) 

func (d DistributionScenario) WithdrawalEntitlement() DecCoins
    entitlement = d.SimplePool() - d.RecipientAdjustment()
    d.ModifyAdjustments(entitlement)
    return entitlement
```

### Distribution scenarios

Note that the distribution scenario structures are found in `state.md`. 

#### Delegation's entitlement to Global.Pool

For delegations (including validator's self-delegation) all fees from fee pool
are subject to commission rate from the operator of the validator. The global
shares should be taken as true number of global bonded shares. The recipients
shares should be taken as the bonded tokens less the validator's commission.

```
type DelegationFromGlobalPool struct {
    DelegationShares              sdk.Dec
    ValidatorCommission           sdk.Dec
    ValidatorBondedTokens         sdk.Dec
    ValidatorDelegatorShareExRate sdk.Dec
    PoolBondedTokens              sdk.Dec
    Global                        Global
    ValDistr                      ValidatorDistribution
    DelDistr                      DelegatorDistribution
}

func (d DelegationFromGlobalPool) DistributorTokens() DecCoins
    return d.Global.Pool

func (d DelegationFromGlobalPool) DistributorCumulativeTokens() DecCoins
    return d.Global.EverReceivedPool

func (d DelegationFromGlobalPool) DistributorPrevReceivedTokens() DecCoins
    return d.Global.PrevReceivedPool
    
func (d DelegationFromGlobalPool) DistributorShares() sdk.Dec
    return d.PoolBondedTokens

func (d DelegationFromGlobalPool) DistributorPrevShares() sdk.Dec
    return d.Global.PrevBondedTokens

func (d DelegationFromGlobalPool) RecipientShares() sdk.Dec
    return d.DelegationShares * d.ValidatorDelegatorShareExRate() * 
           d.ValidatorBondedTokens() * (1 - d.ValidatorCommission)

func (d DelegationFromGlobalPool) RecipientPrevShares() sdk.Dec
    return d.DelDistr.PrevTokens

func (d DelegationFromGlobalPool) RecipientAdjustment() sdk.Dec
    return d.DelDistr.Adjustment

func (d DelegationFromGlobalPool) ModifyAdjustments(withdrawal sdk.Dec)
    d.ValDistr.Adjustment += withdrawal
    d.DelDistr.Adjustment += withdrawal
    d.global.Adjustment += withdrawal
    SetValidatorDistribution(d.ValDistr)
    SetDelegatorDistribution(d.DelDistr)
    SetGlobal(d.Global)
```

#### Delegation's entitlement to ValidatorDistribution.ProposerPool

Delegations (including validator's self-delegation) are still subject
commission on the rewards gained from the proposer pool.  Global shares in this
context is actually the validators total delegations shares.  The recipient's
shares is taken as the effective delegation shares less the validator's
commission. 

```
type DelegationFromProposerPool struct {
    DelegationShares         sdk.Dec
    ValidatorCommission      sdk.Dec
    ValidatorDelegatorShares sdk.Dec
    ValDistr                 ValidatorDistribution
    DelDistr                 DelegatorDistribution
}

func (d DelegationFromProposerPool) DistributorTokens() DecCoins
    return d.ValDistr.ProposerPool

func (d DelegationFromProposerPool) DistributorCumulativeTokens() DecCoins
    return d.ValDistr.EverReceivedProposerReward

func (d DelegationFromProposerPool) DistributorPrevReceivedTokens() DecCoins
    return d.ValDistr.PrevReceivedProposerReward

func (d DelegationFromProposerPool) DistributorShares() sdk.Dec
    return d.ValidatorDelegatorShares

func (d DelegationFromProposerPool) DistributorPrevShares() sdk.Dec
    return d.ValDistr.PrevDelegatorShares

func (d DelegationFromProposerPool) RecipientShares() sdk.Dec
    return d.DelegationShares * (1 - d.ValidatorCommission)

func (d DelegationFromProposerPool) RecipientPrevShares() sdk.Dec
    return d.DelDistr.PrevShares

func (d DelegationFromProposerPool) RecipientAdjustment() sdk.Dec
    return d.DelDistr.AdjustmentProposer

func (d DelegationFromProposerPool) ModifyAdjustments(withdrawal sdk.Dec)
    d.ValDistr.AdjustmentProposer += withdrawal
    d.DelDistr.AdjustmentProposer += withdrawal
    SetValidatorDistribution(d.ValDistr)
    SetDelegatorDistribution(d.DelDistr)
```

#### Validators's commission entitlement to Global.Pool

Similar to a delegator's entitlement, but with recipient shares based on the
commission portion of bonded tokens.

```
type CommissionFromGlobalPool struct {
    ValidatorBondedTokens sdk.Dec
    ValidatorCommission   sdk.Dec
    PoolBondedTokens      sdk.Dec
    Global                Global
    ValDistr              ValidatorDistribution
}

func (c CommissionFromGlobalPool) DistributorTokens() DecCoins
    return c.Global.Pool

func (c CommissionFromGlobalPool) DistributorCumulativeTokens() DecCoins
    return c.Global.EverReceivedPool

func (c CommissionFromGlobalPool) DistributorPrevReceivedTokens() DecCoins
    return c.Global.PrevReceivedPool
    
func (c CommissionFromGlobalPool) DistributorShares() sdk.Dec
    return c.PoolBondedTokens

func (c CommissionFromGlobalPool) DistributorPrevShares() sdk.Dec
    return c.Global.PrevBondedTokens

func (c CommissionFromGlobalPool) RecipientShares() sdk.Dec
    return c.ValidatorBondedTokens() * c.ValidatorCommission

func (c CommissionFromGlobalPool) RecipientPrevShares() sdk.Dec
    return c.ValDistr.PrevBondedTokens * c.ValidatorCommission

func (c CommissionFromGlobalPool) RecipientAdjustment() sdk.Dec
    return c.ValDistr.Adjustment

func (c CommissionFromGlobalPool) ModifyAdjustments(withdrawal sdk.Dec)
    c.ValDistr.Adjustment += withdrawal
    c.Global.Adjustment += withdrawal
    SetValidatorDistribution(c.ValDistr)
    SetGlobal(c.Global)
```

#### Validators's commission entitlement to ValidatorDistribution.ProposerPool

Similar to a delegators entitlement to the proposer pool, but with recipient
shares based on the commission portion of the total delegator shares.

```
type CommissionFromProposerPool struct {
    ValidatorDelegatorShares sdk.Dec
    ValidatorCommission      sdk.Dec
    ValDistr                 ValidatorDistribution
}

func (c CommissionFromProposerPool) DistributorTokens() DecCoins
    return c.ValDistr.ProposerPool

func (c CommissionFromProposerPool) DistributorCumulativeTokens() DecCoins
    return c.ValDistr.EverReceivedProposerReward

func (c CommissionFromProposerPool) DistributorPrevReceivedTokens() DecCoins
    return c.ValDistr.PrevReceivedProposerReward

func (c CommissionFromProposerPool) DistributorShares() sdk.Dec
    return c.ValidatorDelegatorShares

func (c CommissionFromProposerPool) DistributorPrevShares() sdk.Dec
    return c.ValDistr.PrevDelegatorShares

func (c CommissionFromProposerPool) RecipientShares() sdk.Dec
    return c.ValidatorDelegatorShares * (c.ValidatorCommission)

func (c CommissionFromProposerPool) RecipientPrevShares() sdk.Dec
    return c.ValDistr.PrevDelegatorShares * (c.ValidatorCommission)

func (c CommissionFromProposerPool) RecipientAdjustment() sdk.Dec
    return c.ValDistr.AdjustmentProposer

func (c CommissionFromProposerPool) ModifyAdjustments(withdrawal sdk.Dec)
    c.ValDistr.AdjustmentProposer += withdrawal
    SetValidatorDistribution(c.ValDistr)
```

