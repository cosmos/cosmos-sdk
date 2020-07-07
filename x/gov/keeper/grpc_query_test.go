package keeper_test

import (
	"bytes"
	gocontext "context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app         *simapp.SimApp
	ctx         sdk.Context
	queryClient types.QueryClient
}

func (suite *KeeperTestSuite) SetupTest() {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.GovKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	suite.app, suite.ctx, suite.queryClient = app, ctx, queryClient
}

func (suite *KeeperTestSuite) TestGRPCQueryProposal() {
	app, ctx, queryClient := suite.app, suite.ctx, suite.queryClient

	res, err := queryClient.Proposal(gocontext.Background(), &types.QueryProposalRequest{})
	suite.Error(err)
	suite.Nil(res)

	res, err = queryClient.Proposal(gocontext.Background(), &types.QueryProposalRequest{ProposalId: 1})
	suite.Error(err)
	suite.Nil(res)

	testProposal := types.NewTextProposal("Proposal", "testing proposal")
	submittedProposal, err := app.GovKeeper.SubmitProposal(ctx, testProposal)
	suite.NoError(err)

	proposal, err := queryClient.Proposal(gocontext.Background(), &types.QueryProposalRequest{ProposalId: submittedProposal.ProposalID})
	suite.NoError(err)
	suite.Equal(proposal.Proposal.ProposalID, uint64(1))

	proposalFromKeeper, found := app.GovKeeper.GetProposal(ctx, submittedProposal.ProposalID)
	suite.True(found)
	suite.Equal(proposal.Proposal.Content.GetValue(), proposalFromKeeper.Content.GetValue())
}

func (suite *KeeperTestSuite) TestGRPCQueryProposals() {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.GovKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	addrs := simapp.AddTestAddrsIncremental(app, ctx, 2, sdk.NewInt(30000000))

	req := &types.QueryProposalsRequest{}
	proposals, err := queryClient.Proposals(gocontext.Background(), req)
	suite.Error(err)
	suite.Empty(proposals)

	suite.T().Log("should return nil if no proposals are created")
	req = &types.QueryProposalsRequest{Req: &query.PageRequest{Limit: 1}}

	proposals, err = queryClient.Proposals(gocontext.Background(), req)
	suite.NoError(err)
	suite.Empty(proposals.Proposals)

	// create 2 test proposals
	for i := 0; i < 2; i++ {
		num := strconv.Itoa(i + 1)
		testProposal := types.NewTextProposal("Proposal"+num, "testing proposal "+num)
		proposal, err := app.GovKeeper.SubmitProposal(ctx, testProposal)
		suite.NotEmpty(proposal)
		suite.NoError(err)
	}

	// Query for proposals after adding 2 proposals to the store.
	// give page limit as 1 and expect NextKey should not to be empty
	proposals, err = queryClient.Proposals(gocontext.Background(), req)
	suite.NoError(err)
	suite.Len(proposals.Proposals, 1)
	suite.NotEmpty(proposals.Res.NextKey)
	suite.Equal(bytes.Trim(proposals.Res.NextKey, "\x00"), []byte{2})

	proposalFromKeeper1, found := app.GovKeeper.GetProposal(ctx, 1)
	suite.True(found)
	suite.NotEmpty(proposalFromKeeper1)
	suite.Equal(proposals.Proposals[0].Content.GetValue(), proposalFromKeeper1.Content.GetValue())

	req = &types.QueryProposalsRequest{
		Req: &query.PageRequest{
			Key:   proposals.Res.NextKey,
			Limit: 1,
		},
	}

	// query for the next page which is 2nd proposal at present context.
	// and expect NextKey should be empty
	proposals, err = queryClient.Proposals(gocontext.Background(), req)

	suite.T().Log(proposals.Res)
	suite.NoError(err)
	suite.Len(proposals.Proposals, 1)
	suite.Empty(proposals.Res.NextKey)

	proposalFromKeeper2, found := app.GovKeeper.GetProposal(ctx, 2)
	suite.True(found)
	suite.NotEmpty(proposalFromKeeper2)
	suite.Equal(proposals.Proposals[0].Content.GetValue(), proposalFromKeeper2.Content.GetValue())

	req = &types.QueryProposalsRequest{Req: &query.PageRequest{Limit: 2}}

	// Query the page with limit 2 and expect NextKey should ne nil
	proposals, err = queryClient.Proposals(gocontext.Background(), req)

	suite.NoError(err)
	suite.Len(proposals.Proposals, 2)
	suite.Empty(proposals.Res)

	proposal, found := app.GovKeeper.GetProposal(ctx, 1)
	suite.True(found)
	suite.NotNil(proposal)

	// filter proposals
	proposal.Status = types.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	req = &types.QueryProposalsRequest{
		ProposalStatus: types.StatusVotingPeriod,
		Req:            &query.PageRequest{Limit: 1},
	}

	proposals, err = queryClient.Proposals(gocontext.Background(), req)
	suite.NoError(err)
	suite.Len(proposals.Proposals, 1)
	suite.Empty(proposals.Res)
	suite.Equal(proposals.Proposals[0].Status, types.StatusVotingPeriod)

	req = &types.QueryProposalsRequest{
		Depositor: addrs[0],
		Req:       &query.PageRequest{Limit: 1},
	}

	depositCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(20)))
	deposit := types.NewDeposit(1, addrs[0], depositCoins)
	app.GovKeeper.SetDeposit(ctx, deposit)

	proposals, err = queryClient.Proposals(gocontext.Background(), req)
	suite.NoError(err)
	suite.Len(proposals.Proposals, 1)
	suite.Empty(proposals.Res)
	suite.Equal(proposals.Proposals[0].ProposalID, uint64(1))

	req = &types.QueryProposalsRequest{
		Voter: addrs[0],
		Req:   &query.PageRequest{Limit: 1},
	}

	suite.NoError(app.GovKeeper.AddVote(ctx, 1, addrs[0], types.OptionAbstain))

	suite.NoError(err)
	suite.Len(proposals.Proposals, 1)
	suite.Empty(proposals.Res)
	suite.Equal(proposals.Proposals[0].ProposalID, uint64(1))
}

