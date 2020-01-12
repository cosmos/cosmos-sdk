package keeper

import (
	"context"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	types1 "github.com/gogo/protobuf/types"
)

type QueryHandler struct {
	keeper Keeper
}

func (q QueryHandler) GetDepositParams(context.Context, *types1.Empty) (*types.DepositParams, error) {
	panic("implement me")
}

func (q QueryHandler) GetVotingParams(context.Context, *types1.Empty) (*types.VotingParams, error) {
	panic("implement me")
}

func (q QueryHandler) GetTallyParams(context.Context, *types1.Empty) (*types.TallyParams, error) {
	panic("implement me")
}

func (q QueryHandler) GetProposal(context.Context, *types.GetProposalRequest) (*types.AnyProposal, error) {
	panic("implement me")
}

func (q QueryHandler) GetDeposit(context.Context, *types.GetDepositRequest) (*types.Deposit, error) {
	panic("implement me")
}

func (q QueryHandler) GetVote(context.Context, *types.GetVoteRequest) (*types.Vote, error) {
	panic("implement me")
}

func (q QueryHandler) GetDeposits(context.Context, *types.GetProposalRequest) (*types.GetDepositsResponse, error) {
	panic("implement me")
}

func (q QueryHandler) GetTally(context.Context, *types.GetProposalRequest) (*types.TallyResult, error) {
	panic("implement me")
}

func (q QueryHandler) GetVotes(context.Context, *types.GetVotesRequest) (*types.GetVotesResponse, error) {
	panic("implement me")
}

func (q QueryHandler) GetProposals(context.Context, *types.GetProposalsRequest) (*types.GetProposalsResponse, error) {
	panic("implement me")
}

var _ types.QueryServer = QueryHandler{}

