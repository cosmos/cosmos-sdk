package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// ClientKeeper expected account IBC client keeper
type ClientKeeper interface {
	GetCommitmentPath() commitment.Prefix
	GetConsensusState(ctx sdk.Context, clientID string) (clientexported.ConsensusState, bool)
	GetClientState(ctx sdk.Context, clientID string) (clienttypes.ClientState, bool)
	VerifyMembership(
		ctx sdk.Context, clientState clienttypes.ClientState, height uint64,
		proof commitment.Proof, path string, value []byte,
	) bool
	VerifyNonMembership(
		ctx sdk.Context, clientState clienttypes.ClientState, height uint64,
		proof commitment.Proof, path string,
	) bool
}
