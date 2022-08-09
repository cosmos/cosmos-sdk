package v2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade/exported"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// Migrate migrates the x/upgrade module state from the consensus version 1 to
// version 2, returning an error upon failure.
func Migrate(ctx sdk.Context, subspace exported.ParamSubspace) error {
	// set KeyTable if it has not already been set
	if !subspace.HasKeyTable() {
		subspace = subspace.WithKeyTable(types.ParamKeyTable())
	}

	params := types.DefaultParams()
	subspace.SetParamSet(ctx, &params)

	return nil
}
