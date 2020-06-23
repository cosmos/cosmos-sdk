package keeper_test

import (
	gocontext "context"
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestAllProposal(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx)
	types.RegisterQueryServer(queryHelper, app.GovKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	// check for the proposals with no proposal added should return null.
	pageReq := &query.PageRequest{
		Key:        nil,
		Limit:      1,
		CountTotal: false,
	}

	req := types.NewQueryProposalsRequest(0, nil, nil, pageReq)

	proposals, err := queryClient.AllProposals(gocontext.Background(), req)
	require.NoError(t, err)
	require.Equal(t, string(proposals.Proposals), "null")

	// create 2 test proposals
	for i := 0; i < 2; i++ {
		num := strconv.Itoa(i + 1)
		testProposal := types.NewTextProposal("Proposal"+num, "testing proposal "+num)
		_, err := app.GovKeeper.SubmitProposal(ctx, testProposal)
		require.NoError(t, err)
	}

	// Query for proposals after adding 2 proposals to the store.
	// give page limit as 1 and expect NextKey should not to be empty
	proposals, err = queryClient.AllProposals(gocontext.Background(), req)
	require.NoError(t, err)
	require.NotEmpty(t, proposals.Proposals)
	require.NotEmpty(t, proposals.Res.NextKey)

	pageReq = &query.PageRequest{
		Key:        proposals.Res.NextKey,
		Limit:      1,
		CountTotal: false,
	}

	req = types.NewQueryProposalsRequest(0, nil, nil, pageReq)

	// query for the next page which is 2nd proposal at present context.
	// and expect NextKey should be empty
	proposals, err = queryClient.AllProposals(gocontext.Background(), req)

	require.NoError(t, err)
	require.NotEmpty(t, proposals.Proposals)
	require.Empty(t, proposals.Res)

	pageReq = &query.PageRequest{
		Key:        nil,
		Limit:      2,
		CountTotal: false,
	}

	req = types.NewQueryProposalsRequest(0, nil, nil, pageReq)

	// Query the page with limit 2 and expect NextKey should ne nil
	proposals, err = queryClient.AllProposals(gocontext.Background(), req)

	require.NoError(t, err)
	require.NotEmpty(t, proposals.Proposals)
	require.Empty(t, proposals.Res)
}

func TestVotesGRPC(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx)
	types.RegisterQueryServer(queryHelper, app.GovKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	addrs := simapp.AddTestAddrsIncremental(app, ctx, 2, sdk.NewInt(30000000))

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID

	proposal.Status = types.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	// query vote when no vote registered and expect null.
	pageReq := &query.PageRequest{
		Key:        nil,
		Limit:      1,
		CountTotal: false,
	}

	req := types.NewQueryVotesRequest(proposalID, pageReq)
	votes, err := queryClient.Votes(gocontext.Background(), req)

	require.NoError(t, err)
	require.Equal(t, string(votes.Votes), "null")
	require.Empty(t, votes.Res)

	// Register two votes with different addresses
	// Test first vote
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[0], types.OptionAbstain))
	_, found := app.GovKeeper.GetVote(ctx, proposalID, addrs[0])
	require.True(t, found)

	// Test second vote
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[1], types.OptionNoWithVeto))
	_, found = app.GovKeeper.GetVote(ctx, proposalID, addrs[1])
	require.True(t, found)

	// query vote with limit 1 and expect next should not be nil.
	pageReq = &query.PageRequest{
		Key:        nil,
		Limit:      1,
		CountTotal: false,
	}

	req = types.NewQueryVotesRequest(proposalID, pageReq)
	votes, err = queryClient.Votes(gocontext.Background(), req)
	require.NoError(t, err)
	require.NotEmpty(t, votes.Votes)
	require.NotEmpty(t, votes.Res)

	// query vote with limit 1, next key and expect NextKey to be nil.
	pageReq = &query.PageRequest{
		Key:        votes.Res.NextKey,
		Limit:      1,
		CountTotal: false,
	}

	req = types.NewQueryVotesRequest(proposalID, pageReq)
	votes, err = queryClient.Votes(gocontext.Background(), req)
	require.NoError(t, err)
	require.NotEmpty(t, votes.Votes)
	require.Empty(t, votes.Res)

	// query vote with limit 2 and expect NextKey to be nil.
	pageReq = &query.PageRequest{
		Key:        nil,
		Limit:      2,
		CountTotal: false,
	}

	req = types.NewQueryVotesRequest(proposalID, pageReq)
	votes, err = queryClient.Votes(gocontext.Background(), req)
	require.NoError(t, err)
	require.NotEmpty(t, votes.Votes)
	require.Empty(t, votes.Res)
}
