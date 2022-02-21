package v046_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v040gov "github.com/cosmos/cosmos-sdk/x/gov/migrations/v040"
	v046gov "github.com/cosmos/cosmos-sdk/x/gov/migrations/v046"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta2"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func TestMigrateStore(t *testing.T) {
	cdc := simapp.MakeTestEncodingConfig().Codec
	govKey := sdk.NewKVStoreKey("gov")
	ctx := testutil.DefaultContext(govKey, sdk.NewTransientStoreKey("transient_test"))
	store := ctx.KVStore(govKey)

	propTime := time.Unix(1e9, 0)

	// Create 2 proposals
	prop1, err := v1beta1.NewProposal(v1beta1.NewTextProposal("my title 1", "my desc 1"), 1, propTime, propTime)
	require.NoError(t, err)
	prop1Bz, err := cdc.Marshal(&prop1)
	require.NoError(t, err)
	prop2, err := v1beta1.NewProposal(upgradetypes.NewSoftwareUpgradeProposal("my title 2", "my desc 2", upgradetypes.Plan{
		Name: "my plan 2",
	}), 2, propTime, propTime)
	require.NoError(t, err)
	prop2Bz, err := cdc.Marshal(&prop2)
	require.NoError(t, err)

	store.Set(v040gov.ProposalKey(prop1.ProposalId), prop1Bz)
	store.Set(v040gov.ProposalKey(prop2.ProposalId), prop2Bz)

	// Run migrations.
	err = v046gov.MigrateStore(ctx, govKey, cdc)
	require.NoError(t, err)

	var newProp1 v1beta2.Proposal
	err = cdc.Unmarshal(store.Get(v040gov.ProposalKey(prop1.ProposalId)), &newProp1)
	require.NoError(t, err)
	compareProps(t, prop1, newProp1)

	var newProp2 v1beta2.Proposal
	err = cdc.Unmarshal(store.Get(v040gov.ProposalKey(prop2.ProposalId)), &newProp2)
	require.NoError(t, err)
	compareProps(t, prop2, newProp2)
}

func compareProps(t *testing.T, oldProp v1beta1.Proposal, newProp v1beta2.Proposal) {
	require.Equal(t, oldProp.ProposalId, newProp.Id)
	require.Equal(t, oldProp.TotalDeposit.String(), sdk.Coins(newProp.TotalDeposit).String())
	require.Equal(t, oldProp.Status.String(), newProp.Status.String())
	require.Equal(t, oldProp.FinalTallyResult.Yes.String(), newProp.FinalTallyResult.YesCount)
	require.Equal(t, oldProp.FinalTallyResult.No.String(), newProp.FinalTallyResult.NoCount)
	require.Equal(t, oldProp.FinalTallyResult.NoWithVeto.String(), newProp.FinalTallyResult.NoWithVetoCount)
	require.Equal(t, oldProp.FinalTallyResult.Abstain.String(), newProp.FinalTallyResult.AbstainCount)

	newContent := newProp.Messages[0].GetCachedValue().(*v1beta2.MsgExecLegacyContent).Content.GetCachedValue().(v1beta1.Content)
	require.Equal(t, oldProp.Content.GetCachedValue().(v1beta1.Content), newContent)

	// Compare UNIX times, as a simple Equal gives difference between Local and
	// UTC times.
	// ref: https://github.com/golang/go/issues/19486#issuecomment-292968278
	require.Equal(t, oldProp.SubmitTime.Unix(), newProp.SubmitTime.Unix())
	require.Equal(t, oldProp.DepositEndTime.Unix(), newProp.DepositEndTime.Unix())
	require.Equal(t, oldProp.VotingStartTime.Unix(), newProp.VotingStartTime.Unix())
	require.Equal(t, oldProp.VotingEndTime.Unix(), newProp.VotingEndTime.Unix())
}
