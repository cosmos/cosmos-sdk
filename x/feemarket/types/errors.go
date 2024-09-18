package types

import (
	sdkerrors "cosmossdk.io/errors"
)

var (
	ErrNoFeeCoins      = sdkerrors.New(ModuleName, 1, "no fee coin provided. Must provide one.")
	ErrTooManyFeeCoins = sdkerrors.New(ModuleName, 2, "too many fee coins provided.  Only one fee coin may be provided")
	ErrResolverNotSet  = sdkerrors.New(ModuleName, 3, "denom resolver interface not set.  Only the feemarket base fee denomination can be used")
)
