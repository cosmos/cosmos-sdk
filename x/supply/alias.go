// nolint
package supply

import (
	"github.com/cosmos/cosmos-sdk/x/supply/keeper"
	"github.com/cosmos/cosmos-sdk/x/supply/types"
)

type (
	Keeper       = keeper.Keeper
	Supplier     = types.Supplier
	CoinsSupply  = types.CoinsSupply
	CoinSupply   = types.CoinSupply
	GenesisState = types.GenesisState
)

var (
	NewKeeper          = keeper.NewKeeper
	RegisterInvariants = keeper.RegisterInvariants

	NewSupplier                = types.NewSupplier
	DefaultSupplier            = types.DefaultSupplier
	NewCoinsSupply             = types.NewCoinsSupply
	NewCoinsSupplyFromSupplier = types.NewCoinsSupplyFromSupplier
	NewCoinSupply              = types.NewCoinSupply
	NewCoinSupplyFromSupplier  = types.NewCoinSupplyFromSupplier
	NewGenesisState            = types.NewGenesisState
	DefaultGenesisState        = types.DefaultGenesisState
)

const (
	StoreKey = keeper.StoreKey

	TypeCirculating = types.TypeCirculating
	TypeVesting     = types.TypeVesting
	TypeModules     = types.TypeModules
	TypeTotal       = types.TypeTotal
)
