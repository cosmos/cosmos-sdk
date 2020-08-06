package keeper

import (
	"fmt"
	"strings"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// Keeper represents a type that grants read and write permissions to any client
// state information
type Keeper struct {
	storeKey      sdk.StoreKey
	cdc           codec.BinaryMarshaler
	stakingKeeper types.StakingKeeper
}

// NewKeeper creates a new NewKeeper instance
func NewKeeper(cdc codec.BinaryMarshaler, key sdk.StoreKey, sk types.StakingKeeper) Keeper {
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

	clientState := k.MustUnmarshalClientState(bz)
	return clientState, true
}

// SetClientState sets a particular Client to the store
func (k Keeper) SetClientState(ctx sdk.Context, clientID string, clientState exported.ClientState) {
	store := k.ClientStore(ctx, clientID)
	store.Set(host.KeyClientState(), k.MustMarshalClientState(clientState))
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
func (k Keeper) GetClientConsensusState(ctx sdk.Context, clientID string, height uint64) (exported.ConsensusState, bool) {
	store := k.ClientStore(ctx, clientID)
	bz := store.Get(host.KeyConsensusState(height))
	if bz == nil {
		return nil, false
	}

	consensusState := k.MustUnmarshalConsensusState(bz)
	return consensusState, true
}

// SetClientConsensusState sets a ConsensusState to a particular client at the given
// height
func (k Keeper) SetClientConsensusState(ctx sdk.Context, clientID string, height uint64, consensusState exported.ConsensusState) {
	store := k.ClientStore(ctx, clientID)
	store.Set(host.KeyConsensusState(height), k.MustMarshalConsensusState(consensusState))
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
		consensusState := k.MustUnmarshalConsensusState(iterator.Value())

		if cb(clientID, consensusState) {
			break
		}
	}
}

// GetAllGenesisClients returns all the clients in state with their client ids returned as GenesisClientState
func (k Keeper) GetAllGenesisClients(ctx sdk.Context) (genClients []types.GenesisClientState) {
	k.IterateClients(ctx, func(clientID string, cs exported.ClientState) bool {
		genClients = append(genClients, types.NewGenesisClientState(clientID, cs))
		return false
	})
	return
}

// GetAllConsensusStates returns all stored client consensus states.
func (k Keeper) GetAllConsensusStates(ctx sdk.Context) types.ClientsConsensusStates {
	clientConsStates := make(types.ClientsConsensusStates, 0)
	mapClientIDToConsStateIdx := make(map[string]int)

	k.IterateConsensusStates(ctx, func(clientID string, cs exported.ConsensusState) bool {
		anyClientState := types.MustPackConsensusState(cs)

		idx, ok := mapClientIDToConsStateIdx[clientID]
		if ok {
			clientConsStates[idx].ConsensusStates = append(clientConsStates[idx].ConsensusStates, anyClientState)
			return false
		}

		clientConsState := types.ClientConsensusStates{
			ClientID:        clientID,
			ConsensusStates: []*codectypes.Any{anyClientState},
		}

		clientConsStates = append(clientConsStates, clientConsState)
		mapClientIDToConsStateIdx[clientID] = len(clientConsStates) - 1
		return false
	})

	return clientConsStates.Sort()
}

// HasClientConsensusState returns if keeper has a ConsensusState for a particular
// client at the given height
func (k Keeper) HasClientConsensusState(ctx sdk.Context, clientID string, height uint64) bool {
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
func (k Keeper) GetClientConsensusStateLTE(ctx sdk.Context, clientID string, maxHeight uint64) (exported.ConsensusState, bool) {
	for i := maxHeight; i > 0; i-- {
		found := k.HasClientConsensusState(ctx, clientID, i)
		if found {
			return k.GetClientConsensusState(ctx, clientID, i)
		}
	}
	return nil, false
}

// GetSelfConsensusState introspects the (self) past historical info at a given height
// and returns the expected consensus state at that height.
func (k Keeper) GetSelfConsensusState(ctx sdk.Context, height uint64) (exported.ConsensusState, bool) {
	histInfo, found := k.stakingKeeper.GetHistoricalInfo(ctx, int64(height))
	if !found {
		return nil, false
	}

	consensusState := &ibctmtypes.ConsensusState{
		Height:             height,
		Timestamp:          histInfo.Header.Time,
		Root:               commitmenttypes.NewMerkleRoot(histInfo.Header.AppHash),
		NextValidatorsHash: histInfo.Header.NextValidatorsHash,
	}
	return consensusState, true
}

// IterateClients provides an iterator over all stored light client State
// objects. For each State object, cb will be called. If the cb returns true,
// the iterator will close and stop.
func (k Keeper) IterateClients(ctx sdk.Context, cb func(clientID string, cs exported.ClientState) bool) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, host.KeyClientStorePrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		keySplit := strings.Split(string(iterator.Key()), "/")
		if keySplit[len(keySplit)-1] != "clientState" {
			continue
		}
		clientState := k.MustUnmarshalClientState(iterator.Value())

		// key is ibc/{clientid}/clientState
		// Thus, keySplit[1] is clientID
		if cb(keySplit[1], clientState) {
			break
		}
	}
}

// GetAllClients returns all stored light client State objects.
func (k Keeper) GetAllClients(ctx sdk.Context) (states []exported.ClientState) {
	k.IterateClients(ctx, func(_ string, state exported.ClientState) bool {
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
