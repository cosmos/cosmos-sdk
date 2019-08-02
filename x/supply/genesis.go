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
	if data.Supply.GetTotal().Empty() {
		var totalSupply sdk.Coins
		ak.IterateAccounts(ctx,
			func(acc authexported.Account) (stop bool) {
				totalSupply = totalSupply.Add(acc.GetCoins())
				return false
			},
		)

		data.Supply = data.Supply.SetTotal(totalSupply)
	}

	keeper.SetSupply(ctx, data.Supply)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
	return NewGenesisState(keeper.GetSupply(ctx))
}

// ValidateGenesis performs basic validation of supply genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {
	return data.Supply.ValidateBasic()
}
