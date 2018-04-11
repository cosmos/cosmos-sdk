package bank

import (
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc"
)

// Handle all "bank" type messages.
func NewHandler(ck CoinKeeper, ibck ibc.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case IBCSendMsg:
			return handleIBCSendMsg(ctx, ibck, ck, msg)
		case SendMsg:
			return handleSendMsg(ctx, ck, msg)
		case IssueMsg:
			return handleIssueMsg(ctx, ck, msg)
		case ibc.ReceiveMsg:
			return ibck.Receive(func(ctx sdk.Context, p ibc.Payload) sdk.Error {
				switch p := p.(type) {
				case SendPayload:
					return handleSendPayload(ctx, ck, p)
				default:
					errMsg := "Unrecognized ibc Payload type: " + reflect.TypeOf(p).Name()
					return sdk.ErrUnknownRequest(errMsg)
				}
			}, ctx, msg)
		case ibc.ReceiptMsg:
			return ibck.Receipt(func(ctx sdk.Context, r ibc.Receipt) sdk.Error {
				switch r := r.(type) {
				case SendFailReceipt:
					return handleSendFailReceipt(ctx, ck, r)
				}
			}, ctx, msg)
		default:
			errMsg := "Unrecognized bank Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle IBCSendMsg
func handleIBCSendMsg(ctx sdk.Context, ibck ibc.Keeper, ck CoinKeeper, msg IBCSendMsg) sdk.Result {
	p := msg.SendPayload
	_, err := ck.SubtractCoins(ctx, p.SrcAddr, p.Coins)
	if err != nil {
		return err.Result()
	}
	ibck.Send(ctx, p, msg.DestChain)
	return sdk.Result{}
}

// Handle SendMsg.
func handleSendMsg(ctx sdk.Context, ck CoinKeeper, msg SendMsg) sdk.Result {
	// NOTE: totalIn == totalOut should already have been checked

	err := ck.InputOutputCoins(ctx, msg.Inputs, msg.Outputs)
	if err != nil {
		return err.Result()
	}

	// TODO: add some tags so we can search it!
	return sdk.Result{} // TODO
}

// Handle IssueMsg.
func handleIssueMsg(ctx sdk.Context, ck CoinKeeper, msg IssueMsg) sdk.Result {
	panic("not implemented yet")
}

func handleSendPayload(ctx sdk.Context, ck CoinKeeper, p SendPayload) sdk.Error {
	_, err := ck.AddCoins(ctx, p.DestAddr, p.Coins)
	return err

}

func handleSendFailReceipt(ctx sdk.Context, ck CoinKeeper, r SendFailReceipt) sdk.Error {
	_, err := ck.AddCoins(ctx, r.SrcAddr, r.Coins)
	return err
}
