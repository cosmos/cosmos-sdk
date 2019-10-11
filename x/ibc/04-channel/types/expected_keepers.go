package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ics03types "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
)

// ConnectionKeeper expected account IBC connection keeper
type ConnectionKeeper interface {
	GetConnection(ctx sdk.Context, connectionID string) (ics03types.ConnectionEnd, bool)
}
