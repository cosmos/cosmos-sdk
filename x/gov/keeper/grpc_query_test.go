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

func TestGRPCProposal(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx)
	types.RegisterQueryServer(queryHelper, app.GovKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	pageReq := &query.PageRequest{
		Key:        nil,
		Limit:      0,
		CountTotal: false,
	}

	req := types.NewQueryProposalRequest(1, pageReq)

	_, err := queryClient.Proposal(gocontext.Background(), req)
	require.Error(t, err)

	_, err = queryClient.Proposal(gocontext.Background(), &types.QueryProposalRequest{})
	require.Error(t, err)

	testProposal := types.NewTextProposal("Proposal", "testing proposal")
	submittedProposal, err := app.GovKeeper.SubmitProposal(ctx, testProposal)
	require.NoError(t, err)

	req = types.NewQueryProposalRequest(submittedProposal.ProposalID, pageReq)

	proposal, err := queryClient.Proposal(gocontext.Background(), req)
	require.NoError(t, err)
	require.NotEmpty(t, proposal)
}

func TestGRPCProposals(t *testing.T) {
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

	proposals, err := queryClient.Proposals(gocontext.Background(), req)
	require.NoError(t, err)
	require.Empty(t, proposals.Proposals)

	// create 2 test proposals
	for i := 0; i < 2; i++ {
		num := strconv.Itoa(i + 1)
		testProposal := types.NewTextProposal("Proposal"+num, "testing proposal "+num)
		_, err := app.GovKeeper.SubmitProposal(ctx, testProposal)
		require.NoError(t, err)
	}

	// Query for proposals after adding 2 proposals to the store.
	// give page limit as 1 and expect NextKey should not to be empty
	proposals, err = queryClient.Proposals(gocontext.Background(), req)
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
	proposals, err = queryClient.Proposals(gocontext.Background(), req)

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
	proposals, err = queryClient.Proposals(gocontext.Background(), req)

	require.NoError(t, err)
	require.NotEmpty(t, proposals.Proposals)
	require.Empty(t, proposals.Res)
}

func TestGRPCVote(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx)
	types.RegisterQueryServer(queryHelper, app.GovKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	addrs := simapp.AddTestAddrsIncremental(app, ctx, 1, sdk.NewInt(30000000))

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID

	req := &types.QueryVoteRequest{
		ProposalId: proposalID,
		Voter:      nil,
	}

	_, err = queryClient.Vote(gocontext.Background(), req)
	require.Error(t, err)

	req = &types.QueryVoteRequest{
		ProposalId: proposalID,
		Voter:      addrs[0],
	}

	_, err = queryClient.Vote(gocontext.Background(), req)
	require.Error(t, err)

	proposal.Status = types.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[0], types.OptionAbstain))
	_, found := app.GovKeeper.GetVote(ctx, proposalID, addrs[0])
	require.True(t, found)

	req = &types.QueryVoteRequest{
		ProposalId: proposalID,
		Voter:      addrs[0],
	}

	vote, err := queryClient.Vote(gocontext.Background(), req)

	require.NoError(t, err)
	require.NotEmpty(t, vote.Vote)
}

func TestGRPCVotes(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx)
	types.RegisterQueryServer(queryHelper, app.GovKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	addrs := simapp.AddTestAddrsIncremental(app, ctx, 2, sdk.NewInt(30000000))

	// query vote when no vote registered and expect null.
	pageReq := &query.PageRequest{
		Key:        nil,
		Limit:      1,
		CountTotal: false,
	}

	req := types.NewQueryProposalRequest(0, pageReq)
	_, err := queryClient.Votes(gocontext.Background(), req)
	require.Error(t, err)

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID

	proposal.Status = types.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	req = types.NewQueryProposalRequest(proposalID, pageReq)
	votes, err := queryClient.Votes(gocontext.Background(), req)

	require.NoError(t, err)
	require.Empty(t, votes.Votes)
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

	req = types.NewQueryProposalRequest(proposalID, pageReq)
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

	req = types.NewQueryProposalRequest(proposalID, pageReq)
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

	req = types.NewQueryProposalRequest(proposalID, pageReq)
	votes, err = queryClient.Votes(gocontext.Background(), req)
	require.NoError(t, err)
	require.NotEmpty(t, votes.Votes)
	require.Empty(t, votes.Res)
}

func TestGRPCParams(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx)
	types.RegisterQueryServer(queryHelper, app.GovKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	req := &types.QueryParamsRequest{
		ParamsType: "",
	}

	_, err := queryClient.Params(gocontext.Background(), req)
	require.Error(t, err)

	req.ParamsType = "wrong parmas type"
	_, err = queryClient.Params(gocontext.Background(), req)
	require.Error(t, err)

	req.ParamsType = types.ParamDeposit
	params, err := queryClient.Params(gocontext.Background(), req)
	require.NoError(t, err)
	require.NotEmpty(t, params.DepositParams)

	req.ParamsType = types.ParamVoting
	params, err = queryClient.Params(gocontext.Background(), req)
	require.NoError(t, err)
	require.NotEmpty(t, params.VotingParams)

	req.ParamsType = types.ParamTallying
	params, err = queryClient.Params(gocontext.Background(), req)
	require.NoError(t, err)
	require.NotEmpty(t, params.TallyParams)
}

