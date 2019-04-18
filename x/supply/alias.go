// nolint
package supply

import (
	"github.com/cosmos/cosmos-sdk/x/supply/keeper"
	"github.com/cosmos/cosmos-sdk/x/supply/types"
)

type (
	Keeper   = keeper.Keeper
	Supplier = types.Supplier
)

var (
	NewKeeper = keeper.NewKeeper
)

const (
	StoreKey = keeper.StoreKey
	// QuerierRoute = keeper.QuerierRoute

	TypeCirculating = types.TypeCirculating
	TypeVesting     = types.TypeVesting
	TypeModules     = types.TypeModules
)
