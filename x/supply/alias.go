package supply

// DONTCOVER
// nolint

import (
	"github.com/cosmos/cosmos-sdk/x/supply/keeper"
	"github.com/cosmos/cosmos-sdk/x/supply/types"
)

const (
	ModuleName   = types.ModuleName
	StoreKey     = types.StoreKey
	RouterKey    = types.RouterKey
	QuerierRoute = types.QuerierRoute
	Minter       = types.Minter
	Burner       = types.Burner
	Staking      = types.Staking
)

var (
	// functions aliases
	RegisterInvariants    = keeper.RegisterInvariants
	AllInvariants         = keeper.AllInvariants
	TotalSupply           = keeper.TotalSupply
	NewKeeper             = keeper.NewKeeper
	NewQuerier            = keeper.NewQuerier
	SupplyKey             = keeper.SupplyKey
	NewModuleAddress      = types.NewModuleAddress
	NewEmptyModuleAccount = types.NewEmptyModuleAccount
	NewModuleAccount      = types.NewModuleAccount
	RegisterCodec         = types.RegisterCodec
	NewGenesisState       = types.NewGenesisState
	DefaultGenesisState   = types.DefaultGenesisState
	NewSupply             = types.NewSupply
	DefaultSupply         = types.DefaultSupply

	// variable aliases
	ModuleCdc = types.ModuleCdc
)

type (
	Keeper        = keeper.Keeper
	ModuleAccount = types.ModuleAccount
	GenesisState  = types.GenesisState
	Supply        = types.Supply
	Codec         = types.Codec
)
