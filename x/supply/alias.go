package supply

import (
	"github.com/cosmos/cosmos-sdk/x/supply/keeper"
	"github.com/cosmos/cosmos-sdk/x/supply/types"
)

// nolint
type (
	Keeper      = keeper.Keeper
	Supplier    = types.Supplier
	TokenHolder = types.TokenHolder
)

// nolint
var (
	NewKeeper         = keeper.NewKeeper
	GetTokenHolderKey = keeper.GetTokenHolderKey
)

// nolint
const (
	StoreKey     = keeper.StoreKey
	QuerierRoute = keeper.QuerierRoute
)
