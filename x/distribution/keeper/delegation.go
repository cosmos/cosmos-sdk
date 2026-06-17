package keeper

import (
	"bytes"
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// initializeDelegation initializes starting info for a new delegation
func (k Keeper) initializeDelegation(ctx context.Context, val sdk.ValAddress, del sdk.AccAddress) error {
	// period has already been incremented - we want to store the period ended by this delegation action
	valCurrentRewards, err := k.GetValidatorCurrentRewards(ctx, val)
	if err != nil {
		return err
	}
	previousPeriod := valCurrentRewards.Period - 1

	// increment reference count for the period we're going to track
	err = k.incrementReferenceCount(ctx, val, previousPeriod)
	if err != nil {
		return err
	}

	validator, err := k.stakingKeeper.Validator(ctx, val)
	if err != nil {
		return err
	}

	delegation, err := k.stakingKeeper.Delegation(ctx, del, val)
	if err != nil {
		return err
	}

	// calculate delegation stake in tokens
	// we don't store directly, so multiply delegation shares * (tokens per share)
	// note: necessary to truncate so we don't allow withdrawing more rewards than owed
	stake := validator.TokensFromSharesTruncated(delegation.GetShares())
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return k.SetDelegatorStartingInfo(ctx, val, del, types.NewDelegatorStartingInfo(previousPeriod, stake, uint64(sdkCtx.BlockHeight())))
}

// calculateDelegationRewardsBetween calculates the rewards accrued by a delegation between two periods
func (k Keeper) calculateDelegationRewardsBetween(ctx context.Context, val stakingtypes.ValidatorI,
	startingPeriod, endingPeriod uint64, stake math.LegacyDec,
) (sdk.DecCoins, error) {
	// sanity check
	if startingPeriod > endingPeriod {
		panic("startingPeriod cannot be greater than endingPeriod")
	}

	// sanity check
	if stake.IsNegative() {
		panic("stake should not be negative")
	}

	valBz, err := k.stakingKeeper.ValidatorAddressCodec().StringToBytes(val.GetOperator())
	if err != nil {
		panic(err)
	}

	// return staking * (ending - starting)
	starting, err := k.getValidatorHistoricalRewards(ctx, valBz, startingPeriod, true)
	if err != nil {
		return sdk.DecCoins{}, err
	}

	ending, err := k.getValidatorHistoricalRewards(ctx, valBz, endingPeriod, true)
	if err != nil {
		return sdk.DecCoins{}, err
	}

	difference := ending.CumulativeRewardRatio.Sub(starting.CumulativeRewardRatio)
	if difference.IsAnyNegative() {
		panic("negative rewards should not be possible")
	}
	// note: necessary to truncate so we don't allow withdrawing more rewards than owed
	rewards := difference.MulDecTruncate(stake)
	return rewards, nil
}

// CalculateDelegationRewards calculates the total rewards accrued by a delegation
func (k Keeper) CalculateDelegationRewards(ctx context.Context, val stakingtypes.ValidatorI, del stakingtypes.DelegationI, endingPeriod uint64) (rewards sdk.DecCoins, err error) {
	addrCodec := k.authKeeper.AddressCodec()
	delAddr, err := addrCodec.StringToBytes(del.GetDelegatorAddr())
	if err != nil {
		return sdk.DecCoins{}, err
	}

	valAddr, err := k.stakingKeeper.ValidatorAddressCodec().StringToBytes(del.GetValidatorAddr())
	if err != nil {
		return sdk.DecCoins{}, err
	}

	// fetch starting info for delegation
	startingInfo, err := k.GetDelegatorStartingInfo(ctx, sdk.ValAddress(valAddr), sdk.AccAddress(delAddr))
	if err != nil {
		return sdk.DecCoins{}, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if startingInfo.Height == uint64(sdkCtx.BlockHeight()) {
		// started this height, no rewards yet
		return sdk.DecCoins{}, err
	}

	startingPeriod := startingInfo.PreviousPeriod
	stake := startingInfo.Stake

	// Iterate through slashes and withdraw with calculated staking for
	// distribution periods. These period offsets are dependent on *when* slashes
	// happen - namely, in BeginBlock, after rewards are allocated...
	// Slashes which happened in the first block would have been before this
	// delegation existed, UNLESS they were slashes of a redelegation to this
	// validator which was itself slashed (from a fault committed by the
	// redelegation source validator) earlier in the same BeginBlock.
	startingHeight := startingInfo.Height
	// Slashes this block happened after reward allocation, but we have to account
	// for them for the stake sanity check below.
	endingHeight := uint64(sdkCtx.BlockHeight())
	if endingHeight > startingHeight {
		k.IterateValidatorSlashEventsBetween(ctx, valAddr, startingHeight, endingHeight,
			func(height uint64, event types.ValidatorSlashEvent) (stop bool) {
				endingPeriod := event.ValidatorPeriod
				if endingPeriod > startingPeriod {
					delRewards, err := k.calculateDelegationRewardsBetween(ctx, val, startingPeriod, endingPeriod, stake)
					if err != nil {
						panic(err)
					}
					rewards = rewards.Add(delRewards...)

					// Note: It is necessary to truncate so we don't allow withdrawing
					// more rewards than owed.
					stake = stake.MulTruncate(math.LegacyOneDec().Sub(event.Fraction))
					startingPeriod = endingPeriod
				}
				return false
			},
		)
	}

	// A total stake sanity check; Recalculated final stake should be less than or
	// equal to current stake here. We cannot use Equals because stake is truncated
	// when multiplied by slash fractions (see above). We could only use equals if
	// we had arbitrary-precision rationals.
	currentStake := val.TokensFromShares(del.GetShares())

	if stake.GT(currentStake) {
		// Account for rounding inconsistencies between:
		//
		//     currentStake: calculated as in staking with a single computation
		//     stake:        calculated as an accumulation of stake
		//                   calculations across validator's distribution periods
		//
		// These inconsistencies are due to differing order of operations which
		// will inevitably have different accumulated rounding and may lead to
		// the smallest decimal place being one greater in stake than
		// currentStake. When we calculated slashing by period, even if we
		// round down for each slash fraction, it's possible due to how much is
		// being rounded that we slash less when slashing by period instead of
		// for when we slash without periods. In other words, the single slash,
		// and the slashing by period could both be rounding down but the
		// slashing by period is simply rounding down less, thus making stake >
		// currentStake
		//
		// A small amount of this error is tolerated and corrected for,
		// however any greater amount should be considered a breach in expected
		// behavior.
		marginOfErr := math.LegacySmallestDec().MulInt64(3)
		if stake.LTE(currentStake.Add(marginOfErr)) {
			stake = currentStake
		} else {
			panic(fmt.Sprintf("calculated final stake for delegator %s greater than current stake"+
				"\n\tfinal stake:\t%s"+
				"\n\tcurrent stake:\t%s",
				del.GetDelegatorAddr(), stake, currentStake))
		}
	}

	// calculate rewards for final period
	delRewards, err := k.calculateDelegationRewardsBetween(ctx, val, startingPeriod, endingPeriod, stake)
	if err != nil {
		return sdk.DecCoins{}, err
	}

	rewards = rewards.Add(delRewards...)
	return rewards, nil
}

// withdrawDestination is the resolved target for a reward withdrawal.
// String yields the value used for the withdraw_address event attribute.
type withdrawDestination interface {
	IsRedirected() bool
	SpecifiedWithdrawAddress() string
	ResolvedWithdrawAddress() string
}

// withdrawToAddr routes rewards to a specific account via the bank module.
type withdrawToAddr struct {
	ResolvedWithdrawAddr  sdk.AccAddress
	SpecifiedWithdrawAddr sdk.AccAddress
}

func newWithdrawToAddress(resolved, specified sdk.AccAddress) withdrawToAddr {
	return withdrawToAddr{
		ResolvedWithdrawAddr:  resolved,
		SpecifiedWithdrawAddr: specified,
	}
}

func (d withdrawToAddr) IsRedirected() bool {
	return !bytes.Equal(d.ResolvedWithdrawAddr, d.SpecifiedWithdrawAddr)
}

func (d withdrawToAddr) SpecifiedWithdrawAddress() string {
	return d.SpecifiedWithdrawAddr.String()
}

func (d withdrawToAddr) ResolvedWithdrawAddress() string {
	return d.ResolvedWithdrawAddr.String()
}

// withdrawToCommunityPool credits rewards to the community pool.
type withdrawToCommunityPool struct {
	SpecifiedWithdrawAddr sdk.AccAddress
}

func newWithdrawToCommunityPool(specified sdk.AccAddress) withdrawToCommunityPool {
	return withdrawToCommunityPool{SpecifiedWithdrawAddr: specified}
}

func (withdrawToCommunityPool) IsRedirected() bool {
	return true
}

func (d withdrawToCommunityPool) SpecifiedWithdrawAddress() string {
	return d.SpecifiedWithdrawAddr.String()
}

func (withdrawToCommunityPool) ResolvedWithdrawAddress() string {
	return types.AttributeValueCommunityPool
}

// resolveWithdrawDestination resolves a withdraw destination based on if the
// withdraw is 'strict' or not. a withdraw being strict or not determines if
// fallback addresses should be used or not. if a withdraw is strict, the
// withdraw must go to the owners specified withdraw address, and if it cannot
// for some reason, an error is returned. if a withdraw is not strict and it
// cannot go to the owner's specified withdraw address, it will fallback to the
// owner's address itself. if the owner's address cannot be used, it will
// fallback to the community pool.
func (k Keeper) resolveWithdrawDestination(
	ctx context.Context,
	owner sdk.AccAddress,
	strict bool,
) (withdrawDestination, error) {
	if strict {
		return k.resolveWithdrawDestinationStrict(ctx, owner)
	}
	return k.resolveWithdrawDestinationFallback(ctx, owner)
}

// resolveWithdrawDestinationStrict returns the destination for owner's stored
// withdraw address. If that address is in the bank module's blocked set, it
// returns ErrWithdrawAddrBlocked.
func (k Keeper) resolveWithdrawDestinationStrict(
	ctx context.Context,
	owner sdk.AccAddress,
) (withdrawToAddr, error) {
	withdrawAddr, err := k.GetDelegatorWithdrawAddr(ctx, owner)
	if err != nil {
		return withdrawToAddr{}, err
	}
	if k.bankKeeper.BlockedAddr(withdrawAddr) {
		// copying error return that the bank module uses when it tries to send
		// to a blocked address
		return withdrawToAddr{}, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive funds", withdrawAddr)
	}
	return newWithdrawToAddress(withdrawAddr, withdrawAddr), nil
}

// resolveWithdrawDestinationFallback returns the destination for owner's
// stored withdraw address. If that address is in the bank module's blocked
// set, it falls back to the owner's address itself. If the owner's address is
// in the bank module's blocked set, it falls back to the community pool.
func (k Keeper) resolveWithdrawDestinationFallback(
	ctx context.Context,
	owner sdk.AccAddress,
) (withdrawDestination, error) {
	withdrawAddr, err := k.GetDelegatorWithdrawAddr(ctx, owner)
	if err != nil {
		return nil, err
	}

	if !k.bankKeeper.BlockedAddr(withdrawAddr) {
		return newWithdrawToAddress(withdrawAddr, withdrawAddr), nil
	}
	if !k.bankKeeper.BlockedAddr(owner) {
		return newWithdrawToAddress(owner, withdrawAddr), nil
	}
	return newWithdrawToCommunityPool(withdrawAddr), nil
}

func emitWithdrawDestinationRedirectedEvent(ctx context.Context, dest withdrawDestination, validatorOp, delegatorAddr string) {
	attrs := []sdk.Attribute{
		sdk.NewAttribute(types.AttributeKeyOriginalWithdrawAddress, dest.SpecifiedWithdrawAddress()),
		sdk.NewAttribute(types.AttributeKeyWithdrawAddress, dest.ResolvedWithdrawAddress()),
		sdk.NewAttribute(types.AttributeKeyValidator, validatorOp),
	}
	if delegatorAddr != "" {
		attrs = append(attrs, sdk.NewAttribute(types.AttributeKeyDelegator, delegatorAddr))
	}
	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(types.EventTypeWithdrawAddrRedirected, attrs...))
}

// withdrawDelegationRewards withdrawals rewards to a given withdraw
// destination.
func (k Keeper) withdrawDelegationRewards(ctx context.Context, val stakingtypes.ValidatorI, del stakingtypes.DelegationI, dest withdrawDestination) (sdk.Coins, error) {
	addrCodec := k.authKeeper.AddressCodec()
	delAddr, err := addrCodec.StringToBytes(del.GetDelegatorAddr())
	if err != nil {
		return nil, err
	}

	valAddr, err := k.stakingKeeper.ValidatorAddressCodec().StringToBytes(del.GetValidatorAddr())
	if err != nil {
		return nil, err
	}

	// check existence of delegator starting info
	hasInfo, err := k.HasDelegatorStartingInfo(ctx, sdk.ValAddress(valAddr), sdk.AccAddress(delAddr))
	if err != nil {
		return nil, err
	}

	if !hasInfo {
		return nil, types.ErrEmptyDelegationDistInfo
	}

	// end current period and calculate rewards
	endingPeriod, err := k.IncrementValidatorPeriod(ctx, val)
	if err != nil {
		return nil, err
	}

	rewardsRaw, err := k.CalculateDelegationRewards(ctx, val, del, endingPeriod)
	if err != nil {
		return nil, err
	}

	outstanding, err := k.GetValidatorOutstandingRewardsCoins(ctx, sdk.ValAddress(valAddr))
	if err != nil {
		return nil, err
	}

	// defensive edge case may happen on the very final digits
	// of the decCoins due to operation order of the distribution mechanism.
	rewards := rewardsRaw.Intersect(outstanding)
	if !rewards.Equal(rewardsRaw) {
		logger := k.Logger(ctx)
		logger.Info(
			"rounding error withdrawing rewards from validator",
			"delegator", del.GetDelegatorAddr(),
			"validator", val.GetOperator(),
			"got", rewards.String(),
			"expected", rewardsRaw.String(),
		)
	}

	finalRewards, err := k.sendCoinsToDestination(ctx, rewards, dest)
	if err != nil {
		return nil, err
	}

	err = k.SetValidatorOutstandingRewards(ctx, sdk.ValAddress(valAddr), types.ValidatorOutstandingRewards{Rewards: outstanding.Sub(rewards)})
	if err != nil {
		return nil, err
	}

	// decrement reference count of starting period
	startingInfo, err := k.GetDelegatorStartingInfo(ctx, sdk.ValAddress(valAddr), sdk.AccAddress(delAddr))
	if err != nil {
		return nil, err
	}

	startingPeriod := startingInfo.PreviousPeriod
	err = k.decrementReferenceCount(ctx, sdk.ValAddress(valAddr), startingPeriod)
	if err != nil {
		return nil, err
	}

	// remove delegator starting info
	err = k.DeleteDelegatorStartingInfo(ctx, sdk.ValAddress(valAddr), sdk.AccAddress(delAddr))
	if err != nil {
		return nil, err
	}

	if finalRewards.IsZero() {
		baseDenom, _ := sdk.GetBaseDenom()
		if baseDenom == "" {
			baseDenom = sdk.DefaultBondDenom
		}

		// Note, we do not call the NewCoins constructor as we do not want the zero
		// coin removed.
		finalRewards = sdk.Coins{sdk.NewCoin(baseDenom, math.ZeroInt())}
	}

	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeWithdrawRewards,
			sdk.NewAttribute(sdk.AttributeKeyAmount, finalRewards.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, val.GetOperator()),
			sdk.NewAttribute(types.AttributeKeyDelegator, del.GetDelegatorAddr()),
		),
	)

	return finalRewards, nil
}

// sendCoinsToDestination sends the truncated coins to dest and routes the
// decimal remainder (and the truncated coins themselves, for the community-pool
// destination) into FeePool.CommunityPool.
func (k Keeper) sendCoinsToDestination(ctx context.Context, coins sdk.DecCoins, dest withdrawDestination) (sdk.Coins, error) {
	toSend, remainder := coins.TruncateDecimal()

	communityPoolFunds := remainder
	switch d := dest.(type) {
	case withdrawToAddr:
		if !toSend.IsZero() {
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, d.ResolvedWithdrawAddr, toSend); err != nil {
				return nil, err
			}
		}
	case withdrawToCommunityPool:
		communityPoolFunds = communityPoolFunds.Add(sdk.NewDecCoinsFromCoins(toSend...)...)
	default:
		return nil, fmt.Errorf("unknown withdraw destination type %T", dest)
	}

	if !communityPoolFunds.IsZero() {
		feePool, err := k.FeePool.Get(ctx)
		if err != nil {
			return nil, err
		}
		feePool.CommunityPool = feePool.CommunityPool.Add(communityPoolFunds...)
		if err := k.FeePool.Set(ctx, feePool); err != nil {
			return nil, err
		}
	}

	return toSend, nil
}
