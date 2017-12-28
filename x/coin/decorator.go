package coin

import (
	sdk "github.com/cosmos/cosmos-sdk"
)

func Decorator(ctx sdk.Context, store sdk.MultiStore, tx sdk.Tx, next sdk.Handler) sdk.Result {
	if msg, ok := tx.(CoinMsg); ok {
		return handleCoinsMsg(ctx, store, msg)
	} else {
		return next(ctx, store, tx)
	}
}

func handleCoinsMsg(ctx sdk.Context, store sdk.MultiStore, tx sdk.Tx) sdk.Result {
	panic("not implemented yet") // XXX
}
