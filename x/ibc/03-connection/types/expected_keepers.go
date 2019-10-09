package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ics02exported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	ics02types "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// ClientKeeper expected account IBC client keeper
type ClientKeeper interface {
	GetConsensusState(ctx sdk.Context, clientID string) (ics02exported.ConsensusState, bool)
	GetClientState(ctx sdk.Context, clientID string) (ics02types.ClientState, bool)
	VerifyMembership(
		ctx sdk.Context, clientState ics02types.ClientState, height uint64,
		proof ics23.Proof, path string, value []byte,
	) bool
	VerifyNonMembership(
		ctx sdk.Context, clientState ics02types.ClientState, height uint64,
		proof ics23.Proof, path string,
	) bool
}
