package v3

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/exported"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/migrations/v1"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// migrateProposals migrates all legacy proposals into MsgExecLegacyContent
// proposals.
func migrateProposals(store sdk.KVStore, cdc codec.BinaryCodec) error {
	propStore := prefix.NewStore(store, v1.ProposalsKeyPrefix)

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

func migrateParams(ctx sdk.Context, storeKey storetypes.StoreKey, legacySubspace exported.ParamSubspace, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)

	dp := govv1.DepositParams{}
	vp := govv1.VotingParams{}
	tp := govv1.TallyParams{}
	legacySubspace.Get(ctx, govv1.ParamStoreKeyDepositParams, &dp)
	legacySubspace.Get(ctx, govv1.ParamStoreKeyVotingParams, &vp)
	legacySubspace.Get(ctx, govv1.ParamStoreKeyTallyParams, &tp)

	params := govv1.NewParams(
		dp.MinDeposit,
		*dp.MaxDepositPeriod,
		*vp.VotingPeriod,
		tp.Quorum,
		tp.Threshold,
		tp.VetoThreshold,
	)

	bz, err := cdc.Marshal(&params)
	if err != nil {
		return err
	}

	store.Set(ParamsKey, bz)

	return nil
}

// MigrateStore performs in-place store migrations from v0.43 to v0.46. The
// migration includes:
//
// - Migrate proposals to be Msg-based.
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, legacySubspace exported.ParamSubspace, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)

	if err := migrateProposals(store, cdc); err != nil {
		return err
	}

	return migrateParams(ctx, storeKey, legacySubspace, cdc)
}
