package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
)

// ClientKeeper expected account IBC client keeper
type ClientKeeper interface {
	GetConsensusState(ctx sdk.Context, clientID string) (clientexported.ConsensusState, bool)
	GetClientState(ctx sdk.Context, clientID string) (client.ClientState, bool)
	VerifyMembership(
		ctx sdk.Context, clientState client.ClientState, height uint64,
		proof commitmentexported.ProofI, path commitmentexported.PathI, value []byte,
	) bool
	VerifyNonMembership(
		ctx sdk.Context, clientState client.ClientState, height uint64,
		proof commitmentexported.ProofI, path commitmentexported.PathI,
	) bool
}
