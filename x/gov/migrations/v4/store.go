package v4

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// Zero implies that the initial deposit requirement is off
// by default.
var minInitialDepositRatio = sdk.ZeroDec().String()

// MigrateStore performs in-place store migrations for consensus version 3
// in the gov module.
// Please note that this is the first version that switches from using
// SDK versioning (v043 etc) for package names to consensus versioning
// of the gov module.
// The migration includes:
//
// - Setting the minimum deposit param in the paramstore.
func MigrateStore(ctx sdk.Context, govParamSpace types.ParamSubspace) error {
	migrateParams(ctx, govParamSpace)
	return nil
}

func migrateParams(ctx sdk.Context, govParamSpace types.ParamSubspace) {
	var depositParams v1.DepositParams
	govParamSpace.Get(ctx, v1.ParamStoreKeyDepositParams, &depositParams)
	depositParams.MinInitialDepositRatio = minInitialDepositRatio
	govParamSpace.Set(ctx, v1.ParamStoreKeyDepositParams, &depositParams)
}
