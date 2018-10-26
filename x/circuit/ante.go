package circuit

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewAnteHandler(k Keeper) sdk.AnteHandler {
	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, sdk.Result, bool) {
		for _, msg := range tx.GetMsgs() {
			var breaker bool
			k.space.GetWithSubkeyIfExists(ctx, MsgRouteKey, []byte(msg.Route()), &breaker)
			if breaker {
				return ctx, sdk.ErrUnauthorized("msg route circuit breaked").Result(), true
			}
			k.space.GetWithSubkeyIfExists(ctx, MsgTypeKey, []byte(msg.Type()), &breaker)
			if breaker {
				return ctx, sdk.ErrUnauthorized("msg type circuit breaked").Result(), true
			}
		}
		return ctx, sdk.Result{}, false
	}
}
