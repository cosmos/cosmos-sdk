package vesting

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func validateAmount(amount sdk.Coins) error {
	if !amount.IsValid() {
		return sdkerrors.ErrInvalidCoins.Wrap(amount.String())
	}

	if !amount.IsAllPositive() {
		return sdkerrors.ErrInvalidCoins.Wrap(amount.String())
	}

	return nil
}
