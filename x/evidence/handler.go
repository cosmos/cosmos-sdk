package evidence

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case MsgSubmitEvidence:
			return handleMsgSubmitEvidence(ctx, k, msg)

		default:
			return sdk.ErrUnknownRequest(fmt.Sprintf("unrecognized %s message type: %T", ModuleName, msg)).Result()
		}
	}
}

func handleMsgSubmitEvidence(ctx sdk.Context, k Keeper, msg MsgSubmitEvidence) sdk.Result {
	if err := k.SubmitEvidence(ctx, msg.Evidence); err != nil {
		return sdk.ConvertError(err).Result()
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Submitter.String()),
		),
	)

	return sdk.Result{
		Data:   msg.Evidence.Hash(),
		Events: ctx.EventManager().Events(),
	}
}
