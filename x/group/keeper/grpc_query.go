package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/group/internal/orm"
)

func (q Keeper) GroupInfo(goCtx context.Context, request *group.QueryGroupInfo) (*group.QueryGroupInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupID := request.GroupId
	groupInfo, err := q.getGroupInfo(ctx.Context(), groupID)
	if err != nil {
		return nil, err
	}

	return &group.QueryGroupInfoResponse{Info: &groupInfo}, nil
}

func (q Keeper) getGroupInfo(goCtx context.Context, id uint64) (group.GroupInfo, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	var obj group.GroupInfo
	_, err := q.server.groupTable.GetOne(ctx.KVStore(q.server.key), id, &obj)
	return obj, err
}

func (q Keeper) GroupAccountInfo(goCtx context.Context, request *group.QueryGroupAccountInfo) (*group.QueryGroupAccountInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(request.Address)
	if err != nil {
		return nil, err
	}
	groupAccountInfo, err := q.getGroupAccountInfo(ctx.Context(), addr)
	if err != nil {
		return nil, err
	}

	return &group.QueryGroupAccountInfoResponse{Info: &groupAccountInfo}, nil
}

func (q Keeper) getGroupAccountInfo(goCtx context.Context, accountAddress sdk.AccAddress) (group.GroupAccountInfo, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	var obj group.GroupAccountInfo
	return obj, q.server.groupAccountTable.GetOne(ctx.KVStore(q.server.key), accountAddress.Bytes(), &obj)
}

func (q Keeper) Proposal(goCtx context.Context, request *group.QueryProposal) (*group.QueryProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	proposalID := request.ProposalId
	proposal, err := q.getProposal(ctx, proposalID)
	if err != nil {
		return nil, err
	}

	return &group.QueryProposalResponse{Proposal: &proposal}, nil
}

func (q Keeper) getProposal(ctx sdk.Context, proposalID uint64) (group.Proposal, error) {
	var p group.Proposal
	if _, err := q.server.proposalTable.GetOne(ctx.KVStore(q.server.key), proposalID, &p); err != nil {
		return group.Proposal{}, sdkerrors.Wrap(err, "load proposal")
	}
	return p, nil
}

func (q Keeper) VoteByProposalVoter(goCtx context.Context, request *group.QueryVoteByProposalVoter) (*group.QueryVoteByProposalVoterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(request.Voter)
	if err != nil {
		return nil, err
	}
	proposalID := request.ProposalId
	vote, err := q.getVote(ctx, proposalID, addr)
	if err != nil {
		return nil, err
	}
	return &group.QueryVoteByProposalVoterResponse{
		Vote: &vote,
	}, nil
}

func (q Keeper) getVote(ctx sdk.Context, proposalID uint64, voter sdk.AccAddress) (group.Vote, error) {
	var v group.Vote
	return v, q.server.voteTable.GetOne(ctx.KVStore(q.server.key), orm.PrimaryKey(&group.Vote{ProposalId: proposalID, Voter: voter.String()}), &v)
}
