package msgstat

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/store"
)

// NewAnteHandler returns an AnteHandler that checks
// whether msg type is activate or not
func NewAnteHandler(store params.Store) sdk.AnteHandler {
	return func(ctx sdk.Context, tx sdk.Tx) (sdk.Context, sdk.Result, bool) {
		for _, msg := range tx.GetMsgs() {
			if !store.Has(ctx, params.NewKey(msg.Type())) {
				return ctx, sdk.ErrUnauthorized("deactivated msg type").Result(), true
			}
			var activated bool
			store.Get(ctx, params.NewKey(msg.Type()), &activated)
			if !activated {
				return ctx, sdk.ErrUnauthorized("deactivated msg type").Result(), true
			}
		}
		return ctx, sdk.Result{}, false
	}
}
