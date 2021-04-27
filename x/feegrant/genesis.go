package feegrant

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	"github.com/cosmos/cosmos-sdk/x/feegrant/types"
)

// GenesisState contains a set of fee allowances, persisted from the store
type GenesisState []types.FeeAllowanceGrant

// ValidateBasic ensures all grants in the genesis state are valid
func (g GenesisState) ValidateBasic() error {
	for _, f := range g {
		grant, err := f.GetFeeGrant()
		if err != nil {
			return err
		}
		err = grant.ValidateBasic()
		if err != nil {
			return err
		}
	}
	return nil
}

// InitGenesis will initialize the keeper from a *previously validated* GenesisState
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data *types.GenesisState) {
	for _, f := range data.FeeAllowances {
		granter, err := sdk.AccAddressFromBech32(f.Granter)
		if err != nil {
			panic(err)
		}
		grantee, err := sdk.AccAddressFromBech32(f.Grantee)
		if err != nil {
			panic(err)
		}

		grant, err := f.GetFeeGrant()
		if err != nil {
			panic(err)
		}

		err = k.GrantFeeAllowance(ctx, granter, grantee, grant)
		if err != nil {
			panic(err)
		}
	}
}

// ExportGenesis will dump the contents of the keeper into a serializable GenesisState.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) (*types.GenesisState, error) {
	var grants []types.FeeAllowanceGrant

	err := k.IterateAllFeeAllowances(ctx, func(grant types.FeeAllowanceGrant) bool {
		grants = append(grants, grant)
		return false
	})

	return &types.GenesisState{
		FeeAllowances: grants,
	}, err
}
