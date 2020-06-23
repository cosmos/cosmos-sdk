package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

// AllProposals implements the Query/AllProposals gRPC method
func (q Keeper) AllProposals(c context.Context, req *types.QueryAllProposalsRequest) (*types.QueryAllProposalsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	var proposals types.Proposals
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(q.storeKey)
	proposalStore := prefix.NewStore(store, types.ProposalsKeyPrefix)

	res, err := query.Paginate(proposalStore, req.Req, func(key []byte, value []byte) error {
		var result types.Proposal
		err := q.cdc.UnmarshalBinaryBare(value, &result)
		if err != nil {
			return err
		}
		proposals = append(proposals, result)
		return nil
	})

	if err != nil {
		return &types.QueryAllProposalsResponse{}, err
	}

	bz, err := q.cdc.MarshalJSON(proposals)
	return &types.QueryAllProposalsResponse{Proposals: bz, Res: res}, nil
}

// Votes returns single proposal's votes
func (q Keeper) Votes(c context.Context, req *types.QueryProposalRequest) (*types.QueryVotesResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	var votes types.Votes
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(q.storeKey)
	votesStore := prefix.NewStore(store, types.VotesKey(req.ProposalId))
	// proposalStore := prefix.NewStore(votesStore, types.VotesKey(proposalID))

	res, err := query.Paginate(votesStore, req.Req, func(key []byte, value []byte) error {
		var result types.Vote
		err := q.cdc.UnmarshalBinaryBare(value, &result)
		if err != nil {
			return err
		}
		votes = append(votes, result)
		return nil
	})

	if err != nil {
		return &types.QueryVotesResponse{}, err
	}

	bz, err := q.cdc.MarshalJSON(votes)
	return &types.QueryVotesResponse{Votes: bz, Res: res}, nil
}

// Deposits returns single proposal's all deposits
func (q Keeper) Deposits(c context.Context, req *types.QueryProposalRequest) (*types.QueryDepositsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	var deposits types.Deposits
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(q.storeKey)
	depositStore := prefix.NewStore(store, types.DepositsKey(req.ProposalId))

	res, err := query.Paginate(depositStore, req.Req, func(key []byte, value []byte) error {
		var result types.Deposit
		err := q.cdc.UnmarshalBinaryBare(value, &result)
		if err != nil {
			return err
		}
		deposits = append(deposits, result)
		return nil
	})

	if err != nil {
		return &types.QueryDepositsResponse{}, err
	}

	bz, err := q.cdc.MarshalJSON(deposits)
	return &types.QueryDepositsResponse{Deposits: bz, Res: res}, nil
}
