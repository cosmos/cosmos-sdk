package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// Keeper defines the IBC connection keeper
type Keeper struct {
	storeKey  sdk.StoreKey
	cdc       *codec.Codec
	codespace sdk.CodespaceType
	prefix    []byte // prefix bytes for accessing the store

	clientKeeper types.ClientKeeper
}

// NewKeeper creates a new IBC connection Keeper instance
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, codespace sdk.CodespaceType, ck types.ClientKeeper) Keeper {
	return Keeper{
		storeKey:  key,
		cdc:       cdc,
		codespace: sdk.CodespaceType(fmt.Sprintf("%s/%s", codespace, types.DefaultCodespace)), // "ibc/connection",
		prefix:    []byte{},
		// prefix:       []byte(types.SubModuleName + "/"),                                          // "connection/"
		clientKeeper: ck,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s/%s", ibctypes.ModuleName, types.SubModuleName))
}

// GetCommitmentPrefix returns the IBC connection store prefix as a commitment
// Prefix
func (k Keeper) GetCommitmentPrefix() commitment.PrefixI {
	return commitment.NewPrefix([]byte(k.storeKey.Name()))
}

// GetConnection returns a connection with a particular identifier
func (k Keeper) GetConnection(ctx sdk.Context, connectionID string) (types.ConnectionEnd, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), k.prefix)
	bz := store.Get(types.KeyConnection(connectionID))
	if bz == nil {
		return types.ConnectionEnd{}, false
	}

	var connection types.ConnectionEnd
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &connection)
	return connection, true
}

// SetConnection sets a connection to the store
func (k Keeper) SetConnection(ctx sdk.Context, connectionID string, connection types.ConnectionEnd) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), k.prefix)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(connection)
	store.Set(types.KeyConnection(connectionID), bz)
}

// GetClientConnectionPaths returns all the connection paths stored under a
// particular client
func (k Keeper) GetClientConnectionPaths(ctx sdk.Context, clientID string) ([]string, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), k.prefix)
	bz := store.Get(types.KeyClientConnections(clientID))
	if bz == nil {
		return nil, false
	}

	var paths []string
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &paths)
	return paths, true
}

// SetClientConnectionPaths sets the connections paths for client
func (k Keeper) SetClientConnectionPaths(ctx sdk.Context, clientID string, paths []string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), k.prefix)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(paths)
	store.Set(types.KeyClientConnections(clientID), bz)
}

// addConnectionToClient is used to add a connection identifier to the set of
// connections associated with a client.
//
// CONTRACT: client must already exist
func (k Keeper) addConnectionToClient(ctx sdk.Context, clientID, connectionID string) sdk.Error {
	conns, found := k.GetClientConnectionPaths(ctx, clientID)
	if !found {
		conns = []string{}
	}

	conns = append(conns, connectionID)
	k.SetClientConnectionPaths(ctx, clientID, conns)
	return nil
}

// removeConnectionFromClient is used to remove a connection identifier from the
// set of connections associated with a client.
//
// CONTRACT: client must already exist
func (k Keeper) removeConnectionFromClient(ctx sdk.Context, clientID, connectionID string) sdk.Error {
	conns, found := k.GetClientConnectionPaths(ctx, clientID)
	if !found {
		return types.ErrClientConnectionPathsNotFound(k.codespace, clientID)
	}

	conns, ok := host.RemovePath(conns, connectionID)
	if !ok {
		return types.ErrConnectionPath(k.codespace)
	}

	k.SetClientConnectionPaths(ctx, clientID, conns)
	return nil
}

// VerifyMembership helper function for state membership verification
func (k Keeper) VerifyMembership(
	ctx sdk.Context,
	connection types.ConnectionEnd,
	height uint64,
	proof commitment.ProofI,
	pathStr string,
	value []byte,
) bool {
	// FIXME: commented out for demo
	/*
		clientState, found := k.clientKeeper.GetClientState(ctx, connection.ClientID)
		if !found {
			return false
		}
	*/
	path, err := commitment.ApplyPrefix(connection.Counterparty.Prefix, pathStr)
	if err != nil {
		return false
	}

	return k.clientKeeper.VerifyMembership(ctx, connection.ClientID, height, proof, path, value)
}

// VerifyNonMembership helper function for state non-membership verification
func (k Keeper) VerifyNonMembership(
	ctx sdk.Context,
	connection types.ConnectionEnd,
	height uint64,
	proof commitment.ProofI,
	pathStr string,
) bool {
	// FIXME: commented out for demo
	/*
		clientState, found := k.clientKeeper.GetClientState(ctx, connection.ClientID)
		if !found {
			return false
		}
	*/

	path, err := commitment.ApplyPrefix(connection.Counterparty.Prefix, pathStr)
	if err != nil {
		return false
	}

	return k.clientKeeper.VerifyNonMembership(ctx, connection.ClientID, height, proof, path)
}
