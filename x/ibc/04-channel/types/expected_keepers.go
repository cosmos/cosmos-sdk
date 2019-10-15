package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// ClientKeeper expected account IBC client keeper
type ClientKeeper interface {
	GetConsensusState(ctx sdk.Context, clientID string) (clientexported.ConsensusState, bool)
}

// ConnectionKeeper expected account IBC connection keeper
type ConnectionKeeper interface {
	GetConnection(ctx sdk.Context, connectionID string) (connectiontypes.ConnectionEnd, bool)
	VerifyMembership(
		ctx sdk.Context, connection connectiontypes.ConnectionEnd, height uint64,
		proof ics23.Proof, path string, value []byte,
	) bool
	VerifyNonMembership(
		ctx sdk.Context, connection connectiontypes.ConnectionEnd, height uint64,
		proof ics23.Proof, path string,
	) bool
}

// PortKeeper expected account IBC port keeper
type PortKeeper interface {
	GetPort(ctx sdk.Context, portID string) (sdk.CapabilityKey, bool)
}
