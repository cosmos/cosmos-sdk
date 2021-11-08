package server

import (
	"context"

	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/group/internal/orm"
)

func (s serverImpl) GroupInfo(goCtx context.Context, request *group.QueryGroupInfo) (*group.QueryGroupInfoResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	groupID := request.GroupId
	groupInfo, err := s.getGroupInfo(ctx.Context(), groupID)
	if err != nil {
		return nil, err
	}

	return &group.QueryGroupInfoResponse{Info: &groupInfo}, nil
}

func (s serverImpl) getGroupInfo(goCtx context.Context, id uint64) (group.GroupInfo, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	var obj group.GroupInfo
	_, err := s.groupTable.GetOne(ctx.KVStore(s.key), id, &obj)
	return obj, err
}

func (s serverImpl) GroupAccountInfo(goCtx context.Context, request *group.QueryGroupAccountInfo) (*group.QueryGroupAccountInfoResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(request.Address)
	if err != nil {
		return nil, err
	}
	groupAccountInfo, err := s.getGroupAccountInfo(ctx.Context(), addr)
	if err != nil {
		return nil, err
	}

	return &group.QueryGroupAccountInfoResponse{Info: &groupAccountInfo}, nil
}

func (s serverImpl) getGroupAccountInfo(goCtx context.Context, accountAddress sdk.AccAddress) (group.GroupAccountInfo, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	var obj group.GroupAccountInfo
	return obj, s.groupAccountTable.GetOne(ctx.KVStore(s.key), accountAddress.Bytes(), &obj)
}

func (s serverImpl) Proposal(goCtx context.Context, request *group.QueryProposal) (*group.QueryProposalResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	proposalID := request.ProposalId
	proposal, err := s.getProposal(ctx, proposalID)
	if err != nil {
		return nil, err
	}

	return &group.QueryProposalResponse{Proposal: &proposal}, nil
}

func (s serverImpl) getProposal(ctx types.Context, proposalID uint64) (group.Proposal, error) {
	var p group.Proposal
	if _, err := s.proposalTable.GetOne(ctx.KVStore(s.key), proposalID, &p); err != nil {
		return group.Proposal{}, sdkerrors.Wrap(err, "load proposal")
	}
	return p, nil
}

func (s serverImpl) VoteByProposalVoter(goCtx context.Context, request *group.QueryVoteByProposalVoter) (*group.QueryVoteByProposalVoterResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(request.Voter)
	if err != nil {
		return nil, err
	}
	proposalID := request.ProposalId
	vote, err := s.getVote(ctx, proposalID, addr)
	if err != nil {
		return nil, err
	}
	return &group.QueryVoteByProposalVoterResponse{
		Vote: &vote,
	}, nil
}

func (s serverImpl) getVote(ctx types.Context, proposalID uint64, voter sdk.AccAddress) (group.Vote, error) {
	var v group.Vote
	return v, s.voteTable.GetOne(ctx.KVStore(s.key), orm.PrimaryKey(&group.Vote{ProposalId: proposalID, Voter: voter.String()}), &v)
}
