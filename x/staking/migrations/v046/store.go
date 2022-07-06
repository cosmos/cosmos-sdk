package v046

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// MigrateStore performs in-place store migrations from v0.43/v0.44/v0.45 to v0.46.
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec, legacySubspace exported.Subspace) error {
	store := ctx.KVStore(storeKey)
	var legacyParams types.Params
	legacySubspace.GetParamSet(ctx, &legacyParams)

	if err := legacyParams.Validate(); err != nil {
		return err
	}

	bz := cdc.MustMarshal(&legacyParams)
	store.Set(types.ParametersKey, bz)

	return nil
}
