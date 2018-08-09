# Transactions

## TxWithdrawDelegation

When a delegator wishes to withdraw their transaction fees it must send
`TxWithdrawDelegation`. Note that parts of this transaction logic are also
triggered each with any change in individual delegations, such as an unbond,
redelegation, or delegation of additional tokens to a specific validator.  

Each time a withdrawal is made by a recipient the adjustment term must be
modified for each block with a change in distributors shares since the time of
last withdrawal.  This is accomplished by iterating over all relevant
`ValidatorUpdate`'s stored in distribution state.


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
    vus = GetValidatorUpdates(DelDistr.WithdrawalHeight)
        
    // update all adjustment factors for each delegation since last withdrawal
    for vu = range vus 
        for delegation = range delegations 
            DelDistr = GetDelegationDistribution(delegation.DelegatorAddr,
                                                 delegation.ValidatorAddr)
            vu.ProcessPowerChangeDelegation(delegation, DelDistr) 
    
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

func (vu ValidatorUpdate) ProcessPowerChangeDelegation(delegation sdk.Delegation, 
                      DelDistr DelegationDistribution) 

    // get the historical scenarios
    scenario1 = vu.DelegationFromGlobalPool(delegation, DelDistr) 
    scenario2 = vu.DelegationFromProvisionPool(delegation, DelDistr) 

    // process the adjustment factors 
    scenario1.UpdateAdjustmentForPowerChange(vu.Height) 
    scenario2.UpdateAdjustmentForPowerChange(vu.Height) 
```

## TxWithdrawValidator

When a validator wishes to withdraw their transaction fees it must send
`TxWithdrawDelegation`. Note that parts of this transaction logic is also
triggered each with any change in individual delegations, such as an unbond,
redelegation, or delegation of additional tokens to a specific validator. This
transaction withdraws the validators commission rewards, as well as any rewards
earning on their self-delegation. 

```
type TxWithdrawValidator struct {
    ownerAddr    sdk.AccAddress // validator address to withdraw from 
    withdrawAddr sdk.AccAddress // address to make the withdrawal to
}

func WithdrawalValidator(ownerAddr, withdrawAddr sdk.AccAddress)

    // update the delegator adjustment factors and also withdrawal delegation fees
    entitlement = GetDelegatorEntitlement(ownerAddr)
    
    // update the validator adjustment factors for commission 
    ValDistr = GetValidatorDistribution(ownerAddr.ValidatorAddr)
    vus = GetValidatorUpdates(ValDistr.CommissionWithdrawalHeight)
    for vu = range vus 
        vu.ProcessPowerChangeCommission()

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

func (vu ValidatorUpdate) ProcessPowerChangeCommission() 

    // get the historical scenarios
    scenario1 = vu.CommissionFromGlobalPool()
    scenario2 = vu.CommissionFromProposerPool()

    // process the adjustment factors 
    scenario1.UpdateAdjustmentForPowerChange(vu.Height) 
    scenario2.UpdateAdjustmentForPowerChange(vu.Height) 
```

## Common Calculations 


### Update total validator accum

The total amount of validator accum must be calculated in order to determine
the amount of pool tokens which a validator is entitled to at a particular block. 
This term is to be updated during a validator withdrawal. 

``` 
func (g Global) UpdateTotalValAccum() 

TODO 

```

### Update total delegator accum

Each delegation holds multiple accumulation factors to specify its entitlement to
the rewards from a validator. `Accum` is used to passively calculate
each bonds entitled rewards from the `RewardPool`. `AccumProposer` is used to
passively calculate each bonds entitled rewards from
`ValidatorDistribution.ProposerRewardPool`

``` 
TODO
```

### Global Pool to Validator Pool

Everytime a validator or delegator make a withdraw or the validator is the
proposer and receives new tokens - the relavent validator must move tokens from
the passive global pool to thier own pool. 

``` 
TODO
```


### Delegation's entitlement to ValidatorDistribution.Pool

For delegations (including validator's self-delegation) all fees from fee pool
are subject to commission rate from the owner of the validator. The global
shares should be taken as true number of global bonded shares. The recipients
shares should be taken as the bonded tokens less the validator's commission.

Each delegation holds multiple accumulation factors to specify its entitlement to
the rewards from a validator. `Accum` is used to passively calculate
each bonds entitled rewards from the `RewardPool`. `AccumProposer` is used to
passively calculate each bonds entitled rewards from
`ValidatorDistribution.ProposerRewardPool`

```
TODO
```

### Validators's commission entitlement to ValidatorDistribution.Pool

Similar to a delegator's entitlement, but with recipient shares based on the
commission portion of bonded tokens.

```
TODO
```
