package bank

import (
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc"
)

// NewHandler returns a handler for "bank" type messages.
func NewHandler(k Keeper, ibcc ibc.Channel) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case IBCSendMsg:
			return handleIBCSendMsg(ctx, ibcc, k, msg)
		case MsgSend:
			return handleMsgSend(ctx, k, msg)
		case MsgIssue:
			return handleMsgIssue(ctx, k, msg)
		case ibc.ReceiveMsg:
			return ibcc.Receive(func(ctx sdk.Context, p ibc.Payload) (ibc.Payload, sdk.Error) {
				switch p := p.(type) {
				case SendPayload:
					return handleSendPayload(ctx, k, p)
				default:
					errMsg := "Unrecognized ibc Payload type: " + reflect.TypeOf(p).Name()
					return nil, sdk.ErrUnknownRequest(errMsg)
				}
			}, ctx, msg)
		case ibc.ReceiptMsg:
			return ibcc.Receipt(func(ctx sdk.Context, p ibc.Payload) {
				switch p := p.(type) {
				case SendFailReceipt:
					handleSendFailReceipt(ctx, k, p)
				}
			}, ctx, msg)
		default:
			errMsg := "Unrecognized bank Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle IBCSendMsg
func handleIBCSendMsg(ctx sdk.Context, ibcc ibc.Channel, k Keeper, msg IBCSendMsg) sdk.Result {
	p := msg.SendPayload
	_, err := k.SubtractCoins(ctx, p.SrcAddr, p.Coins)
	if err != nil {
		return err.Result()
	}
	ibcc.Send(ctx, p, msg.DestChain)
	return sdk.Result{}
}

// Handle MsgSend.
func handleMsgSend(ctx sdk.Context, k Keeper, msg MsgSend) sdk.Result {

	// NOTE: totalIn == totalOut should already have been checked

	err := k.InputOutputCoins(ctx, msg.Inputs, msg.Outputs)
	if err != nil {
		return err.Result()
	}

	// TODO: add some tags so we can search it!
	return sdk.Result{} // TODO
}

// Handle MsgIssue.
func handleMsgIssue(ctx sdk.Context, k Keeper, msg MsgIssue) sdk.Result {
	panic("not implemented yet")
}

func handleSendPayload(ctx sdk.Context, k Keeper, p SendPayload) (ibc.Payload, sdk.Error) {
	_, err := k.AddCoins(ctx, p.DestAddr, p.Coins)
	if err != nil {
		return SendFailReceipt{p}, err
	}

	return nil, nil
}

func handleSendFailReceipt(ctx sdk.Context, k Keeper, r SendFailReceipt) {
	_, err := k.AddCoins(ctx, r.SrcAddr, r.Coins)
	if err != nil {
		panic(err)
	}
}
