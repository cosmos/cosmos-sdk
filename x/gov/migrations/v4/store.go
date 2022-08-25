package v4

import (
	"sort"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/exported"
	types "github.com/cosmos/cosmos-sdk/x/gov/types"
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
	)

	bz, err := cdc.Marshal(&params)
	if err != nil {
		return err
	}

	store.Set(ParamsKey, bz)

	return nil
}

// AddProposerAddressToProposal will add proposer to proposal
// and set to the store
func AddProposerAddressToProposal(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec, proposals map[uint64]string) error {
	proposalIDS := make([]uint64, 0, len(proposals))

	for proposerID := range proposals {
		proposalIDS = append(proposalIDS, proposerID)
	}
	// sort the proposalIDS
	sort.Slice(proposalIDS, func(i, j int) bool { return proposalIDS[i] < proposalIDS[j] })

	store := ctx.KVStore(storeKey)

	for _, proposerID := range proposalIDS {
		bz := store.Get(types.ProposalKey(proposerID))
		var proposal govv1.Proposal
		if err := cdc.Unmarshal(bz, &proposal); err != nil {
			panic(err)
		}

		proposal.Proposer = proposals[proposerID]

		// set the new proposal with proposer
		bz, err := cdc.Marshal(&proposal)
		if err != nil {
			panic(err)
		}
		store.Set(types.ProposalKey(proposal.Id), bz)
	}

	return nil
}

// MigrateStore performs in-place store migrations from v3 (v0.46) to v4 (v0.47). The
// migration includes:
//
// Params migrations from x/params to gov
// Addition of the new min initial deposit ratio parameter that is set to 0 by default.
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, legacySubspace exported.ParamSubspace, cdc codec.BinaryCodec) error {
	return migrateParams(ctx, storeKey, legacySubspace, cdc)
}
