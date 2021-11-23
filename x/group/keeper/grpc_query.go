package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/group/internal/orm"
)

var _ group.QueryServer = Keeper{}

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
	_, err := q.groupTable.GetOne(ctx.KVStore(q.key), id, &obj)
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
	return obj, q.groupAccountTable.GetOne(ctx.KVStore(q.key), accountAddress.Bytes(), &obj)
}

func (q Keeper) GroupMembers(goCtx context.Context, request *group.QueryGroupMembers) (*group.QueryGroupMembersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupID := request.GroupId
	it, err := q.getGroupMembers(ctx.Context(), groupID, request.Pagination)
	if err != nil {
		return nil, err
	}

	var members []*group.GroupMember
	pageRes, err := orm.Paginate(it, request.Pagination, &members)
	if err != nil {
		return nil, err
	}

	return &group.QueryGroupMembersResponse{
		Members:    members,
		Pagination: pageRes,
	}, nil
}

func (q Keeper) getGroupMembers(goCtx context.Context, id uint64, pageRequest *query.PageRequest) (orm.Iterator, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return q.groupMemberByGroupIndex.GetPaginated(ctx.KVStore(q.key), id, pageRequest)
}

func (q Keeper) GroupsByAdmin(goCtx context.Context, request *group.QueryGroupsByAdmin) (*group.QueryGroupsByAdminResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(request.Admin)
	if err != nil {
		return nil, err
	}
	it, err := q.getGroupsByAdmin(ctx.Context(), addr, request.Pagination)
	if err != nil {
		return nil, err
	}

	var groups []*group.GroupInfo
	pageRes, err := orm.Paginate(it, request.Pagination, &groups)
	if err != nil {
		return nil, err
	}

	return &group.QueryGroupsByAdminResponse{
		Groups:     groups,
		Pagination: pageRes,
	}, nil
}

func (q Keeper) getGroupsByAdmin(goCtx context.Context, admin sdk.AccAddress, pageRequest *query.PageRequest) (orm.Iterator, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return q.groupByAdminIndex.GetPaginated(ctx.KVStore(q.key), admin.Bytes(), pageRequest)
}

func (q Keeper) GroupAccountsByGroup(goCtx context.Context, request *group.QueryGroupAccountsByGroup) (*group.QueryGroupAccountsByGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupID := request.GroupId
	it, err := q.getGroupAccountsByGroup(ctx, groupID, request.Pagination)
	if err != nil {
		return nil, err
	}

	var accounts []*group.GroupAccountInfo
	pageRes, err := orm.Paginate(it, request.Pagination, &accounts)
	if err != nil {
		return nil, err
	}

	return &group.QueryGroupAccountsByGroupResponse{
		GroupAccounts: accounts,
		Pagination:    pageRes,
	}, nil
}

func (q Keeper) getGroupAccountsByGroup(ctx sdk.Context, id uint64, pageRequest *query.PageRequest) (orm.Iterator, error) {
	return q.groupAccountByGroupIndex.GetPaginated(ctx.KVStore(q.key), id, pageRequest)
}

func (q Keeper) GroupAccountsByAdmin(goCtx context.Context, request *group.QueryGroupAccountsByAdmin) (*group.QueryGroupAccountsByAdminResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(request.Admin)
	if err != nil {
		return nil, err
	}
	it, err := q.getGroupAccountsByAdmin(ctx, addr, request.Pagination)
	if err != nil {
		return nil, err
	}

	var accounts []*group.GroupAccountInfo
	pageRes, err := orm.Paginate(it, request.Pagination, &accounts)
	if err != nil {
		return nil, err
	}

	return &group.QueryGroupAccountsByAdminResponse{
		GroupAccounts: accounts,
		Pagination:    pageRes,
	}, nil
}

func (q Keeper) getGroupAccountsByAdmin(ctx sdk.Context, admin sdk.AccAddress, pageRequest *query.PageRequest) (orm.Iterator, error) {
	return q.groupAccountByAdminIndex.GetPaginated(ctx.KVStore(q.key), admin.Bytes(), pageRequest)
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

func (q Keeper) ProposalsByGroupAccount(goCtx context.Context, request *group.QueryProposalsByGroupAccount) (*group.QueryProposalsByGroupAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(request.Address)
	if err != nil {
		return nil, err
	}
	it, err := q.getProposalsByGroupAccount(ctx, addr, request.Pagination)
	if err != nil {
		return nil, err
	}

	var proposals []*group.Proposal
	pageRes, err := orm.Paginate(it, request.Pagination, &proposals)
	if err != nil {
		return nil, err
	}

	return &group.QueryProposalsByGroupAccountResponse{
		Proposals:  proposals,
		Pagination: pageRes,
	}, nil
}

func (q Keeper) getProposalsByGroupAccount(ctx sdk.Context, account sdk.AccAddress, pageRequest *query.PageRequest) (orm.Iterator, error) {
	return q.proposalByGroupAccountIndex.GetPaginated(ctx.KVStore(q.key), account.Bytes(), pageRequest)
}

func (q Keeper) getProposal(ctx sdk.Context, proposalID uint64) (group.Proposal, error) {
	var p group.Proposal
	if _, err := q.proposalTable.GetOne(ctx.KVStore(q.key), proposalID, &p); err != nil {
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

func (q Keeper) VotesByProposal(goCtx context.Context, request *group.QueryVotesByProposal) (*group.QueryVotesByProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	proposalID := request.ProposalId
	it, err := q.getVotesByProposal(ctx, proposalID, request.Pagination)
	if err != nil {
		return nil, err
	}

	var votes []*group.Vote
	pageRes, err := orm.Paginate(it, request.Pagination, &votes)
	if err != nil {
		return nil, err
	}

	return &group.QueryVotesByProposalResponse{
		Votes:      votes,
		Pagination: pageRes,
	}, nil
}

func (q Keeper) VotesByVoter(goCtx context.Context, request *group.QueryVotesByVoter) (*group.QueryVotesByVoterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(request.Voter)
	if err != nil {
		return nil, err
	}
	it, err := q.getVotesByVoter(ctx, addr, request.Pagination)
	if err != nil {
		return nil, err
	}

	var votes []*group.Vote
	pageRes, err := orm.Paginate(it, request.Pagination, &votes)
	if err != nil {
		return nil, err
	}

	return &group.QueryVotesByVoterResponse{
		Votes:      votes,
		Pagination: pageRes,
	}, nil
}

func (q Keeper) getVote(ctx sdk.Context, proposalID uint64, voter sdk.AccAddress) (group.Vote, error) {
	var v group.Vote
	return v, q.voteTable.GetOne(ctx.KVStore(q.key), orm.PrimaryKey(&group.Vote{ProposalId: proposalID, Voter: voter.String()}), &v)
}

func (q Keeper) getVotesByProposal(ctx types.Context, proposalID uint64, pageRequest *query.PageRequest) (orm.Iterator, error) {
	return q.voteByProposalIndex.GetPaginated(ctx.KVStore(q.key), proposalID, pageRequest)
}

func (q Keeper) getVotesByVoter(ctx types.Context, voter sdk.AccAddress, pageRequest *query.PageRequest) (orm.Iterator, error) {
	return q.voteByVoterIndex.GetPaginated(ctx.KVStore(q.key), voter.Bytes(), pageRequest)
}