func (suite *KeeperTestSuite) TestGRPCQueryVote() {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.GovKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	addrs := simapp.AddTestAddrsIncremental(app, ctx, 1, sdk.NewInt(30000000))

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	suite.NoError(err)
	proposalID := proposal.ProposalID

	req := &types.QueryVoteRequest{ProposalId: proposalID}

	vote, err := queryClient.Vote(gocontext.Background(), req)
	suite.Error(err)
	suite.Nil(vote)

	req = &types.QueryVoteRequest{
		ProposalId: proposalID,
		Voter:      addrs[0],
	}

	vote, err = queryClient.Vote(gocontext.Background(), req)
	suite.Error(err)
	suite.Nil(vote)

	proposal.Status = types.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	suite.NoError(app.GovKeeper.AddVote(ctx, proposalID, addrs[0], types.OptionAbstain))
	voteFromKeeper, found := app.GovKeeper.GetVote(ctx, proposalID, addrs[0])
	suite.True(found)
	suite.NotEmpty(voteFromKeeper)

	req = &types.QueryVoteRequest{
		ProposalId: proposalID,
		Voter:      addrs[0],
	}

	vote, err = queryClient.Vote(gocontext.Background(), req)

	suite.NoError(err)
	suite.Equal(vote.Vote.Option, types.OptionAbstain)
	suite.Equal(voteFromKeeper, vote.Vote)
}

