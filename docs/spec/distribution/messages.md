# Messages

## MsgWithdrawDelegationRewardsAll

When a delegator wishes to withdraw their rewards it must send
`MsgWithdrawDelegationRewardsAll`. Note that parts of this transaction logic are also
triggered each with any change in individual delegations, such as an unbond,
redelegation, or delegation of additional tokens to a specific validator.  

```golang
type MsgWithdrawDelegationRewardsAll struct {
    DelegatorAddr sdk.AccAddress
}

func WithdrawDelegationRewardsAll(delegatorAddr, withdrawAddr sdk.AccAddress) 
    height = GetHeight()
    withdraw = GetDelegatorRewardsAll(delegatorAddr, height)
    AddCoins(withdrawAddr, withdraw.TruncateDecimal())

func GetDelegatorRewardsAll(delegatorAddr sdk.AccAddress, height int64) DecCoins
    
    // get all distribution scenarios
    delegations = GetDelegations(delegatorAddr)
        
    // collect all entitled rewards
    withdraw = 0
    pool = staking.GetPool() 
    feePool = GetFeePool() 
    for delegation = range delegations 
        delInfo = GetDelegationDistInfo(delegation.DelegatorAddr,
                        delegation.ValidatorAddr)
        valInfo = GetValidatorDistInfo(delegation.ValidatorAddr)
        validator = GetValidator(delegation.ValidatorAddr)

        feePool, diWithdraw = delInfo.WithdrawRewards(feePool, valInfo, height, pool.BondedTokens, 
                   validator.Tokens, validator.DelegatorShares, validator.Commission)
        withdraw += diWithdraw

    SetFeePool(feePool) 
    return withdraw
```

## MsgWithdrawDelegationReward

under special circumstances a delegator may wish to withdraw rewards from only
a single validator. 

```golang
type MsgWithdrawDelegationReward struct {
    DelegatorAddr sdk.AccAddress
    ValidatorAddr sdk.ValAddress
}

func WithdrawDelegationReward(delegatorAddr, validatorAddr, withdrawAddr sdk.AccAddress) 
    height = GetHeight()
    
    // get all distribution scenarios
    pool = staking.GetPool() 
    feePool = GetFeePool() 
    delInfo = GetDelegationDistInfo(delegatorAddr,
                    validatorAddr)
    valInfo = GetValidatorDistInfo(validatorAddr)
    validator = GetValidator(validatorAddr)

    feePool, withdraw = delInfo.WithdrawRewards(feePool, valInfo, height, pool.BondedTokens, 
               validator.Tokens, validator.DelegatorShares, validator.Commission)

    SetFeePool(feePool) 
    AddCoins(withdrawAddr, withdraw.TruncateDecimal())
```


## MsgWithdrawValidatorRewardsAll

When a validator wishes to withdraw their rewards it must send
`MsgWithdrawValidatorRewardsAll`. Note that parts of this transaction logic are also
triggered each with any change in individual delegations, such as an unbond,
redelegation, or delegation of additional tokens to a specific validator. This
transaction withdraws the validators commission fee, as well as any rewards
earning on their self-delegation. 

```
type MsgWithdrawValidatorRewardsAll struct {
    OperatorAddr sdk.ValAddress // validator address to withdraw from 
}

func WithdrawValidatorRewardsAll(operatorAddr, withdrawAddr sdk.AccAddress)

    height = GetHeight()
    feePool = GetFeePool() 
    pool = GetPool() 
    ValInfo = GetValidatorDistInfo(delegation.ValidatorAddr)
    validator = GetValidator(delegation.ValidatorAddr)

    // withdraw self-delegation
    withdraw = GetDelegatorRewardsAll(validator.OperatorAddr, height)

    // withdrawal validator commission rewards
    feePool, commission = valInfo.WithdrawCommission(feePool, valInfo, height, pool.BondedTokens, 
               validator.Tokens, validator.Commission)
    withdraw += commission
    SetFeePool(feePool) 

    AddCoins(withdrawAddr, withdraw.TruncateDecimal())
```
    
