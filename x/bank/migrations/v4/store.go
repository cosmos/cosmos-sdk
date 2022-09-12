package v4

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/bank/exported"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

const ModuleName = "bank"

var ParamsKey = []byte{0x05}

// migrateParams takes the parameters that are currently stored
// and managed by the x/params module and stores them directly into the x/bank
// module state.
func migrateParams(ctx sdk.Context, legacySubspace exported.Subspace, store storetypes.KVStore, cdc codec.BinaryCodec) error {
	var currParams types.Params
	legacySubspace.GetParamSet(ctx, &currParams)

	if err := currParams.Validate(); err != nil {
		return err
	}

	bz, err := cdc.Marshal(&currParams)
	if err != nil {
		return err
	}

	store.Set(ParamsKey, bz)

	return nil
}

// verifyDenoms doesn't do any store modifications, it only iterates through
// all existing coin metadata in state, and verify that their denoms are valid.
func verifyDenoms(ctx sdk.Context, store storetypes.KVStore, cdc codec.BinaryCodec) error {
	denomMetaDataStore := prefix.NewStore(store, types.DenomMetadataPrefix)

	iterator := denomMetaDataStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var metadata types.Metadata
		cdc.MustUnmarshal(iterator.Value(), &metadata)

		if err := metadata.Validate(); err != nil {
			return sdkerrors.Wrapf(err, "metadata contains an invalid denom, you should run a custom migration to all existing denoms compliant to the new format")
		}
	}

	return nil
}

// MigrateStore migrates the x/bank module state from the consensus version 3 to
// version 4. Specifically, it:
// - migrates the legacy param store into the bank module,
// - verify all denoms in state are valid.
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, legacySubspace exported.Subspace, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)

	if err := migrateParams(ctx, legacySubspace, store, cdc); err != nil {
		return err
	}

	return verifyDenoms(ctx, store, cdc)
}
