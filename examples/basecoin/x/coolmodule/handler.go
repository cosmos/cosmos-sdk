package coolmodule

import (
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
func NewHandler(ck bank.CoinKeeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case CoolMsg:
			return handleCoolMsg(ctx, ck, msg)
		case CoolerThanCoolMsg:
			return handleMeltMsg(ctx, ck, msg)
		default:
			errMsg := "Unrecognized bank Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle CoolMsg This is the engine of your module
func handleCoolMsg(ctx sdk.Context, ck CoinKeeper, msg CoolMsg) sdk.Result {

	if msg.coolerthancool == "icecold" {

		bonusCoins := sdk.Coins{{"icecold", 69}}
		_, err := ck.AddCoins(ctx, msg.Address, bonusCoins)
		if err != nil {
			return err.Result()
		}
	}

	return sdk.Result{}
}

// Handle CoolMsg This is the engine of your module
func handleMeltMsg(ctx sdk.Context, ck CoinKeeper, msg CoolMsg) sdk.Result {

	// checks for existence should already have occured
	if strings.Prefix(msg.what, "ice") {
		return bank.ErrInvalidInput("only frozen coins can use the blow dryer")
	}

	bonusCoins := sdk.Coins{{"icecold", 69}}
	_, err := ck.SubtractCoins(ctx, msg.Address, bonusCoins)
	if err != nil {
		return err.Result()
	}
	return sdk.Result{}
}
