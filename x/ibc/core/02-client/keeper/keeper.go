package keeper

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/light"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
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

// GetClientConsensusState gets the stored consensus state from a client at a given height.
func (k Keeper) GetClientConsensusState(ctx sdk.Context, clientID string, height exported.Height) (exported.ConsensusState, bool) {
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
func (k Keeper) SetClientConsensusState(ctx sdk.Context, clientID string, height exported.Height, consensusState exported.ConsensusState) {
	store := k.ClientStore(ctx, clientID)
	store.Set(host.KeyConsensusState(height), k.MustMarshalConsensusState(consensusState))
}

// IterateConsensusStates provides an iterator over all stored consensus states.
// objects. For each State object, cb will be called. If the cb returns true,
// the iterator will close and stop.
func (k Keeper) IterateConsensusStates(ctx sdk.Context, cb func(clientID string, cs types.ConsensusStateWithHeight) bool) {
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
		height := types.MustParseHeight(keySplit[3])
		consensusState := k.MustUnmarshalConsensusState(iterator.Value())

		consensusStateWithHeight := types.NewConsensusStateWithHeight(height, consensusState)

		if cb(clientID, consensusStateWithHeight) {
			break
		}
	}
}

// GetAllGenesisClients returns all the clients in state with their client ids returned as IdentifiedClientState
func (k Keeper) GetAllGenesisClients(ctx sdk.Context) (genClients []types.IdentifiedClientState) {
	k.IterateClients(ctx, func(clientID string, cs exported.ClientState) bool {
		genClients = append(genClients, types.NewIdentifiedClientState(clientID, cs))
		return false
	})
	return
}

