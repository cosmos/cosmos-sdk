package transfer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ics04 "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
)

func handleMsgTransfer(ctx sdk.Context, k Keeper, msg MsgTransfer) (res sdk.Result) {
	err := k.SendTransfer(ctx, msg.SrcPort, msg.SrcChannel, msg.Denomination, msg.Amount, msg.Sender, msg.Receiver, msg.Source)
	if err != nil {
		return err.Result()
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, ics04.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		))

	return sdk.Result{Events: ctx.EventManager().Events()}
}
