package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/evidence/types"
)

var _ types.MsgServer = Keeper{}

// SubmitEvidence implements the MsgServer.SubmitEvidence method.
func (k Keeper) SubmitEvidence(goCtx context.Context, msg *types.MsgSubmitEvidence) (*types.MsgSubmitEvidenceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	evidence := msg.GetEvidence()
	if err := k.SubmitEvidenceI(ctx, evidence); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.GetSubmitter().String()),
		),
	)

	return &types.MsgSubmitEvidenceResponse{}, nil
}