func (suite *KeeperTestSuite) TestGRPCQueryVotes() {
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
	suite.Error(err)

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	suite.NoError(err)
	proposalID := proposal.ProposalID

	proposal.Status = types.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	req = &types.QueryVotesRequest{
		ProposalId: proposalID,
		Req:        pageReq,
	}

	votes, err := queryClient.Votes(gocontext.Background(), req)

	suite.NoError(err)
	suite.Empty(votes.Votes)
	suite.Empty(votes.Res)

	// Register two votes with different addresses
	// Test first vote
	suite.NoError(app.GovKeeper.AddVote(ctx, proposalID, addrs[0], types.OptionAbstain))
	_, found := app.GovKeeper.GetVote(ctx, proposalID, addrs[0])
	suite.True(found)

	// Test second vote
	suite.NoError(app.GovKeeper.AddVote(ctx, proposalID, addrs[1], types.OptionNoWithVeto))
	_, found = app.GovKeeper.GetVote(ctx, proposalID, addrs[1])
	suite.True(found)

	// query vote with limit 1 and expect next should not be nil.
	pageReq = &query.PageRequest{Limit: 1}

	req = &types.QueryVotesRequest{
		ProposalId: proposalID,
		Req:        pageReq,
	}

	votes, err = queryClient.Votes(gocontext.Background(), req)
	suite.NoError(err)
	suite.Len(votes.Votes, 1)
	suite.Equal(types.OptionAbstain, votes.Votes[0].Option)
	suite.NotEmpty(votes.Res)

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
	suite.NoError(err)
	suite.Len(votes.Votes, 1)
	suite.Equal(types.OptionNoWithVeto, votes.Votes[0].Option)
	suite.Empty(votes.Res)

	// query vote with limit 2 and expect NextKey to be nil.
	pageReq = &query.PageRequest{Limit: 2}

	req = &types.QueryVotesRequest{
		ProposalId: proposalID,
		Req:        pageReq,
	}

	votes, err = queryClient.Votes(gocontext.Background(), req)
	suite.NoError(err)
	suite.Len(votes.Votes, 2)
	suite.Empty(votes.Res)
}

func (suite *KeeperTestSuite) TestGRPCQueryParams() {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.GovKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	req := &types.QueryParamsRequest{
		ParamsType: "",
	}

	_, err := queryClient.Params(gocontext.Background(), req)
	suite.Error(err)

	req.ParamsType = "wrong parmas type"
	_, err = queryClient.Params(gocontext.Background(), req)
	suite.Error(err)

	req.ParamsType = types.ParamDeposit
	params, err := queryClient.Params(gocontext.Background(), req)
	suite.NoError(err)
	suite.False(params.DepositParams.MinDeposit.IsZero())
	suite.Equal(app.GovKeeper.GetDepositParams(ctx), params.DepositParams)

	req.ParamsType = types.ParamVoting
	params, err = queryClient.Params(gocontext.Background(), req)
	suite.NoError(err)
	suite.NotZero(params.VotingParams.VotingPeriod)
	suite.Equal(app.GovKeeper.GetVotingParams(ctx), params.VotingParams)

	req.ParamsType = types.ParamTallying
	params, err = queryClient.Params(gocontext.Background(), req)
	suite.NoError(err)
	suite.False(params.TallyParams.Quorum.IsZero())
	suite.False(params.TallyParams.Threshold.IsZero())
	suite.False(params.TallyParams.Veto.IsZero())
	suite.Equal(app.GovKeeper.GetTallyParams(ctx), params.TallyParams)
}

func (suite *KeeperTestSuite) TestGRPCQueryDeposit() {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.GovKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	addrs := simapp.AddTestAddrsIncremental(app, ctx, 1, sdk.NewInt(30000000))

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	suite.NoError(err)
	proposalID := proposal.ProposalID

	req := &types.QueryDepositRequest{ProposalId: proposalID}

	_, err = queryClient.Deposit(gocontext.Background(), req)
	suite.Error(err)

	req = &types.QueryDepositRequest{
		ProposalId: proposalID,
		Depositor:  addrs[0],
	}

	_, err = queryClient.Deposit(gocontext.Background(), req)
	suite.Error(err)

	depositCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(20)))
	deposit := types.NewDeposit(proposalID, addrs[0], depositCoins)
	app.GovKeeper.SetDeposit(ctx, deposit)

	req = &types.QueryDepositRequest{
		ProposalId: proposalID,
		Depositor:  addrs[0],
	}

	d, err := queryClient.Deposit(gocontext.Background(), req)

	suite.NoError(err)
	suite.False(d.Deposit.Empty())
	suite.Equal(d.Deposit.Amount, depositCoins)

	depositFromKeeper, found := app.GovKeeper.GetDeposit(ctx, proposalID, addrs[0])
	suite.True(found)
	suite.Equal(depositFromKeeper, d.Deposit)
}

