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

// Handle all "coolmodule" type objects
func (k Keeper) Handler(ctx sdk.Context, msg sdk.Msg) sdk.Result {
	switch msg := msg.(type) {
	case SetTrendMsg:
		return handleSetTrendMsg(ctx, k, msg)
	case QuizMsg:
		return handleQuizMsg(ctx, k, msg)
	default:
		errMsg := fmt.Sprintf("Unrecognized cool Msg type: %v", reflect.TypeOf(msg).Name())
		return sdk.ErrUnknownRequest(errMsg).Result()
	}
}

// Handle QuizMsg This is the engine of your module
func handleSetTrendMsg(ctx sdk.Context, k Keeper, msg SetTrendMsg) sdk.Result {
	k.setTrend(ctx, msg.Cool)
	return sdk.Result{}
}

// Handle QuizMsg This is the engine of your module
func handleQuizMsg(ctx sdk.Context, k Keeper, msg QuizMsg) sdk.Result {

	correct := k.CheckTrend(ctx, msg.CoolAnswer)

	if !correct {
		return ErrIncorrectCoolAnswer(msg.CoolAnswer).Result()
	}

	if ctx.IsCheckTx() {
		return sdk.Result{} // TODO
	}

	bonusCoins := sdk.Coins{{msg.CoolAnswer, 69}}

	_, err := k.ck.AddCoins(ctx, msg.Sender, bonusCoins)
	if err != nil {
		return err.Result()
	}

	return sdk.Result{}
}
