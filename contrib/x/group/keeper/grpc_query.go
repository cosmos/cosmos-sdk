package keeper

import (
	"context"
	"math"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	errorsmod "cosmossdk.io/errors"

	group2 "github.com/cosmos/cosmos-sdk/contrib/x/group"
	"github.com/cosmos/cosmos-sdk/contrib/x/group/errors"
	orm2 "github.com/cosmos/cosmos-sdk/contrib/x/group/internal/orm"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
)

var _ group2.QueryServer = Keeper{}

// GroupInfo queries info about a group.
func (k Keeper) GroupInfo(goCtx context.Context, request *group2.QueryGroupInfoRequest) (*group2.QueryGroupInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupID := request.GroupId
	groupInfo, err := k.getGroupInfo(ctx, groupID)
	if err != nil {
		return nil, errorsmod.Wrap(err, "group")
	}

	return &group2.QueryGroupInfoResponse{Info: &groupInfo}, nil
}

// getGroupInfo gets the group info of the given group id.
func (k Keeper) getGroupInfo(ctx sdk.Context, id uint64) (group2.GroupInfo, error) {
	var obj group2.GroupInfo
	_, err := k.groupTable.GetOne(ctx.KVStore(k.key), id, &obj)
	return obj, err
}

// GroupPolicyInfo queries info about a group policy.
func (k Keeper) GroupPolicyInfo(goCtx context.Context, request *group2.QueryGroupPolicyInfoRequest) (*group2.QueryGroupPolicyInfoResponse, error) {
	_, err := k.accKeeper.AddressCodec().StringToBytes(request.Address)
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	groupPolicyInfo, err := k.getGroupPolicyInfo(ctx, request.Address)
	if err != nil {
		return nil, errorsmod.Wrap(err, "group policy")
	}

	return &group2.QueryGroupPolicyInfoResponse{Info: &groupPolicyInfo}, nil
}

// getGroupPolicyInfo gets the group policy info of the given account address.
func (k Keeper) getGroupPolicyInfo(ctx sdk.Context, accountAddress string) (group2.GroupPolicyInfo, error) {
	var obj group2.GroupPolicyInfo
	return obj, k.groupPolicyTable.GetOne(ctx.KVStore(k.key), orm2.PrimaryKey(&group2.GroupPolicyInfo{Address: accountAddress}), &obj)
}

// GroupMembers queries all members of a group.
func (k Keeper) GroupMembers(goCtx context.Context, request *group2.QueryGroupMembersRequest) (*group2.QueryGroupMembersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupID := request.GroupId
	it, err := k.getGroupMembers(ctx, groupID, request.Pagination)
	if err != nil {
		return nil, err
	}

	var members []*group2.GroupMember
	pageRes, err := orm2.Paginate(it, request.Pagination, &members)
	if err != nil {
		return nil, err
	}

	return &group2.QueryGroupMembersResponse{
		Members:    members,
		Pagination: pageRes,
	}, nil
}

// getGroupMembers returns an iterator for the given group id and page request.
func (k Keeper) getGroupMembers(ctx sdk.Context, id uint64, pageRequest *query.PageRequest) (orm2.Iterator, error) {
	return k.groupMemberByGroupIndex.GetPaginated(ctx.KVStore(k.key), id, pageRequest)
}

// GroupsByAdmin queries all groups where a given address is admin.
func (k Keeper) GroupsByAdmin(goCtx context.Context, request *group2.QueryGroupsByAdminRequest) (*group2.QueryGroupsByAdminResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := k.accKeeper.AddressCodec().StringToBytes(request.Admin)
	if err != nil {
		return nil, err
	}
	it, err := k.getGroupsByAdmin(ctx, addr, request.Pagination)
	if err != nil {
		return nil, err
	}

	var groups []*group2.GroupInfo
	pageRes, err := orm2.Paginate(it, request.Pagination, &groups)
	if err != nil {
		return nil, err
	}

	return &group2.QueryGroupsByAdminResponse{
		Groups:     groups,
		Pagination: pageRes,
	}, nil
}

// getGroupsByAdmin returns an iterator for the given admin account address and page request.
func (k Keeper) getGroupsByAdmin(ctx sdk.Context, admin sdk.AccAddress, pageRequest *query.PageRequest) (orm2.Iterator, error) {
	return k.groupByAdminIndex.GetPaginated(ctx.KVStore(k.key), admin.Bytes(), pageRequest)
}

// GroupPoliciesByGroup queries all groups policies of a given group.
func (k Keeper) GroupPoliciesByGroup(goCtx context.Context, request *group2.QueryGroupPoliciesByGroupRequest) (*group2.QueryGroupPoliciesByGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupID := request.GroupId
	it, err := k.getGroupPoliciesByGroup(ctx, groupID, request.Pagination)
	if err != nil {
		return nil, err
	}

	var policies []*group2.GroupPolicyInfo
	pageRes, err := orm2.Paginate(it, request.Pagination, &policies)
	if err != nil {
		return nil, err
	}

	return &group2.QueryGroupPoliciesByGroupResponse{
		GroupPolicies: policies,
		Pagination:    pageRes,
	}, nil
}

