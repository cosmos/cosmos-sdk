package delegate

import (
	"fmt"
	cosmos "github.com/cosmos/cosmos-sdk/types"
)

func NewHandler(k Dispatcher) cosmos.Handler {
	return func(ctx cosmos.Context, msg cosmos.Msg) cosmos.Result {
		switch msg := msg.(type) {
		case MsgDelegatedAction:
			return k.DispatchAction(ctx, msg.Actor, msg.Action)
		default:
			errMsg := fmt.Sprintf("Unrecognized data Msg type: %v", msg.Type())
			return cosmos.ErrUnknownRequest(errMsg).Result()
		}
	}
}
