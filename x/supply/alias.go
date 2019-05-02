// nolint
package supply

import (
	"github.com/cosmos/cosmos-sdk/x/supply/keeper"
)

type (
	Keeper = keeper.Keeper
)

var (
	NewKeeper = keeper.NewKeeper

	RegisterInvariants     = keeper.RegisterInvariants
	AllInvariants          = keeper.AllInvariants
	StakingTokensInvariant = keeper.StakingTokensInvariant
)
