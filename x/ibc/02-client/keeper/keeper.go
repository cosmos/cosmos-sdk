package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// Keeper represents a type that grants read and write permissions to any client
// state information
type Keeper struct {
	storeKey      sdk.StoreKey
	cdc           *codec.Codec
	stakingKeeper types.StakingKeeper
}

// NewKeeper creates a new NewKeeper instance
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, sk types.StakingKeeper) Keeper {
	return Keeper{
		storeKey:      key,
		cdc:           cdc,
		stakingKeeper: sk,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s/%s", ibctypes.ModuleName, types.SubModuleName))
}

// GetClientState gets a particular client from the store
func (k Keeper) GetClientState(ctx sdk.Context, clientID string) (types.State, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyClientState(clientID))
	if bz == nil {
		return types.State{}, false
	}

	var clientState types.State
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &clientState)
	return clientState, true
}

// SetClientState sets a particular Client to the store
func (k Keeper) SetClientState(ctx sdk.Context, clientState types.State) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(clientState)
	store.Set(types.KeyClientState(clientState.ID), bz)
}

// GetClientType gets the consensus type for a specific client
func (k Keeper) GetClientType(ctx sdk.Context, clientID string) (exported.ClientType, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyClientType(clientID))
	if bz == nil {
		return 0, false
	}

	return exported.ClientType(bz[0]), true
}

// SetClientType sets the specific client consensus type to the provable store
func (k Keeper) SetClientType(ctx sdk.Context, clientID string, clientType exported.ClientType) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyClientType(clientID), []byte{byte(clientType)})
}

// GetClientConsensusState gets the latest stored consensus state from a given client.
func (k Keeper) GetClientConsensusState(ctx sdk.Context, clientID string) (exported.ConsensusState, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyConsensusState(clientID))
	if bz == nil {
		return nil, false
	}

	var consensusState exported.ConsensusState
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &consensusState)
	return consensusState, true
}

// SetClientConsensusState sets a ConsensusState to a particular client at the latest
// height
func (k Keeper) SetClientConsensusState(ctx sdk.Context, clientID string, consensusState exported.ConsensusState) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(consensusState)
	store.Set(types.KeyConsensusState(clientID), bz)
}

// GetVerifiedRoot gets a verified commitment Root from a particular height to
// a client
func (k Keeper) GetVerifiedRoot(ctx sdk.Context, clientID string, height uint64) (commitment.RootI, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.KeyRoot(clientID, height))
	if bz == nil {
		return nil, false
	}

	var root commitment.RootI
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &root)
	return root, true
}

// SetVerifiedRoot sets a verified commitment Root from a particular height to
// a client
func (k Keeper) SetVerifiedRoot(ctx sdk.Context, clientID string, height uint64, root commitment.RootI) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(root)
	store.Set(types.KeyRoot(clientID, height), bz)
}

// IterateClients provides an iterator over all stored light client State
// objects. For each State object, cb will be called. If the cb returns true,
// the iterator will close and stop.
func (k Keeper) IterateClients(ctx sdk.Context, cb func(types.State) bool) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetClientKeysPrefix(ibctypes.KeyClientPrefix))

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

// GetCommitter will get the Committer of a particular client at the oldest height
// that is less than or equal to the height passed in
func (k Keeper) GetCommitter(ctx sdk.Context, clientID string, height uint64) (exported.Committer, bool) {
	store := ctx.KVStore(k.storeKey)

	var committer exported.Committer

	// TODO: Replace this for-loop with a ReverseIterator for efficiency
	for i := height; i > 0; i-- {
		bz := store.Get(types.KeyCommitter(clientID, i))
		if bz != nil {
			k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &committer)
			return committer, true
		}
	}
	return nil, false
}

// SetCommitter sets a committer from a particular height to
// a particular client
func (k Keeper) SetCommitter(ctx sdk.Context, clientID string, height uint64, committer exported.Committer) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(committer)
	store.Set(types.KeyCommitter(clientID, height), bz)
}

// GetConsensusState introspects the past historical info at a given height and
// returns the expected consensus state at that height.
func (k Keeper) GetConsensusState(ctx sdk.Context, height int64) (exported.ConsensusState, bool) {
	histInfo, found := k.stakingKeeper.GetHistoricalInfo(ctx, height)
	if !found {
		return nil, false
	}

	// query the following historical info to get the next validator set
	nextHistInfo, found := k.stakingKeeper.GetHistoricalInfo(ctx, height+1)
	if !found {
		return nil, false
	}

	consensusState := tendermint.ConsensusState{
		ChainID:          histInfo.Header.ChainID,
		Height:           uint64(histInfo.Header.Height),
		Root:             commitment.NewRoot(histInfo.Header.AppHash),
		ValidatorSet:     tmtypes.NewValidatorSet(histInfo.ValSet.ToTmValidators()),
		NextValidatorSet: tmtypes.NewValidatorSet(nextHistInfo.ValSet.ToTmValidators()),
	}
	return consensusState, true
}

// State returns a new client state with a given id as defined in
// https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#example-implementation
func (k Keeper) initialize(ctx sdk.Context, clientID string, consensusState exported.ConsensusState) types.State {
	clientState := types.NewClientState(clientID)
	k.SetClientConsensusState(ctx, clientID, consensusState)
	return clientState
}

// freeze updates the state of the client in the event of a misbehaviour
func (k Keeper) freeze(ctx sdk.Context, clientState types.State) (types.State, error) {
	if clientState.Frozen {
		return types.State{}, sdkerrors.Wrap(errors.ErrClientFrozen, clientState.ID)
	}

	clientState.Frozen = true
	return clientState, nil
}

// VerifyMembership state membership verification function defined by the client type
func (k Keeper) VerifyMembership(
	ctx sdk.Context,
	clientID string,
	height uint64, // sequence
	proof commitment.ProofI,
	path commitment.PathI,
	value []byte,
) bool {
	clientState, found := k.GetClientState(ctx, clientID)
	if !found {
		return false
	}

	if clientState.Frozen {
		return false
	}

	root, found := k.GetVerifiedRoot(ctx, clientID, height)
	if !found {
		return false
	}

	return proof.VerifyMembership(root, path, value)
}

// VerifyNonMembership state non-membership function defined by the client type
func (k Keeper) VerifyNonMembership(
	ctx sdk.Context,
	clientID string,
	height uint64, // sequence
	proof commitment.ProofI,
	path commitment.PathI,
) bool {
	clientState, found := k.GetClientState(ctx, clientID)
	if !found {
		return false
	}

	if clientState.Frozen {
		return false
	}

	root, found := k.GetVerifiedRoot(ctx, clientID, height)
	if !found {
		return false
	}

	return proof.VerifyNonMembership(root, path)
}