func TestGRPCDeposit(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx)
	types.RegisterQueryServer(queryHelper, app.GovKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	addrs := simapp.AddTestAddrsIncremental(app, ctx, 1, sdk.NewInt(30000000))

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID

	req := &types.QueryDepositRequest{
		ProposalId: proposalID,
		Depositor:  nil,
	}

	_, err = queryClient.Deposit(gocontext.Background(), req)
	require.Error(t, err)

	req = &types.QueryDepositRequest{
		ProposalId: proposalID,
		Depositor:  addrs[0],
	}

	_, err = queryClient.Deposit(gocontext.Background(), req)
	require.Error(t, err)

	deposit := types.NewDeposit(proposalID, addrs[0], sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(20))))
	app.GovKeeper.SetDeposit(ctx, deposit)

	req = &types.QueryDepositRequest{
		ProposalId: proposalID,
		Depositor:  addrs[0],
	}

	d, err := queryClient.Deposit(gocontext.Background(), req)

	require.NoError(t, err)
	require.NotEmpty(t, d.Deposit)
}

func TestGRPCDeposits(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx)
	types.RegisterQueryServer(queryHelper, app.GovKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	addrs := simapp.AddTestAddrsIncremental(app, ctx, 2, sdk.NewInt(30000000))

	pageReq := &query.PageRequest{
		Key:        nil,
		Limit:      1,
		CountTotal: false,
	}

	req := types.NewQueryProposalRequest(0, pageReq)
	_, err := queryClient.Deposits(gocontext.Background(), req)
	require.Error(t, err)

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID

	req = types.NewQueryProposalRequest(proposalID, pageReq)
	deposits, err := queryClient.Deposits(gocontext.Background(), req)
	require.NoError(t, err)
	require.Empty(t, deposits.Deposits)
	require.Empty(t, deposits.Res)

	deposit1 := types.NewDeposit(proposalID, addrs[0], sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(20))))
	app.GovKeeper.SetDeposit(ctx, deposit1)

	deposit2 := types.NewDeposit(proposalID, addrs[1], sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(30))))
	app.GovKeeper.SetDeposit(ctx, deposit2)

	proposal.Status = types.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	pageReq = &query.PageRequest{
		Key:        nil,
		Limit:      1,
		CountTotal: false,
	}

	req = types.NewQueryProposalRequest(proposalID, pageReq)
	deposits, err = queryClient.Deposits(gocontext.Background(), req)

	require.NoError(t, err)
	require.NotEmpty(t, deposits.Deposits)
	require.NotEmpty(t, deposits.Res.NextKey)

	// query vote with limit 1, next key and expect NextKey to be nil.
	pageReq = &query.PageRequest{
		Key:        deposits.Res.NextKey,
		Limit:      1,
		CountTotal: false,
	}

	req = types.NewQueryProposalRequest(proposalID, pageReq)
	deposits, err = queryClient.Deposits(gocontext.Background(), req)
	require.NoError(t, err)
	require.NotEmpty(t, deposits.Deposits)
	require.Empty(t, deposits.Res)

	// query vote with limit 2 and expect NextKey to be nil.
	pageReq = &query.PageRequest{
		Key:        nil,
		Limit:      2,
		CountTotal: false,
	}

	req = types.NewQueryProposalRequest(proposalID, pageReq)
	deposits, err = queryClient.Deposits(gocontext.Background(), req)
	require.NoError(t, err)
	require.NotEmpty(t, deposits.Deposits)
	require.Empty(t, deposits.Res)
}

func TestGRPCTally(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx)
	types.RegisterQueryServer(queryHelper, app.GovKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	addrs, _ := createValidators(ctx, app, []int64{5, 5, 5})
	tp := TestProposal

	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID

	pageReq := &query.PageRequest{
		Key:        nil,
		Limit:      0,
		CountTotal: false,
	}

	req := types.NewQueryProposalRequest(0, pageReq)
	_, err = queryClient.TallyResult(gocontext.Background(), req)
	require.Error(t, err)

	req = types.NewQueryProposalRequest(2, pageReq)
	_, err = queryClient.TallyResult(gocontext.Background(), req)
	require.Error(t, err)

	req = types.NewQueryProposalRequest(proposalID, pageReq)
	tally, err := queryClient.TallyResult(gocontext.Background(), req)
	require.NoError(t, err)
	require.Equal(t, tally.Tally, types.EmptyTallyResult())

	proposal.Status = types.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[0], types.OptionYes))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)

	req = types.NewQueryProposalRequest(proposalID, pageReq)
	tally, err = queryClient.TallyResult(gocontext.Background(), req)

	require.NoError(t, err)
	require.NotEmpty(t, tally.Tally)

	proposal.Status = types.StatusPassed
	app.GovKeeper.SetProposal(ctx, proposal)

	req = types.NewQueryProposalRequest(proposalID, pageReq)
	tally, err = queryClient.TallyResult(gocontext.Background(), req)

	require.NoError(t, err)
	require.NotEmpty(t, tally.Tally)
}
