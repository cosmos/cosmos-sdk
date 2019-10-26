package transfer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
)

// HandleMsgTransfer defines the sdk.Handler for MsgTransfer
func HandleMsgTransfer(ctx sdk.Context, k Keeper, msg MsgTransfer) (res sdk.Result) {
	err := k.SendTransfer(ctx, msg.SrcPort, msg.SrcChannel, msg.Denomination, msg.Amount, msg.Sender, msg.Receiver, msg.Source)
	if err != nil {
		return err.Result()
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeKeyReceiver, msg.Receiver),
			sdk.NewAttribute(types.AttributeKeyAmount, sdk.NewCoin(msg.Denomination, msg.Amount).String()),
		))

	return sdk.Result{Events: ctx.EventManager().Events()}
}
