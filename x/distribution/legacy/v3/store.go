package v3

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// If expedited, the deposit to enter voting period will be
// increased to 5000 OSMO. The proposal will have 24 hours to achieve
// a two-thirds majority of all staked OSMO voting power voting YES.
var (
	MinimumRestakeThreshold = sdk.NewDec(10_000_000)
	RestakePeriod           = sdk.NewInt(1000)
)

// MigrateStore performs in-place store migrations for consensus version 4
// in the gov module.
// The migration includes:
//
// - Setting the expedited proposals params in the paramstore.
func MigrateStore(ctx sdk.Context, subspace paramtypes.Subspace) error {
	migrateParamsStore(ctx, subspace)
	return nil
}

func migrateParamsStore(ctx sdk.Context, subspace paramtypes.Subspace) {
	subspace.Set(ctx, types.ParamRestakePeriod, RestakePeriod)
	subspace.Set(ctx, types.ParamMinimumRestakeThreshold, MinimumRestakeThreshold)
}
