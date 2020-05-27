package evidence

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/evidence/exported"
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case exported.MsgSubmitEvidence:
			return handleMsgSubmitEvidence(ctx, k, msg)

		default:

			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", ModuleName, msg)
		}
	}
}

func handleMsgSubmitEvidence(ctx sdk.Context, k Keeper, msg exported.MsgSubmitEvidence) (*sdk.Result, error) {
	evidence := msg.GetEvidence()
	if err := k.SubmitEvidence(ctx, evidence); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.GetSubmitter().String()),
		),
	)

	return &sdk.Result{
		Data:   evidence.Hash(),
		Events: ctx.EventManager().ABCIEvents(),
	}, nil
}
