package v046

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// MigrateStore performs in-place store migrations from v0.43/v0.44 to v0.45.
// The migration includes:
//
// - Setting the MinCommissionRate param in the paramstore
func MigrateStore(ctx sdk.Context, paramstore paramtypes.Subspace) error {
	migrateParamsStore(ctx, paramstore)

	return nil
}

func migrateParamsStore(ctx sdk.Context, paramstore paramtypes.Subspace) {
	DefaultMinSelfDelegation := sdk.ZeroInt()
	paramstore.WithKeyTable(types.ParamKeyTable())
	paramstore.Set(ctx, types.KeyMinSelfDelegation, DefaultMinSelfDelegation)
}
