package circuit

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/space"
)

// NewAnteHandler returns an AnteHandler that checks
// whether msg type is circuit brake or not
func NewAnteHandler(space params.Space) sdk.AnteHandler {
	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, sdk.Result, bool) {
		for _, msg := range tx.GetMsgs() {
			key := CircuitBrakeKey(msg.Type())
			var brake bool
			space.GetIfExists(ctx, key, &brake)
			if brake {
				return ctx, sdk.ErrUnauthorized("msg type circuit brake").Result(), true
			}
		}
		return ctx, sdk.Result{}, false
	}
}
