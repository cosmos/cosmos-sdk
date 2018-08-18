package cool

import (
	"fmt"
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// This is just an example to demonstrate a functional custom module
// with full feature set functionality.
//
//  /$$$$$$$  /$$$$$$   /$$$$$$  /$$
// /$$_____/ /$$__  $$ /$$__  $$| $$
//| $$      | $$  \ $$| $$  \ $$| $$
//| $$      | $$  | $$| $$  | $$| $$
//|  $$$$$$$|  $$$$$$/|  $$$$$$/| $$$$$$$
// \_______/ \______/  \______/ |______/

// NewHandler returns a handler for "cool" type messages.
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgSetTrend:
			return handleMsgSetTrend(ctx, k, msg)
		case MsgQuiz:
			return handleMsgQuiz(ctx, k, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized cool Msg type: %v", reflect.TypeOf(msg).Name())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle MsgQuiz This is the engine of your module
func handleMsgSetTrend(ctx sdk.Context, k Keeper, msg MsgSetTrend) sdk.Result {
	k.setTrend(ctx, msg.Cool)
	return sdk.Result{}
}

// Handle MsgQuiz This is the engine of your module
func handleMsgQuiz(ctx sdk.Context, k Keeper, msg MsgQuiz) sdk.Result {

	correct := k.CheckTrend(ctx, msg.CoolAnswer)

	if !correct {
		return ErrIncorrectCoolAnswer(k.codespace, msg.CoolAnswer).Result()
	}

	if ctx.IsCheckTx() {
		return sdk.Result{} // TODO
	}

	bonusCoins := sdk.Coins{sdk.NewCoin(msg.CoolAnswer, 69)}

	_, _, err := k.ck.AddCoins(ctx, msg.Sender, bonusCoins)
	if err != nil {
		return err.Result()
	}

	return sdk.Result{}
}
