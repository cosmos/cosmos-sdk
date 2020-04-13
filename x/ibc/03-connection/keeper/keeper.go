package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// Keeper defines the IBC connection keeper
type Keeper struct {
	storeKey     sdk.StoreKey
	cdc          *codec.Codec
	clientKeeper types.ClientKeeper
}

// NewKeeper creates a new IBC connection Keeper instance
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, ck types.ClientKeeper) Keeper {
	return Keeper{
		storeKey:     key,
		cdc:          cdc,
		clientKeeper: ck,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s/%s", ibctypes.ModuleName, types.SubModuleName))
}

// GetCommitmentPrefix returns the IBC connection store prefix as a commitment
// Prefix
func (k Keeper) GetCommitmentPrefix() commitmentexported.Prefix {
	return commitmenttypes.NewMerklePrefix([]byte(k.storeKey.Name()))
}

// GetConnection returns a connection with a particular identifier
func (k Keeper) GetConnection(ctx sdk.Context, connectionID string) (types.ConnectionEnd, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(ibctypes.KeyConnection(connectionID))
	if bz == nil {
		return types.ConnectionEnd{}, false
	}

	var connection types.ConnectionEnd
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &connection)
	return connection, true
}

// SetConnection sets a connection to the store
func (k Keeper) SetConnection(ctx sdk.Context, connectionID string, connection types.ConnectionEnd) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(connection)
	store.Set(ibctypes.KeyConnection(connectionID), bz)
}

// GetClientConnectionPaths returns all the connection paths stored under a
// particular client
func (k Keeper) GetClientConnectionPaths(ctx sdk.Context, clientID string) ([]string, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(ibctypes.KeyClientConnections(clientID))
	if bz == nil {
		return nil, false
	}

	var paths []string
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &paths)
	return paths, true
}

// SetClientConnectionPaths sets the connections paths for client
func (k Keeper) SetClientConnectionPaths(ctx sdk.Context, clientID string, paths []string) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(paths)
	store.Set(ibctypes.KeyClientConnections(clientID), bz)
}

// IterateConnections provides an iterator over all ConnectionEnd objects.
// For each ConnectionEnd, cb will be called. If the cb returns true, the
// iterator will close and stop.
func (k Keeper) IterateConnections(ctx sdk.Context, cb func(types.IdentifiedConnectionEnd) bool) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, ibctypes.KeyConnectionPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var connection types.ConnectionEnd
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &connection)
		identifier := string(iterator.Key()[len(ibctypes.KeyConnectionPrefix)+1:])

		conn := types.IdentifiedConnectionEnd{
			Connection: connection,
			Identifier: identifier,
		}

		if cb(conn) {
			break
		}
	}
}

// GetAllConnections returns all stored ConnectionEnd objects.
func (k Keeper) GetAllConnections(ctx sdk.Context) (connections []types.IdentifiedConnectionEnd) {
	k.IterateConnections(ctx, func(connection types.IdentifiedConnectionEnd) bool {
		connections = append(connections, connection)
		return false
	})
	return connections
}

// addConnectionToClient is used to add a connection identifier to the set of
// connections associated with a client.
func (k Keeper) addConnectionToClient(ctx sdk.Context, clientID, connectionID string) error {
	_, found := k.clientKeeper.GetClientState(ctx, clientID)
	if !found {
		return clienttypes.ErrClientNotFound
	}

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
// nolint: unused
func (k Keeper) removeConnectionFromClient(ctx sdk.Context, clientID, connectionID string) error {
	conns, found := k.GetClientConnectionPaths(ctx, clientID)
	if !found {
		return sdkerrors.Wrap(types.ErrClientConnectionPathsNotFound, clientID)
	}

	conns, ok := host.RemovePath(conns, connectionID)
	if !ok {
		return sdkerrors.Wrap(types.ErrConnectionPath, clientID)
	}

	k.SetClientConnectionPaths(ctx, clientID, conns)
	return nil
}
