package vesting

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/core/address"
	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	vestingtypes "cosmossdk.io/x/accounts/vesting/types/v1"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Base Vesting Account

// NewBaseVestingAccount creates a new BaseVestingAccount object. It is the
// callers responsibility to ensure the base account has sufficient funds with
// regards to the original vesting amount.
func NewBaseVestingAccount(d accountstd.Dependencies) (*BaseVestingAccount, error) {
	baseVestingAccount := &BaseVestingAccount{
		OriginalVesting:  sdk.NewCoins(),
		DelegatedFree:    sdk.NewCoins(),
		DelegatedVesting: sdk.NewCoins(),
		AddressCodec:     d.AddressCodec,
		EndTime:          0,
	}

	return baseVestingAccount, nil
}

type BaseVestingAccount struct {
	OriginalVesting  sdk.Coins
	DelegatedFree    sdk.Coins
	DelegatedVesting sdk.Coins
	AddressCodec     address.Codec
	// Vesting end time, as unix timestamp (in seconds).
	EndTime int64
}

// --------------- execute -----------------

// LockedCoinsFromVesting returns all the coins that are not spendable (i.e. locked)
// for a vesting account given the current vesting coins. If no coins are locked,
// an empty slice of Coins is returned.
//
// CONTRACT: Delegated vesting coins and vestingCoins must be sorted.
func (bva BaseVestingAccount) LockedCoinsFromVesting(vestingCoins sdk.Coins) sdk.Coins {
	lockedCoins := vestingCoins.Sub(vestingCoins.Min(bva.DelegatedVesting)...)
	if lockedCoins == nil {
		return sdk.Coins{}
	}
	return lockedCoins
}

// TrackDelegation tracks a delegation amount for any given vesting account type
// given the amount of coins currently vesting and the current account balance
// of the delegation denominations.
//
// CONTRACT: The account's coins, delegation coins, vesting coins, and delegated
// vesting coins must be sorted.
func (bva *BaseVestingAccount) TrackDelegation(balance, vestingCoins, amount sdk.Coins) {
	for _, coin := range amount {
		baseAmt := balance.AmountOf(coin.Denom)
		vestingAmt := vestingCoins.AmountOf(coin.Denom)
		delVestingAmt := bva.DelegatedVesting.AmountOf(coin.Denom)

		// Panic if the delegation amount is zero or if the base coins does not
		// exceed the desired delegation amount.
		if coin.Amount.IsZero() || baseAmt.LT(coin.Amount) {
			panic("delegation attempt with zero coins or insufficient funds")
		}

		// compute x and y per the specification, where:
		// X := min(max(V - DV, 0), D)
		// Y := D - X
		x := math.MinInt(math.MaxInt(vestingAmt.Sub(delVestingAmt), math.ZeroInt()), coin.Amount)
		y := coin.Amount.Sub(x)

		if !x.IsZero() {
			xCoin := sdk.NewCoin(coin.Denom, x)
			bva.DelegatedVesting = bva.DelegatedVesting.Add(xCoin)
		}

		if !y.IsZero() {
			yCoin := sdk.NewCoin(coin.Denom, y)
			bva.DelegatedFree = bva.DelegatedFree.Add(yCoin)
		}
	}
}

// TrackUndelegation tracks an undelegation amount by setting the necessary
// values by which delegated vesting and delegated vesting need to decrease and
// by which amount the base coins need to increase.
//
// NOTE: The undelegation (bond refund) amount may exceed the delegated
// vesting (bond) amount due to the way undelegation truncates the bond refund,
// which can increase the validator's exchange rate (tokens/shares) slightly if
// the undelegated tokens are non-integral.
//
// CONTRACT: The account's coins and undelegation coins must be sorted.
func (bva *BaseVestingAccount) TrackUndelegation(amount sdk.Coins) {
	for _, coin := range amount {
		// panic if the undelegation amount is zero
		if coin.Amount.IsZero() {
			panic("undelegation attempt with zero coins")
		}
		delegatedFree := bva.DelegatedFree.AmountOf(coin.Denom)
		delegatedVesting := bva.DelegatedVesting.AmountOf(coin.Denom)

		// compute x and y per the specification, where:
		// X := min(DF, D)
		// Y := min(DV, D - X)
		x := math.MinInt(delegatedFree, coin.Amount)
		y := math.MinInt(delegatedVesting, coin.Amount.Sub(x))

		if !x.IsZero() {
			xCoin := sdk.NewCoin(coin.Denom, x)
			bva.DelegatedFree = bva.DelegatedFree.Sub(xCoin)
		}

		if !y.IsZero() {
			yCoin := sdk.NewCoin(coin.Denom, y)
			bva.DelegatedVesting = bva.DelegatedVesting.Sub(yCoin)
		}
	}
}

// --------------- Query -----------------

// QueryOriginalVesting returns a vesting account's original vesting amount
func (bva BaseVestingAccount) QueryOriginalVesting(ctx context.Context, _ *vestingtypes.QueryOriginalVestingRequest) (
	*vestingtypes.QueryOriginalVestingResponse, error,
) {
	return &vestingtypes.QueryOriginalVestingResponse{
		OriginalVesting: bva.OriginalVesting,
	}, nil
}

// QueryDelegatedFree returns a vesting account's delegation amount that is not
// vesting.
func (bva BaseVestingAccount) QueryDelegatedFree(ctx context.Context, _ *vestingtypes.QueryDelegatedFreeRequest) (
	*vestingtypes.QueryDelegatedFreeResponse, error,
) {
	return &vestingtypes.QueryDelegatedFreeResponse{
		DelegatedFree: bva.DelegatedFree,
	}, nil
}

// QueryDelegatedVesting returns a vesting account's delegation amount that is
// still vesting.
func (bva BaseVestingAccount) QueryDelegatedVesting(ctx context.Context, _ *vestingtypes.QueryDelegatedVestingRequest) (
	*vestingtypes.QueryDelegatedVestingResponse, error,
) {
	return &vestingtypes.QueryDelegatedVestingResponse{
		DelegatedVesting: bva.DelegatedVesting,
	}, nil
}

// QueryEndTime returns a vesting account's end time
func (bva BaseVestingAccount) QueryEndTime(ctx context.Context, _ *vestingtypes.QueryEndTimeRequest) (
	*vestingtypes.QueryEndTimeResponse, error,
) {
	return &vestingtypes.QueryEndTimeResponse{
		EndTime: bva.EndTime,
	}, nil
}

// Validate checks for errors on the account fields
func (bva BaseVestingAccount) Validate() error {
	if bva.EndTime < 0 {
		return errors.New("end time cannot be negative")
	}

	if !bva.OriginalVesting.IsValid() || !bva.OriginalVesting.IsAllPositive() {
		return fmt.Errorf("invalid coins: %s", bva.OriginalVesting.String())
	}

	if !(bva.DelegatedVesting.IsAllLTE(bva.OriginalVesting)) {
		return errors.New("delegated vesting amount cannot be greater than original vesting amount")
	}

	return nil
}
