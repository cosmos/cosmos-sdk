package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandler defines the IBC handler
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		// TODO:
		return sdk.Result{}
	}
}