// GetAllConsensusStates returns all stored client consensus states.
func (k Keeper) GetAllConsensusStates(ctx sdk.Context) types.ClientsConsensusStates {
	clientConsStates := make(types.ClientsConsensusStates, 0)
	mapClientIDToConsStateIdx := make(map[string]int)

	k.IterateConsensusStates(ctx, func(clientID string, cs types.ConsensusStateWithHeight) bool {
		idx, ok := mapClientIDToConsStateIdx[clientID]
		if ok {
			clientConsStates[idx].ConsensusStates = append(clientConsStates[idx].ConsensusStates, cs)
			return false
		}

		clientConsState := types.ClientConsensusStates{
			ClientId:        clientID,
			ConsensusStates: []types.ConsensusStateWithHeight{cs},
		}

		clientConsStates = append(clientConsStates, clientConsState)
		mapClientIDToConsStateIdx[clientID] = len(clientConsStates) - 1
		return false
	})

	return clientConsStates.Sort()
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
// It will only search for heights within the same version
func (k Keeper) GetClientConsensusStateLTE(ctx sdk.Context, clientID string, maxHeight exported.Height) (exported.ConsensusState, bool) {
	h := maxHeight
	ok := true
	for ok {
		found := k.HasClientConsensusState(ctx, clientID, h)
		if found {
			return k.GetClientConsensusState(ctx, clientID, h)
		}
		h, ok = h.Decrement()
	}
	return nil, false
}

// GetSelfConsensusState introspects the (self) past historical info at a given height
// and returns the expected consensus state at that height.
// For now, can only retrieve self consensus states for the current version
func (k Keeper) GetSelfConsensusState(ctx sdk.Context, height exported.Height) (exported.ConsensusState, bool) {
	selfHeight, ok := height.(types.Height)
	if !ok {
		return nil, false
	}
	// check that height version matches chainID version
	version := types.ParseChainID(ctx.ChainID())
	if version != height.GetVersionNumber() {
		return nil, false
	}
	histInfo, found := k.stakingKeeper.GetHistoricalInfo(ctx, int64(selfHeight.VersionHeight))
	if !found {
		return nil, false
	}

	consensusState := &ibctmtypes.ConsensusState{
		Timestamp:          histInfo.Header.Time,
		Root:               commitmenttypes.NewMerkleRoot(histInfo.Header.GetAppHash()),
		NextValidatorsHash: histInfo.Header.NextValidatorsHash,
	}
	return consensusState, true
}

// ValidateSelfClient validates the client parameters for a client of the running chain
// This function is only used to validate the client state the counterparty stores for this chain
// Client must be in same version as the executing chain
func (k Keeper) ValidateSelfClient(ctx sdk.Context, clientState exported.ClientState) error {
	tmClient, ok := clientState.(*ibctmtypes.ClientState)
	if !ok {
		return sdkerrors.Wrapf(types.ErrInvalidClient, "client must be a Tendermint client, expected: %T, got: %T",
			&ibctmtypes.ClientState{}, tmClient)
	}

	if clientState.IsFrozen() {
		return types.ErrClientFrozen
	}

	if ctx.ChainID() != tmClient.ChainId {
		return sdkerrors.Wrapf(types.ErrInvalidClient, "invalid chain-id. expected: %s, got: %s",
			ctx.ChainID(), tmClient.ChainId)
	}

	version := types.ParseChainID(ctx.ChainID())

	// client must be in the same version as executing chain
	if tmClient.LatestHeight.VersionNumber != version {
		return sdkerrors.Wrapf(types.ErrInvalidClient, "client is not in the same version as the chain. expected version: %d, got: %d",
			tmClient.LatestHeight.VersionNumber, version)
	}

	selfHeight := types.NewHeight(version, uint64(ctx.BlockHeight()))
	if tmClient.LatestHeight.GT(selfHeight) {
		return sdkerrors.Wrapf(types.ErrInvalidClient, "client has LatestHeight %d greater than chain height %d",
			tmClient.LatestHeight, ctx.BlockHeight())
	}

	// consensus params must match consensus params on executing chain
	expectedConsensusParams := ctx.ConsensusParams()
	if !reflect.DeepEqual(expectedConsensusParams, tmClient.ConsensusParams) {
		return sdkerrors.Wrapf(types.ErrInvalidClient, "client has invalid consensus params, expected: %v got: %v",
			expectedConsensusParams, tmClient.ConsensusParams)
	}

	expectedProofSpecs := commitmenttypes.GetSDKSpecs()
	if !reflect.DeepEqual(expectedProofSpecs, tmClient.ProofSpecs) {
		return sdkerrors.Wrapf(types.ErrInvalidClient, "client has invalid proof specs. expected: %v got: %v",
			expectedProofSpecs, tmClient.ProofSpecs)
	}

	if err := light.ValidateTrustLevel(tmClient.TrustLevel.ToTendermint()); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidClient, "trust-level invalid: %v", err)
	}

	expectedUbdPeriod := k.stakingKeeper.UnbondingTime(ctx)
	if expectedUbdPeriod != tmClient.UnbondingPeriod {
		return sdkerrors.Wrapf(types.ErrInvalidClient, "invalid unbonding period. expected: %s, got: %s",
			expectedUbdPeriod, tmClient.UnbondingPeriod)
	}

	if tmClient.UnbondingPeriod < tmClient.TrustingPeriod {
		return sdkerrors.Wrapf(types.ErrInvalidClient, "unbonding period must be greater than trusting period. unbonding period (%d) < trusting period (%d)",
			tmClient.UnbondingPeriod, tmClient.TrustingPeriod)
	}

	if tmClient.UpgradePath != "" {
		// For now, SDK IBC implementation assumes that upgrade path (if defined) is defined by SDK upgrade module
		// Must escape any merkle key before adding it to upgrade path
		upgradeKey := url.PathEscape(upgradetypes.KeyUpgradedClient)
		expectedUpgradePath := fmt.Sprintf("%s/%s", upgradetypes.StoreKey, upgradeKey)
		if tmClient.UpgradePath != expectedUpgradePath {
			return sdkerrors.Wrapf(types.ErrInvalidClient, "upgrade path must be the upgrade path defined by upgrade module. expected %s, got %s",
				expectedUpgradePath, tmClient.UpgradePath)
		}
	}
	return nil
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
