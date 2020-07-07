package keeper_test

import (
	"bytes"
	gocontext "context"
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
)

type GRPCTestSuite struct {
	suite.Suite

	app         *simapp.SimApp
	ctx         sdk.Context
	queryClient types.QueryClient
}

func (suite *GRPCTestSuite) SetupTest() {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.GovKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	suite.app = app
	suite.ctx = ctx
	suite.queryClient = queryClient
}

func (suite *GRPCTestSuite) TestGRPCQueryProposal() {
	app, ctx, queryClient := suite.app, suite.ctx, suite.queryClient

	res, err := queryClient.Proposal(gocontext.Background(), &types.QueryProposalRequest{})
	suite.Require().Error(err)
	suite.Require().Nil(res)

	res, err = queryClient.Proposal(gocontext.Background(), &types.QueryProposalRequest{ProposalId: 1})
	suite.Require().Error(err)
	suite.Require().Nil(res)

	testProposal := types.NewTextProposal("Proposal", "testing proposal")
	submittedProposal, err := app.GovKeeper.SubmitProposal(ctx, testProposal)
	suite.Require().NoError(err)

	proposal, err := queryClient.Proposal(gocontext.Background(), &types.QueryProposalRequest{ProposalId: submittedProposal.ProposalID})
	suite.Require().NoError(err)
	suite.Require().Equal(proposal.Proposal.ProposalID, uint64(1))

	proposalFromKeeper, found := app.GovKeeper.GetProposal(ctx, submittedProposal.ProposalID)
	suite.Require().True(found)
	suite.Require().Equal(proposal.Proposal.Content.GetValue(), proposalFromKeeper.Content.GetValue())
}

func (suite *GRPCTestSuite) TestGRPCQueryProposals() {
	app, ctx, queryClient := suite.app, suite.ctx, suite.queryClient

	addrs := simapp.AddTestAddrsIncremental(app, ctx, 2, sdk.NewInt(30000000))

	req := &types.QueryProposalsRequest{}
	proposals, err := queryClient.Proposals(gocontext.Background(), req)
	suite.Require().Error(err)
	suite.Require().Empty(proposals)

	suite.T().Log("should return nil if no proposals are created")
	req = &types.QueryProposalsRequest{Req: &query.PageRequest{Limit: 1}}

	proposals, err = queryClient.Proposals(gocontext.Background(), req)
	suite.Require().NoError(err)
	suite.Require().Empty(proposals.Proposals)

	// create 2 test proposals
	for i := 0; i < 2; i++ {
		num := strconv.Itoa(i + 1)
		testProposal := types.NewTextProposal("Proposal"+num, "testing proposal "+num)
		proposal, err := app.GovKeeper.SubmitProposal(ctx, testProposal)
		suite.Require().NotEmpty(proposal)
		suite.Require().NoError(err)
	}

	// Query for proposals after adding 2 proposals to the store.
	// give page limit as 1 and expect NextKey should not to be empty
	proposals, err = queryClient.Proposals(gocontext.Background(), req)
	suite.Require().NoError(err)
	suite.Require().Len(proposals.Proposals, 1)
	suite.Require().NotEmpty(proposals.Res.NextKey)
	suite.Require().Equal(bytes.Trim(proposals.Res.NextKey, "\x00"), []byte{2})

	proposalFromKeeper1, found := app.GovKeeper.GetProposal(ctx, 1)
	suite.Require().True(found)
	suite.Require().NotEmpty(proposalFromKeeper1)
	suite.Require().Equal(proposals.Proposals[0].Content.GetValue(), proposalFromKeeper1.Content.GetValue())

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
	suite.Require().NoError(err)
	suite.Require().Len(proposals.Proposals, 1)
	suite.Require().Empty(proposals.Res.NextKey)

	proposalFromKeeper2, found := app.GovKeeper.GetProposal(ctx, 2)
	suite.Require().True(found)
	suite.Require().NotEmpty(proposalFromKeeper2)
	suite.Require().Equal(proposals.Proposals[0].Content.GetValue(), proposalFromKeeper2.Content.GetValue())

	req = &types.QueryProposalsRequest{Req: &query.PageRequest{Limit: 2}}

	// Query the page with limit 2 and expect NextKey should ne nil
	proposals, err = queryClient.Proposals(gocontext.Background(), req)

	suite.Require().NoError(err)
	suite.Require().Len(proposals.Proposals, 2)
	suite.Require().Empty(proposals.Res)

	proposal, found := app.GovKeeper.GetProposal(ctx, 1)
	suite.Require().True(found)
	suite.Require().NotNil(proposal)

	// filter proposals
	proposal.Status = types.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	req = &types.QueryProposalsRequest{
		ProposalStatus: types.StatusVotingPeriod,
		Req:            &query.PageRequest{Limit: 1},
	}

	proposals, err = queryClient.Proposals(gocontext.Background(), req)
	suite.Require().NoError(err)
	suite.Require().Len(proposals.Proposals, 1)
	suite.Require().Empty(proposals.Res)
	suite.Require().Equal(proposals.Proposals[0].Status, types.StatusVotingPeriod)

	req = &types.QueryProposalsRequest{
		Depositor: addrs[0],
		Req:       &query.PageRequest{Limit: 1},
	}

	depositCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(20)))
	deposit := types.NewDeposit(1, addrs[0], depositCoins)
	app.GovKeeper.SetDeposit(ctx, deposit)

	proposals, err = queryClient.Proposals(gocontext.Background(), req)
	suite.Require().NoError(err)
	suite.Require().Len(proposals.Proposals, 1)
	suite.Require().Empty(proposals.Res)
	suite.Require().Equal(proposals.Proposals[0].ProposalID, uint64(1))

	req = &types.QueryProposalsRequest{
		Voter: addrs[0],
		Req:   &query.PageRequest{Limit: 1},
	}

	suite.Require().NoError(app.GovKeeper.AddVote(ctx, 1, addrs[0], types.OptionAbstain))

	suite.Require().NoError(err)
	suite.Require().Len(proposals.Proposals, 1)
	suite.Require().Empty(proposals.Res)
	suite.Require().Equal(proposals.Proposals[0].ProposalID, uint64(1))
}

