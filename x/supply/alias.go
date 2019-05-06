// nolint
package supply

import (
	"github.com/cosmos/cosmos-sdk/x/supply/keeper"
	"github.com/cosmos/cosmos-sdk/x/supply/types"
)

type (
	SupplyKeeper        = keeper.SupplyKeeper
	SendKeeper          = keeper.SendKeeper
	BaseSendKeeper      = keeper.SendKeeper
	ModuleAccount       = types.ModuleAccount
	ModuleHolderAccount = types.ModuleHolderAccount
	ModuleMinterAccount = types.ModuleMinterAccount
	Supply              = types.Supply
	GenesisState        = types.GenesisState
)

var (
	NewSupplyKeeper   = keeper.NewSupplyKeeper
	NewBaseSendKeeper = keeper.NewBaseSendKeeper

	RegisterInvariants     = keeper.RegisterInvariants
	AllInvariants          = keeper.AllInvariants
	StakingTokensInvariant = keeper.StakingTokensInvariant
	DefaultCodespace       = keeper.DefaultCodespace

	NewModuleHolderAccount = types.NewModuleHolderAccount
	NewModuleMinterAccount = types.NewModuleMinterAccount
	NewSupply              = types.NewSupply
	DefaultSupply          = types.DefaultSupply
	NewGenesisState        = types.NewGenesisState
	DefaultGenesisState    = types.DefaultGenesisState
)

const (
	DefaultParamspace = keeper.DefaultParamspace
	StoreKey          = keeper.StoreKey
)
