package keeper

import (
	"github.com/armon/go-metrics"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
)

// ClientUpdateProposal will try to update the client with the new header if and only if
// the proposal passes. The localhost client is not allowed to be modified with a proposal.
func (k Keeper) ClientUpdateProposal(ctx sdk.Context, p *types.ClientUpdateProposal) error {
	if p.ClientId == exported.Localhost {
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

	k.Logger(ctx).Info("client updated after governance proposal passed", "client-id", p.ClientId, "height", clientState.GetLatestHeight().String())

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{"ibc", "client", "update"},
			1,
			[]metrics.Label{
				telemetry.NewLabel(types.LabelClientType, clientState.ClientType()),
				telemetry.NewLabel(types.LabelClientID, p.ClientId),
				telemetry.NewLabel(types.LabelUpdateType, "proposal"),
			},
		)
	}()

	// emitting events in the keeper for proposal updates to clients
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUpdateClientProposal,
			sdk.NewAttribute(types.AttributeKeyClientID, p.ClientId),
			sdk.NewAttribute(types.AttributeKeyClientType, clientState.ClientType()),
			sdk.NewAttribute(types.AttributeKeyConsensusHeight, header.GetHeight().String()),
		),
	)

	return nil
}
