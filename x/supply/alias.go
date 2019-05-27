// nolint
package supply

import (
	"github.com/cosmos/cosmos-sdk/x/supply/keeper"
	"github.com/cosmos/cosmos-sdk/x/supply/types"
)

type (
	Keeper              = keeper.Keeper
	AccountKeeper       = keeper.AccountKeeper
	ModuleAccount       = types.ModuleAccount
	ModuleHolderAccount = types.ModuleHolderAccount
	ModuleMinterAccount = types.ModuleMinterAccount
	Supply              = types.Supply
	GenesisState        = types.GenesisState
)

var (
	NewKeeper          = keeper.NewKeeper
	RegisterInvariants = keeper.RegisterInvariants
	AllInvariants      = keeper.AllInvariants
	DefaultCodespace   = keeper.DefaultCodespace

	NewModuleHolderAccount = types.NewModuleHolderAccount
	NewModuleMinterAccount = types.NewModuleMinterAccount
	NewSupply              = types.NewSupply
	DefaultSupply          = types.DefaultSupply
	NewGenesisState        = types.NewGenesisState
	DefaultGenesisState    = types.DefaultGenesisState
	RegisterCodec          = types.RegisterCodec
	ModuleCdc              = types.ModuleCdc
)

const (
	ModuleName   = types.ModuleName
	StoreKey     = types.StoreKey
	RouterKey    = types.RouterKey
	QuerierRoute = types.QuerierRoute
)
