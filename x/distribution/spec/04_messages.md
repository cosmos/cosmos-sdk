<!--
order: 4
-->

# Messages

## MsgSetWithdrawAddress

By default a withdrawal address is delegator address. If a delegator wants to change it's
withdrawal address it must send `MsgSetWithdrawAddress`.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/proto/cosmos/distribution/v1beta1/tx.proto#L29-L37

```go

func (k Keeper) SetWithdrawAddr(ctx sdk.Context, delegatorAddr sdk.AccAddress, withdrawAddr sdk.AccAddress) error 
	if k.blockedAddrs[withdrawAddr.String()] {
		fail with "`{withdrawAddr}` is not allowed to receive external funds"
	}

	if !k.GetWithdrawAddrEnabled(ctx) {
		fail with `ErrSetWithdrawAddrDisabled`
	}

	k.SetDelegatorWithdrawAddr(ctx, delegatorAddr, withdrawAddr)
```

## MsgWithdrawDelegatorReward

Under special circumstances a delegator may wish to withdraw rewards from only
a single validator. 

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/proto/cosmos/distribution/v1beta1/tx.proto#L42-L50

```go
// withdraw rewards from a delegation
func (k Keeper) WithdrawDelegationRewards(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (sdk.Coins, error) {
	val := k.stakingKeeper.Validator(ctx, valAddr)
	if val == nil {
		return nil, types.ErrNoValidatorDistInfo
	}

	del := k.stakingKeeper.Delegation(ctx, delAddr, valAddr)
	if del == nil {
		return nil, types.ErrEmptyDelegationDistInfo
	}

	// withdraw rewards
	rewards, err := k.withdrawDelegationRewards(ctx, val, del)
	if err != nil {
		return nil, err
	}

	// reinitialize the delegation
	k.initializeDelegation(ctx, valAddr, delAddr)
	return rewards, nil
}
```

## Withdraw Validator Rewards All

When a validator wishes to withdraw their rewards it must send an
array of `MsgWithdrawDelegatorReward`. Note that parts of this transaction logic are also
triggered each with any change in individual delegations, such as an unbond,
redelegation, or delegation of additional tokens to a specific validator. This
transaction withdraws the validators commission fee, as well as any rewards
earning on their self-delegation.

```go

for _, valAddr := range validators {
    val, err := sdk.ValAddressFromBech32(valAddr)
    if err != nil {
        return err
    }

    msg := types.NewMsgWithdrawDelegatorReward(delAddr, val)
    if err := msg.ValidateBasic(); err != nil {
        return err
    }
    msgs = append(msgs, msg)
}
```

## Common calculations 

### Update total validator accum

The total amount of validator accum must be calculated in order to determine
the amount of pool tokens which a validator is entitled to at a particular
block. The accum is always additive to the existing accum. This term is to be
updated each time rewards are withdrawn from the system. 

```go
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

``` go
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

```go
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

```go
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

```go
func (vi ValidatorDistInfo) WithdrawCommission(g FeePool, height int64, 
          totalBonded, vdTokens, commissionRate Dec) (
          vi ValidatorDistInfo, g FeePool, withdrawn DecCoins)

    g = vi.TakeFeePoolRewards(g, height, totalBonded, vdTokens, commissionRate) 
    
    withdrawalTokens := vi.PoolCommission 
    vi.PoolCommission = 0

    return vi, g, withdrawalTokens
```
