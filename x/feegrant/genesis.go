package feegrant

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	"github.com/cosmos/cosmos-sdk/x/feegrant/types"
)

// InitGenesis will initialize the keeper from a *previously validated* GenesisState
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data *types.GenesisState) error {
	for _, f := range data.Allowances {
		granter, err := sdk.AccAddressFromBech32(f.Granter)
		if err != nil {
			return err
		}
		grantee, err := sdk.AccAddressFromBech32(f.Grantee)
		if err != nil {
			return err
		}

		grant, err := f.GetGrant()
		if err != nil {
			return err
		}

		err = k.GrantAllowance(ctx, granter, grantee, grant)
		if err != nil {
			return err
		}
	}
	return nil
}

// ExportGenesis will dump the contents of the keeper into a serializable GenesisState.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) (*types.GenesisState, error) {
	var grants []types.Grant

	err := k.IterateAllFeeAllowances(ctx, func(grant types.Grant) bool {
		grants = append(grants, grant)
		return false
	})

	return &types.GenesisState{
		Allowances: grants,
	}, err
}
