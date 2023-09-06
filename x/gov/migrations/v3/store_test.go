package v3_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/gov"
	v1gov "github.com/cosmos/cosmos-sdk/x/gov/migrations/v1"
	v3gov "github.com/cosmos/cosmos-sdk/x/gov/migrations/v3"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func TestMigrateStore(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig(gov.AppModuleBasic{}).Codec
	govKey := storetypes.NewKVStoreKey("gov")
	ctx := testutil.DefaultContext(govKey, storetypes.NewTransientStoreKey("transient_test"))
	store := ctx.KVStore(govKey)

	propTime := time.Unix(1e9, 0)

	// Create 2 proposals
	prop1, err := v1beta1.NewProposal(v1beta1.NewTextProposal("my title 1", "my desc 1"), 1, propTime, propTime)
	require.NoError(t, err)
	prop1Bz, err := cdc.Marshal(&prop1)
	require.NoError(t, err)
	prop2, err := v1beta1.NewProposal(v1beta1.NewTextProposal("my title 2", "my desc 2"), 2, propTime, propTime)
	require.NoError(t, err)
	require.NoError(t, err)
	prop2Bz, err := cdc.Marshal(&prop2)
	require.NoError(t, err)

	store.Set(v1gov.ProposalKey(prop1.ProposalId), prop1Bz)
	store.Set(v1gov.ProposalKey(prop2.ProposalId), prop2Bz)

	// Vote on prop 1
	options := []v1beta1.WeightedVoteOption{
		{Option: v1beta1.OptionNo, Weight: math.LegacyMustNewDecFromStr("0.3")},
		{Option: v1beta1.OptionYes, Weight: math.LegacyMustNewDecFromStr("0.7")},
	}
	vote1 := v1beta1.Vote{ProposalId: 1, Voter: voter.String(), Options: options}
	vote1Bz := cdc.MustMarshal(&vote1)
	store.Set(v1gov.VoteKey(1, voter), vote1Bz)

	// Run migrations.
	storeService := runtime.NewKVStoreService(govKey)
	err = v3gov.MigrateStore(ctx, storeService, cdc)
	require.NoError(t, err)

	var newProp1 v1.Proposal
	err = cdc.Unmarshal(store.Get(v1gov.ProposalKey(prop1.ProposalId)), &newProp1)
	require.NoError(t, err)
	compareProps(t, prop1, newProp1)

	var newProp2 v1.Proposal
	err = cdc.Unmarshal(store.Get(v1gov.ProposalKey(prop2.ProposalId)), &newProp2)
	require.NoError(t, err)
	compareProps(t, prop2, newProp2)

	var newVote1 v1.Vote
	err = cdc.Unmarshal(store.Get(v1gov.VoteKey(prop1.ProposalId, voter)), &newVote1)
	require.NoError(t, err)
	// Without the votes migration, we would have 300000000000000000 in state,
	// because of how sdk.Dec stores itself in state.
	require.Equal(t, "0.300000000000000000", newVote1.Options[0].Weight)
	require.Equal(t, "0.700000000000000000", newVote1.Options[1].Weight)
}

func compareProps(t *testing.T, oldProp v1beta1.Proposal, newProp v1.Proposal) {
	t.Helper()
	require.Equal(t, oldProp.ProposalId, newProp.Id)
	require.Equal(t, oldProp.TotalDeposit.String(), sdk.Coins(newProp.TotalDeposit).String())
	require.Equal(t, oldProp.Status.String(), newProp.Status.String())
	require.Equal(t, oldProp.FinalTallyResult.Yes.String(), newProp.FinalTallyResult.YesCount)
	require.Equal(t, oldProp.FinalTallyResult.No.String(), newProp.FinalTallyResult.NoCount)
	require.Equal(t, oldProp.FinalTallyResult.NoWithVeto.String(), newProp.FinalTallyResult.NoWithVetoCount)
	require.Equal(t, oldProp.FinalTallyResult.Abstain.String(), newProp.FinalTallyResult.AbstainCount)

	newContent := newProp.Messages[0].GetCachedValue().(*v1.MsgExecLegacyContent).Content.GetCachedValue().(v1beta1.Content)
	require.Equal(t, oldProp.Content.GetCachedValue().(v1beta1.Content), newContent)

	// Compare UNIX times, as a simple Equal gives difference between Local and
	// UTC times.
	// ref: https://github.com/golang/go/issues/19486#issuecomment-292968278
	require.Equal(t, oldProp.SubmitTime.Unix(), newProp.SubmitTime.Unix())
	require.Equal(t, oldProp.DepositEndTime.Unix(), newProp.DepositEndTime.Unix())
	require.Equal(t, oldProp.VotingStartTime.Unix(), newProp.VotingStartTime.Unix())
	require.Equal(t, oldProp.VotingEndTime.Unix(), newProp.VotingEndTime.Unix())
}
