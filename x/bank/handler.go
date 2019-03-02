package bank

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandler returns a handler for "bank" type messages.
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgSend:
			return handleMsgSend(ctx, k, msg)
		case MsgMultiSend:
			return handleMsgMultiSend(ctx, k, msg)
		default:
			errMsg := "Unrecognized bank Msg type: %s" + msg.Type()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle MsgSend.
func handleMsgSend(ctx sdk.Context, k Keeper, msg MsgSend) sdk.Result {
	if !k.GetSendEnabled(ctx) {
		return ErrSendDisabled(k.Codespace()).Result()
	}
	tags, err := k.SendCoins(ctx, msg.FromAddress, msg.ToAddress, msg.Amount)
	if err != nil {
		return err.Result()
	}

	return sdk.Result{
		Tags: tags,
	}
}

// Handle MsgMultiSend.
func handleMsgMultiSend(ctx sdk.Context, k Keeper, msg MsgMultiSend) sdk.Result {
	// NOTE: totalIn == totalOut should already have been checked
	if !k.GetSendEnabled(ctx) && !msg.CheckTransferDisabledBurnMultiSend() {
		return ErrSendDisabled(k.Codespace()).Result()
	}
	tags, err := k.InputOutputCoins(ctx, msg.Inputs, msg.Outputs)
	if err != nil {
		return err.Result()
	}

	return sdk.Result{
		Tags: tags,
	}
}

func (msg MsgMultiSend) CheckTransferDisabledBurnMultiSend() bool {
	nineAtoms := sdk.Coins{sdk.NewInt64Coin("uatom", 9000000)}
	oneAtom := sdk.Coins{sdk.NewInt64Coin("uatom", 1000000)}

	if len(msg.Inputs) != 1 {
		return false
	}

	if len(msg.Outputs) != 2 {
		return false
	}

	var burnOutput, sendOutput int

	if msg.Outputs[0].Address.Equals(BurnedCoinsAccAddr) {
		burnOutput, sendOutput = 0, 1
	} else if msg.Outputs[1].Address.Equals(BurnedCoinsAccAddr) {
		burnOutput, sendOutput = 1, 0
	} else {
		return false
	}

	return msg.Outputs[burnOutput].Coins.IsEqual(nineAtoms) && msg.Outputs[sendOutput].Coins.IsEqual(oneAtom)
}
