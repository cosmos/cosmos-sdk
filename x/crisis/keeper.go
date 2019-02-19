package crisis

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Crisis keeper
type Keeper struct {
	routes     []InvarRoute
	bankKeeper BankKeeper
}

// register routes for the
func (k *Keeper) RegisterRoute(ctx sdk.Context, invarRoute InvarRoute) {
	k.routes = append(k.routes, invarRoute)
}

// expected bank keeper
type BankKeeper interface {
	AddCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Tags, sdk.Error)
}
