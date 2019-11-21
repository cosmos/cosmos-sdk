package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// ClientKeeper expected account IBC client keeper
type ClientKeeper interface {
	GetConsensusState(ctx sdk.Context, clientID string) (clientexported.ConsensusState, bool)
	GetClientState(ctx sdk.Context, clientID string) (client.State, bool)
	GetVerifiedRoot(ctx sdk.Context, clientID string, height uint64) (commitment.RootI, bool)
	VerifyClientConsensusState(
		ctx sdk.Context,
		clientState client.State,
		height uint64, // sequence
		proof commitment.ProofI,
		prefix commitment.PrefixI,
		consensusState clientexported.ConsensusState,
	) (bool, error)
}
