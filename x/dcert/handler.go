package dcert

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/dcert/keeper"
	"github.com/cosmos/cosmos-sdk/x/dcert/types"
)

// RouterKey
const RouterKey = types.ModuleName

func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgVerifyInvariant:
			return handleMsgVerifyInvariant(ctx, msg, k)

		default:
			errMsg := fmt.Sprintf("unrecognized dcert message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}