func (suite *GRPCTestSuite) TestGRPCQueryVote() {
	app, ctx, queryClient := suite.app, suite.ctx, suite.queryClient

	addrs := simapp.AddTestAddrsIncremental(app, ctx, 1, sdk.NewInt(30000000))

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	suite.Require().NoError(err)
	proposalID := proposal.ProposalID

	req := &types.QueryVoteRequest{ProposalId: proposalID}

	vote, err := queryClient.Vote(gocontext.Background(), req)
	suite.Require().Error(err)
	suite.Require().Nil(vote)

	req = &types.QueryVoteRequest{
		ProposalId: proposalID,
		Voter:      addrs[0],
	}

	vote, err = queryClient.Vote(gocontext.Background(), req)
	suite.Require().Error(err)
	suite.Require().Nil(vote)

	proposal.Status = types.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	suite.Require().NoError(app.GovKeeper.AddVote(ctx, proposalID, addrs[0], types.OptionAbstain))
	voteFromKeeper, found := app.GovKeeper.GetVote(ctx, proposalID, addrs[0])
	suite.Require().True(found)
	suite.Require().NotEmpty(voteFromKeeper)

	req = &types.QueryVoteRequest{
		ProposalId: proposalID,
		Voter:      addrs[0],
	}

	vote, err = queryClient.Vote(gocontext.Background(), req)

	suite.Require().NoError(err)
	suite.Require().Equal(vote.Vote.Option, types.OptionAbstain)
	suite.Require().Equal(voteFromKeeper, vote.Vote)
}