// getGroupPoliciesByGroup returns an iterator for the given group id and page request.
func (k Keeper) getGroupPoliciesByGroup(ctx sdk.Context, id uint64, pageRequest *query.PageRequest) (orm2.Iterator, error) {
	return k.groupPolicyByGroupIndex.GetPaginated(ctx.KVStore(k.key), id, pageRequest)
}

// GroupPoliciesByAdmin queries all groups policies where a given address is
// admin.
func (k Keeper) GroupPoliciesByAdmin(goCtx context.Context, request *group2.QueryGroupPoliciesByAdminRequest) (*group2.QueryGroupPoliciesByAdminResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := k.accKeeper.AddressCodec().StringToBytes(request.Admin)
	if err != nil {
		return nil, err
	}
	it, err := k.getGroupPoliciesByAdmin(ctx, addr, request.Pagination)
	if err != nil {
		return nil, err
	}

	var policies []*group2.GroupPolicyInfo
	pageRes, err := orm2.Paginate(it, request.Pagination, &policies)
	if err != nil {
		return nil, err
	}

	return &group2.QueryGroupPoliciesByAdminResponse{
		GroupPolicies: policies,
		Pagination:    pageRes,
	}, nil
}

// getGroupPoliciesByAdmin returns an iterator for the given admin account address and page request.
func (k Keeper) getGroupPoliciesByAdmin(ctx sdk.Context, admin sdk.AccAddress, pageRequest *query.PageRequest) (orm2.Iterator, error) {
	return k.groupPolicyByAdminIndex.GetPaginated(ctx.KVStore(k.key), admin.Bytes(), pageRequest)
}

// Proposal queries a proposal.
func (k Keeper) Proposal(goCtx context.Context, request *group2.QueryProposalRequest) (*group2.QueryProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	proposalID := request.ProposalId
	proposal, err := k.getProposal(ctx, proposalID)
	if err != nil {
		return nil, err
	}

	return &group2.QueryProposalResponse{Proposal: &proposal}, nil
}

// ProposalsByGroupPolicy queries all proposals of a group policy.
func (k Keeper) ProposalsByGroupPolicy(goCtx context.Context, request *group2.QueryProposalsByGroupPolicyRequest) (*group2.QueryProposalsByGroupPolicyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := k.accKeeper.AddressCodec().StringToBytes(request.Address)
	if err != nil {
		return nil, err
	}
	it, err := k.getProposalsByGroupPolicy(ctx, addr, request.Pagination)
	if err != nil {
		return nil, err
	}

	var proposals []*group2.Proposal
	pageRes, err := orm2.Paginate(it, request.Pagination, &proposals)
	if err != nil {
		return nil, err
	}

	return &group2.QueryProposalsByGroupPolicyResponse{
		Proposals:  proposals,
		Pagination: pageRes,
	}, nil
}

// getProposalsByGroupPolicy returns an iterator for the given account address and page request.
func (k Keeper) getProposalsByGroupPolicy(ctx sdk.Context, account sdk.AccAddress, pageRequest *query.PageRequest) (orm2.Iterator, error) {
	return k.proposalByGroupPolicyIndex.GetPaginated(ctx.KVStore(k.key), account.Bytes(), pageRequest)
}

// getProposal gets the proposal info of the given proposal id.
func (k Keeper) getProposal(ctx sdk.Context, proposalID uint64) (group2.Proposal, error) {
	var p group2.Proposal
	if _, err := k.proposalTable.GetOne(ctx.KVStore(k.key), proposalID, &p); err != nil {
		return group2.Proposal{}, errorsmod.Wrap(err, "load proposal")
	}
	return p, nil
}

// VoteByProposalVoter queries a vote given a voter and a proposal ID.
func (k Keeper) VoteByProposalVoter(goCtx context.Context, request *group2.QueryVoteByProposalVoterRequest) (*group2.QueryVoteByProposalVoterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := k.accKeeper.AddressCodec().StringToBytes(request.Voter)
	if err != nil {
		return nil, err
	}
	proposalID := request.ProposalId
	vote, err := k.getVote(ctx, proposalID, addr)
	if err != nil {
		return nil, err
	}
	return &group2.QueryVoteByProposalVoterResponse{
		Vote: &vote,
	}, nil
}

// VotesByProposal queries all votes on a proposal.
func (k Keeper) VotesByProposal(goCtx context.Context, request *group2.QueryVotesByProposalRequest) (*group2.QueryVotesByProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	proposalID := request.ProposalId
	it, err := k.getVotesByProposal(ctx, proposalID, request.Pagination)
	if err != nil {
		return nil, err
	}

	var votes []*group2.Vote
	pageRes, err := orm2.Paginate(it, request.Pagination, &votes)
	if err != nil {
		return nil, err
	}

	return &group2.QueryVotesByProposalResponse{
		Votes:      votes,
		Pagination: pageRes,
	}, nil
}

