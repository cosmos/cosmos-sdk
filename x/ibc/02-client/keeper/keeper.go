package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// Keeper represents a type that grants read and write permissions to any client
// state information
type Keeper struct {
	storeKey  sdk.StoreKey
	cdc       *codec.Codec
	codespace sdk.CodespaceType
	prefix    []byte // prefix bytes for accessing the store
}

// NewKeeper creates a new NewKeeper instance
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, codespace sdk.CodespaceType) Keeper {
	return Keeper{
		storeKey:  key,
		cdc:       cdc,
		codespace: sdk.CodespaceType(fmt.Sprintf("%s/%s", codespace, types.DefaultCodespace)), // "ibc/client",
		prefix:    []byte{},
		// prefix:    []byte(types.SubModuleName + "/"),                                          // "client/"
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s/%s", ibctypes.ModuleName, types.SubModuleName))
}

// GetClientState gets a particular client from the store
func (k Keeper) GetClientState(ctx sdk.Context, clientID string) (types.State, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), k.prefix)
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
	store := prefix.NewStore(ctx.KVStore(k.storeKey), k.prefix)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(clientState)
	store.Set(types.KeyClientState(clientState.ID()), bz)
}

// GetClientType gets the consensus type for a specific client
func (k Keeper) GetClientType(ctx sdk.Context, clientID string) (exported.ClientType, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), k.prefix)
	bz := store.Get(types.KeyClientType(clientID))
	if bz == nil {
		return 0, false
	}

	return exported.ClientType(bz[0]), true
}

// SetClientType sets the specific client consensus type to the provable store
func (k Keeper) SetClientType(ctx sdk.Context, clientID string, clientType exported.ClientType) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), k.prefix)
	store.Set(types.KeyClientType(clientID), []byte{byte(clientType)})
}

// GetConsensusState creates a new client state and populates it with a given consensus state
func (k Keeper) GetConsensusState(ctx sdk.Context, clientID string) (exported.ConsensusState, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), k.prefix)
	bz := store.Get(types.KeyConsensusState(clientID))
	if bz == nil {
		return nil, false
	}

	var consensusState exported.ConsensusState
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &consensusState)
	return consensusState, true
}

// SetConsensusState sets a ConsensusState to a particular client
func (k Keeper) SetConsensusState(ctx sdk.Context, clientID string, consensusState exported.ConsensusState) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), k.prefix)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(consensusState)
	store.Set(types.KeyConsensusState(clientID), bz)
}

// GetVerifiedRoot gets a verified commitment Root from a particular height to
// a client
func (k Keeper) GetVerifiedRoot(ctx sdk.Context, clientID string, height uint64) (commitment.RootI, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), k.prefix)

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
	store := prefix.NewStore(ctx.KVStore(k.storeKey), k.prefix)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(root)
	store.Set(types.KeyRoot(clientID, height), bz)
}

// State returns a new client state with a given id as defined in
// https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#example-implementation
func (k Keeper) initialize(ctx sdk.Context, clientID string, consensusState exported.ConsensusState) types.State {
	clientState := types.NewClientState(clientID)
	k.SetConsensusState(ctx, clientID, consensusState)
	return clientState
}

func (k Keeper) checkMisbehaviour(ctx sdk.Context, evidence exported.Evidence) error {
	// switch evidence.H1().ClientType() {
	// case exported.Tendermint:
	// 	var tmEvidence tendermint.Evidence
	// 	_, ok := evidence.(tendermint.Evidence)
	// 	if !ok {
	// 		return sdkerrors.Wrap(types.ErrInvalidClientType(k.codespace), "consensus type is not Tendermint")
	// 	}
	// 	// TODO: pass past consensus states
	// 	return tendermint.CheckMisbehaviour(tmEvidence)
	// default:
	// 	panic("unregistered consensus type")
	// }
	return nil
}

// freeze updates the state of the client in the event of a misbehaviour
func (k Keeper) freeze(ctx sdk.Context, clientState types.State) (types.State, error) {
	if clientState.Frozen {
		return types.State{}, sdkerrors.Wrap(types.ErrClientFrozen(k.codespace, clientState.ID()), "already frozen")
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
	// XXX: commented out for demo
	/*
		if clientState.Frozen {
			return false
		}
	*/

	root, found := k.GetVerifiedRoot(ctx, clientID, height)
	if !found {
		return false
	}

	res := proof.VerifyMembership(root, path, value)

	return res
}

// VerifyNonMembership state non-membership function defined by the client type
func (k Keeper) VerifyNonMembership(
	ctx sdk.Context,
	clientID string,
	height uint64, // sequence
	proof commitment.ProofI,
	path commitment.PathI,
) bool {
	// XXX: commented out for demo
	/*
		if clientState.Frozen {
			return false
		}
	*/
	root, found := k.GetVerifiedRoot(ctx, clientID, height)
	if !found {
		return false
	}

	return proof.VerifyNonMembership(root, path)
}
