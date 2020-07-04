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

func TestGRPCQueryProposal(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.GovKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	res, err := queryClient.Proposal(gocontext.Background(), &types.QueryProposalRequest{})
	require.Error(t, err)
	require.Nil(t, res)

	res, err = queryClient.Proposal(gocontext.Background(), &types.QueryProposalRequest{ProposalId: 1})
	require.Error(t, err)
	require.Nil(t, res)

	testProposal := types.NewTextProposal("Proposal", "testing proposal")
	submittedProposal, err := app.GovKeeper.SubmitProposal(ctx, testProposal)
	require.NoError(t, err)

	proposal, err := queryClient.Proposal(gocontext.Background(), &types.QueryProposalRequest{ProposalId: submittedProposal.ProposalID})
	require.NoError(t, err)
	require.Equal(t, proposal.Proposal.ProposalID, uint64(1))

	proposalFromKeeper, found := app.GovKeeper.GetProposal(ctx, submittedProposal.ProposalID)
	require.True(t, found)
	require.Equal(t, proposal.Proposal.Content.GetValue(), proposalFromKeeper.Content.GetValue())
}

func TestGRPCQueryProposals(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.GovKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	addrs := simapp.AddTestAddrsIncremental(app, ctx, 2, sdk.NewInt(30000000))

	// req := &types.QueryProposalsRequest{Req: &query.PageRequest{}}
	req := &types.QueryProposalsRequest{}
	proposals, err := queryClient.Proposals(gocontext.Background(), req)
	require.Error(t, err)
	require.Empty(t, proposals)

	// check for the proposals with no proposal added should return null.
	req = &types.QueryProposalsRequest{Req: &query.PageRequest{Limit: 1}}

	proposals, err = queryClient.Proposals(gocontext.Background(), req)
	require.NoError(t, err)
	require.Empty(t, proposals.Proposals)

	// create 2 test proposals
	for i := 0; i < 2; i++ {
		num := strconv.Itoa(i + 1)
		testProposal := types.NewTextProposal("Proposal"+num, "testing proposal "+num)
		proposal, err := app.GovKeeper.SubmitProposal(ctx, testProposal)
		require.NotEmpty(t, proposal)
		require.NoError(t, err)
	}

	// Query for proposals after adding 2 proposals to the store.
	// give page limit as 1 and expect NextKey should not to be empty
	proposals, err = queryClient.Proposals(gocontext.Background(), req)
	require.NoError(t, err)
	require.Len(t, proposals.Proposals, 1)
	require.NotEmpty(t, proposals.Res.NextKey)

	proposalFromKeeper1, found := app.GovKeeper.GetProposal(ctx, 1)
	require.True(t, found)
	require.NotEmpty(t, proposalFromKeeper1)
	require.Equal(t, proposals.Proposals[0].Content.GetValue(), proposalFromKeeper1.Content.GetValue())

	req = &types.QueryProposalsRequest{
		Req: &query.PageRequest{
			Key:   proposals.Res.NextKey,
			Limit: 1,
		},
	}

	// query for the next page which is 2nd proposal at present context.
	// and expect NextKey should be empty
	proposals, err = queryClient.Proposals(gocontext.Background(), req)

	t.Log(proposals.Res)
	require.NoError(t, err)
	require.Len(t, proposals.Proposals, 1)
	require.Empty(t, proposals.Res.NextKey)

	proposalFromKeeper2, found := app.GovKeeper.GetProposal(ctx, 2)
	require.True(t, found)
	require.NotEmpty(t, proposalFromKeeper2)
	require.Equal(t, proposals.Proposals[0].Content.GetValue(), proposalFromKeeper2.Content.GetValue())

	req = &types.QueryProposalsRequest{Req: &query.PageRequest{Limit: 2}}

	// Query the page with limit 2 and expect NextKey should ne nil
	proposals, err = queryClient.Proposals(gocontext.Background(), req)

	require.NoError(t, err)
	require.Len(t, proposals.Proposals, 2)
	require.Empty(t, proposals.Res)

	proposal, found := app.GovKeeper.GetProposal(ctx, 1)
	require.True(t, found)
	require.NotNil(t, proposal)

	// filter proposals
	proposal.Status = types.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	req = &types.QueryProposalsRequest{
		ProposalStatus: types.StatusVotingPeriod,
		Req:            &query.PageRequest{Limit: 1},
	}

	proposals, err = queryClient.Proposals(gocontext.Background(), req)
	require.NoError(t, err)
	require.Len(t, proposals.Proposals, 1)
	require.Empty(t, proposals.Res)
	require.Equal(t, proposals.Proposals[0].Status, types.StatusVotingPeriod)

	req = &types.QueryProposalsRequest{
		Depositor: addrs[0],
		Req:       &query.PageRequest{Limit: 1},
	}

	depositCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(20)))
	deposit := types.NewDeposit(1, addrs[0], depositCoins)
	app.GovKeeper.SetDeposit(ctx, deposit)

	proposals, err = queryClient.Proposals(gocontext.Background(), req)
	require.NoError(t, err)
	require.Len(t, proposals.Proposals, 1)
	require.Empty(t, proposals.Res)
	require.Equal(t, proposals.Proposals[0].ProposalID, uint64(1))

	req = &types.QueryProposalsRequest{
		Voter: addrs[0],
		Req:   &query.PageRequest{Limit: 1},
	}

	require.NoError(t, app.GovKeeper.AddVote(ctx, 1, addrs[0], types.OptionAbstain))

	require.NoError(t, err)
	require.Len(t, proposals.Proposals, 1)
	require.Empty(t, proposals.Res)
	require.Equal(t, proposals.Proposals[0].ProposalID, uint64(1))
}

