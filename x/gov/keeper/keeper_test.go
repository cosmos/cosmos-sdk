package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app         *simapp.SimApp
	ctx         sdk.Context
	queryClient types.QueryClient
	addrs       []sdk.AccAddress
}

func (suite *KeeperTestSuite) SetupTest() {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.GovKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	suite.app = app
	suite.ctx = ctx
	suite.queryClient = queryClient
	suite.addrs = simapp.AddTestAddrsIncremental(app, ctx, 2, sdk.NewInt(30000000))
}

func TestIncrementProposalNumber(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	tp := TestProposal
	_, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	_, err = app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	_, err = app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	_, err = app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	_, err = app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposal6, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)

	require.Equal(t, uint64(6), proposal6.ProposalId)
}

func TestProposalQueues(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// create test proposals
	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)

	inactiveIterator := app.GovKeeper.InactiveProposalQueueIterator(ctx, proposal.DepositEndTime)
	require.True(t, inactiveIterator.Valid())

	proposalID := types.GetProposalIDFromBytes(inactiveIterator.Value())
	require.Equal(t, proposalID, proposal.ProposalId)
	inactiveIterator.Close()

	app.GovKeeper.ActivateVotingPeriod(ctx, proposal)

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposal.ProposalId)
	require.True(t, ok)

	activeIterator := app.GovKeeper.ActiveProposalQueueIterator(ctx, proposal.VotingEndTime)
	require.True(t, activeIterator.Valid())

	proposalID, _ = types.SplitActiveProposalQueueKey(activeIterator.Key())
	require.Equal(t, proposalID, proposal.ProposalId)

	activeIterator.Close()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
