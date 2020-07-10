package keeper

import (
	"fmt"
	"strings"

	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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
	return ctx.Logger().With("module", fmt.Sprintf("x/%s/%s", host.ModuleName, types.SubModuleName))
}

// GetClientState gets a particular client from the store
func (k Keeper) GetClientState(ctx sdk.Context, clientID string) (exported.ClientState, bool) {
	store := k.ClientStore(ctx, clientID)
	bz := store.Get(host.KeyClientState())
	if bz == nil {
		return nil, false
	}

	var clientState exported.ClientState
	k.cdc.MustUnmarshalBinaryBare(bz, &clientState)
	return clientState, true
}

// SetClientState sets a particular Client to the store
func (k Keeper) SetClientState(ctx sdk.Context, clientState exported.ClientState) {
	store := k.ClientStore(ctx, clientState.GetID())
	bz := k.cdc.MustMarshalBinaryBare(clientState)
	store.Set(host.KeyClientState(), bz)
}

// GetClientType gets the consensus type for a specific client
func (k Keeper) GetClientType(ctx sdk.Context, clientID string) (exported.ClientType, bool) {
	store := k.ClientStore(ctx, clientID)
	bz := store.Get(host.KeyClientType())
	if bz == nil {
		return 0, false
	}

	return exported.ClientType(bz[0]), true
}

// SetClientType sets the specific client consensus type to the provable store
func (k Keeper) SetClientType(ctx sdk.Context, clientID string, clientType exported.ClientType) {
	store := k.ClientStore(ctx, clientID)
	store.Set(host.KeyClientType(), []byte{byte(clientType)})
}

// GetClientConsensusState gets the stored consensus state from a client at a given height.
func (k Keeper) GetClientConsensusState(ctx sdk.Context, clientID string, height exported.Height) (exported.ConsensusState, bool) {
	store := k.ClientStore(ctx, clientID)
	bz := store.Get(host.KeyConsensusState(height))
	if bz == nil {
		return nil, false
	}

	var consensusState exported.ConsensusState
	k.cdc.MustUnmarshalBinaryBare(bz, &consensusState)
	return consensusState, true
}

// SetClientConsensusState sets a ConsensusState to a particular client at the given
// height
func (k Keeper) SetClientConsensusState(ctx sdk.Context, clientID string, height exported.Height, consensusState exported.ConsensusState) {
	store := k.ClientStore(ctx, clientID)
	bz := k.cdc.MustMarshalBinaryBare(consensusState)
	store.Set(host.KeyConsensusState(height), bz)
}

// IterateConsensusStates provides an iterator over all stored consensus states.
// objects. For each State object, cb will be called. If the cb returns true,
// the iterator will close and stop.
func (k Keeper) IterateConsensusStates(ctx sdk.Context, cb func(clientID string, cs exported.ConsensusState) bool) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, host.KeyClientStorePrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		keySplit := strings.Split(string(iterator.Key()), "/")
		// consensus key is in the format "clients/<clientID>/consensusState/<height>"
		if len(keySplit) != 4 || keySplit[2] != "consensusState" {
			continue
		}
		clientID := keySplit[1]
		var consensusState exported.ConsensusState
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &consensusState)

		if cb(clientID, consensusState) {
			break
		}
	}
}

// GetAllConsensusStates returns all stored client consensus states.
// NOTE: non deterministic.
func (k Keeper) GetAllConsensusStates(ctx sdk.Context) (clientConsStates []types.ClientConsensusStates) {
	var clientIDs []string
	// create map to add consensus states to the existing clients
	cons := make(map[string][]exported.ConsensusState)

	k.IterateConsensusStates(ctx, func(clientID string, cs exported.ConsensusState) bool {
		consensusStates, ok := cons[clientID]
		if !ok {
			clientIDs = append(clientIDs, clientID)
			cons[clientID] = []exported.ConsensusState{cs}
			return false
		}

		cons[clientID] = append(consensusStates, cs)
		return false
	})

	// create ClientConsensusStates in the same order of iteration to prevent non-determinism
	for len(clientIDs) > 0 {
		id := clientIDs[len(clientIDs)-1]
		consensusStates, ok := cons[id]
		if !ok {
			panic(fmt.Sprintf("consensus states from client id %s not found", id))
		}

		clientConsState := types.NewClientConsensusStates(id, consensusStates)
		clientConsStates = append(clientConsStates, clientConsState)

		// remove the last element
		clientIDs = clientIDs[:len(clientIDs)-1]
	}

	return clientConsStates
}

