package circuit

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

func NewAnteHandler(space params.Subspace) sdk.AnteHandler {
	space = space.WithTypeTable(ParamTypeTable())
	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, sdk.Result, bool) {
		for _, msg := range tx.GetMsgs() {
			var breaker bool
			space.GetWithSubkeyIfExists(ctx, MsgTypeKey, []byte(msg.Type()), &breaker)
			if breaker {
				return ctx, sdk.ErrUnauthorized("msg type circuit breaked").Result(), true
			}
			space.GetWithSubkeyIfExists(ctx, MsgNameKey, []byte(msg.Name()), &breaker)
			if breaker {
				return ctx, sdk.ErrUnauthorized("msg name circuit breaked").Result(), true
			}
		}
		return ctx, sdk.Result{}, false
	}
}
