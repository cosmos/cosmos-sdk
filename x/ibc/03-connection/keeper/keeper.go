package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
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
		storeKey:     key,
		cdc:          cdc,
		codespace:    sdk.CodespaceType(fmt.Sprintf("%s/%s", codespace, types.DefaultCodespace)), // "ibc/connections",
		prefix:       []byte(types.SubModuleName + "/"),                                          // "connections/"
		clientKeeper: ck,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s/%s", ibctypes.ModuleName, types.SubModuleName))
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
		return types.ErrClientConnectionPathsNotFound(k.codespace)
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
		return types.ErrClientConnectionPathsNotFound(k.codespace)
	}

	conns, ok := removePath(conns, connectionID)
	if !ok {
		return types.ErrConnectionPath(k.codespace)
	}

	k.SetClientConnectionPaths(ctx, clientID, conns)
	return nil
}

func (k Keeper) verifyMembership(
	ctx sdk.Context,
	connection types.ConnectionEnd,
	height uint64,
	proof ics23.Proof,
	path string,
	value interface{}, // value: Value
) bool {
	_, found := k.clientKeeper.GetClientState(ctx, connection.ClientID)
	if !found {
		return false
	}
	// k.clientKeeper.VerifyMembership(ctx, clientState, height, proof, applyPrefix(connection.Counterparty.Prefix, path), value)
	return true
}

func (k Keeper) verifyNonMembership(
	ctx sdk.Context,
	connection types.ConnectionEnd,
	height uint64,
	proof ics23.Proof,
	path string,
) bool {
	_, found := k.clientKeeper.GetClientState(ctx, connection.ClientID)
	if !found {
		return false
	}
	// k.clientKeeper.VerifyNonMembership(ctx, clientState, height, proof, applyPrefix(connection.Counterparty.Prefix, path))
	return true
}

func (k Keeper) getCompatibleVersions() []string {
	// TODO:
	return nil
}

func (k Keeper) pickVersion(counterpartyVersions []string) string {
	// TODO:
	return ""
}

// checkVersion is an opaque function defined by the host state machine which
// determines if two versions are compatible
func checkVersion(version, counterpartyVersion string) bool {
	// TODO:
	return true
}

// removePath is an util function to remove a path from a set.
//
// TODO: move to ICS24
func removePath(paths []string, path string) ([]string, bool) {
	for i, p := range paths {
		if p == path {
			return append(paths[:i], paths[i+1:]...), true
		}
	}
	return paths, false
}
