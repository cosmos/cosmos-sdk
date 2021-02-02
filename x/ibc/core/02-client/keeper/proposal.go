package keeper

import (
	"github.com/armon/go-metrics"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
)

// ClientUpdateProposal will retrieve the subject and substitute client.
// The initial height must be greater than the latest height of the subject
// client. A callback will occur to the subject client state with the client
// prefixed store being provided for both the subject and the substitute client.
// The localhost client is not allowed to be modified with a proposal. The IBC
// client implementations are responsible for validating the parameters of the
// subtitute (enusring they match the subject's parameters) as well as copying
// the necessary consensus states from the subtitute to the subject client
// store.
//
// NOTE: Substitute clients with revision numbers not equal to the revision
// number of the subject client is explicitly disallowed. We cannot support
// this until there is a way to range query for the all the consensus
// states which occurred between two IBC Revision heights.
// https://github.com/cosmos/cosmos-sdk/issues/7712
func (k Keeper) ClientUpdateProposal(ctx sdk.Context, p *types.ClientUpdateProposal) error {
	if p.SubjectClientId == exported.Localhost || p.SubstituteClientId == exported.Localhost {
		return sdkerrors.Wrap(types.ErrInvalidUpdateClientProposal, "cannot update localhost client with proposal")
	}

	subjectClientState, found := k.GetClientState(ctx, p.SubjectClientId)
	if !found {
		return sdkerrors.Wrapf(types.ErrClientNotFound, "subject client with ID %s", p.SubjectClientId)
	}

	if subjectClientState.GetLatestHeight().GTE(p.InitialHeight) {
		return sdkerrors.Wrapf(types.ErrInvalidHeight, "subject client state latest height is greater or equal to initial height (%s >= %s)", subjectClientState.GetLatestHeight(), p.InitialHeight)
	}

	substituteClientState, found := k.GetClientState(ctx, p.SubstituteClientId)
	if !found {
		return sdkerrors.Wrapf(types.ErrClientNotFound, "substitute client with ID %s", p.SubstituteClientId)
	}

	// substitute clients with height across revision numbers is not allowed
	if subjectClientState.GetLatestHeight().GetRevisionNumber() != substituteClientState.GetLatestHeight().GetRevisionNumber() {
		return sdkerrors.Wrapf(types.ErrInvalidHeight, "subject client state and substitute client state must have the same revision number (%d != %d)", subjectClientState.GetLatestHeight().GetRevisionNumber(), substituteClientState.GetLatestHeight().GetRevisionNumber())
	}

	clientState, err := subjectClientState.CheckSubstituteAndUpdateState(ctx, k.cdc, k.ClientStore(ctx, p.SubjectClientId), k.ClientStore(ctx, p.SubstituteClientId), substituteClientState, p.InitialHeight)
	if err != nil {
		return err
	}
	k.SetClientState(ctx, p.SubjectClientId, clientState)

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
			sdk.NewAttribute(types.AttributeKeySubjectClientID, p.SubjectClientId),
			sdk.NewAttribute(types.AttributeKeyClientType, clientState.ClientType()),
			sdk.NewAttribute(types.AttributeKeyConsensusHeight, clientState.GetLatestHeight().String()),
		),
	)

	return nil
}
