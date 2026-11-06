package keeper_test

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func setupDelegatorTally(tb testing.TB, delegations int, options []v1.WeightedVoteOptions) (*fixture, v1.Proposal, []sdk.AccAddress) {
	tb.Helper()
	f := initFixture(tb)
	validators := make([]sdk.ValAddress, delegations)
	for i := range validators {
		addr := sdk.ValAddress([]byte(fmt.Sprintf("tally-validator-%04d", i)))
		pk := ed25519.GenPrivKeyFromSecret([]byte(fmt.Sprintf("tally-validator-%d", i))).PubKey()
		val, err := stakingtypes.NewValidator(addr.String(), pk, stakingtypes.Description{})
		assert.NilError(tb, err)
		val.Status = stakingtypes.Bonded
		val.Tokens = math.NewInt(1_000_000)
		val.DelegatorShares = math.LegacyNewDec(1_000_000)
		assert.NilError(tb, f.stakingKeeper.SetValidator(f.ctx, val))
		assert.NilError(tb, f.stakingKeeper.SetNewValidatorByPowerIndex(f.ctx, val))
		validators[i] = addr
	}
	proposal := v1.Proposal{Id: 1}
	voters := make([]sdk.AccAddress, len(options))
	for i, opts := range options {
		voter := sdk.AccAddress([]byte(fmt.Sprintf("tally-voter-%08d", i)))
		for _, validator := range validators {
			delegation := stakingtypes.NewDelegation(voter.String(), validator.String(), math.LegacyNewDec(10_000))
			assert.NilError(tb, f.stakingKeeper.SetDelegation(f.ctx, delegation))
		}
		vote := v1.NewVote(proposal.Id, voter, opts, "")
		assert.NilError(tb, f.govKeeper.Votes.Set(f.ctx, collections.Join(proposal.Id, voter), vote))
		voters[i] = voter
	}
	return f, proposal, voters
}

func TestTallyDelegatorWeightsRemainSeparate(t *testing.T) {
	t.Parallel()
	// Voter order changes both the number of options and their ordering. Each
	// voter delegates to three validators, exercising repeated use of its weights.
	options := []v1.WeightedVoteOptions{
		{
			{Option: v1.OptionYes, Weight: "0.137"},
			{Option: v1.OptionAbstain, Weight: "0.263"},
			{Option: v1.OptionNo, Weight: "0.4"},
			{Option: v1.OptionNoWithVeto, Weight: "0.2"},
		},
		v1.NewNonSplitVoteOption(v1.OptionYes),
		{
			{Option: v1.OptionNo, Weight: "0.8"},
			{Option: v1.OptionYes, Weight: "0.2"},
		},
	}
	f, proposal, voters := setupDelegatorTally(t, 3, options)
	ctx, _ := f.ctx.CacheContext()
	passes, burns, tally, err := f.govKeeper.Tally(ctx, proposal)
	assert.NilError(t, err)
	assert.Assert(t, !passes)
	assert.Assert(t, !burns)
	expected := v1.NewTallyResult(math.NewInt(40_110), math.NewInt(7_890), math.NewInt(36_000), math.NewInt(6_000))
	assert.Assert(t, tally.Equals(expected))
	for _, voter := range voters {
		key := collections.Join(proposal.Id, voter)
		present, err := f.govKeeper.Votes.Has(ctx, key)
		assert.NilError(t, err)
		assert.Assert(t, !present)
		present, err = f.govKeeper.Votes.Has(f.ctx, key)
		assert.NilError(t, err)
		assert.Assert(t, present)
	}
}

func BenchmarkTallyDelegations(b *testing.B) {
	for _, weighted := range []bool{false, true} {
		for _, delegations := range []int{0, 1, 10, 100} {
			for _, voters := range []int{1, 100} {
				b.Run(fmt.Sprintf("weighted=%t/delegations=%d/voters=%d", weighted, delegations, voters), func(b *testing.B) {
					options := make([]v1.WeightedVoteOptions, voters)
					for i := range options {
						options[i] = v1.NewNonSplitVoteOption(v1.OptionYes)
						if weighted {
							options[i] = v1.WeightedVoteOptions{
								{Option: v1.OptionYes, Weight: "0.137"},
								{Option: v1.OptionAbstain, Weight: "0.263"},
								{Option: v1.OptionNo, Weight: "0.4"},
								{Option: v1.OptionNoWithVeto, Weight: "0.2"},
							}
						}
					}
					f, proposal, _ := setupDelegatorTally(b, delegations, options)
					b.ReportAllocs()
					for b.Loop() {
						ctx, _ := f.ctx.CacheContext()
						if _, _, _, err := f.govKeeper.Tally(ctx, proposal); err != nil {
							b.Fatal(err)
						}
					}
				})
			}
		}
	}
}