func (suite *GRPCTestSuite) TestGRPCQueryVotes() {
	app, ctx, queryClient := suite.app, suite.ctx, suite.queryClient

	addrs := simapp.AddTestAddrsIncremental(app, ctx, 2, sdk.NewInt(30000000))

	// query vote when no vote registered and expect null.
	pageReq := &query.PageRequest{Limit: 1}

	req := &types.QueryVotesRequest{
		ProposalId: 0,
		Req:        pageReq,
	}

	votes, err := queryClient.Votes(gocontext.Background(), req)
	suite.Require().Error(err)
	suite.Require().Nil(votes)

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	suite.Require().NoError(err)
	proposalID := proposal.ProposalID

	proposal.Status = types.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	req = &types.QueryVotesRequest{
		ProposalId: proposalID,
		Req:        pageReq,
	}

	votes, err = queryClient.Votes(gocontext.Background(), req)

	suite.Require().NoError(err)
	suite.Require().Empty(votes.Votes)
	suite.Require().Empty(votes.Res)

	// Register two votes with different addresses
	// Test first vote
	suite.Require().NoError(app.GovKeeper.AddVote(ctx, proposalID, addrs[0], types.OptionAbstain))
	_, found := app.GovKeeper.GetVote(ctx, proposalID, addrs[0])
	suite.Require().True(found)

	// Test second vote
	suite.Require().NoError(app.GovKeeper.AddVote(ctx, proposalID, addrs[1], types.OptionNoWithVeto))
	_, found = app.GovKeeper.GetVote(ctx, proposalID, addrs[1])
	suite.Require().True(found)

	// query vote with limit 1 and expect next should not be nil.
	pageReq = &query.PageRequest{Limit: 1}

	req = &types.QueryVotesRequest{
		ProposalId: proposalID,
		Req:        pageReq,
	}

	votes, err = queryClient.Votes(gocontext.Background(), req)
	suite.Require().NoError(err)
	suite.Require().Len(votes.Votes, 1)
	suite.Require().Equal(types.OptionAbstain, votes.Votes[0].Option)
	suite.Require().NotEmpty(votes.Res)

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
	suite.Require().NoError(err)
	suite.Require().Len(votes.Votes, 1)
	suite.Require().Equal(types.OptionNoWithVeto, votes.Votes[0].Option)
	suite.Require().Empty(votes.Res)

	// query vote with limit 2 and expect NextKey to be nil.
	pageReq = &query.PageRequest{Limit: 2}

	req = &types.QueryVotesRequest{
		ProposalId: proposalID,
		Req:        pageReq,
	}

	votes, err = queryClient.Votes(gocontext.Background(), req)
	suite.Require().NoError(err)
	suite.Require().Len(votes.Votes, 2)
	suite.Require().Empty(votes.Res)
}

func (suite *GRPCTestSuite) TestGRPCQueryParams() {
	app, ctx, queryClient := suite.app, suite.ctx, suite.queryClient

	req := &types.QueryParamsRequest{
		ParamsType: "",
	}

	params, err := queryClient.Params(gocontext.Background(), req)
	suite.Require().Error(err)
	suite.Require().Nil(params)

	req.ParamsType = "wrong parmas type"
	params, err = queryClient.Params(gocontext.Background(), req)
	suite.Require().Error(err)
	suite.Require().Nil(params)

	req.ParamsType = types.ParamDeposit
	params, err = queryClient.Params(gocontext.Background(), req)
	suite.Require().NoError(err)
	suite.Require().False(params.DepositParams.MinDeposit.IsZero())
	suite.Require().Equal(app.GovKeeper.GetDepositParams(ctx), params.DepositParams)

	req.ParamsType = types.ParamVoting
	params, err = queryClient.Params(gocontext.Background(), req)
	suite.Require().NoError(err)
	suite.Require().NotZero(params.VotingParams.VotingPeriod)
	suite.Require().Equal(app.GovKeeper.GetVotingParams(ctx), params.VotingParams)

	req.ParamsType = types.ParamTallying
	params, err = queryClient.Params(gocontext.Background(), req)
	suite.Require().NoError(err)
	suite.Require().False(params.TallyParams.Quorum.IsZero())
	suite.Require().False(params.TallyParams.Threshold.IsZero())
	suite.Require().False(params.TallyParams.Veto.IsZero())
	suite.Require().Equal(app.GovKeeper.GetTallyParams(ctx), params.TallyParams)
}

func (suite *GRPCTestSuite) TestGRPCQueryDeposit() {
	app, ctx, queryClient := suite.app, suite.ctx, suite.queryClient

	addrs := simapp.AddTestAddrsIncremental(app, ctx, 1, sdk.NewInt(30000000))

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	suite.Require().NoError(err)
	proposalID := proposal.ProposalID

	req := &types.QueryDepositRequest{ProposalId: proposalID}

	d, err := queryClient.Deposit(gocontext.Background(), req)
	suite.Require().Error(err)
	suite.Require().Nil(d)

	req = &types.QueryDepositRequest{
		ProposalId: proposalID,
		Depositor:  addrs[0],
	}

	d, err = queryClient.Deposit(gocontext.Background(), req)
	suite.Require().Error(err)
	suite.Require().Nil(d)

	depositCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(20)))
	deposit := types.NewDeposit(proposalID, addrs[0], depositCoins)
	app.GovKeeper.SetDeposit(ctx, deposit)

	req = &types.QueryDepositRequest{
		ProposalId: proposalID,
		Depositor:  addrs[0],
	}

	d, err = queryClient.Deposit(gocontext.Background(), req)

	suite.Require().NoError(err)
	suite.Require().False(d.Deposit.Empty())
	suite.Require().Equal(d.Deposit.Amount, depositCoins)

	depositFromKeeper, found := app.GovKeeper.GetDeposit(ctx, proposalID, addrs[0])
	suite.Require().True(found)
	suite.Require().Equal(depositFromKeeper, d.Deposit)
}

