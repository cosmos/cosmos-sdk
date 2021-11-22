package group

import (
	"context"

	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/group/internal/orm"
)

func (q Keeper) GroupInfo(goCtx context.Context, request *QueryGroupInfo) (*QueryGroupInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupID := request.GroupId
	groupInfo, err := q.getGroupInfo(ctx.Context(), groupID)
	if err != nil {
		return nil, err
	}

	return &QueryGroupInfoResponse{Info: &groupInfo}, nil
}

func (q Keeper) getGroupInfo(goCtx context.Context, id uint64) (GroupInfo, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	var obj GroupInfo
	_, err := q.groupTable.GetOne(ctx.KVStore(q.key), id, &obj)
	return obj, err
}

func (q Keeper) GroupAccountInfo(goCtx context.Context, request *QueryGroupAccountInfo) (*QueryGroupAccountInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(request.Address)
	if err != nil {
		return nil, err
	}
	groupAccountInfo, err := q.getGroupAccountInfo(ctx.Context(), addr)
	if err != nil {
		return nil, err
	}

	return &QueryGroupAccountInfoResponse{Info: &groupAccountInfo}, nil
}

func (q Keeper) getGroupAccountInfo(goCtx context.Context, accountAddress sdk.AccAddress) (GroupAccountInfo, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	var obj GroupAccountInfo
	return obj, q.groupAccountTable.GetOne(ctx.KVStore(q.key), accountAddress.Bytes(), &obj)
}

func (q Keeper) GroupMembers(goCtx context.Context, request *QueryGroupMembers) (*QueryGroupMembersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupID := request.GroupId
	it, err := q.getGroupMembers(ctx.Context(), groupID, request.Pagination)
	if err != nil {
		return nil, err
	}

	var members []*GroupMember
	pageRes, err := orm.Paginate(it, request.Pagination, &members)
	if err != nil {
		return nil, err
	}

	return &QueryGroupMembersResponse{
		Members:    members,
		Pagination: pageRes,
	}, nil
}

func (q Keeper) getGroupMembers(goCtx context.Context, id uint64, pageRequest *query.PageRequest) (orm.Iterator, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return q.groupMemberByGroupIndex.GetPaginated(ctx.KVStore(q.key), id, pageRequest)
}

func (q Keeper) GroupsByAdmin(goCtx context.Context, request *QueryGroupsByAdmin) (*QueryGroupsByAdminResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(request.Admin)
	if err != nil {
		return nil, err
	}
	it, err := q.getGroupsByAdmin(ctx.Context(), addr, request.Pagination)
	if err != nil {
		return nil, err
	}

	var groups []*GroupInfo
	pageRes, err := orm.Paginate(it, request.Pagination, &groups)
	if err != nil {
		return nil, err
	}

	return &QueryGroupsByAdminResponse{
		Groups:     groups,
		Pagination: pageRes,
	}, nil
}

func (q Keeper) getGroupsByAdmin(goCtx context.Context, admin sdk.AccAddress, pageRequest *query.PageRequest) (orm.Iterator, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return q.groupByAdminIndex.GetPaginated(ctx.KVStore(q.key), admin.Bytes(), pageRequest)
}

func (q Keeper) GroupAccountsByGroup(goCtx context.Context, request *QueryGroupAccountsByGroup) (*QueryGroupAccountsByGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupID := request.GroupId
	it, err := q.getGroupAccountsByGroup(ctx, groupID, request.Pagination)
	if err != nil {
		return nil, err
	}

	var accounts []*GroupAccountInfo
	pageRes, err := orm.Paginate(it, request.Pagination, &accounts)
	if err != nil {
		return nil, err
	}

	return &QueryGroupAccountsByGroupResponse{
		GroupAccounts: accounts,
		Pagination:    pageRes,
	}, nil
}

func (q Keeper) getGroupAccountsByGroup(ctx sdk.Context, id uint64, pageRequest *query.PageRequest) (orm.Iterator, error) {
	return q.groupAccountByGroupIndex.GetPaginated(ctx.KVStore(q.key), id, pageRequest)
}

func (q Keeper) GroupAccountsByAdmin(goCtx context.Context, request *QueryGroupAccountsByAdmin) (*QueryGroupAccountsByAdminResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(request.Admin)
	if err != nil {
		return nil, err
	}
	it, err := q.getGroupAccountsByAdmin(ctx, addr, request.Pagination)
	if err != nil {
		return nil, err
	}

	var accounts []*GroupAccountInfo
	pageRes, err := orm.Paginate(it, request.Pagination, &accounts)
	if err != nil {
		return nil, err
	}

	return &QueryGroupAccountsByAdminResponse{
		GroupAccounts: accounts,
		Pagination:    pageRes,
	}, nil
}

