//nolint
package mint

import (
	"github.com/cosmos/cosmos-sdk/x/mint/keeper"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

type (
	Keeper       = keeper.Keeper
	Minter       = types.Minter
	Params       = types.Params
	GenesisState = types.GenesisState
)

var (
	NewKeeper  = keeper.NewKeeper
	NewQuerier = keeper.NewQuerier

	NewGenesisState      = types.NewGenesisState
	DefaultGenesisState  = types.DefaultGenesisState
	InitialMinter        = types.InitialMinter
	DefaultInitialMinter = types.DefaultInitialMinter
	NewParams            = types.NewParams
)

const (
	StoreKey          = keeper.StoreKey
	QuerierRoute      = keeper.QuerierRoute
	DefaultParamspace = keeper.DefaultParamspace
)
