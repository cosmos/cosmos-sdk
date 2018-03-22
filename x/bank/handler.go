package bank

import (
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibc "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// Handle all "bank" type messages.
func NewHandler(ck CoinKeeper, ibcs ibc.Sender) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case IBCSendMsg:
			return handleIBCSendMsg(ctx, ibcs, ck, msg)
		case SendMsg:
			return handleSendMsg(ctx, ck, msg)
		case IssueMsg:
			return handleIssueMsg(ctx, ck, msg)
		default:
			errMsg := "Unrecognized bank Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle IBCSendMsg
func handleIBCSendMsg(ctx sdk.Context, ibcs ibc.Sender, ck CoinKeeper, msg IBCSendMsg) sdk.Result {
	p := msg.SendPayload
	_, err := ck.SubtractCoins(ctx, p.SrcAddr, p.Coins)
	if err != nil {
		return err.Result()
	}
	ibcs.Push(ctx, p, msg.DestChain)
	return sdk.Result{}
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

// Handle all "bank" type IBC payloads

func NewIBCHandler(ck CoinKeeper) ibc.Handler {
	return func(ctx sdk.Context, p ibc.Payload) sdk.Result {
		switch p := p.(type) {
		case SendPayload:
			return handleTransferMsg(ctx, ck, p)
		default:
			errMsg := "Unrecognized bank Payload type: " + reflect.TypeOf(p).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleTransferMsg(ctx sdk.Context, ck CoinKeeper, p SendPayload) sdk.Result {
	_, err := ck.AddCoins(ctx, p.DestAddr, p.Coins)
	return err.Result()

}
