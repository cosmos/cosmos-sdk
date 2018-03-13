package cool

import (
	"fmt"
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
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
func NewHandler(ck bank.CoinKeeper, cm Mapper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case SetTrendMsg:
			return handleSetTrendMsg(ctx, cm, msg)
		case QuizMsg:
			return handleQuizMsg(ctx, ck, cm, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized cool Msg type: %v", reflect.TypeOf(msg).Name())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle QuizMsg This is the engine of your module
func handleSetTrendMsg(ctx sdk.Context, cm Mapper, msg SetTrendMsg) sdk.Result {
	cm.SetTrend(ctx, msg.Cool)
	return sdk.Result{}
}

// Handle QuizMsg This is the engine of your module
func handleQuizMsg(ctx sdk.Context, ck bank.CoinKeeper, cm Mapper, msg QuizMsg) sdk.Result {

	whatsCool := cm.GetCool(ctx)

	// set default if nothing is set
	//if whatsCool == "" {
	//cm.SetTrend(ctx, "icecold")
	//}

	if msg.CoolAnswer == whatsCool {

		bonusCoins := sdk.Coins{{whatsCool, 69}}
		_, err := ck.AddCoins(ctx, msg.Sender, bonusCoins)
		if err != nil {
			return err.Result()
		}
	}

	return sdk.Result{}
}
