package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// ClientKeeper expected account IBC client keeper
type ClientKeeper interface {
	GetConsensusState(ctx sdk.Context, clientID string) (clientexported.ConsensusState, bool)
}

// ConnectionKeeper expected account IBC connection keeper
type ConnectionKeeper interface {
	GetConnection(ctx sdk.Context, connectionID string) (connection.End, bool)
	VerifyMembership(
		ctx sdk.Context, connection connection.End, height uint64,
		proof commitment.ProofI, path string, value []byte,
	) bool
	VerifyNonMembership(
		ctx sdk.Context, connection connection.End, height uint64,
		proof commitment.ProofI, path string,
	) bool
}

// PortKeeper expected account IBC port keeper
type PortKeeper interface {
	GetPort(ctx sdk.Context, portID string) (sdk.CapabilityKey, bool)
	Authenticate(key sdk.CapabilityKey, portID string) bool
}