// VotesByVoter queries all votes of a voter.
func (k Keeper) VotesByVoter(goCtx context.Context, request *group2.QueryVotesByVoterRequest) (*group2.QueryVotesByVoterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := k.accKeeper.AddressCodec().StringToBytes(request.Voter)
	if err != nil {
		return nil, err
	}
	it, err := k.getVotesByVoter(ctx, addr, request.Pagination)
	if err != nil {
		return nil, err
	}

	var votes []*group2.Vote
	pageRes, err := orm2.Paginate(it, request.Pagination, &votes)
	if err != nil {
		return nil, err
	}

	return &group2.QueryVotesByVoterResponse{
		Votes:      votes,
		Pagination: pageRes,
	}, nil
}

// GroupsByMember queries all groups where the given address is a member of.
func (k Keeper) GroupsByMember(goCtx context.Context, request *group2.QueryGroupsByMemberRequest) (*group2.QueryGroupsByMemberResponse, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	member, err := k.accKeeper.AddressCodec().StringToBytes(request.Address)
	if err != nil {
		return nil, err
	}

	iter, err := k.groupMemberByMemberIndex.GetPaginated(ctx.KVStore(k.key), member, request.Pagination)
	if err != nil {
		return nil, err
	}

	var members []*group2.GroupMember
	pageRes, err := orm2.Paginate(iter, request.Pagination, &members)
	if err != nil {
		return nil, err
	}

	var groups []*group2.GroupInfo
	for _, gm := range members {
		groupInfo, err := k.getGroupInfo(ctx, gm.GroupId)
		if err != nil {
			return nil, err
		}
		groups = append(groups, &groupInfo)
	}

	return &group2.QueryGroupsByMemberResponse{
		Groups:     groups,
		Pagination: pageRes,
	}, nil
}

// getVote gets the vote info for the given proposal id and voter address.
func (k Keeper) getVote(ctx sdk.Context, proposalID uint64, voter sdk.AccAddress) (group2.Vote, error) {
	var v group2.Vote
	return v, k.voteTable.GetOne(ctx.KVStore(k.key), orm2.PrimaryKey(&group2.Vote{ProposalId: proposalID, Voter: voter.String()}), &v)
}

// getVotesByProposal returns an iterator for the given proposal id and page request.
func (k Keeper) getVotesByProposal(ctx sdk.Context, proposalID uint64, pageRequest *query.PageRequest) (orm2.Iterator, error) {
	return k.voteByProposalIndex.GetPaginated(ctx.KVStore(k.key), proposalID, pageRequest)
}

// getVotesByVoter returns an iterator for the given voter address and page request.
func (k Keeper) getVotesByVoter(ctx sdk.Context, voter sdk.AccAddress, pageRequest *query.PageRequest) (orm2.Iterator, error) {
	return k.voteByVoterIndex.GetPaginated(ctx.KVStore(k.key), voter.Bytes(), pageRequest)
}

// TallyResult computes the live tally result of a proposal.
func (k Keeper) TallyResult(goCtx context.Context, request *group2.QueryTallyResultRequest) (*group2.QueryTallyResultResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	proposalID := request.ProposalId

	proposal, err := k.getProposal(ctx, proposalID)
	if err != nil {
		return nil, err
	}

	if proposal.Status == group2.PROPOSAL_STATUS_WITHDRAWN || proposal.Status == group2.PROPOSAL_STATUS_ABORTED {
		return nil, errorsmod.Wrapf(errors.ErrInvalid, "can't get the tally of a proposal with status %s", proposal.Status)
	}

	var policyInfo group2.GroupPolicyInfo
	if policyInfo, err = k.getGroupPolicyInfo(ctx, proposal.GroupPolicyAddress); err != nil {
		return nil, errorsmod.Wrap(err, "load group policy")
	}

	tallyResult, err := k.Tally(ctx, proposal, policyInfo.GroupId)
	if err != nil {
		return nil, err
	}

	return &group2.QueryTallyResultResponse{
		Tally: tallyResult,
	}, nil
}

// Groups returns all the groups present in the state.
func (k Keeper) Groups(goCtx context.Context, request *group2.QueryGroupsRequest) (*group2.QueryGroupsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	it, err := k.groupTable.PrefixScan(ctx.KVStore(k.key), 1, math.MaxUint64)
	if err != nil {
		return nil, err
	}
	defer it.Close()

	var groups []*group2.GroupInfo
	pageRes, err := orm2.Paginate(it, request.Pagination, &groups)
	if err != nil {
		return nil, err
	}

	return &group2.QueryGroupsResponse{
		Groups:     groups,
		Pagination: pageRes,
	}, nil
}
