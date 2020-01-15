package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
)

// ClientKeeper expected account IBC client keeper
type ClientKeeper interface {
	// GetConsensusState(ctx sdk.Context, clientID string) (clientexported.ConsensusState, bool)
}

// ConnectionKeeper expected account IBC connection keeper
type ConnectionKeeper interface {
	GetConnection(ctx sdk.Context, connectionID string) (connectionexported.ConnectionI, bool)
}

// PortKeeper expected account IBC port keeper
type PortKeeper interface {
	Authenticate(key sdk.CapabilityKey, portID string) bool
}
