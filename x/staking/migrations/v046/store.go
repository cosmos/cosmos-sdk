package v046

import (
	"github.com/pointnetwork/cosmos-point-sdk/codec"
	storetypes "github.com/pointnetwork/cosmos-point-sdk/store/types"
	sdk "github.com/pointnetwork/cosmos-point-sdk/types"
	paramtypes "github.com/pointnetwork/cosmos-point-sdk/x/params/types"
	"github.com/pointnetwork/cosmos-point-sdk/x/staking/types"
)

// MigrateStore performs in-place store migrations from v0.43/v0.44/v0.45 to v0.46.
// The migration includes:
//
// - Setting the MinCommissionRate param in the paramstore
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec, paramstore paramtypes.Subspace) error {
	migrateParamsStore(ctx, paramstore)

	return nil
}

func migrateParamsStore(ctx sdk.Context, paramstore paramtypes.Subspace) {
	if paramstore.HasKeyTable() {
		paramstore.Set(ctx, types.KeyMinCommissionRate, types.DefaultMinCommissionRate)
	} else {
		paramstore.WithKeyTable(types.ParamKeyTable())
		paramstore.Set(ctx, types.KeyMinCommissionRate, types.DefaultMinCommissionRate)
	}
}
