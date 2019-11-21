package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// ClientKeeper expected account IBC client keeper
type ClientKeeper interface {
	GetConsensusState(ctx sdk.Context, clientID string) (clientexported.ConsensusState, bool)
	GetClientState(ctx sdk.Context, clientID string) (client.State, bool)
	GetVerifiedRoot(ctx sdk.Context, clientID string, height uint64) (commitment.RootI, bool)
}

// ConnectionKeeper expected account IBC connection keeper
type ConnectionKeeper interface {
	GetConnection(ctx sdk.Context, connectionID string) (connection.ConnectionEnd, bool)
}

// PortKeeper expected account IBC port keeper
type PortKeeper interface {
	Authenticate(key sdk.CapabilityKey, portID string) bool
}