func TestGRPCQueryVote(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.GovKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	addrs := simapp.AddTestAddrsIncremental(app, ctx, 1, sdk.NewInt(30000000))

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID

	req := &types.QueryVoteRequest{ProposalId: proposalID}

	vote, err := queryClient.Vote(gocontext.Background(), req)
	require.Error(t, err)
	require.Nil(t, vote)

	req = &types.QueryVoteRequest{
		ProposalId: proposalID,
		Voter:      addrs[0],
	}

	vote, err = queryClient.Vote(gocontext.Background(), req)
	require.Error(t, err)
	require.Nil(t, vote)

	proposal.Status = types.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[0], types.OptionAbstain))
	voteFromKeeper, found := app.GovKeeper.GetVote(ctx, proposalID, addrs[0])
	require.True(t, found)
	require.NotEmpty(t, voteFromKeeper)

	req = &types.QueryVoteRequest{
		ProposalId: proposalID,
		Voter:      addrs[0],
	}

	vote, err = queryClient.Vote(gocontext.Background(), req)

	require.NoError(t, err)
	require.Equal(t, vote.Vote.Option, types.OptionAbstain)
	require.Equal(t, voteFromKeeper, vote.Vote)
}

func TestGRPCQueryVotes(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.GovKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	addrs := simapp.AddTestAddrsIncremental(app, ctx, 2, sdk.NewInt(30000000))

	// query vote when no vote registered and expect null.
	pageReq := &query.PageRequest{Limit: 1}

	req := &types.QueryVotesRequest{
		ProposalId: 0,
		Req:        pageReq,
	}

	_, err := queryClient.Votes(gocontext.Background(), req)
	require.Error(t, err)

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID

	proposal.Status = types.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	req = &types.QueryVotesRequest{
		ProposalId: proposalID,
		Req:        pageReq,
	}

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
	pageReq = &query.PageRequest{Limit: 1}

	req = &types.QueryVotesRequest{
		ProposalId: proposalID,
		Req:        pageReq,
	}

	votes, err = queryClient.Votes(gocontext.Background(), req)
	require.NoError(t, err)
	require.Len(t, votes.Votes, 1)
	require.Equal(t, types.OptionAbstain, votes.Votes[0].Option)
	require.NotEmpty(t, votes.Res)

	// query vote with limit 1, next key and expect NextKey to be nil.
	pageReq = &query.PageRequest{
		Key:   votes.Res.NextKey,
		Limit: 1,
	}

	req = &types.QueryVotesRequest{
		ProposalId: proposalID,
		Req:        pageReq,
	}

	votes, err = queryClient.Votes(gocontext.Background(), req)
	require.NoError(t, err)
	require.Len(t, votes.Votes, 1)
	require.Equal(t, types.OptionNoWithVeto, votes.Votes[0].Option)
	require.Empty(t, votes.Res)

	// query vote with limit 2 and expect NextKey to be nil.
	pageReq = &query.PageRequest{Limit: 2}

	req = &types.QueryVotesRequest{
		ProposalId: proposalID,
		Req:        pageReq,
	}

	votes, err = queryClient.Votes(gocontext.Background(), req)
	require.NoError(t, err)
	require.Len(t, votes.Votes, 2)
	require.Empty(t, votes.Res)
}

func TestGRPCQueryParams(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
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
	require.False(t, params.DepositParams.MinDeposit.IsZero())
	require.Equal(t, app.GovKeeper.GetDepositParams(ctx), params.DepositParams)

	req.ParamsType = types.ParamVoting
	params, err = queryClient.Params(gocontext.Background(), req)
	require.NoError(t, err)
	require.NotZero(t, params.VotingParams.VotingPeriod)
	require.Equal(t, app.GovKeeper.GetVotingParams(ctx), params.VotingParams)

	req.ParamsType = types.ParamTallying
	params, err = queryClient.Params(gocontext.Background(), req)
	require.NoError(t, err)
	require.False(t, params.TallyParams.Quorum.IsZero())
	require.False(t, params.TallyParams.Threshold.IsZero())
	require.False(t, params.TallyParams.Veto.IsZero())
	require.Equal(t, app.GovKeeper.GetTallyParams(ctx), params.TallyParams)
}

