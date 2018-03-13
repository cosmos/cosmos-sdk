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
		case SetCoolMsg:
			return handleSetCoolMsg(ctx, cm, msg)
		case CoolMsg:
			return handleCoolMsg(ctx, ck, cm, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized cool Msg type: %v", reflect.TypeOf(msg).Name())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle CoolMsg This is the engine of your module
func handleSetCoolMsg(ctx sdk.Context, cm Mapper, msg SetCoolMsg) sdk.Result {
	cm.SetCool(ctx, msg.Cool)
	return sdk.Result{}
}

// Handle CoolMsg This is the engine of your module
func handleCoolMsg(ctx sdk.Context, ck bank.CoinKeeper, cm Mapper, msg CoolMsg) sdk.Result {

	whatsCool := cm.GetCool(ctx)

	// set default if nothing is set
	//if whatsCool == "" {
	//cm.SetCool(ctx, "icecold")
	//}

	if msg.CoolerThanCool == whatsCool {

		bonusCoins := sdk.Coins{{whatsCool, 69}}
		_, err := ck.AddCoins(ctx, msg.Sender, bonusCoins)
		if err != nil {
			return err.Result()
		}
	}

	return sdk.Result{}
}
