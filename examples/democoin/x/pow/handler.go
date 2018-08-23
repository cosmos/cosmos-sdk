package pow

import (
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// POW handler
func (pk Keeper) Handler(ctx sdk.Context, msg sdk.Msg) sdk.Result {
	switch msg := msg.(type) {
	case MsgMine:
		return handleMsgMine(ctx, pk, msg)
	default:
		errMsg := "Unrecognized pow Msg type: " + reflect.TypeOf(msg).Name()
		return sdk.ErrUnknownRequest(errMsg).Result()
	}
}

func handleMsgMine(ctx sdk.Context, pk Keeper, msg MsgMine) sdk.Result {

	// precondition: msg has passed ValidateBasic

	newDiff, newCount, err := pk.CheckValid(ctx, msg.Difficulty, msg.Count)
	if err != nil {
		return err.Result()
	}

	// commented for now, makes testing difficult
	// TODO figure out a better test method that allows early CheckTx return
	/*
		if ctx.IsCheckTx() {
			return sdk.Result{} // TODO
		}
	*/

	err = pk.ApplyValid(ctx, msg.Sender, newDiff, newCount)
	if err != nil {
		return err.Result()
	}

	return sdk.Result{}
}
