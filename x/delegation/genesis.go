package delegation

import sdk "github.com/cosmos/cosmos-sdk/types"

// GenesisState contains a set of fee allowances, persisted from the store
type GenesisState []FeeAllowanceGrant

// ValidateBasic ensures all grants in the genesis state are valid
func (g GenesisState) ValidateBasic() error {
	for _, f := range g {
		err := f.ValidateBasic()
		if err != nil {
			return err
		}
	}
	return nil
}

// InitGenesis will initialize the keeper from a *previously validated* GenesisState
func InitGenesis(ctx sdk.Context, k Keeper, gen GenesisState) error {
	for _, f := range gen {
		err := k.DelegateFeeAllowance(ctx, f)
		if err != nil {
			return err
		}
	}
	return nil
}

// ExportGenesis will dump the contents of the keeper into a serializable GenesisState
func ExportGenesis(ctx sdk.Context, k Keeper) (GenesisState, error) {
	return k.GetAllFeeAllowances(ctx)
}
