package v4

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/exported"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/migrations/v1"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

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
		sdk.ZeroDec().String(),
		sdk.ZeroDec().String(),
		"",
	)

	bz, err := cdc.Marshal(&params)
	if err != nil {
		return err
	}

	store.Set(ParamsKey, bz)

	return nil
}

func migrateProposalVotingPeriod(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)
	propStore := prefix.NewStore(store, v1.ProposalsKeyPrefix)

	iter := propStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var prop govv1.Proposal
		err := cdc.Unmarshal(iter.Value(), &prop)
		if err != nil {
			return err
		}

		if prop.Status == govv1.StatusVotingPeriod {
			store.Set(VotingPeriodProposalKey(prop.Id), []byte{1})
		}
	}

	return nil
}

// MigrateStore performs in-place store migrations from v3 (v0.46) to v4 (v0.47). The
// migration includes:
//
// Params migrations from x/params to gov
// Addition of the new min initial deposit ratio parameter that is set to 0 by default.
// Proposals in voting period are tracked in a separate index.
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, legacySubspace exported.ParamSubspace, cdc codec.BinaryCodec) error {
	if err := migrateProposalVotingPeriod(ctx, storeKey, cdc); err != nil {
		return err
	}

	return migrateParams(ctx, storeKey, legacySubspace, cdc)
}
