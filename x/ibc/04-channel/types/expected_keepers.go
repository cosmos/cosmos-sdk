package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	// connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
)

// ConnectionKeeper expected account IBC connection keeper
type ConnectionKeeper interface {
	// GetConnection(ctx sdk.Context, connectionID string) (connectiontypes.ConnectionEnd, bool)
}

// PortKeeper expected account IBC port keeper
type PortKeeper interface {
	Authenticate(key sdk.CapabilityKey, portID string) bool
}
