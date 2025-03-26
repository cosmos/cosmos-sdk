package types

import (
	"errors"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PercentageCoinMul multiplies each coin in an sdk.Coins struct by the given percentage and returns the new
// value.
//
// When performing multiplication, the resulting values are truncated to an sdk.Int.
func PercentageCoinMul(percentage math.LegacyDec, coins sdk.Coins) sdk.Coins {
	ret := sdk.NewCoins()

	for _, denom := range coins.Denoms() {
		am := sdk.NewCoin(denom, percentage.MulInt(coins.AmountOf(denom)).TruncateInt())
		ret = ret.Add(am)
	}

	return ret
}

func (cf *ContinuousFund) Validate() error {
	if cf.Recipient == "" {
		return errors.New("recipient cannot be empty")
	}

	// Validate percentage
	if cf.Percentage.IsNil() || cf.Percentage.IsZero() {
		return errors.New("percentage cannot be zero or empty")
	}
	if cf.Percentage.IsNegative() {
		return errors.New("percentage cannot be negative")
	}
	if cf.Percentage.GT(math.LegacyOneDec()) {
		return errors.New("percentage cannot be greater than one")
	}
	return nil
}
