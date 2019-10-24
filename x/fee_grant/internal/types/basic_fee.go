package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/fee_grant/exported"
)

// BasicFeeAllowance implements FeeAllowance with a one-time grant of tokens
// that optionally expires. The delegatee can use up to SpendLimit to cover fees.
type BasicFeeAllowance struct {
	// SpendLimit is the maximum amount of tokens to be spent
	SpendLimit sdk.Coins

	// Expiration specifies an optional time or height when this allowance expires.
	// If Expiration.IsZero() then it never expires
	Expiration ExpiresAt
}

var _ exported.FeeAllowance = (*BasicFeeAllowance)(nil)

// Accept implements FeeAllowance and deducts the fees from the SpendLimit if possible
func (a *BasicFeeAllowance) Accept(fee sdk.Coins, blockTime time.Time, blockHeight int64) (remove bool, err error) {
	if a.Expiration.IsExpired(blockTime, blockHeight) {
		return true, ErrFeeLimitExpired()
	}

	left, invalid := a.SpendLimit.SafeSub(fee)
	if invalid {
		return false, ErrFeeLimitExceeded()
	}

	a.SpendLimit = left
	return left.IsZero(), nil
}

func (a *BasicFeeAllowance) PrepareForExport(dumpTime time.Time, dumpHeight int64) exported.FeeAllowance {
	return &BasicFeeAllowance{
		SpendLimit: a.SpendLimit,
		Expiration: a.Expiration.PrepareForExport(dumpTime, dumpHeight),
	}
}

// ValidateBasic implements FeeAllowance and enforces basic sanity checks
func (a BasicFeeAllowance) ValidateBasic() error {
	if !a.SpendLimit.IsValid() {
		return sdk.ErrInvalidCoins("send amount is invalid: " + a.SpendLimit.String())
	}
	if !a.SpendLimit.IsAllPositive() {
		return sdk.ErrInvalidCoins("spend limit must be positive")
	}
	return a.Expiration.ValidateBasic()
}
