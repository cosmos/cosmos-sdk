package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// Keeper represents a type that grants read and write permissions to any client
// state information
type Keeper struct {
	storeKey sdk.StoreKey
	cdc      *codec.Codec
}

// NewKeeper creates a new NewKeeper instance
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey) Keeper {
	return Keeper{
		storeKey: key,
		cdc:      cdc,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s/%s", ibctypes.ModuleName, types.SubModuleName))
}

// GetClientState gets a particular client from the store
func (k Keeper) GetClientState(ctx sdk.Context, clientID string) (exported.ClientState, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(ibctypes.KeyClientState(clientID))
	if bz == nil {
		return nil, false
	}

	var clientState exported.ClientState
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &clientState)
	return clientState, true
}

// SetClientState sets a particular Client to the store
func (k Keeper) SetClientState(ctx sdk.Context, clientState exported.ClientState) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(clientState)
	store.Set(ibctypes.KeyClientState(clientState.GetID()), bz)
}

// GetClientType gets the consensus type for a specific client
func (k Keeper) GetClientType(ctx sdk.Context, clientID string) (exported.ClientType, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(ibctypes.KeyClientType(clientID))
	if bz == nil {
		return 0, false
	}

	return exported.ClientType(bz[0]), true
}

// SetClientType sets the specific client consensus type to the provable store
func (k Keeper) SetClientType(ctx sdk.Context, clientID string, clientType exported.ClientType) {
	store := ctx.KVStore(k.storeKey)
	store.Set(ibctypes.KeyClientType(clientID), []byte{byte(clientType)})
}

// GetConsensusState creates a new client state and populates it with a given consensus state
func (k Keeper) GetConsensusState(ctx sdk.Context, clientID string, height uint64) (exported.ConsensusState, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(ibctypes.KeyConsensusState(clientID, height))
	if bz == nil {
		return nil, false
	}

	var consensusState exported.ConsensusState
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &consensusState)
	return consensusState, true
}

// SetConsensusState sets a ConsensusState to a particular client
func (k Keeper) SetConsensusState(ctx sdk.Context, clientID string, height uint64, consensusState exported.ConsensusState) {
	fmt.Printf("set consensus state for client ID %v and height %v\n", clientID, height)
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(consensusState)
	store.Set(ibctypes.KeyConsensusState(clientID, height), bz)
}

// IterateClients provides an iterator over all stored light client State
// objects. For each State object, cb will be called. If the cb returns true,
// the iterator will close and stop.
func (k Keeper) IterateClients(ctx sdk.Context, cb func(exported.ClientState) bool) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, ibctypes.KeyClientPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var clientState exported.ClientState
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &clientState)

		if cb(clientState) {
			break
		}
	}
}

// GetAllClients returns all stored light client State objects.
func (k Keeper) GetAllClients(ctx sdk.Context) (states []exported.ClientState) {
	k.IterateClients(ctx, func(state exported.ClientState) bool {
		states = append(states, state)
		return false
	})
	return states
}

// State returns a new client state with a given id as defined in
// https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#example-implementation
func (k Keeper) initialize(
	ctx sdk.Context, clientID string, clientType exported.ClientType,
	consensusState exported.ConsensusState,
) (exported.ClientState, error) {
	var clientState exported.ClientState
	height := uint64(ctx.BlockHeight())

	switch clientType {
	case exported.Tendermint:
		clientState = tendermint.NewClientState(clientID, height)
	default:
		return nil, types.ErrInvalidClientType
	}

	k.SetConsensusState(ctx, clientID, height, consensusState)
	return clientState, nil
}
