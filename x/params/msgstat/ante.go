package msgstat

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/space"
)

// NewAnteHandler returns an AnteHandler that checks
// whether msg type is activate or not
func NewAnteHandler(space params.Space) sdk.AnteHandler {
	return func(ctx sdk.Context, tx sdk.Tx) (sdk.Context, sdk.Result, bool) {
		for _, msg := range tx.GetMsgs() {
			key := params.NewKey(msg.Type())
			if !space.Has(ctx, key) {
				return ctx, sdk.ErrUnauthorized("deactivated msg type").Result(), true
			}
			var activated bool
			space.Get(ctx, key, &activated)
			if !activated {
				return ctx, sdk.ErrUnauthorized("deactivated msg type").Result(), true
			}
		}
		return ctx, sdk.Result{}, false
	}
}
