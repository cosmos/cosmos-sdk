// nolint
package supply

import (
	"github.com/cosmos/cosmos-sdk/x/supply/keeper"
	"github.com/cosmos/cosmos-sdk/x/supply/types"
)

type (
	Keeper              = keeper.Keeper
	ModuleAccount       = types.ModuleAccount
	ModuleHolderAccount = types.ModuleHolderAccount
	ModuleMinterAccount = types.ModuleMinterAccount
)

var (
	NewKeeper = keeper.NewKeeper

	RegisterInvariants     = keeper.RegisterInvariants
	AllInvariants          = keeper.AllInvariants
	StakingTokensInvariant = keeper.StakingTokensInvariant

	NewModuleHolderAccount = types.NewModuleHolderAccount
	NewModuleMinterAccount = types.NewModuleMinterAccount
)
