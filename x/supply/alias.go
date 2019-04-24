// nolint
package supply

import (
	"github.com/cosmos/cosmos-sdk/x/supply/keeper"
	"github.com/cosmos/cosmos-sdk/x/supply/types"
)

type (
	Keeper       = keeper.Keeper
	Supplier     = types.Supplier
	GenesisState = types.GenesisState
)

var (
	NewKeeper           = keeper.NewKeeper
	NewGenesisState     = types.NewGenesisState
	DefaultGenesisState = types.DefaultGenesisState
)

const (
	StoreKey = keeper.StoreKey
	// QuerierRoute = keeper.QuerierRoute

	TypeCirculating = types.TypeCirculating
	TypeVesting     = types.TypeVesting
	TypeModules     = types.TypeModules
	TypeTotal       = types.TypeTotal
)