// HasClientConsensusState returns if keeper has a ConsensusState for a particular
// client at the given height
func (k Keeper) HasClientConsensusState(ctx sdk.Context, clientID string, height exported.Height) bool {
	store := k.ClientStore(ctx, clientID)
	return store.Has(host.KeyConsensusState(height))
}

// GetLatestClientConsensusState gets the latest ConsensusState stored for a given client
func (k Keeper) GetLatestClientConsensusState(ctx sdk.Context, clientID string) (exported.ConsensusState, bool) {
	clientState, ok := k.GetClientState(ctx, clientID)
	if !ok {
		return nil, false
	}
	return k.GetClientConsensusState(ctx, clientID, clientState.GetLatestHeight())
}

// GetClientConsensusStateLTE will get the latest ConsensusState of a particular client at the latest height
// less than or equal to the given height
func (k Keeper) GetClientConsensusStateLTE(ctx sdk.Context, clientID string, maxHeight exported.Height) (exported.ConsensusState, bool) {
	// NOTE: For tendermint Heights this is only implemented for a single epoch
	height := maxHeight
	var err error
	for height.Valid() {
		found := k.HasClientConsensusState(ctx, clientID, height)
		if found {
			return k.GetClientConsensusState(ctx, clientID, height)
		}
		height, err = height.Decrement()
		if err != nil {
			break
		}
	}
	return nil, false
}

// GetSelfConsensusState introspects the (self) past historical info at a given height
// and returns the expected consensus state at that height.
func (k Keeper) GetSelfConsensusState(ctx sdk.Context, height exported.Height) (exported.ConsensusState, bool) {
	tmHeight, ok := height.(ibctmtypes.Height)
	if !ok {
		return nil, false
	}
	// Only support retrieving HistoricalInfo from the current epoch for now
	// TODO: check EpochNumber matches expected
	histInfo, found := k.stakingKeeper.GetHistoricalInfo(ctx, int64(tmHeight.EpochHeight))
	if !found {
		return nil, false
	}

	valSet := stakingtypes.Validators(histInfo.Valset)

	consensusState := ibctmtypes.ConsensusState{
		Height:       tmHeight,
		Timestamp:    histInfo.Header.Time,
		Root:         commitmenttypes.NewMerkleRoot(histInfo.Header.AppHash),
		ValidatorSet: tmtypes.NewValidatorSet(valSet.ToTmValidators()),
	}
	return consensusState, true
}

// IterateClients provides an iterator over all stored light client State
// objects. For each State object, cb will be called. If the cb returns true,
// the iterator will close and stop.
func (k Keeper) IterateClients(ctx sdk.Context, cb func(exported.ClientState) bool) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, host.KeyClientStorePrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		keySplit := strings.Split(string(iterator.Key()), "/")
		if keySplit[len(keySplit)-1] != "clientState" {
			continue
		}
		var clientState exported.ClientState
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &clientState)

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

// ClientStore returns isolated prefix store for each client so they can read/write in separate
// namespace without being able to read/write other client's data
func (k Keeper) ClientStore(ctx sdk.Context, clientID string) sdk.KVStore {
	// append here is safe, appends within a function won't cause
	// weird side effects when its singlethreaded
	clientPrefix := append([]byte("clients/"+clientID), '/')
	return prefix.NewStore(ctx.KVStore(k.storeKey), clientPrefix)
}
