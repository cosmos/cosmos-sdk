package circuit

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewAnteHandler(k Keeper) sdk.AnteHandler {
	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, sdk.Result, bool) {
		for _, msg := range tx.GetMsgs() {
			if k.CheckMsgBreak(ctx, msg) {
				return ctx, sdk.ErrUnauthorized("msg circuit break").Result(), true
			}
		}
		return ctx, sdk.Result{}, false
	}
}
