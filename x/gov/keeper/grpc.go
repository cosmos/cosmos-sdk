package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	proto2 "github.com/gogo/protobuf/proto"
	proto "github.com/gogo/protobuf/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type QueryHandler struct {
	keeper Keeper
}

func SDKContext(ctx context.Context) sdk.Context {
	panic("TODO")
}

func (q QueryHandler) GetDepositParams(ctx context.Context, _ *proto.Empty) (*types.DepositParams, error) {
	res := q.keeper.GetDepositParams(SDKContext(ctx))
	return &res, nil
}

func (q QueryHandler) GetVotingParams(ctx context.Context, _ *proto.Empty) (*types.VotingParams, error) {
	res := q.keeper.GetVotingParams(SDKContext(ctx))
	return &res, nil
}

func (q QueryHandler) GetTallyParams(ctx context.Context, _ *proto.Empty) (*types.TallyParams, error) {
	res := q.keeper.GetTallyParams(SDKContext(ctx))
	return &res, nil
}

func (q QueryHandler) GetProposal(ctx context.Context, req *types.GetProposalRequest) (*types.AnyProposal, error) {
	res, found := q.keeper.GetProposalI(SDKContext(ctx), req.ProposalID)
	if !found {
		return nil, status.Error(codes.NotFound, "proposal not found")
	}
	content := res.GetContent()
	contentMsg, found := content.(proto2.Message)
	if !found {
		return nil, status.Error(codes.Internal, "invalid codec")
	}
	any, err := proto.MarshalAny(contentMsg)
	if err != nil {
		return nil, status.Error(codes.Internal, "invalid codec")
	}
	return &types.AnyProposal{
		ProposalBase: res.GetProposalBase(),
		Content:      any,
	}, nil
}

func (q QueryHandler) GetDeposit(ctx context.Context, req *types.GetDepositRequest) (*types.Deposit, error) {
	res, found := q.keeper.GetDeposit(SDKContext(ctx), req.ProposalID, req.Depositor)
	if !found {
		return nil, status.Error(codes.NotFound, "proposal not found")
	}
	return &res, nil
}

func (q QueryHandler) GetVote(ctx context.Context, req *types.GetVoteRequest) (*types.Vote, error) {
	res, found := q.keeper.GetVote(SDKContext(ctx), req.ProposalID, req.Voter)
	if !found {
		return nil, status.Error(codes.NotFound, "proposal not found")
	}
	return &res, nil
}

func (q QueryHandler) GetDeposits(ctx context.Context, req *types.GetProposalRequest) (*types.GetDepositsResponse, error) {
	res := q.keeper.GetDeposits(SDKContext(ctx), req.ProposalID)
	return &types.GetDepositsResponse{Deposits: res}, nil
}

func (q QueryHandler) GetTally(ctx context.Context, req *types.GetProposalRequest) (*types.TallyResult, error) {
	proposal, found := q.keeper.GetProposal(SDKContext(ctx), req.ProposalID)
	if !found {
		return nil, status.Error(codes.NotFound, "proposal not found")
	}

	var tallyResult types.TallyResult
	switch {
	case proposal.Status == types.StatusDepositPeriod:
		tallyResult = types.EmptyTallyResult()

	case proposal.Status == types.StatusPassed || proposal.Status == types.StatusRejected:
		tallyResult = proposal.FinalTallyResult

	default:
		// proposal is in voting period
		_, _, tallyResult = q.keeper.Tally(SDKContext(ctx), proposal)
	}
	return &tallyResult, nil
}

func (q QueryHandler) GetVotes(context.Context, *types.GetVotesRequest) (*types.GetVotesResponse, error) {
	panic("implement me")
}

func (q QueryHandler) GetProposals(context.Context, *types.GetProposalsRequest) (*types.GetProposalsResponse, error) {
	panic("implement me")
}

var _ types.QueryServer = QueryHandler{}
