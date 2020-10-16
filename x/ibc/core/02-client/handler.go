package client

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/keeper"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
)

// HandleMsgUpgradeClient defines the sdk.Handler for MsgUpgradeClient
func HandleMsgUpgradeClient(ctx sdk.Context, k keeper.Keeper, msg *types.MsgUpgradeClient) (*sdk.Result, error) {
	upgradedClient, err := types.UnpackClientState(msg.ClientState)
	if err != nil {
		return nil, err
	}

	if err := upgradedClient.Validate(); err != nil {
		return nil, err
	}

	if err = k.UpgradeClient(ctx, msg.ClientId, upgradedClient, msg.UpgradeHeight, msg.ProofUpgrade); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	)

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, nil
}

// NewClientUpdateProposalHandler defines the client update proposal handler
func NewClientUpdateProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.ClientUpdateProposal:
			return k.ClientUpdateProposal(ctx, c)

		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized ibc proposal content type: %T", c)
		}
	}
}