func (suite *GRPCTestSuite) TestGRPCQueryDeposits() {
	app, ctx, queryClient := suite.app, suite.ctx, suite.queryClient

	addrs := simapp.AddTestAddrsIncremental(app, ctx, 2, sdk.NewInt(30000000))

	pageReq := &query.PageRequest{Limit: 1}

	req := &types.QueryDepositsRequest{
		ProposalId: 0,
		Req:        pageReq,
	}

	deposits, err := queryClient.Deposits(gocontext.Background(), req)
	suite.Require().Error(err)
	suite.Require().Nil(deposits)

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	suite.Require().NoError(err)
	proposalID := proposal.ProposalID

	req = &types.QueryDepositsRequest{
		ProposalId: proposalID,
		Req:        pageReq,
	}

	deposits, err = queryClient.Deposits(gocontext.Background(), req)
	suite.Require().NoError(err)
	suite.Require().Empty(deposits.Deposits)
	suite.Require().Empty(deposits.Res)

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

	suite.Require().NoError(err)
	suite.Require().Len(deposits.Deposits, 1)
	suite.Require().Equal(depositAmount1, deposits.Deposits[0].Amount)
	suite.Require().NotEmpty(deposits.Res.NextKey)

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
	suite.Require().NoError(err)
	suite.Require().Len(deposits.Deposits, 1)
	suite.Require().Equal(depositAmount2, deposits.Deposits[0].Amount)
	suite.Require().Empty(deposits.Res)

	// query vote with limit 2 and expect NextKey to be nil.
	pageReq = &query.PageRequest{Limit: 2}

	req = &types.QueryDepositsRequest{
		ProposalId: proposalID,
		Req:        pageReq,
	}

	deposits, err = queryClient.Deposits(gocontext.Background(), req)
	suite.Require().NoError(err)
	suite.Require().Len(deposits.Deposits, 2)
	suite.Require().Empty(deposits.Res)
}

func (suite *GRPCTestSuite) TestGRPCQueryTally() {
	app, ctx, queryClient := suite.app, suite.ctx, suite.queryClient

	addrs, _ := createValidators(ctx, app, []int64{5, 5, 5})
	tp := TestProposal

	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	suite.Require().NoError(err)
	proposalID := proposal.ProposalID

	req := &types.QueryTallyResultRequest{ProposalId: 0}
	tally, err := queryClient.TallyResult(gocontext.Background(), req)
	suite.Require().Error(err)
	suite.Require().Nil(tally)

	req = &types.QueryTallyResultRequest{ProposalId: 2}
	tally, err = queryClient.TallyResult(gocontext.Background(), req)
	suite.Require().Error(err)
	suite.Require().Nil(tally)

	req = &types.QueryTallyResultRequest{ProposalId: proposalID}
	tally, err = queryClient.TallyResult(gocontext.Background(), req)
	suite.Require().NoError(err)
	suite.Require().Equal(tally.Tally, types.EmptyTallyResult())

	proposal.Status = types.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	suite.Require().NoError(app.GovKeeper.AddVote(ctx, proposalID, addrs[0], types.OptionYes))
	suite.Require().NoError(app.GovKeeper.AddVote(ctx, proposalID, addrs[1], types.OptionYes))
	suite.Require().NoError(app.GovKeeper.AddVote(ctx, proposalID, addrs[2], types.OptionYes))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	suite.Require().True(ok)

	req = &types.QueryTallyResultRequest{ProposalId: proposalID}
	tally, err = queryClient.TallyResult(gocontext.Background(), req)

	suite.Require().NoError(err)
	suite.Require().NotEmpty(tally.Tally)
	suite.Require().NotZero(tally.Tally.Yes)

	proposal.Status = types.StatusPassed
	app.GovKeeper.SetProposal(ctx, proposal)

	req = &types.QueryTallyResultRequest{ProposalId: proposalID}
	tally, err = queryClient.TallyResult(gocontext.Background(), req)

	suite.Require().NoError(err)
	suite.Require().NotEmpty(tally.Tally)
	suite.Require().NotZero(tally.Tally.Yes)
}

func TestGRPCTestSuite(t *testing.T) {
	suite.Run(t, new(GRPCTestSuite))
}
