package circuit

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

func NewAnteHandler(space params.Subspace) sdk.AnteHandler {
	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, sdk.Result, bool) {
		for _, msg := range tx.GetMsgs() {
			var breaker bool
			space.GetWithSubkeyIfExists(ctx, MsgTypeKey, []byte(msg.Type()), &breaker)
			if breaker {
				return ctx, sdk.ErrUnauthorized("msg type circuit breaked").Result(), true
			}
		}
		return ctx, sdk.Result{}, false
	}
}
