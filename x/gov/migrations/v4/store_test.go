package v4_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	v1gov "github.com/cosmos/cosmos-sdk/x/gov/migrations/v1"
	v4 "github.com/cosmos/cosmos-sdk/x/gov/migrations/v4"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
)

var (
	_, _, addr   = testdata.KeyTestPubAddr()
	govAcct      = authtypes.NewModuleAddress(types.ModuleName)
	TestProposal = getTestProposal()
)

type mockSubspace struct {
	dp v1.DepositParams
	vp v1.VotingParams
	tp v1.TallyParams
}

func newMockSubspace(p v1.Params) mockSubspace {
	return mockSubspace{
		dp: v1.DepositParams{
			MinDeposit:       p.MinDeposit,
			MaxDepositPeriod: p.MaxDepositPeriod,
		},
		vp: v1.VotingParams{
			VotingPeriod: p.VotingPeriod,
		},
		tp: v1.TallyParams{
			Quorum:        p.Quorum,
			Threshold:     p.Threshold,
			VetoThreshold: p.VetoThreshold,
		},
	}
}

func (ms mockSubspace) Get(ctx sdk.Context, key []byte, ptr interface{}) {
	switch string(key) {
	case string(v1.ParamStoreKeyDepositParams):
		*ptr.(*v1.DepositParams) = ms.dp
	case string(v1.ParamStoreKeyVotingParams):
		*ptr.(*v1.VotingParams) = ms.vp
	case string(v1.ParamStoreKeyTallyParams):
		*ptr.(*v1.TallyParams) = ms.tp
	}
}

func TestMigrateStore(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig(upgrade.AppModuleBasic{}, gov.AppModuleBasic{}, bank.AppModuleBasic{}).Codec
	govKey := sdk.NewKVStoreKey("gov")
	ctx := testutil.DefaultContext(govKey, sdk.NewTransientStoreKey("transient_test"))
	store := ctx.KVStore(govKey)

	legacySubspace := newMockSubspace(v1.DefaultParams())

	propTime := time.Unix(1e9, 0)

	// Create 2 proposals
	prop1Content, err := v1.NewLegacyContent(v1beta1.NewTextProposal("Test", "description"), authtypes.NewModuleAddress("gov").String())
	require.NoError(t, err)
	proposal1, err := v1.NewProposal([]sdk.Msg{prop1Content}, 1, "some metadata for the legacy content", propTime, propTime)
	require.NoError(t, err)
	prop1Bz, err := cdc.Marshal(&proposal1)
	require.NoError(t, err)
	store.Set(v1gov.ProposalKey(proposal1.Id), prop1Bz)

	proposal2, err := v1.NewProposal(getTestProposal(), 2, "some metadata for the legacy content", propTime, propTime)
	require.NoError(t, err)
	prop2Bz, err := cdc.Marshal(&proposal2)
	require.NoError(t, err)
	store.Set(v1gov.ProposalKey(proposal2.Id), prop2Bz)

	// Run migrations.
	err = v4.MigrateStore(ctx, govKey, legacySubspace, cdc)
	require.NoError(t, err)

	// Check params
	var params v1.Params
	bz := store.Get(v4.ParamsKey)
	require.NoError(t, cdc.Unmarshal(bz, &params))
	require.NotNil(t, params)
	require.Equal(t, legacySubspace.dp.MinDeposit, params.MinDeposit)
	require.Equal(t, legacySubspace.dp.MaxDepositPeriod, params.MaxDepositPeriod)
	require.Equal(t, legacySubspace.vp.VotingPeriod, params.VotingPeriod)
	require.Equal(t, legacySubspace.tp.Quorum, params.Quorum)
	require.Equal(t, legacySubspace.tp.Threshold, params.Threshold)
	require.Equal(t, legacySubspace.tp.VetoThreshold, params.VetoThreshold)
	require.Equal(t, sdk.ZeroDec().String(), params.MinInitialDepositRatio)

	// Check proposals and contents (1)
	var newProposal v1.Proposal
	bz = store.Get(v1gov.ProposalKey(proposal1.Id))
	require.NoError(t, cdc.Unmarshal(bz, &newProposal))

	var newPropContents v1.ProposalContents
	bz = store.Get(v4.ProposalContentsKey(newProposal.Id))
	require.NoError(t, cdc.Unmarshal(bz, &newPropContents))

	err = sdktx.UnpackInterfaces(cdc, newPropContents.Messages)
	require.NoError(t, err)

	checkMigratedProp(t, proposal1, newProposal, newPropContents)

	// Check proposals and contents (2)
	var newProposal2 v1.Proposal
	bz = store.Get(v1gov.ProposalKey(proposal2.Id))
	require.NoError(t, cdc.Unmarshal(bz, &newProposal2))

	var newPropContents2 v1.ProposalContents
	bz = store.Get(v4.ProposalContentsKey(newProposal2.Id))
	require.NoError(t, cdc.Unmarshal(bz, &newPropContents2))

	err = sdktx.UnpackInterfaces(cdc, newPropContents2.Messages)
	require.NoError(t, err)

	checkMigratedProp(t, proposal2, newProposal2, newPropContents2)
}

func getTestProposal() []sdk.Msg {
	legacyProposalMsg, err := v1.NewLegacyContent(v1beta1.NewTextProposal("Title", "description"), authtypes.NewModuleAddress(types.ModuleName).String())
	if err != nil {
		panic(err)
	}

	return []sdk.Msg{
		banktypes.NewMsgSend(govAcct, addr, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1000)))),
		legacyProposalMsg,
	}
}

func checkMigratedProp(t *testing.T, oldProp v1.Proposal, newProp v1.Proposal, newContents v1.ProposalContents) {
	// We are not changing any of these fields but we are checking them to make sure
	require.Equal(t, oldProp.Id, newProp.Id)
	require.Equal(t, oldProp.Status.String(), newProp.Status.String())
	require.Equal(t, sdk.Coins(newProp.TotalDeposit).String(), sdk.Coins(newProp.TotalDeposit).String())
	require.Equal(t, oldProp.FinalTallyResult, newProp.FinalTallyResult)

	// Compare UNIX times, as a simple Equal gives difference between Local and
	// UTC times.
	// ref: https://github.com/golang/go/issues/19486#issuecomment-292968278
	require.Equal(t, oldProp.SubmitTime.Unix(), newProp.SubmitTime.Unix())
	require.Equal(t, oldProp.DepositEndTime.Unix(), newProp.DepositEndTime.Unix())
	require.Equal(t, oldProp.VotingStartTime, newProp.VotingStartTime)
	require.Equal(t, oldProp.VotingEndTime, newProp.VotingEndTime)

	// Check contents

	// These 2 should be empty now
	require.Empty(t, newProp.Messages)
	require.Empty(t, newProp.Metadata)

	require.Equal(t, oldProp.Messages, newContents.Messages)
	require.Equal(t, oldProp.Metadata, newContents.Metadata)
}
