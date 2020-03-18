package feegrant

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant/exported"
)

// GenesisState contains a set of fee allowances, persisted from the store
type GenesisState []exported.FeeAllowanceGrant

// ValidateBasic ensures all grants in the genesis state are valid
func (g GenesisState) ValidateBasic() error {
	for _, f := range g {
		err := f.GetFeeGrant().ValidateBasic()
		if err != nil {
			return err
		}
	}
	return nil
}

// InitGenesis will initialize the keeper from a *previously validated* GenesisState
func InitGenesis(ctx sdk.Context, k Keeper, gen GenesisState) {
	for _, f := range gen {
		k.GrantFeeAllowance(ctx, f)
	}
}

// ExportGenesis will dump the contents of the keeper into a serializable GenesisState
//
// All expiration heights will be thrown off if we dump state and start at a new
// chain at height 0. Thus, we allow the Allowances to "prepare themselves"
// for export, like if they have expiry at 5000 and current is 4000, they export with
// expiry of 1000. Every FeeAllowance has a method `PrepareForExport` that allows
// them to perform any changes needed prior to export.
func ExportGenesis(ctx sdk.Context, k Keeper) (GenesisState, error) {
	time, height := ctx.BlockTime(), ctx.BlockHeight()
	var grants []exported.FeeAllowanceGrant

	err := k.IterateAllFeeAllowances(ctx, func(grant exported.FeeAllowanceGrant) bool {
		grants = append(grants, grant.GetFeeGrant().PrepareForExport(time, height))
		return false
	})

	return grants, err
}
