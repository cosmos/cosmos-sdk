package coin

import (
	sdk "github.com/cosmos/cosmos-sdk"
)

func Decorator(ctx sdk.Context, store sdk.MultiStore, tx sdk.Tx, next sdk.Handler) sdk.Result {
	if msg, ok := tx.(CoinsMsg); ok {
		handleCoinsMsg(ctx, store, msg)
	} else {
		next(ctx, store, tx)
	}
}
