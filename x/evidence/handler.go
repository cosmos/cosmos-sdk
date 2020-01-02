package evidence

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case MsgSubmitEvidence:
			return handleMsgSubmitEvidence(ctx, k, msg)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", ModuleName, msg)
		}
	}
}

func handleMsgSubmitEvidence(ctx sdk.Context, k Keeper, msg MsgSubmitEvidence) (*sdk.Result, error) {
	if err := k.SubmitEvidence(ctx, msg.Evidence); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Submitter.String()),
		),
	)

	return &sdk.Result{
		Data:   msg.Evidence.Hash(),
		Events: ctx.EventManager().Events(),
	}, nil
}
