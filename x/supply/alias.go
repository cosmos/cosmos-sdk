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
	NewKeeper          = keeper.NewKeeper
	RegisterInvariants = keeper.RegisterInvariants

	NewSupplier         = types.NewSupplier
	DefaultSupplier     = types.DefaultSupplier
	NewGenesisState     = types.NewGenesisState
	DefaultGenesisState = types.DefaultGenesisState
)

const (
	StoreKey = keeper.StoreKey

	TypeCirculating = types.TypeCirculating
	TypeVesting     = types.TypeVesting
	TypeModules     = types.TypeModules
	TypeLiquid      = types.TypeLiquid
	TypeTotal       = types.TypeTotal
)
