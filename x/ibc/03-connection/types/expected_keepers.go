package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

// ClientKeeper expected account IBC client keeper
type ClientKeeper interface {
	GetConsensusState(ctx sdk.Context, clientID string) (clientexported.ConsensusState, bool)
	GetClientState(ctx sdk.Context, clientID string) (clientexported.ClientState, bool)
}