## Common calculations 

### Update total validator accum

The total amount of validator accum must be calculated in order to determine
the amount of pool tokens which a validator is entitled to at a particular
block. The accum is always additive to the existing accum. This term is to be
updated each time rewards are withdrawn from the system. 

``` 
func (g FeePool) UpdateTotalValAccum(height int64, totalBondedTokens Dec) FeePool
    blocks = height - g.TotalValAccumUpdateHeight
    g.TotalValAccum += totalDelShares * blocks
    g.TotalValAccumUpdateHeight = height
    return g
```

### Update validator's accums

The total amount of delegator accum must be updated in order to determine the
amount of pool tokens which each delegator is entitled to, relative to the
other delegators for that validator. The accum is always additive to
the existing accum. This term is to be updated each time a
withdrawal is made from a validator. 

``` 
func (vi ValidatorDistInfo) UpdateTotalDelAccum(height int64, totalDelShares Dec) ValidatorDistInfo
    blocks = height - vi.TotalDelAccumUpdateHeight
    vi.TotalDelAccum += totalDelShares * blocks
    vi.TotalDelAccumUpdateHeight = height
    return vi
```

### FeePool pool to validator pool

Every time a validator or delegator executes a withdrawal or the validator is
the proposer and receives new tokens, the relevant validator must move tokens
from the passive global pool to their own pool. It is at this point that the
commission is withdrawn

``` 
func (vi ValidatorDistInfo) TakeFeePoolRewards(g FeePool, height int64, totalBonded, vdTokens, commissionRate Dec) (
                                vi ValidatorDistInfo, g FeePool)

    g.UpdateTotalValAccum(height, totalBondedShares)
    
    // update the validators pool
    blocks = height - vi.FeePoolWithdrawalHeight
    vi.FeePoolWithdrawalHeight = height
    accum = blocks * vdTokens
    withdrawalTokens := g.Pool * accum / g.TotalValAccum 
    commission := withdrawalTokens * commissionRate
    
    g.TotalValAccum -= accumm
    vi.PoolCommission += commission
    vi.PoolCommissionFree += withdrawalTokens - commission
    g.Pool -= withdrawalTokens

    return vi, g
```


### Delegation reward withdrawal

For delegations (including validator's self-delegation) all rewards from reward
pool have already had the validator's commission taken away.

```
func (di DelegationDistInfo) WithdrawRewards(g FeePool, vi ValidatorDistInfo,
    height int64, totalBonded, vdTokens, totalDelShares, commissionRate Dec) (
    di DelegationDistInfo, g FeePool, withdrawn DecCoins)

    vi.UpdateTotalDelAccum(height, totalDelShares) 
    g = vi.TakeFeePoolRewards(g, height, totalBonded, vdTokens, commissionRate) 
    
    blocks = height - di.WithdrawalHeight
    di.WithdrawalHeight = height
    accum = delegatorShares * blocks 
     
    withdrawalTokens := vi.Pool * accum / vi.TotalDelAccum
    vi.TotalDelAccum -= accum

    vi.Pool -= withdrawalTokens
    vi.TotalDelAccum -= accum
    return di, g, withdrawalTokens

```

### Validator commission withdrawal

Commission is calculated each time rewards enter into the validator.

```
func (vi ValidatorDistInfo) WithdrawCommission(g FeePool, height int64, 
          totalBonded, vdTokens, commissionRate Dec) (
          vi ValidatorDistInfo, g FeePool, withdrawn DecCoins)

    g = vi.TakeFeePoolRewards(g, height, totalBonded, vdTokens, commissionRate) 
    
    withdrawalTokens := vi.PoolCommission 
    vi.PoolCommission = 0

    return vi, g, withdrawalTokens
```
