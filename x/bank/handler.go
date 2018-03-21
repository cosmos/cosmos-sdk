package bank

import (
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Handle all "bank" type messages.
func (ck CoinKeeper) Handler(ctx sdk.Context, msg sdk.Msg) sdk.Result {
	switch msg := msg.(type) {
	case SendMsg:
		return handleSendMsg(ctx, ck, msg)
	case IssueMsg:
		return handleIssueMsg(ctx, ck, msg)
	default:
		errMsg := "Unrecognized bank Msg type: " + reflect.TypeOf(msg).Name()
		return sdk.ErrUnknownRequest(errMsg).Result()
	}
}

// Handle SendMsg.
func handleSendMsg(ctx sdk.Context, ck CoinKeeper, msg SendMsg) sdk.Result {
	// NOTE: totalIn == totalOut should already have been checked

	for _, in := range msg.Inputs {
		_, err := ck.SubtractCoins(ctx, in.Address, in.Coins)
		if err != nil {
			return err.Result()
		}
	}

	for _, out := range msg.Outputs {
		_, err := ck.AddCoins(ctx, out.Address, out.Coins)
		if err != nil {
			return err.Result()
		}
	}

	// TODO: add some tags so we can search it!
	return sdk.Result{} // TODO
}

// Handle IssueMsg.
func handleIssueMsg(ctx sdk.Context, ck CoinKeeper, msg IssueMsg) sdk.Result {
	panic("not implemented yet")
}
