package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/delegation/exported"
	abci "github.com/tendermint/tendermint/abci/types"
)

// BasicFeeAllowance implements FeeAllowance with a one-time grant of tokens
// that optionally expires. The delegatee can use up to SpendLimit to cover fees.
type BasicFeeAllowance struct {
	// SpendLimit is the maximum amount of tokens to be spent
	SpendLimit sdk.Coins

	// TODO: make this time or height
	// Expiration specifies an optional time when this allowance expires
	// is Expiration.IsZero() then it never expires
	Expiration time.Time
}

var _ exported.FeeAllowance = (*BasicFeeAllowance)(nil)

// Accept implements FeeAllowance and deducts the fees from the SpendLimit if possible
func (a *BasicFeeAllowance) Accept(fee sdk.Coins, block abci.Header) (remove bool, err error) {
	// TODO: handle expiry

	left, invalid := a.SpendLimit.SafeSub(fee)
	if invalid {
		return false, ErrFeeLimitExceeded()
	}
	if left.IsZero() {
		return true, nil
	}
	a.SpendLimit = left
	return false, nil
}
