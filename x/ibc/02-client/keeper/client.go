package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
)

// CreateClient creates a new client state and populates it with a given consensus
// state as defined in https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#create
func (k Keeper) CreateClient(
	ctx sdk.Context, clientID string,
	clientType exported.ClientType, consensusState exported.ConsensusState,
) (types.State, error) {
	_, found := k.GetClientState(ctx, clientID)
	if found {
		return types.State{}, sdkerrors.Wrapf(errors.ErrClientExists, "cannot create client with ID %s", clientID)
	}

	_, found = k.GetClientType(ctx, clientID)
	if found {
		panic(fmt.Sprintf("consensus type is already defined for client %s", clientID))
	}

	clientState := k.initialize(ctx, clientID, consensusState)
	k.SetCommitter(ctx, clientID, consensusState.GetHeight(), consensusState.GetCommitter())
	k.SetVerifiedRoot(ctx, clientID, consensusState.GetHeight(), consensusState.GetRoot())
	k.SetClientState(ctx, clientState)
	k.SetClientType(ctx, clientID, clientType)
	k.Logger(ctx).Info(fmt.Sprintf("client %s created at height %d", clientID, consensusState.GetHeight()))

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreateClient,
			sdk.NewAttribute(types.AttributeKeyClientID, clientID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	})

	return clientState, nil
}

// UpdateClient updates the consensus state and the state root from a provided header
func (k Keeper) UpdateClient(ctx sdk.Context, clientID string, header exported.Header) error {
	clientType, found := k.GetClientType(ctx, clientID)
	if !found {
		return sdkerrors.Wrapf(errors.ErrClientTypeNotFound, "cannot update client with ID %s", clientID)
	}

	// check that the header consensus matches the client one
	if header.ClientType() != clientType {
		return sdkerrors.Wrapf(errors.ErrInvalidConsensus, "cannot update client with ID %s", clientID)
	}

	clientState, found := k.GetClientState(ctx, clientID)
	if !found {
		return sdkerrors.Wrapf(errors.ErrClientNotFound, "cannot update client with ID %s", clientID)
	}

	if clientState.Frozen {
		return sdkerrors.Wrapf(errors.ErrClientFrozen, "cannot update client with ID %s", clientID)
	}

	consensusState, found := k.GetConsensusState(ctx, clientID)
	if !found {
		return sdkerrors.Wrapf(errors.ErrConsensusStateNotFound, "cannot update client with ID %s", clientID)
	}

	consensusState, err := consensusState.CheckValidityAndUpdateState(header)
	if err != nil {
		return sdkerrors.Wrapf(err, "cannot update client with ID %s", clientID)
	}

	k.SetConsensusState(ctx, clientID, consensusState)
	k.SetCommitter(ctx, clientID, consensusState.GetHeight(), consensusState.GetCommitter())
	k.SetVerifiedRoot(ctx, clientID, consensusState.GetHeight(), consensusState.GetRoot())
	k.Logger(ctx).Info(fmt.Sprintf("client %s updated to height %d", clientID, consensusState.GetHeight()))

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeUpdateClient,
			sdk.NewAttribute(types.AttributeKeyClientID, clientID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	})

	return nil
}

// CheckMisbehaviourAndUpdateState checks for client misbehaviour and freezes the
// client if so.
//
// NOTE: In the first implementation, only Tendermint misbehaviour evidence is
// supported.
func (k Keeper) CheckMisbehaviourAndUpdateState(ctx sdk.Context, evidence evidenceexported.Evidence) error {
	misbehaviour, ok := evidence.(tendermint.Misbehaviour)
	if !ok {
		return sdkerrors.Wrap(errors.ErrInvalidClientType, "consensus type is not Tendermint")
	}

	clientState, found := k.GetClientState(ctx, misbehaviour.ClientID)
	if !found {
		return sdkerrors.Wrap(errors.ErrClientNotFound, misbehaviour.ClientID)
	}

	committer, found := k.GetCommitter(ctx, misbehaviour.ClientID, uint64(misbehaviour.GetHeight()))
	if !found {
		return errors.ErrCommitterNotFound
	}
	tmCommitter, ok := committer.(tendermint.Committer)
	if !ok {
		return sdkerrors.Wrap(errors.ErrInvalidCommitter, "committer type is not Tendermint")
	}

	if err := tendermint.CheckMisbehaviour(tmCommitter, misbehaviour); err != nil {
		return sdkerrors.Wrap(errors.ErrInvalidEvidence, err.Error())
	}

	clientState, err := k.freeze(ctx, clientState)
	if err != nil {
		return err
	}

	k.SetClientState(ctx, clientState)
	k.Logger(ctx).Info(fmt.Sprintf("client %s frozen due to misbehaviour", misbehaviour.ClientID))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSubmitMisbehaviour,
			sdk.NewAttribute(types.AttributeKeyClientID, misbehaviour.ClientID),
		),
	)

	return nil
}
