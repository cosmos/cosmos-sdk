package supply

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/supply/internal/types"
)

// InitGenesis sets supply information for genesis.
//
// CONTRACT: all types of accounts must have been already initialized/created
func InitGenesis(ctx sdk.Context, keeper Keeper, ak types.AccountKeeper, data GenesisState) {
	// manually set the total supply based on accounts if not provided
	if data.Supply.Empty() {
		var totalSupply sdk.Coins
		ak.IterateAccounts(ctx,
			func(acc authexported.Account) (stop bool) {
				totalSupply = totalSupply.Add(acc.GetCoins()...)
				return false
			},
		)

		data.Supply = totalSupply
	}

	keeper.SetSupply(ctx, types.NewSupply(data.Supply))
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
	var coins sdk.DecCoins
	for _, coin := range keeper.GetSupply(ctx).GetTotal() {
		if coin.Amount.IsZero() {
			continue
		}
		coins = coins.Add(coin)
	}
	return NewGenesisState(coins)
}

// ValidateGenesis performs basic validation of supply genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {
	return types.NewSupply(data.Supply).ValidateBasic()
}
