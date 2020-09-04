package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
)

// ClientUpdateProposal will try to update the client with the new header if and only if
// the proposal passes. The localhost client is not allowed to be modified with a proposal.
func (k Keeper) ClientUpdateProposal(ctx sdk.Context, p *types.ClientUpdateProposal) error {
	clientType, found := k.GetClientType(ctx, p.ClientId)
	if !found {
		return sdkerrors.Wrapf(types.ErrClientTypeNotFound, "cannot update client with ID %s", p.ClientId)
	}

	if clientType == exported.Localhost {
		return sdkerrors.Wrap(types.ErrInvalidUpdateClientProposal, "cannot update localhost client with proposal")
	}

	clientState, found := k.GetClientState(ctx, p.ClientId)
	if !found {
		return sdkerrors.Wrapf(types.ErrClientNotFound, "cannot update client with ID %s", p.ClientId)
	}

	header, err := types.UnpackHeader(p.Header)
	if err != nil {
		return err
	}

	clientState, consensusState, err := clientState.CheckProposedHeaderAndUpdateState(ctx, k.cdc, k.ClientStore(ctx, p.ClientId), header)
	if err != nil {
		return err
	}

	k.SetClientState(ctx, p.ClientId, clientState)
	k.SetClientConsensusState(ctx, p.ClientId, header.GetHeight(), consensusState)

	k.Logger(ctx).Info("client updated after governance proposal passed", "client-id", p.ClientId, "height", clientState.GetLatestHeight())

	// emitting events in the keeper for proposal updates to clients
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUpdateClientProposal,
			sdk.NewAttribute(types.AttributeKeyClientID, p.ClientId),
			sdk.NewAttribute(types.AttributeKeyClientType, clientType.String()),
			sdk.NewAttribute(types.AttributeKeyConsensusHeight, fmt.Sprintf("%d", header.GetHeight())),
		),
	)

	return nil
}
