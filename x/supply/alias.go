// nolint
package supply

import (
	"github.com/cosmos/cosmos-sdk/x/supply/keeper"
)

type (
	Keeper = keeper.Keeper
)

var (
	NewKeeper      = keeper.NewKeeper
	AccountsSupply = keeper.AccountsSupply
	EscrowedSupply = keeper.EscrowedSupply
	TotalSupply    = keeper.TotalSupply

	RegisterInvariants     = keeper.RegisterInvariants
	AllInvariants          = keeper.AllInvariants
	StakingTokensInvariant = StakingTokensInvariant
)
