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
	cdc           codec.BinaryMarshaler
	aminoCdc      *codec.Codec
	stakingKeeper types.StakingKeeper
}

// NewKeeper creates a new NewKeeper instance
func NewKeeper(cdc codec.BinaryMarshaler, aminoCdc *codec.Codec, key sdk.StoreKey, sk types.StakingKeeper) Keeper {
	return Keeper{
		storeKey:      key,
		cdc:           cdc,
		aminoCdc:      aminoCdc,
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

	var consensusState exported.ConsensusState
	k.aminoCdc.MustUnmarshalBinaryBare(bz, &consensusState)
	return consensusState, true
}

// SetClientConsensusState sets a ConsensusState to a particular client at the given
// height
func (k Keeper) SetClientConsensusState(ctx sdk.Context, clientID string, height uint64, consensusState exported.ConsensusState) {
	store := k.ClientStore(ctx, clientID)
	bz := k.aminoCdc.MustMarshalBinaryBare(consensusState)
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
		k.aminoCdc.MustUnmarshalBinaryBare(iterator.Value(), &consensusState)

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

	valSet := stakingtypes.Validators(histInfo.Valset)

	consensusState := ibctmtypes.ConsensusState{
		Height:             height,
		Timestamp:          histInfo.Header.Time,
		Root:               commitmenttypes.NewMerkleRoot(histInfo.Header.AppHash),
		NextValidatorsHash: histInfo.Header.NextValidatorsHash,
		ValidatorSet:       tmtypes.NewValidatorSet(valSet.ToTmValidators()),
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
