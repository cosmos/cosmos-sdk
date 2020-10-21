package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

var _ types.QueryServer = Keeper{}

// Proposal returns proposal details based on ProposalID
func (q Keeper) Proposal(c context.Context, req *types.QueryProposalRequest) (*types.QueryProposalResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.ProposalId == 0 {
		return nil, status.Error(codes.InvalidArgument, "proposal id can not be 0")
	}

	ctx := sdk.UnwrapSDKContext(c)

	proposal, found := q.GetProposal(ctx, req.ProposalId)
	if !found {
		return nil, status.Errorf(codes.NotFound, "proposal %d doesn't exist", req.ProposalId)
	}

	return &types.QueryProposalResponse{Proposal: proposal}, nil
}

// Proposals implements the Query/Proposals gRPC method
func (q Keeper) Proposals(c context.Context, req *types.QueryProposalsRequest) (*types.QueryProposalsResponse, error) {
	var filteredProposals types.Proposals
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(q.storeKey)
	proposalStore := prefix.NewStore(store, types.ProposalsKeyPrefix)

	pageRes, err := query.FilteredPaginate(proposalStore, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		var p types.Proposal
		if err := q.cdc.UnmarshalBinaryBare(value, &p); err != nil {
			return false, status.Error(codes.Internal, err.Error())
		}

		matchVoter, matchDepositor, matchStatus := true, true, true

		// match status (if supplied/valid)
		if types.ValidProposalStatus(req.ProposalStatus) {
			matchStatus = p.Status == req.ProposalStatus
		}

		// match voter address (if supplied)
		if len(req.Voter) > 0 {
			voter, err := sdk.AccAddressFromBech32(req.Voter)
			if err != nil {
				return false, err
			}

			_, matchVoter = q.GetVote(ctx, p.ProposalId, voter)
		}

		// match depositor (if supplied)
		if len(req.Depositor) > 0 {
			depositor, err := sdk.AccAddressFromBech32(req.Depositor)
			if err != nil {
				return false, err
			}
			_, matchDepositor = q.GetDeposit(ctx, p.ProposalId, depositor)
		}

		if matchVoter && matchDepositor && matchStatus {
			if accumulate {
				filteredProposals = append(filteredProposals, p)
			}

			return true, nil
		}

		return false, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryProposalsResponse{Proposals: filteredProposals, Pagination: pageRes}, nil
}

// Vote returns Voted information based on proposalID, voterAddr
func (q Keeper) Vote(c context.Context, req *types.QueryVoteRequest) (*types.QueryVoteResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.ProposalId == 0 {
		return nil, status.Error(codes.InvalidArgument, "proposal id can not be 0")
	}

	if req.Voter == "" {
		return nil, status.Error(codes.InvalidArgument, "empty voter address")
	}

	ctx := sdk.UnwrapSDKContext(c)

	voter, err := sdk.AccAddressFromBech32(req.Voter)
	if err != nil {
		return nil, err
	}
	vote, found := q.GetVote(ctx, req.ProposalId, voter)
	if !found {
		return nil, status.Errorf(codes.InvalidArgument,
			"voter: %v not found for proposal: %v", req.Voter, req.ProposalId)
	}

	return &types.QueryVoteResponse{Vote: vote}, nil
}

// Votes returns single proposal's votes
func (q Keeper) Votes(c context.Context, req *types.QueryVotesRequest) (*types.QueryVotesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.ProposalId == 0 {
		return nil, status.Error(codes.InvalidArgument, "proposal id can not be 0")
	}

	var votes types.Votes
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(q.storeKey)
	votesStore := prefix.NewStore(store, types.VotesKey(req.ProposalId))

	pageRes, err := query.Paginate(votesStore, req.Pagination, func(key []byte, value []byte) error {
		var vote types.Vote
		if err := q.cdc.UnmarshalBinaryBare(value, &vote); err != nil {
			return err
		}

		votes = append(votes, vote)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryVotesResponse{Votes: votes, Pagination: pageRes}, nil
}

// Params queries all params
func (q Keeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	switch req.ParamsType {
	case types.ParamDeposit:
		depositParmas := q.GetDepositParams(ctx)
		return &types.QueryParamsResponse{DepositParams: depositParmas}, nil

	case types.ParamVoting:
		votingParmas := q.GetVotingParams(ctx)
		return &types.QueryParamsResponse{VotingParams: votingParmas}, nil

	case types.ParamTallying:
		tallyParams := q.GetTallyParams(ctx)
		return &types.QueryParamsResponse{TallyParams: tallyParams}, nil

	default:
		return nil, status.Errorf(codes.InvalidArgument,
			"%s is not a valid parameter type", req.ParamsType)
	}
}

// Deposit queries single deposit information based proposalID, depositAddr
func (q Keeper) Deposit(c context.Context, req *types.QueryDepositRequest) (*types.QueryDepositResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.ProposalId == 0 {
		return nil, status.Error(codes.InvalidArgument, "proposal id can not be 0")
	}

	if req.Depositor == "" {
		return nil, status.Error(codes.InvalidArgument, "empty depositor address")
	}

	ctx := sdk.UnwrapSDKContext(c)

	depositor, err := sdk.AccAddressFromBech32(req.Depositor)
	if err != nil {
		return nil, err
	}
	deposit, found := q.GetDeposit(ctx, req.ProposalId, depositor)
	if !found {
		return nil, status.Errorf(codes.InvalidArgument,
			"depositer: %v not found for proposal: %v", req.Depositor, req.ProposalId)
	}

	return &types.QueryDepositResponse{Deposit: deposit}, nil
}

// Deposits returns single proposal's all deposits
func (q Keeper) Deposits(c context.Context, req *types.QueryDepositsRequest) (*types.QueryDepositsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.ProposalId == 0 {
		return nil, status.Error(codes.InvalidArgument, "proposal id can not be 0")
	}

	var deposits types.Deposits
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(q.storeKey)
	depositStore := prefix.NewStore(store, types.DepositsKey(req.ProposalId))

	pageRes, err := query.Paginate(depositStore, req.Pagination, func(key []byte, value []byte) error {
		var deposit types.Deposit
		if err := q.cdc.UnmarshalBinaryBare(value, &deposit); err != nil {
			return err
		}

		deposits = append(deposits, deposit)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryDepositsResponse{Deposits: deposits, Pagination: pageRes}, nil
}

// TallyResult queries the tally of a proposal vote
func (q Keeper) TallyResult(c context.Context, req *types.QueryTallyResultRequest) (*types.QueryTallyResultResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.ProposalId == 0 {
		return nil, status.Error(codes.InvalidArgument, "proposal id can not be 0")
	}

	ctx := sdk.UnwrapSDKContext(c)

	proposal, ok := q.GetProposal(ctx, req.ProposalId)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "proposal %d doesn't exist", req.ProposalId)
	}

	var tallyResult types.TallyResult

	switch {
	case proposal.Status == types.StatusDepositPeriod:
		tallyResult = types.EmptyTallyResult()

	case proposal.Status == types.StatusPassed || proposal.Status == types.StatusRejected:
		tallyResult = proposal.FinalTallyResult

	default:
		// proposal is in voting period
		_, _, tallyResult = q.Tally(ctx, proposal)
	}

	return &types.QueryTallyResultResponse{Tally: tallyResult}, nil
}
