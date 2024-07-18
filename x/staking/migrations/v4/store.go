package v4

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// subspace contains the method needed for migrations of the legacy Params subspace
type subspace interface {
	GetParamSet(ctx sdk.Context, ps paramtypes.ParamSet)
	Set(ctx sdk.Context, key []byte, value interface{})
}

// MigrateStore performs in-place store migrations from v3 to v4.
func MigrateStore(ctx sdk.Context, store storetypes.KVStore, cdc codec.BinaryCodec, legacySubspace exported.Subspace) error {
	// migrate params
	if err := migrateParams(ctx, store, cdc, legacySubspace); err != nil {
		return err
	}

	return nil
}

// migrateParams will set the params to store from legacySubspace
func migrateParams(ctx sdk.Context, store storetypes.KVStore, cdc codec.BinaryCodec, legacySubspace exported.Subspace) error {
	var legacyParams types.Params

	// cast to type that implements .Set()
	// KeyMinCommissionRate was not added in v3 migration in the LSM line
	// we need to be set it to avoid unmarshalling errs
	extendedLegacySubspace := legacySubspace.(subspace)
	extendedLegacySubspace.Set(ctx, types.KeyMinCommissionRate, types.DefaultMinCommissionRate)

	// get all params after setting MinCommissionRate
	extendedLegacySubspace.GetParamSet(ctx, &legacyParams)

	if err := legacyParams.Validate(); err != nil {
		return err
	}

	bz := cdc.MustMarshal(&legacyParams)
	store.Set(types.ParamsKey, bz)
	return nil
}
