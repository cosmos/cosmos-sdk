<!--
order: 4
-->

# Messages

## MsgSetWithdrawAddress

By default, the withdraw address is delegator address. If a delegator wants to change its withdraw address it must send `MsgSetWithdrawAddress`.
This is only possible if the parameter `WithdrawAddrEnabled` is set to `true`.

The withdraw address cannot be any of the module accounts. These are blocked from being withdraw addresses by being added to the distribution keeper's `blockedAddrs` array at initialization.

Response:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.42.4/proto/cosmos/distribution/v1beta1/tx.proto#L29-L37

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

A delegator can withdraw its rewards for a specific delegation.
Internally in the distribution module, this is treated exactly as if the delegator simply started a new delegation of the same value, simultaneously removing the previous delegation.
The rewards are sent immediately from the distribution `ModuleAccount` to the withdraw address.
Any remainder (truncated decimals) are sent to the community pool.
The starting height of the delegation is set to the current validator period, and the reference count for the previous period is decremented.
The amount withdrawn is deducted from the `ValidatorOutstandingRewards` variable for the validator.

In the F1 distribution, the total rewards are calculated per validator period, and a delegator receives a piece of those rewards in proportion to their stake in the validator.
In basic F1, the total rewards that all the delegators are entitled to between to periods is calculated the following way.
Let `R(X)` be the total accumulated rewards up to period `X` divided by the tokens staked at that time, i.e. the reward ratio to multiply an individual delegators stake with to get their reward.
Then the rewards for all the delegators for staking between periods `A` and `B` are `(R(B) - R(A)) * total stake`.
However, this doesn't take slashing into account.

Taking the slashes into account requires iteration.
Let `F(X)` be the fraction a validator is to be slashed for the slashing event that happened at period `X`.
If the validator was slashed at periods `P1, ..., PN`, where `A < P1`, `PN < B`, we calculate the individual delegator's rewards, `T(A, B)`, as follows:

```
stake := initial stake
rewards := 0
previous := A
for P in P1, ..., PN`:
    rewards = (R(P) - previous) * stake
    stake = stake * F(P)
    previous = P
rewards = rewards + (R(B) - R(PN)) * stake
```

Put another way, the historical rewards are calculated retroactively by playing back all the slashes, and attenuating the delegator's stake at each step.
The final calculated stake should be equivalent to the actual staked coins in the delegation, within a margin of error due to rounding errors.

Response:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.42.4/proto/cosmos/distribution/v1beta1/tx.proto#L42-L50

## WithdrawValidatorCommission

The validator can send this message to withdraw their accumulated commission.
The commission is calculated in every block during `BeginBlock`, so no iteration is required to withdraw.
The amount withdrawn is deducted from the `ValidatorOutstandingRewards` variable for the validator.
Only integer amounts can be sent, so if the accumulated awards have any decimals, the amount is truncated before it's sent, and the remainder is left to be withdrawn later.

## FundCommunityPool

This message sends coins directly from the sender to the community pool.

Expected to fail if for some reason the amount cannot be transferred from the sender to the distribution module account.

```go
func (k Keeper) FundCommunityPool(ctx sdk.Context, amount sdk.Coins, sender sdk.AccAddress) error {
    if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, amount); err != nil {
        return err
    }

	feePool := k.GetFeePool(ctx)
	feePool.CommunityPool = feePool.CommunityPool.Add(sdk.NewDecCoinsFromCoins(amount...)...)
	k.SetFeePool(ctx, feePool)

	return nil
}
```

## Common operations

### Initialize delegation

Every time a delegation is changed, the rewards are withdrawn, and the delegation is reinitialized.
Initializing a delegation means incrementing the validator period and keeping track of the starting period of the delegation.

```go
// initialize starting info for a new delegation
func (k Keeper) initializeDelegation(ctx sdk.Context, val sdk.ValAddress, del sdk.AccAddress) {
    // period has already been incremented - we want to store the period ended by this delegation action
    previousPeriod := k.GetValidatorCurrentRewards(ctx, val).Period - 1

	// increment reference count for the period we're going to track
	k.incrementReferenceCount(ctx, val, previousPeriod)

	validator := k.stakingKeeper.Validator(ctx, val)
	delegation := k.stakingKeeper.Delegation(ctx, del, val)

	// calculate delegation stake in tokens
	// we don't store directly, so multiply delegation shares * (tokens per share)
	// note: necessary to truncate so we don't allow withdrawing more rewards than owed
	stake := validator.TokensFromSharesTruncated(delegation.GetShares())
	k.SetDelegatorStartingInfo(ctx, val, del, types.NewDelegatorStartingInfo(previousPeriod, stake, uint64(ctx.BlockHeight())))
}
```
