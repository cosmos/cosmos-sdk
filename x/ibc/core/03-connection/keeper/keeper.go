package keeper

import (
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
)

// Keeper defines the IBC connection keeper
type Keeper struct {
	// implements gRPC QueryServer interface
	types.QueryServer

	storeKey     sdk.StoreKey
	cdc          codec.BinaryMarshaler
	clientKeeper types.ClientKeeper
}

// NewKeeper creates a new IBC connection Keeper instance
func NewKeeper(cdc codec.BinaryMarshaler, key sdk.StoreKey, ck types.ClientKeeper) Keeper {
	return Keeper{
		storeKey:     key,
		cdc:          cdc,
		clientKeeper: ck,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+host.ModuleName+"/"+types.SubModuleName)
}

// GetCommitmentPrefix returns the IBC connection store prefix as a commitment
// Prefix
func (k Keeper) GetCommitmentPrefix() exported.Prefix {
	return commitmenttypes.NewMerklePrefix([]byte(k.storeKey.Name()))
}

// GenerateConnectionIdentifier returns the next connection identifier.
func (k Keeper) GenerateConnectionIdentifier(ctx sdk.Context) string {
	nextConnSeq := k.GetNextConnectionSequence(ctx)
	connectionID := types.FormatConnectionIdentifier(nextConnSeq)

	nextConnSeq++
	k.SetNextConnectionSequence(ctx, nextConnSeq)
	return connectionID
}

// GetConnection returns a connection with a particular identifier
func (k Keeper) GetConnection(ctx sdk.Context, connectionID string) (types.ConnectionEnd, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(host.ConnectionKey(connectionID))
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
	store.Set(host.ConnectionKey(connectionID), bz)
}

// GetTimestampAtHeight returns the timestamp in nanoseconds of the consensus state at the
// given height.
func (k Keeper) GetTimestampAtHeight(ctx sdk.Context, connection types.ConnectionEnd, height exported.Height) (uint64, error) {
	consensusState, found := k.clientKeeper.GetClientConsensusState(
		ctx, connection.GetClientID(), height,
	)

	if !found {
		return 0, sdkerrors.Wrapf(
			clienttypes.ErrConsensusStateNotFound,
			"clientID (%s), height (%s)", connection.GetClientID(), height,
		)
	}

	return consensusState.GetTimestamp(), nil
}

// GetClientConnectionPaths returns all the connection paths stored under a
// particular client
func (k Keeper) GetClientConnectionPaths(ctx sdk.Context, clientID string) ([]string, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(host.ClientConnectionsKey(clientID))
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
	store.Set(host.ClientConnectionsKey(clientID), bz)
}

// GetNextConnectionSequence gets the next connection sequence from the store.
func (k Keeper) GetNextConnectionSequence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(types.KeyNextConnectionSequence))
	if bz == nil {
		panic("next connection sequence is nil")
	}

	return sdk.BigEndianToUint64(bz)
}

// SetNextConnectionSequence sets the next connection sequence to the store.
func (k Keeper) SetNextConnectionSequence(ctx sdk.Context, sequence uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := sdk.Uint64ToBigEndian(sequence)
	store.Set([]byte(types.KeyNextConnectionSequence), bz)
}

// GetAllClientConnectionPaths returns all stored clients connection id paths. It
// will ignore the clients that haven't initialized a connection handshake since
// no paths are stored.
func (k Keeper) GetAllClientConnectionPaths(ctx sdk.Context) []types.ConnectionPaths {
	var allConnectionPaths []types.ConnectionPaths
	k.clientKeeper.IterateClients(ctx, func(clientID string, cs exported.ClientState) bool {
		paths, found := k.GetClientConnectionPaths(ctx, clientID)
		if !found {
			// continue when connection handshake is not initialized
			return false
		}
		connPaths := types.NewConnectionPaths(clientID, paths)
		allConnectionPaths = append(allConnectionPaths, connPaths)
		return false
	})

	return allConnectionPaths
}

// IterateConnections provides an iterator over all ConnectionEnd objects.
// For each ConnectionEnd, cb will be called. If the cb returns true, the
// iterator will close and stop.
func (k Keeper) IterateConnections(ctx sdk.Context, cb func(types.IdentifiedConnection) bool) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, []byte(host.KeyConnectionPrefix))

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var connection types.ConnectionEnd
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &connection)

		connectionID := host.MustParseConnectionPath(string(iterator.Key()))
		identifiedConnection := types.NewIdentifiedConnection(connectionID, connection)
		if cb(identifiedConnection) {
			break
		}
	}
}

// GetAllConnections returns all stored ConnectionEnd objects.
func (k Keeper) GetAllConnections(ctx sdk.Context) (connections []types.IdentifiedConnection) {
	k.IterateConnections(ctx, func(connection types.IdentifiedConnection) bool {
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
		return sdkerrors.Wrap(clienttypes.ErrClientNotFound, clientID)
	}

	conns, found := k.GetClientConnectionPaths(ctx, clientID)
	if !found {
		conns = []string{}
	}

	conns = append(conns, connectionID)
	k.SetClientConnectionPaths(ctx, clientID, conns)
	return nil
}
