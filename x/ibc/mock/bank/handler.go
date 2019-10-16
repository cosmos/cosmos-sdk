package mockbank

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgTransfer:
			return handleMsgTransfer(ctx, k, msg)
		default:
			return sdk.ErrUnknownRequest("failed to parse message").Result()
		}
	}
}

func handleMsgTransfer(ctx sdk.Context, k Keeper, msg MsgTransfer) (res sdk.Result) {
	if msg.Proof != nil {
		// send packet
		err := k.SendTransfer(ctx, msg.SrcPort, msg.SrcChannel, msg.Amount, msg.Sender, msg.Receiver, msg.Source, msg.Timeout)
		if err != nil {
			return err.Result()
		}
	} else {
		// receive packet
		err := k.ReceiveTransfer(ctx, msg.SrcPort, msg.SrcChannel, msg.Amount, msg.Sender, msg.Receiver, msg.Source, msg.Timeout, msg.Proof, msg.ProofHeight)
		if err != nil {
			return err.Result()
		}
	}

	return sdk.Result{Events: ctx.EventManager().Events()}
}