func (suite *KeeperTestSuite) TestGRPCQueryDeposits() {
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
	suite.Error(err)

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	suite.NoError(err)
	proposalID := proposal.ProposalID

	req = &types.QueryDepositsRequest{
		ProposalId: proposalID,
		Req:        pageReq,
	}

	deposits, err := queryClient.Deposits(gocontext.Background(), req)
	suite.NoError(err)
	suite.Empty(deposits.Deposits)
	suite.Empty(deposits.Res)

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

	suite.NoError(err)
	suite.Len(deposits.Deposits, 1)
	suite.Equal(depositAmount1, deposits.Deposits[0].Amount)
	suite.NotEmpty(deposits.Res.NextKey)

	// query vote with limit 1, next key and expect NextKey to be nil.
	pageReq = &query.PageRequest{
		Key:   deposits.Res.NextKey,
		Limit: 1,
	}

	req = &types.QueryDepositsRequest{
		ProposalId: proposalID,
		Req:        pageReq,
	}

	deposits, err = queryClient.Deposits(gocontext.Background(), req)
	suite.NoError(err)
	suite.Len(deposits.Deposits, 1)
	suite.Equal(depositAmount2, deposits.Deposits[0].Amount)
	suite.Empty(deposits.Res)

	// query vote with limit 2 and expect NextKey to be nil.
	pageReq = &query.PageRequest{Limit: 2}

	req = &types.QueryDepositsRequest{
		ProposalId: proposalID,
		Req:        pageReq,
	}

	deposits, err = queryClient.Deposits(gocontext.Background(), req)
	suite.NoError(err)
	suite.Len(deposits.Deposits, 2)
	suite.Empty(deposits.Res)
}

func (suite *KeeperTestSuite) TestGRPCQueryTally() {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.GovKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	addrs, _ := createValidators(ctx, app, []int64{5, 5, 5})
	tp := TestProposal

	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	suite.NoError(err)
	proposalID := proposal.ProposalID

	req := &types.QueryTallyResultRequest{ProposalId: 0}
	_, err = queryClient.TallyResult(gocontext.Background(), req)
	suite.Error(err)

	req = &types.QueryTallyResultRequest{ProposalId: 2}
	_, err = queryClient.TallyResult(gocontext.Background(), req)
	suite.Error(err)

	req = &types.QueryTallyResultRequest{ProposalId: proposalID}
	tally, err := queryClient.TallyResult(gocontext.Background(), req)
	suite.NoError(err)
	suite.Equal(tally.Tally, types.EmptyTallyResult())

	proposal.Status = types.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	suite.NoError(app.GovKeeper.AddVote(ctx, proposalID, addrs[0], types.OptionYes))
	suite.NoError(app.GovKeeper.AddVote(ctx, proposalID, addrs[1], types.OptionYes))
	suite.NoError(app.GovKeeper.AddVote(ctx, proposalID, addrs[2], types.OptionYes))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	suite.True(ok)

	req = &types.QueryTallyResultRequest{ProposalId: proposalID}
	tally, err = queryClient.TallyResult(gocontext.Background(), req)

	suite.NoError(err)
	suite.NotEmpty(tally.Tally)
	suite.NotZero(tally.Tally.Yes)

	proposal.Status = types.StatusPassed
	app.GovKeeper.SetProposal(ctx, proposal)

	req = &types.QueryTallyResultRequest{ProposalId: proposalID}
	tally, err = queryClient.TallyResult(gocontext.Background(), req)

	suite.NoError(err)
	suite.NotEmpty(tally.Tally)
	suite.NotZero(tally.Tally.Yes)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
