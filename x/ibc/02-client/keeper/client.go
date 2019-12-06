package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
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
		return types.State{}, errors.ErrClientExists(k.codespace, clientID)
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
	return clientState, nil
}

// IterateClients returns an iterator that allows for returning a list of clients
func (k Keeper) IterateClients(ctx sdk.Context, cb func(types.State) bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixClient)
	iterator := sdk.KVStorePrefixIterator(store, nil)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var clientState types.State
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &clientState)

		if cb(clientState) {
			break
		}
	}
}

// GetAllClients returns all stored light client State objects.
func (k Keeper) GetAllClients(ctx sdk.Context) (states []types.State) {
	k.IterateClients(ctx, func(state types.State) bool {
		states = append(states, state)
		return false
	})
	return states
}

// UpdateClient updates the consensus state and the state root from a provided header
func (k Keeper) UpdateClient(ctx sdk.Context, clientID string, header exported.Header) error {
	clientType, found := k.GetClientType(ctx, clientID)
	if !found {
		return sdkerrors.Wrap(errors.ErrClientTypeNotFound(k.codespace), "cannot update client")
	}

	// check that the header consensus matches the client one
	if header.ClientType() != clientType {
		return sdkerrors.Wrap(errors.ErrInvalidConsensus(k.codespace), "cannot update client")
	}

	clientState, found := k.GetClientState(ctx, clientID)
	if !found {
		return sdkerrors.Wrap(errors.ErrClientNotFound(k.codespace, clientID), "cannot update client")
	}

	if clientState.Frozen {
		return sdkerrors.Wrap(errors.ErrClientFrozen(k.codespace, clientID), "cannot update client")
	}

	consensusState, found := k.GetConsensusState(ctx, clientID)
	if !found {
		return sdkerrors.Wrap(errors.ErrConsensusStateNotFound(k.codespace), "cannot update client")
	}

	consensusState, err := consensusState.CheckValidityAndUpdateState(header)
	if err != nil {
		return sdkerrors.Wrap(err, "cannot update client")
	}

	k.SetConsensusState(ctx, clientID, consensusState)
	k.SetCommitter(ctx, clientID, consensusState.GetHeight(), consensusState.GetCommitter())
	k.SetVerifiedRoot(ctx, clientID, consensusState.GetHeight(), consensusState.GetRoot())
	k.Logger(ctx).Info(fmt.Sprintf("client %s updated to height %d", clientID, consensusState.GetHeight()))
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
		return errors.ErrInvalidClientType(k.codespace, "consensus type is not Tendermint")
	}

	clientState, found := k.GetClientState(ctx, misbehaviour.ClientID)
	if !found {
		return errors.ErrClientNotFound(k.codespace, misbehaviour.ClientID)
	}

	committer, found := k.GetCommitter(ctx, misbehaviour.ClientID, uint64(misbehaviour.GetHeight()))
	if !found {
		return errors.ErrCommitterNotFound(k.codespace, fmt.Sprintf("committer not found for height %d", misbehaviour.GetHeight()))
	}
	tmCommitter, ok := committer.(tendermint.Committer)
	if !ok {
		return errors.ErrInvalidCommitter(k.codespace, "committer type is not Tendermint")
	}

	if err := tendermint.CheckMisbehaviour(tmCommitter, misbehaviour); err != nil {
		return errors.ErrInvalidEvidence(k.codespace, err.Error())
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
