package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/group/errors"
	"github.com/cosmos/cosmos-sdk/x/group/internal/orm"
)

var _ group.QueryServer = Keeper{}

func (q Keeper) GroupInfo(goCtx context.Context, request *group.QueryGroupInfoRequest) (*group.QueryGroupInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupID := request.GroupId
	groupInfo, err := q.getGroupInfo(ctx, groupID)
	if err != nil {
		return nil, err
	}

	return &group.QueryGroupInfoResponse{Info: &groupInfo}, nil
}

func (q Keeper) getGroupInfo(ctx sdk.Context, id uint64) (group.GroupInfo, error) {
	var obj group.GroupInfo
	_, err := q.groupTable.GetOne(ctx.KVStore(q.key), id, &obj)
	return obj, err
}

func (q Keeper) GroupPolicyInfo(goCtx context.Context, request *group.QueryGroupPolicyInfoRequest) (*group.QueryGroupPolicyInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupPolicyInfo, err := q.GetGroupPolicyInfo(ctx, request.Address)
	if err != nil {
		return nil, err
	}

	return &group.QueryGroupPolicyInfoResponse{Info: &groupPolicyInfo}, nil
}

func (q Keeper) GetGroupPolicyInfo(ctx sdk.Context, accountAddress string) (group.GroupPolicyInfo, error) {
	var obj group.GroupPolicyInfo
	return obj, q.groupPolicyTable.GetOne(ctx.KVStore(q.key), orm.PrimaryKey(&group.GroupPolicyInfo{Address: accountAddress}), &obj)
}

