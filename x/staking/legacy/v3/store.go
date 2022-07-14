package v3

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// MigrateStore performs in-place store migrations for consensus version 3
// in the staking module.
// Please note that this is the first version that switches from using
// SDK versioning (v046 etc) for package names to consensus versioning
// of the staking module.
// The migration includes:
//
// - Setting the MinCommissionRate param in the paramstore
func MigrateStore(ctx sdk.Context, paramstore paramtypes.Subspace) error {
	migrateParamsStore(ctx, paramstore)

	return nil
}

func migrateParamsStore(ctx sdk.Context, paramstore paramtypes.Subspace) {
	DefaultMinSelfDelegation := sdk.ZeroInt()
	paramstore.Set(ctx, types.KeyMinSelfDelegation, DefaultMinSelfDelegation)
}
