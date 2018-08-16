# Transactions

## TxWithdrawDelegation

When a delegator wishes to withdraw their rewards it must send
`TxWithdrawDelegation`. Note that parts of this transaction logic are also
triggered each with any change in individual delegations, such as an unbond,
redelegation, or delegation of additional tokens to a specific validator.  

```golang
type TxWithdrawDelegationRewards struct {
    delegatorAddr sdk.AccAddress
    withdrawAddr  sdk.AccAddress // address to make the withdrawal to
}

func WithdrawDelegationRewards(delegatorAddr, withdrawAddr sdk.AccAddress) 
    height = GetHeight()
    withdraw = GetDelegatorAllWithdraws(delegatorAddr, height)
    AddCoins(withdrawAddr, totalEntitlment.TruncateDecimal())

func GetDelegatorAllWithdraws(delegatorAddr sdk.AccAddress, height int64) DecCoins
    
    // get all distribution scenarios
    delegations = GetDelegations(delegatorAddr)
        
    // collect all entitled rewards
    withdraw = 0
    pool = stake.GetPool() 
    global = GetGlobal() 
    for delegation = range delegations 
        delInfo = GetDelegationDistInfo(delegation.DelegatorAddr,
                        delegation.ValidatorAddr)
        valInfo = GetValidatorDistInfo(delegation.ValidatorAddr)
        validator = GetValidator(delegation.ValidatorAddr)

        // get all commission rate changes since last withdraw
        vus = GetValidatorUpdates(delegation.ValidatorAddr, delInfo.WithdrawalHeight) 
        
        for vu = range vus {
            global, diWithdraw = delInfo.WithdrawRewards(global, valInfo, vu.Height, pool.BondedTokens, 
                       validator.Tokens, validator.DelegatorShares, vu.OldCommissionRate)
            withdraw += diWithdraw
        }

        global, diWithdraw = delInfo.WithdrawRewards(global, valInfo, height, pool.BondedTokens, 
                   validator.Tokens, validator.DelegatorShares, validator.Commission)
        withdraw += diWithdraw

    SetGlobal(global) 
    return withdraw
```

## TxWithdrawValidator

When a validator wishes to withdraw their rewards it must send
`TxWithdrawDelegation`. Note that parts of this transaction logic is also
triggered each with any change in individual delegations, such as an unbond,
redelegation, or delegation of additional tokens to a specific validator. This
transaction withdraws the validators commission fee, as well as any rewards
earning on their self-delegation. 

```
type TxWithdrawValidatorRewards struct {
    operatorAddr    sdk.AccAddress // validator address to withdraw from 
    withdrawAddr sdk.AccAddress // address to make the withdrawal to
}

func WithdrawValidatorRewards(operatorAddr, withdrawAddr sdk.AccAddress)

    height = GetHeight()
    global = GetGlobal() 
    pool = GetPool() 
    ValInfo = GetValidatorDistInfo(delegation.ValidatorAddr)
    validator = GetValidator(delegation.ValidatorAddr)

    // withdraw self-delegation
    withdraw = GetDelegatorAllWithdraws(validator.OperatorAddr, height)

    // withdrawal validator commission rewards
    global, commission = valInfo.WithdrawCommission(global, valInfo, height, pool.BondedTokens, 
               validator.Tokens, validator.DelegatorShares, validator.Commission)
    withdraw += commission
    SetGlobal(global) 

    AddCoins(withdrawAddr, totalEntitlment.TruncateDecimal())
```
    
## Common calculations 

### Update total validator accum

The total amount of validator accum must be calculated in order to determine
the amount of pool tokens which a validator is entitled to at a particular
block. The accum is always additive to the existing accum. This term is to be
updates each time rewards are withdrawn from the system. 

``` 
func (g Global) UpdateTotalValAccum(height int64, totalBondedTokens Dec) 
    blocks = height - g.TotalValAccumUpdateHeight
    g.TotalValAccum += totalDelShares * blocks
    g.TotalValAccumUpdateHeight = height
```

### Update validator's accums

The total amount of delegator accum must be updated in order to determine the
amount of pool tokens which each delegator is entitled to, relative to the
other delegators for that validator. The accum is always additive to
the existing accum. This term is to be updated each time a
withdrawal is made from a validator. 

``` 
func (vi ValidatorDistInfo) UpdateTotalDelAccum(height int64, totalDelShares Dec) 
    blocks = height - vi.TotalDelAccumUpdateHeight
    vi.TotalDelAccum += totalDelShares * blocks
    vi.TotalDelAccumUpdateHeight = height
```

### Global pool to validator pool

Every time a validator or delegator make a withdraw or the validator is the
proposer and receives new tokens - the relevant validator must move tokens from
the passive global pool to their own pool. 

``` 
func (vi ValidatorDistInfo) TakeAccum(g Global, height int64, totalBonded, vdTokens Dec) g Global
    g.UpdateTotalValAccum(height, totalBondedShares)
    g.UpdateValAccum(height, totalBondedShares)
    
    // update the validators pool
    blocks = height - vi.GlobalWithdrawalHeight
    vi.GlobalWithdrawalHeight = height
    accum = blocks * vdTokens
    withdrawalTokens := g.Pool * accum / g.TotalValAccum 
    
    g.TotalValAccum -= accumm
    vi.Pool += withdrawalTokens
    g.Pool -= withdrawalTokens

    return g
```


### Delegation's withdrawal

For delegations (including validator's self-delegation) all rewards from reward pool
are subject to commission rate from the operator of the validator. 

```
func (di DelegatorDistInfo) WithdrawRewards(g Global, vi ValidatorDistInfo,
    height int64, totalBonded, vdTokens, totalDelShares, commissionRate Dec) (g Global, withdrawn DecCoins)

    vi.UpdateTotalDelAccum(height, totalDelShares) 
    g = vi.TakeAccum(g, height, totalBonded, vdTokens) 
    
    blocks = height - di.WithdrawalHeight
    di.WithdrawalHeight = height
    accum = delegatorShares * blocks * (1 - commissionRate)
     
    withdrawalTokens := vi.Pool * accum / vi.TotalDelAccum
    vi.TotalDelAccum -= accum

    vi.Pool -= withdrawalTokens
    vi.TotalDelAccum -= accum
    return g, withdrawalTokens

```

### Validators's commission withdrawal

Similar to a delegator's entitlement, but with recipient shares based on the
commission portion of bonded tokens.

```
func (vi ValidatorDistInfo) WithdrawCommission(g Global, height int64, 
          totalBonded, vdTokens, totalDelShares, commissionRate Dec) (g Global, withdrawn DecCoins)

    vi.UpdateTotalDelAccum(height, totalDelShares) 
    g = vi.TakeAccum(g, height, totalBonded, vdTokens) 
    
    blocks = height - vi.CommissionWithdrawalHeight
    vi.CommissionWithdrawalHeight = height
    accum = delegatorShares * blocks * (commissionRate)
     
    withdrawalTokens := vi.Pool * accum / vi.TotalDelAccum
    vi.TotalDelAccum -= accum

    vi.Pool -= withdrawalTokens
    vi.TotalDelAccum -= accum

    return g, withdrawalTokens
```
