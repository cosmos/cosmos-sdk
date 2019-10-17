package delegation

import sdk "github.com/cosmos/cosmos-sdk/types"

type GenesisState []FeeAllowanceGrant

func (g GenesisState) ValidateBasic() error {
	for _, f := range g {
		err := f.ValidateBasic()
		if err != nil {
			return err
		}
	}
	return nil
}

func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) {
	// TODO
}

func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	// TODO
	return {}
}