func TestGRPCQueryDeposit(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.GovKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	addrs := simapp.AddTestAddrsIncremental(app, ctx, 1, sdk.NewInt(30000000))

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID

	req := &types.QueryDepositRequest{ProposalId: proposalID}

	_, err = queryClient.Deposit(gocontext.Background(), req)
	require.Error(t, err)

	req = &types.QueryDepositRequest{
		ProposalId: proposalID,
		Depositor:  addrs[0],
	}

	_, err = queryClient.Deposit(gocontext.Background(), req)
	require.Error(t, err)

	depositCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(20)))
	deposit := types.NewDeposit(proposalID, addrs[0], depositCoins)
	app.GovKeeper.SetDeposit(ctx, deposit)

	req = &types.QueryDepositRequest{
		ProposalId: proposalID,
		Depositor:  addrs[0],
	}

	d, err := queryClient.Deposit(gocontext.Background(), req)

	require.NoError(t, err)
	require.False(t, d.Deposit.Empty())
	require.Equal(t, d.Deposit.Amount, depositCoins)

	depositFromKeeper, found := app.GovKeeper.GetDeposit(ctx, proposalID, addrs[0])
	require.True(t, found)
	require.Equal(t, depositFromKeeper, d.Deposit)
}

func TestGRPCQueryDeposits(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.GovKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	addrs := simapp.AddTestAddrsIncremental(app, ctx, 2, sdk.NewInt(30000000))

	pageReq := &query.PageRequest{Limit: 1}

	req := &types.QueryDepositsRequest{
		ProposalId: 0,
		Req:        pageReq,
	}

	_, err := queryClient.Deposits(gocontext.Background(), req)
	require.Error(t, err)

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID

	req = &types.QueryDepositsRequest{
		ProposalId: proposalID,
		Req:        pageReq,
	}

	deposits, err := queryClient.Deposits(gocontext.Background(), req)
	require.NoError(t, err)
	require.Empty(t, deposits.Deposits)
	require.Empty(t, deposits.Res)

	depositAmount1 := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(20)))
	deposit1 := types.NewDeposit(proposalID, addrs[0], depositAmount1)
	app.GovKeeper.SetDeposit(ctx, deposit1)

	depositAmount2 := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(30)))
	deposit2 := types.NewDeposit(proposalID, addrs[1], depositAmount2)
	app.GovKeeper.SetDeposit(ctx, deposit2)

	proposal.Status = types.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	pageReq = &query.PageRequest{Limit: 1}

	req = &types.QueryDepositsRequest{
		ProposalId: proposalID,
		Req:        pageReq,
	}

	deposits, err = queryClient.Deposits(gocontext.Background(), req)

	require.NoError(t, err)
	require.Len(t, deposits.Deposits, 1)
	require.Equal(t, depositAmount1, deposits.Deposits[0].Amount)
	require.NotEmpty(t, deposits.Res.NextKey)

	// query vote with limit 1, next key and expect NextKey to be nil.
	pageReq = &query.PageRequest{
		Key:        deposits.Res.NextKey,
		Limit:      1,
		CountTotal: false,
	}

	req = &types.QueryDepositsRequest{
		ProposalId: proposalID,
		Req:        pageReq,
	}

	deposits, err = queryClient.Deposits(gocontext.Background(), req)
	require.NoError(t, err)
	require.Len(t, deposits.Deposits, 1)
	require.Equal(t, depositAmount2, deposits.Deposits[0].Amount)
	require.Empty(t, deposits.Res)

	// query vote with limit 2 and expect NextKey to be nil.
	pageReq = &query.PageRequest{Limit: 2}

	req = &types.QueryDepositsRequest{
		ProposalId: proposalID,
		Req:        pageReq,
	}

	deposits, err = queryClient.Deposits(gocontext.Background(), req)
	require.NoError(t, err)
	require.Len(t, deposits.Deposits, 2)
	require.Empty(t, deposits.Res)
}

func TestGRPCQueryTally(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.GovKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	addrs, _ := createValidators(ctx, app, []int64{5, 5, 5})
	tp := TestProposal

	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID

	req := &types.QueryTallyResultRequest{ProposalId: 0}
	_, err = queryClient.TallyResult(gocontext.Background(), req)
	require.Error(t, err)

	req = &types.QueryTallyResultRequest{ProposalId: 2}
	_, err = queryClient.TallyResult(gocontext.Background(), req)
	require.Error(t, err)

	req = &types.QueryTallyResultRequest{ProposalId: proposalID}
	tally, err := queryClient.TallyResult(gocontext.Background(), req)
	require.NoError(t, err)
	require.Equal(t, tally.Tally, types.EmptyTallyResult())

	proposal.Status = types.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[0], types.OptionYes))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[1], types.OptionYes))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[2], types.OptionYes))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)

	req = &types.QueryTallyResultRequest{ProposalId: proposalID}
	tally, err = queryClient.TallyResult(gocontext.Background(), req)

	require.NoError(t, err)
	require.NotEmpty(t, tally.Tally)
	require.NotZero(t, tally.Tally.Yes)

	proposal.Status = types.StatusPassed
	app.GovKeeper.SetProposal(ctx, proposal)

	req = &types.QueryTallyResultRequest{ProposalId: proposalID}
	tally, err = queryClient.TallyResult(gocontext.Background(), req)

	require.NoError(t, err)
	require.NotEmpty(t, tally.Tally)
	require.NotZero(t, tally.Tally.Yes)
}
