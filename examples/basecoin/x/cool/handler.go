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
		case SetWhatCoolMsg:
			return handleSetWhatCoolMsg(ctx, cm, msg)
		case WhatCoolMsg:
			return handleWhatCoolMsg(ctx, ck, cm, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized cool Msg type: %v", reflect.TypeOf(msg).Name())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle WhatCoolMsg This is the engine of your module
func handleSetWhatCoolMsg(ctx sdk.Context, cm Mapper, msg SetWhatCoolMsg) sdk.Result {
	cm.SetWhatCool(ctx, msg.WhatCool)
	return sdk.Result{}
}

// Handle WhatCoolMsg This is the engine of your module
func handleWhatCoolMsg(ctx sdk.Context, ck bank.CoinKeeper, cm Mapper, msg WhatCoolMsg) sdk.Result {

	whatsCool := cm.GetWhatCool(ctx)

	if msg.CoolerThanCool == whatsCool {

		bonusCoins := sdk.Coins{{whatsCool, 69}}
		_, err := ck.AddCoins(ctx, msg.Sender, bonusCoins)
		if err != nil {
			return err.Result()
		}
	}

	return sdk.Result{}
}
