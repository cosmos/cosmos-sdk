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

// ExportGenesis will dump the contents of the keeper into a serializable GenesisState
//
// All expiration heights will be thrown off if we dump state and start at a new
// chain at height 0. Thus, we allow the Allowances to "prepare themselves"
// for export, like if they have expiry at 5000 and current is 4000, they export with
// expiry of 1000. Every FeeAllowance has a method `PrepareForExport` that allows
// them to perform any changes needed prior to export.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) (*types.GenesisState, error) {
	time, height := ctx.BlockTime(), ctx.BlockHeight()
	var grants []types.Grant

	err := k.IterateAllFeeAllowances(ctx, func(grant types.Grant) bool {
		grants = append(grants, grant.PrepareForExport(time, height))
		return false
	})

	return &types.GenesisState{
		Allowances: grants,
	}, err
}
