package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ics03types "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// ConnectionKeeper expected account IBC connection keeper
type ConnectionKeeper interface {
	GetConnection(ctx sdk.Context, connectionID string) (ics03types.ConnectionEnd, bool)
	VerifyMembership(
		ctx sdk.Context, connection ics03types.ConnectionEnd, height uint64,
		proof ics23.Proof, path string, value []byte,
	) bool
	VerifyNonMembership(
		ctx sdk.Context, connection ics03types.ConnectionEnd, height uint64,
		proof ics23.Proof, path string,
	) bool
}
