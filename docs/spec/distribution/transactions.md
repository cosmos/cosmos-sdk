# Transactions

## TxWithdrawDelegationRewards

When a delegator wishes to withdraw their rewards it must send
`TxWithdrawDelegationRewards`. Note that parts of this transaction logic are also
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
    AddCoins(withdrawAddr, withdraw.TruncateDecimal())

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

        global, diWithdraw = delInfo.WithdrawRewards(global, valInfo, height, pool.BondedTokens, 
                   validator.Tokens, validator.DelegatorShares, validator.Commission)
        withdraw += diWithdraw

    SetGlobal(global) 
    return withdraw
```

## TxWithdrawDelegationReward

under special circumstances a delegator may wish to withdraw rewards from only
a single validator. 

```golang
type TxWithdrawDelegationReward struct {
    delegatorAddr sdk.AccAddress
    validatorAddr sdk.AccAddress
    withdrawAddr  sdk.AccAddress // address to make the withdrawal to
}

func WithdrawDelegationReward(delegatorAddr, validatorAddr, withdrawAddr sdk.AccAddress) 
    height = GetHeight()
    
    // get all distribution scenarios
    pool = stake.GetPool() 
    global = GetGlobal() 
    delInfo = GetDelegationDistInfo(delegatorAddr,
                    validatorAddr)
    valInfo = GetValidatorDistInfo(validatorAddr)
    validator = GetValidator(validatorAddr)

    global, withdraw = delInfo.WithdrawRewards(global, valInfo, height, pool.BondedTokens, 
               validator.Tokens, validator.DelegatorShares, validator.Commission)

    SetGlobal(global) 
    AddCoins(withdrawAddr, withdraw.TruncateDecimal())
```


## TxWithdrawValidatorRewards

When a validator wishes to withdraw their rewards it must send
`TxWithdrawValidatorRewards`. Note that parts of this transaction logic are also
triggered each with any change in individual delegations, such as an unbond,
redelegation, or delegation of additional tokens to a specific validator. This
transaction withdraws the validators commission fee, as well as any rewards
earning on their self-delegation. 

```
type TxWithdrawValidatorRewards struct {
    operatorAddr sdk.AccAddress // validator address to withdraw from 
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
               validator.Tokens, validator.Commission)
    withdraw += commission
    SetGlobal(global) 

    AddCoins(withdrawAddr, withdraw.TruncateDecimal())
```
    
## Common calculations 

### Update total validator accum

The total amount of validator accum must be calculated in order to determine
the amount of pool tokens which a validator is entitled to at a particular
block. The accum is always additive to the existing accum. This term is to be
updated each time rewards are withdrawn from the system. 

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

Every time a validator or delegator executes a withdrawal or the validator is
the proposer and receives new tokens, the relevant validator must move tokens
from the passive global pool to their own pool. It is at this point that the
commission is withdrawn

``` 
func (vi ValidatorDistInfo) TakeAccum(g Global, height int64, totalBonded, vdTokens, commissionRate Dec) g Global
    g.UpdateTotalValAccum(height, totalBondedShares)
    g.UpdateValAccum(height, totalBondedShares)
    
    // update the validators pool
    blocks = height - vi.GlobalWithdrawalHeight
    vi.GlobalWithdrawalHeight = height
    accum = blocks * vdTokens
    withdrawalTokens := g.Pool * accum / g.TotalValAccum 
    commission := withdrawalTokens * commissionRate
    
    g.TotalValAccum -= accumm
    vi.PoolCommission += commission
    vi.PoolCommissionFree += withdrawalTokens - commission
    g.Pool -= withdrawalTokens

    return g
```


### Delegation reward withdrawal

For delegations (including validator's self-delegation) all rewards from reward
pool have already had the validator's commission taken away.

```
func (di DelegatorDistInfo) WithdrawRewards(g Global, vi ValidatorDistInfo,
    height int64, totalBonded, vdTokens, totalDelShares, commissionRate Dec) (g Global, withdrawn DecCoins)

    vi.UpdateTotalDelAccum(height, totalDelShares) 
    g = vi.TakeAccum(g, height, totalBonded, vdTokens, commissionRate) 
    
    blocks = height - di.WithdrawalHeight
    di.WithdrawalHeight = height
    accum = delegatorShares * blocks 
     
    withdrawalTokens := vi.Pool * accum / vi.TotalDelAccum
    vi.TotalDelAccum -= accum

    vi.Pool -= withdrawalTokens
    vi.TotalDelAccum -= accum
    return g, withdrawalTokens

```

### Validator commission withdrawal

Commission is calculated each time rewards enter into the validator.

```
func (vi ValidatorDistInfo) WithdrawCommission(g Global, height int64, 
          totalBonded, vdTokens, commissionRate Dec) (g Global, withdrawn DecCoins)

    g = vi.TakeAccum(g, height, totalBonded, vdTokens, commissionRate) 
    
    withdrawalTokens := vi.PoolCommission 
    vi.PoolCommission = 0

    return g, withdrawalTokens
```
