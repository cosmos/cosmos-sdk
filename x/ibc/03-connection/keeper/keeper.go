package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// Keeper defines the IBC connection keeper
type Keeper struct {
	storeKey     sdk.StoreKey
	aminoCdc     *codec.Codec    // amino codec. TODO: remove after clients have been migrated to proto
	cdc          codec.Marshaler // hybrid codec
	clientKeeper types.ClientKeeper
}

// NewKeeper creates a new IBC connection Keeper instance
func NewKeeper(aminoCdc *codec.Codec, cdc codec.Marshaler, key sdk.StoreKey, ck types.ClientKeeper) Keeper {
	return Keeper{
		storeKey:     key,
		aminoCdc:     aminoCdc,
		cdc:          cdc,
		clientKeeper: ck,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s/%s", host.ModuleName, types.SubModuleName))
}

// GetCommitmentPrefix returns the IBC connection store prefix as a commitment
// Prefix
func (k Keeper) GetCommitmentPrefix(t commitmentexported.Type) commitmentexported.Prefix {
	switch t {
	case commitmentexported.Merkle:
		return commitmenttypes.NewMerklePrefix([]byte(k.storeKey.Name()))
	case commitmentexported.Signature:
		return commitmenttypes.NewSignaturePrefix([]byte(k.storeKey.Name()))
	default:
		return nil
	}
}

// GetConnection returns a connection with a particular identifier
func (k Keeper) GetConnection(ctx sdk.Context, connectionID string) (types.ConnectionEnd, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(host.KeyConnection(connectionID))
	if bz == nil {
		return types.ConnectionEnd{}, false
	}

	var connection types.ConnectionEnd
	k.cdc.MustUnmarshalBinaryBare(bz, &connection)

	return connection, true
}

// SetConnection sets a connection to the store
func (k Keeper) SetConnection(ctx sdk.Context, connectionID string, connection types.ConnectionEnd) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(&connection)
	store.Set(host.KeyConnection(connectionID), bz)
}

// GetTimestampAtHeight returns the timestamp in nanoseconds of the consensus state at the
// given height.
func (k Keeper) GetTimestampAtHeight(ctx sdk.Context, connection types.ConnectionEnd, height uint64) (uint64, error) {
	consensusState, found := k.clientKeeper.GetClientConsensusState(
		ctx, connection.GetClientID(), height,
	)

	if !found {
		return 0, sdkerrors.Wrapf(
			clienttypes.ErrConsensusStateNotFound,
			"clientID (%s), height (%d)", connection.GetClientID(), height,
		)
	}

	return consensusState.GetTimestamp(), nil
}

// GetClientConnectionPaths returns all the connection paths stored under a
// particular client
func (k Keeper) GetClientConnectionPaths(ctx sdk.Context, clientID string) ([]string, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(host.KeyClientConnections(clientID))
	if bz == nil {
		return nil, false
	}

	var clientPaths types.ClientPaths
	k.cdc.MustUnmarshalBinaryBare(bz, &clientPaths)
	return clientPaths.Paths, true
}

// SetClientConnectionPaths sets the connections paths for client
func (k Keeper) SetClientConnectionPaths(ctx sdk.Context, clientID string, paths []string) {
	store := ctx.KVStore(k.storeKey)
	clientPaths := types.ClientPaths{Paths: paths}
	bz := k.cdc.MustMarshalBinaryBare(&clientPaths)
	store.Set(host.KeyClientConnections(clientID), bz)
}

// GetAllClientConnectionPaths returns all stored clients connection id paths. It
// will ignore the clients that haven't initialized a connection handshake since
// no paths are stored.
func (k Keeper) GetAllClientConnectionPaths(ctx sdk.Context) []types.ConnectionPaths {
	var allConnectionPaths []types.ConnectionPaths
	k.clientKeeper.IterateClients(ctx, func(cs clientexported.ClientState) bool {
		paths, found := k.GetClientConnectionPaths(ctx, cs.GetID())
		if !found {
			// continue when connection handshake is not initialized
			return false
		}
		connPaths := types.NewConnectionPaths(cs.GetID(), paths)
		allConnectionPaths = append(allConnectionPaths, connPaths)
		return false
	})

	return allConnectionPaths
}

// IterateConnections provides an iterator over all ConnectionEnd objects.
// For each ConnectionEnd, cb will be called. If the cb returns true, the
// iterator will close and stop.
func (k Keeper) IterateConnections(ctx sdk.Context, cb func(types.ConnectionEnd) bool) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, host.KeyConnectionPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var connection types.ConnectionEnd
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &connection)

		if cb(connection) {
			break
		}
	}
}

// GetAllConnections returns all stored ConnectionEnd objects.
func (k Keeper) GetAllConnections(ctx sdk.Context) (connections []types.ConnectionEnd) {
	k.IterateConnections(ctx, func(connection types.ConnectionEnd) bool {
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
