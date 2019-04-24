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

	NewGenesisState     = types.NewGenesisState
	DefaultGenesisState = types.DefaultGenesisState
)

const (
	StoreKey     = keeper.StoreKey
	QuerierRoute = keeper.QuerierRoute
)
