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
// intitial height must be after latest height of subject client

// TODO:
// light clients handle copying and verification of the substitute
// don't allow substitutes which have a different revision number
func (k Keeper) ClientUpdateProposal(ctx sdk.Context, p *types.ClientUpdateProposal) error {
	if p.SubjectClientId == exported.Localhost {
		return sdkerrors.Wrap(types.ErrInvalidUpdateClientProposal, "cannot update localhost client with proposal")
	}

	subjectClientState, found := k.GetClientState(ctx, p.SubjectClientId)
	if !found {
		return sdkerrors.Wrapf(types.ErrClientNotFound, "subject client with ID %s", p.SubjectClientId)
	}

	if subjectClientState.GetLatestHeight().GTE(p.InitialHeight) {
		return sdkerrors.Wrapf(types.ErrInvalidHeight, "subject client state latest height is greater or equal to initital height (%s >= %s)", subjectClientState.GetLatestHeight(), p.InitialHeight)
	}

	substituteClientState, found := k.GetClientState(ctx, p.SubstituteClientId)
	if !found {
		return sdkerrors.Wrapf(types.ErrClientNotFound, "substitute client with ID %s", p.SubstituteClientId)
	}

	clientState, consensusState, err := clientState.CheckSubstituteAndUpdateState(ctx, k.cdc, k.ClientStore(ctx, p.SubjectClientId), k.ClientStore(ctx, p.SubstituteClientId), p.InitialHeight)
	if err != nil {
		return err
	}

	k.Logger(ctx).Info("client updated after governance proposal passed", "client-id", p.SubjectClientId, "height", clientState.GetLatestHeight().String())

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{"ibc", "client", "update"},
			1,
			[]metrics.Label{
				telemetry.NewLabel("client-type", clientState.ClientType()),
				telemetry.NewLabel("client-id", p.SubjectClientId),
				telemetry.NewLabel("update-type", "proposal"),
			},
		)
	}()

	// emitting events in the keeper for proposal updates to clients
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUpdateClientProposal,
			sdk.NewAttribute(types.AttributeKeySubjectClientId, p.SubjectClientId),
			sdk.NewAttribute(types.AttributeKeyClientType, clientState.ClientType()),
			sdk.NewAttribute(types.AttributeKeyConsensusHeight, header.GetHeight().String()),
		),
	)

	return nil
}
