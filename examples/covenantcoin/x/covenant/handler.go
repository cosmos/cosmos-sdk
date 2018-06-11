package covenant

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"reflect"
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgCreateCovenant:
			return handleMsgCreate(ctx, k, msg)
		case MsgSettleCovenant:
			return handleMsgSettle(ctx, k, msg)
		default:
			errMsg := "Unrecognized Escrow Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgCreate(ctx sdk.Context, keeper Keeper, msg MsgCreateCovenant) sdk.Result {
	id, err := keeper.createCovenant(ctx, msg.Sender, msg.Settlers, msg.Receivers, msg.Amount)
	if err != nil {
		return err.Result()
	}
	d, _ := keeper.cdc.MarshalBinary(id)
	return sdk.Result{
		Data: d,
	}
}

func handleMsgSettle(ctx sdk.Context, keeper Keeper, msg MsgSettleCovenant) sdk.Result {
	err := keeper.settleCovenant(ctx, msg.CovID, msg.Settler, msg.Receiver)
	if err != nil {
		return err.Result()
	}
	return sdk.Result{}
}
