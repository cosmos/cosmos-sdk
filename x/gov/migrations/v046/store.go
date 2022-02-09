package v046

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v040 "github.com/cosmos/cosmos-sdk/x/gov/migrations/v040"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta2"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// migrateProposals migrates all legacy proposals into MsgExecLegacyContent
// proposals.
func migrateProposals(store sdk.KVStore, cdc codec.BinaryCodec) error {
	propStore := prefix.NewStore(store, v040.ProposalsKeyPrefix)

	iter := propStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var oldProp v1beta1.Proposal
		err := cdc.Unmarshal(iter.Value(), &oldProp)
		if err != nil {
			return err
		}

		newProp, err := convertToNewProposal(oldProp)
		if err != nil {
			return err
		}
		bz, err := cdc.Marshal(&newProp)
		if err != nil {
			return err
		}

		// Set new value on store.
		propStore.Set(iter.Key(), bz)
	}

	return nil
}

// migrateParams migrates legacy depositParams  params into the new deposit params.
func migrateParams(ctx sdk.Context, paramstore paramtypes.Subspace) {

	oldDp := v1beta2.DepositParams{}
	paramstore.Get(ctx, v1beta1.ParamStoreKeyVotingParams, &oldDp)

	// depParams := convertToNewDepParams(oldDp)

	paramstore.WithKeyTable(v1beta2.ParamKeyTable())
	// paramstore.Set(ctx, v1beta2.ParamStoreKeyVotingParams, depParams)
}

// MigrateStore performs in-place store migrations from v0.43 to v0.46. The
// migration includes:
//
// - Migrate proposals to be Msg-based.
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec, paramStore paramtypes.Subspace) error {
	store := ctx.KVStore(storeKey)

	migrateParams(ctx, paramStore)

	return migrateProposals(store, cdc)
}
