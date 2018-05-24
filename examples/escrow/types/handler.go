package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"reflect"
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case CreateCovenantMessage:
			return handleMsgCreate(ctx, k, msg)
		case SettleCovenantMessage:
			return handleMsgSettle(ctx, k, msg)
		default:
			errMsg := "Unrecognized escrow Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgCreate(ctx sdk.Context, keeper Keeper, msg CreateCovenantMessage) (int64, bool) {
	return keeper.createCovenant(ctx, msg.Sender, msg.Settlers, msg.Receivers, msg.Amount)
}

func handleMsgSettle(ctx sdk.Context, keeper Keeper, msg SettleCovenantMessage) bool {
	return keeper.settleCovenant(ctx, msg.CovID, msg.Settler, msg.Receiver)
}
