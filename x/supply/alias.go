// nolint
package supply

import (
	"github.com/cosmos/cosmos-sdk/x/supply/keeper"
	"github.com/cosmos/cosmos-sdk/x/supply/types"
)

type (
	Keeper            = keeper.Keeper
	PoolAccount       = types.PoolAccount
	PoolHolderAccount = types.PoolHolderAccount
	PoolMinterAccount = types.PoolMinterAccount
	Supply            = types.Supply
	GenesisState      = types.GenesisState
)

var (
	NewKeeper              = keeper.NewKeeper
	RegisterInvariants     = keeper.RegisterInvariants
	AllInvariants          = keeper.AllInvariants
	StakingTokensInvariant = keeper.StakingTokensInvariant
	DefaultCodespace       = keeper.DefaultCodespace

	NewPoolHolderAccount = types.NewPoolHolderAccount
	NewPoolMinterAccount = types.NewPoolMinterAccount
	NewSupply            = types.NewSupply
	DefaultSupply        = types.DefaultSupply
	NewGenesisState      = types.NewGenesisState
	DefaultGenesisState  = types.DefaultGenesisState
)

const (
	DefaultParamspace = keeper.DefaultParamspace
	StoreKey          = keeper.StoreKey
)