func (q Keeper) GroupMembers(goCtx context.Context, request *group.QueryGroupMembersRequest) (*group.QueryGroupMembersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupID := request.GroupId
	it, err := q.getGroupMembers(ctx, groupID, request.Pagination)
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

func (q Keeper) getGroupMembers(ctx sdk.Context, id uint64, pageRequest *query.PageRequest) (orm.Iterator, error) {
	return q.groupMemberByGroupIndex.GetPaginated(ctx.KVStore(q.key), id, pageRequest)
}

func (q Keeper) GroupsByAdmin(goCtx context.Context, request *group.QueryGroupsByAdminRequest) (*group.QueryGroupsByAdminResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(request.Admin)
	if err != nil {
		return nil, err
	}
	it, err := q.getGroupsByAdmin(ctx, addr, request.Pagination)
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

func (q Keeper) getGroupsByAdmin(ctx sdk.Context, admin sdk.AccAddress, pageRequest *query.PageRequest) (orm.Iterator, error) {
	return q.groupByAdminIndex.GetPaginated(ctx.KVStore(q.key), admin.Bytes(), pageRequest)
}

func (q Keeper) GroupPoliciesByGroup(goCtx context.Context, request *group.QueryGroupPoliciesByGroupRequest) (*group.QueryGroupPoliciesByGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupID := request.GroupId
	it, err := q.getGroupPoliciesByGroup(ctx, groupID, request.Pagination)
	if err != nil {
		return nil, err
	}

	var policies []*group.GroupPolicyInfo
	pageRes, err := orm.Paginate(it, request.Pagination, &policies)
	if err != nil {
		return nil, err
	}

	return &group.QueryGroupPoliciesByGroupResponse{
		GroupPolicies: policies,
		Pagination:    pageRes,
	}, nil
}

func (q Keeper) getGroupPoliciesByGroup(ctx sdk.Context, id uint64, pageRequest *query.PageRequest) (orm.Iterator, error) {
	return q.groupPolicyByGroupIndex.GetPaginated(ctx.KVStore(q.key), id, pageRequest)
}

func (q Keeper) GroupPoliciesByAdmin(goCtx context.Context, request *group.QueryGroupPoliciesByAdminRequest) (*group.QueryGroupPoliciesByAdminResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(request.Admin)
	if err != nil {
		return nil, err
	}
	it, err := q.getGroupPoliciesByAdmin(ctx, addr, request.Pagination)
	if err != nil {
		return nil, err
	}

	var policies []*group.GroupPolicyInfo
	pageRes, err := orm.Paginate(it, request.Pagination, &policies)
	if err != nil {
		return nil, err
	}

	return &group.QueryGroupPoliciesByAdminResponse{
		GroupPolicies: policies,
		Pagination:    pageRes,
	}, nil
}

func (q Keeper) getGroupPoliciesByAdmin(ctx sdk.Context, admin sdk.AccAddress, pageRequest *query.PageRequest) (orm.Iterator, error) {
	return q.groupPolicyByAdminIndex.GetPaginated(ctx.KVStore(q.key), admin.Bytes(), pageRequest)
}

func (q Keeper) Proposal(goCtx context.Context, request *group.QueryProposalRequest) (*group.QueryProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	proposalID := request.ProposalId
	proposal, err := q.getProposal(ctx, proposalID)
	if err != nil {
		return nil, err
	}

	return &group.QueryProposalResponse{Proposal: &proposal}, nil
}

func (q Keeper) ProposalsByGroupPolicy(goCtx context.Context, request *group.QueryProposalsByGroupPolicyRequest) (*group.QueryProposalsByGroupPolicyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(request.Address)
	if err != nil {
		return nil, err
	}
	it, err := q.getProposalsByGroupPolicy(ctx, addr, request.Pagination)
	if err != nil {
		return nil, err
	}

	var proposals []*group.Proposal
	pageRes, err := orm.Paginate(it, request.Pagination, &proposals)
	if err != nil {
		return nil, err
	}

	return &group.QueryProposalsByGroupPolicyResponse{
		Proposals:  proposals,
		Pagination: pageRes,
	}, nil
}

func (q Keeper) getProposalsByGroupPolicy(ctx sdk.Context, account sdk.AccAddress, pageRequest *query.PageRequest) (orm.Iterator, error) {
	return q.proposalByGroupPolicyIndex.GetPaginated(ctx.KVStore(q.key), account.Bytes(), pageRequest)
}

func (q Keeper) getProposal(ctx sdk.Context, proposalID uint64) (group.Proposal, error) {
	var p group.Proposal
	if _, err := q.proposalTable.GetOne(ctx.KVStore(q.key), proposalID, &p); err != nil {
		return group.Proposal{}, sdkerrors.Wrap(err, "load proposal")
	}
	return p, nil
}

func (q Keeper) VoteByProposalVoter(goCtx context.Context, request *group.QueryVoteByProposalVoterRequest) (*group.QueryVoteByProposalVoterResponse, error) {
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

func (q Keeper) VotesByProposal(goCtx context.Context, request *group.QueryVotesByProposalRequest) (*group.QueryVotesByProposalResponse, error) {
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

func (q Keeper) VotesByVoter(goCtx context.Context, request *group.QueryVotesByVoterRequest) (*group.QueryVotesByVoterResponse, error) {
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

func (q Keeper) GroupsByMember(goCtx context.Context, request *group.QueryGroupsByMemberRequest) (*group.QueryGroupsByMemberResponse, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	member, err := sdk.AccAddressFromBech32(request.Address)
	if err != nil {
		return nil, err
	}

	iter, err := q.groupMemberByMemberIndex.GetPaginated(ctx.KVStore(q.key), member.Bytes(), request.Pagination)
	if err != nil {
		return nil, err
	}

	var members []*group.GroupMember
	pageRes, err := orm.Paginate(iter, request.Pagination, &members)
	if err != nil {
		return nil, err
	}

	var groups []*group.GroupInfo
	for _, gm := range members {
		groupInfo, err := q.getGroupInfo(ctx, gm.GroupId)
		if err != nil {
			return nil, err
		}
		groups = append(groups, &groupInfo)
	}

	return &group.QueryGroupsByMemberResponse{
		Groups:     groups,
		Pagination: pageRes,
	}, nil
}

func (q Keeper) getVote(ctx sdk.Context, proposalID uint64, voter sdk.AccAddress) (group.Vote, error) {
	var v group.Vote
	return v, q.voteTable.GetOne(ctx.KVStore(q.key), orm.PrimaryKey(&group.Vote{ProposalId: proposalID, Voter: voter.String()}), &v)
}

func (q Keeper) getVotesByProposal(ctx sdk.Context, proposalID uint64, pageRequest *query.PageRequest) (orm.Iterator, error) {
	return q.voteByProposalIndex.GetPaginated(ctx.KVStore(q.key), proposalID, pageRequest)
}

func (q Keeper) getVotesByVoter(ctx sdk.Context, voter sdk.AccAddress, pageRequest *query.PageRequest) (orm.Iterator, error) {
	return q.voteByVoterIndex.GetPaginated(ctx.KVStore(q.key), voter.Bytes(), pageRequest)
}

// Tally is a function that tallies a proposal by iterating through its votes,
// and returns the tally result without modifying the proposal or any state.
// TODO Merge with https://github.com/cosmos/cosmos-sdk/issues/11151
func (q Keeper) Tally(ctx sdk.Context, p group.Proposal, groupId uint64) (group.TallyResult, error) {
	// If proposal has already been tallied and updated, then its status is
	// closed, in which case we just return the previously stored result.
	if p.Status == group.PROPOSAL_STATUS_CLOSED {
		return p.FinalTallyResult, nil
	}

	it, err := q.voteByProposalIndex.Get(ctx.KVStore(q.key), p.Id)
	if err != nil {
		return group.TallyResult{}, err
	}
	defer it.Close()

	tallyResult := group.DefaultTallyResult()

	var vote group.Vote
	for {
		_, err = it.LoadNext(&vote)
		if errors.ErrORMIteratorDone.Is(err) {
			break
		}
		if err != nil {
			return group.TallyResult{}, err
		}

		var member group.GroupMember
		err := q.groupMemberTable.GetOne(ctx.KVStore(q.key), orm.PrimaryKey(&group.GroupMember{
			GroupId: groupId,
			Member:  &group.Member{Address: vote.Voter},
		}), &member)

		switch {
		case sdkerrors.ErrNotFound.Is(err):
			// If the member left the group after voting, then we simply skip the
			// vote.
			continue
		case err != nil:
			// For any other errors, we stop and return the error.
			return group.TallyResult{}, err
		}

		if err := tallyResult.Add(vote, member.Member.Weight); err != nil {
			return group.TallyResult{}, sdkerrors.Wrap(err, "add new vote")
		}
	}

	return tallyResult, nil
}

func (q Keeper) UpdateTallyOfVPEndProposals(ctx sdk.Context) error {
	timeBytes := sdk.FormatTimeBytes(ctx.BlockTime())
	it, _ := q.ProposalsByVotingPeriodEnd.Get(ctx.KVStore(q.key), sdk.PrefixEndBytes(timeBytes))

	var proposal group.Proposal
	for {
		_, err := it.LoadNext(&proposal)
		if errors.ErrORMIteratorDone.Is(err) {
			break
		}
		if err != nil {
			return err
		}

		// check whether the proposal can be tallied.
		if proposal.Status != group.PROPOSAL_STATUS_SUBMITTED {
			continue
		}

		var policyInfo group.GroupPolicyInfo
		if policyInfo, err = q.GetGroupPolicyInfo(ctx, proposal.Address); err != nil {
			return err
		}

		tallyRes, err := q.Tally(ctx, proposal, policyInfo.GroupId)
		if err != nil {
			return err
		}

		proposal.FinalTallyResult = tallyRes
		storeUpdates := func() error {
			if err := q.proposalTable.Update(ctx.KVStore(q.key), proposal.Id, &proposal); err != nil {
				return err
			}
			return nil
		}

		if err := storeUpdates(); err != nil {
			return err
		}
	}

	return nil
}
