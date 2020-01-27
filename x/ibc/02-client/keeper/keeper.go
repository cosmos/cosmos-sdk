package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint"
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

// GetClientConsensusState gets the stored consensus state from a client at a given height.
func (k Keeper) GetClientConsensusState(ctx sdk.Context, clientID string, height uint64) (exported.ConsensusState, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(ibctypes.KeyConsensusState(clientID, height))
	if bz == nil {
		return nil, false
	}

	var consensusState exported.ConsensusState
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &consensusState)
	return consensusState, true
}

// SetClientConsensusState sets a ConsensusState to a particular client at the given
// height
func (k Keeper) SetClientConsensusState(ctx sdk.Context, clientID string, height uint64, consensusState exported.ConsensusState) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(consensusState)
	store.Set(ibctypes.KeyConsensusState(clientID, height), bz)
}

// GetSelfConsensusState introspects the (self) past historical info at a given height
// and returns the expected consensus state at that height.
func (k Keeper) GetSelfConsensusState(ctx sdk.Context, height uint64) (exported.ConsensusState, bool) {
	histInfo, found := k.stakingKeeper.GetHistoricalInfo(ctx, int64(height))
	if !found {
		return nil, false
	}

	consensusState := tendermint.ConsensusState{
		Root:             commitment.NewRoot(histInfo.Header.AppHash),
		ValidatorSetHash: tmtypes.NewValidatorSet(histInfo.ValSet.ToTmValidators()).Hash(),
	}
	return consensusState, true
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

	k.SetClientConsensusState(ctx, clientID, height, consensusState)
	return clientState, nil
}
