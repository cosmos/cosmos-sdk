package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

// ClientKeeper expected account IBC client keeper
type ClientKeeper interface {
	GetClientState(ctx sdk.Context, clientID string) (clientexported.ClientState, bool)
	GetClientConsensusState(ctx sdk.Context, clientID string, height clientexported.Height) (clientexported.ConsensusState, bool)
	GetSelfConsensusState(ctx sdk.Context, height clientexported.Height) (clientexported.ConsensusState, bool)
	ValidateSelfClient(ctx sdk.Context, clientState clientexported.ClientState) error
	IterateClients(ctx sdk.Context, cb func(string, clientexported.ClientState) bool)
	ClientStore(ctx sdk.Context, clientID string) sdk.KVStore
}