func (q Keeper) getGroupAccountsByAdmin(ctx sdk.Context, admin sdk.AccAddress, pageRequest *query.PageRequest) (orm.Iterator, error) {
	return q.groupAccountByAdminIndex.GetPaginated(ctx.KVStore(q.key), admin.Bytes(), pageRequest)
}

func (q Keeper) Proposal(goCtx context.Context, request *QueryProposal) (*QueryProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	proposalID := request.ProposalId
	proposal, err := q.getProposal(ctx, proposalID)
	if err != nil {
		return nil, err
	}

	return &QueryProposalResponse{Proposal: &proposal}, nil
}

func (q Keeper) ProposalsByGroupAccount(goCtx context.Context, request *QueryProposalsByGroupAccount) (*QueryProposalsByGroupAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(request.Address)
	if err != nil {
		return nil, err
	}
	it, err := q.getProposalsByGroupAccount(ctx, addr, request.Pagination)
	if err != nil {
		return nil, err
	}

	var proposals []*Proposal
	pageRes, err := orm.Paginate(it, request.Pagination, &proposals)
	if err != nil {
		return nil, err
	}

	return &QueryProposalsByGroupAccountResponse{
		Proposals:  proposals,
		Pagination: pageRes,
	}, nil
}

func (q Keeper) getProposalsByGroupAccount(ctx sdk.Context, account sdk.AccAddress, pageRequest *query.PageRequest) (orm.Iterator, error) {
	return q.proposalByGroupAccountIndex.GetPaginated(ctx.KVStore(q.key), account.Bytes(), pageRequest)
}

func (q Keeper) getProposal(ctx sdk.Context, proposalID uint64) (Proposal, error) {
	var p Proposal
	if _, err := q.proposalTable.GetOne(ctx.KVStore(q.key), proposalID, &p); err != nil {
		return Proposal{}, sdkerrors.Wrap(err, "load proposal")
	}
	return p, nil
}

func (q Keeper) VoteByProposalVoter(goCtx context.Context, request *QueryVoteByProposalVoter) (*QueryVoteByProposalVoterResponse, error) {
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
	return &QueryVoteByProposalVoterResponse{
		Vote: &vote,
	}, nil
}

func (q Keeper) VotesByProposal(goCtx context.Context, request *QueryVotesByProposal) (*QueryVotesByProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	proposalID := request.ProposalId
	it, err := q.getVotesByProposal(ctx, proposalID, request.Pagination)
	if err != nil {
		return nil, err
	}

	var votes []*Vote
	pageRes, err := orm.Paginate(it, request.Pagination, &votes)
	if err != nil {
		return nil, err
	}

	return &QueryVotesByProposalResponse{
		Votes:      votes,
		Pagination: pageRes,
	}, nil
}

func (q Keeper) VotesByVoter(goCtx context.Context, request *QueryVotesByVoter) (*QueryVotesByVoterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(request.Voter)
	if err != nil {
		return nil, err
	}
	it, err := q.getVotesByVoter(ctx, addr, request.Pagination)
	if err != nil {
		return nil, err
	}

	var votes []*Vote
	pageRes, err := orm.Paginate(it, request.Pagination, &votes)
	if err != nil {
		return nil, err
	}

	return &QueryVotesByVoterResponse{
		Votes:      votes,
		Pagination: pageRes,
	}, nil
}

func (q Keeper) getVote(ctx sdk.Context, proposalID uint64, voter sdk.AccAddress) (Vote, error) {
	var v Vote
	return v, q.voteTable.GetOne(ctx.KVStore(q.key), orm.PrimaryKey(&Vote{ProposalId: proposalID, Voter: voter.String()}), &v)
}

func (q Keeper) getVotesByProposal(ctx types.Context, proposalID uint64, pageRequest *query.PageRequest) (orm.Iterator, error) {
	return q.voteByProposalIndex.GetPaginated(ctx.KVStore(q.key), proposalID, pageRequest)
}

func (q Keeper) getVotesByVoter(ctx types.Context, voter sdk.AccAddress, pageRequest *query.PageRequest) (orm.Iterator, error) {
	return q.voteByVoterIndex.GetPaginated(ctx.KVStore(q.key), voter.Bytes(), pageRequest)
}
