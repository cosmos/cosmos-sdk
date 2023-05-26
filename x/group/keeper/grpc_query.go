package keeper

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	groupv1 "cosmossdk.io/api/cosmos/group/v1"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/orm/model/ormlist"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/ormutil"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/group/errors"
)

var _ group.QueryServer = Keeper{}

// GroupInfo queries info about a group.
func (k Keeper) GroupInfo(goCtx context.Context, request *group.QueryGroupInfoRequest) (*group.QueryGroupInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupID := request.GroupId
	groupInfo, err := k.getGroupInfo(ctx, groupID)
	if err != nil {
		return nil, errorsmod.Wrap(err, "group")
	}

	return &group.QueryGroupInfoResponse{Info: &groupInfo}, nil
}

// getGroupInfo gets the group info of the given group id.
func (k Keeper) getGroupInfo(ctx sdk.Context, id uint64) (group.GroupInfo, error) {
	groupInfo, err := k.state.GroupInfoTable().Get(ctx, id)
	if err != nil {
		return group.GroupInfo{}, err
	}

	return group.GroupInfoFromPulsar(groupInfo), nil
}

// GroupPolicyInfo queries info about a group policy.
func (k Keeper) GroupPolicyInfo(goCtx context.Context, request *group.QueryGroupPolicyInfoRequest) (*group.QueryGroupPolicyInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupPolicyInfo, err := k.getGroupPolicyInfo(ctx, request.Address)
	if err != nil {
		return nil, errorsmod.Wrap(err, "group policy")
	}

	return &group.QueryGroupPolicyInfoResponse{Info: &groupPolicyInfo}, nil
}

// getGroupPolicyInfo gets the group policy info of the given account address.
func (k Keeper) getGroupPolicyInfo(ctx sdk.Context, accountAddress string) (group.GroupPolicyInfo, error) {
	groupPolicyInfo, err := k.state.GroupPolicyInfoTable().Get(ctx, accountAddress)
	if err != nil {
		return group.GroupPolicyInfo{}, err
	}

	return group.GroupPolicyInfoFromPulsar(groupPolicyInfo), nil
}

// GroupMembers queries all members of a group.
func (k Keeper) GroupMembers(goCtx context.Context, request *group.QueryGroupMembersRequest) (*group.QueryGroupMembersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupID := request.GroupId
	it, err := k.getGroupMembers(ctx, groupID, request.Pagination)
	if err != nil {
		return nil, err
	}
	defer it.Close()

	var res group.QueryGroupMembersResponse
	for it.Next() {
		member, err := it.Value()
		if err != nil {
			return nil, err
		}

		m := group.GroupMemberFromPulsar(member)
		res.Members = append(res.Members, &m)
	}

	res.Pagination, err = ormutil.PulsarPageResToGogoPageRes(it.PageResponse())
	if err != nil {
		return nil, sdkerrors.ErrLogic.Wrap(err.Error())
	}

	return &res, nil
}

// getGroupMembers returns an iterator for the given group id and page request.
func (k Keeper) getGroupMembers(ctx sdk.Context, id uint64, pageRequest *query.PageRequest) (groupv1.GroupMemberIterator, error) {
	pg, err := ormutil.GogoPageReqToPulsarPageReq(pageRequest)
	if err != nil {
		return groupv1.GroupMemberIterator{}, fmt.Errorf("invalid page request: %w", err)
	}

	return k.state.GroupMemberTable().List(ctx, groupv1.GroupMemberGroupIdMemberAddressIndexKey{}.WithGroupId(id), ormlist.Paginate(pg))
}

// GroupsByAdmin queries all groups where a given address is admin.
func (k Keeper) GroupsByAdmin(goCtx context.Context, request *group.QueryGroupsByAdminRequest) (*group.QueryGroupsByAdminResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := k.accKeeper.AddressCodec().StringToBytes(request.Admin)
	if err != nil {
		return nil, err
	}

	it, err := k.getGroupsByAdmin(ctx, addr, request.Pagination)
	if err != nil {
		return nil, err
	}
	defer it.Close()

	var res group.QueryGroupsByAdminResponse
	for it.Next() {
		groupInfo, err := it.Value()
		if err != nil {
			return nil, err
		}

		g := group.GroupInfoFromPulsar(groupInfo)
		res.Groups = append(res.Groups, &g)
	}

	res.Pagination, err = ormutil.PulsarPageResToGogoPageRes(it.PageResponse())
	if err != nil {
		return nil, sdkerrors.ErrLogic.Wrap(err.Error())
	}

	return &res, nil
}

// getGroupsByAdmin returns an iterator for the given admin account address and page request.
func (k Keeper) getGroupsByAdmin(ctx sdk.Context, admin sdk.AccAddress, pageRequest *query.PageRequest) (groupv1.GroupInfoIterator, error) {
	pg, err := ormutil.GogoPageReqToPulsarPageReq(pageRequest)
	if err != nil {
		return groupv1.GroupInfoIterator{}, fmt.Errorf("invalid page request: %w", err)
	}

	return k.state.GroupInfoTable().List(ctx, groupv1.GroupInfoAdminIndexKey{}.WithAdmin(admin.String()), ormlist.Paginate(pg))
}

// GroupPoliciesByGroup queries all groups policies of a given group.
func (k Keeper) GroupPoliciesByGroup(goCtx context.Context, request *group.QueryGroupPoliciesByGroupRequest) (*group.QueryGroupPoliciesByGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupID := request.GroupId

	it, err := k.getGroupPoliciesByGroup(ctx, groupID, request.Pagination)
	if err != nil {
		return nil, err
	}
	defer it.Close()

	var res group.QueryGroupPoliciesByGroupResponse
	for it.Next() {
		policy, err := it.Value()
		if err != nil {
			return nil, err
		}

		p := group.GroupPolicyInfoFromPulsar(policy)
		res.GroupPolicies = append(res.GroupPolicies, &p)
	}

	res.Pagination, err = ormutil.PulsarPageResToGogoPageRes(it.PageResponse())
	if err != nil {
		return nil, sdkerrors.ErrLogic.Wrap(err.Error())
	}

	return &res, nil
}

// getGroupPoliciesByGroup returns an iterator for the given group id and page request.
func (k Keeper) getGroupPoliciesByGroup(ctx sdk.Context, id uint64, pageRequest *query.PageRequest) (groupv1.GroupPolicyInfoIterator, error) {
	pg, err := ormutil.GogoPageReqToPulsarPageReq(pageRequest)
	if err != nil {
		return groupv1.GroupPolicyInfoIterator{}, fmt.Errorf("invalid page request: %w", err)
	}

	return k.state.GroupPolicyInfoTable().List(ctx, groupv1.GroupPolicyInfoGroupIdIndexKey{}.WithGroupId(id), ormlist.Paginate(pg))
}

// GroupPoliciesByAdmin queries all groups policies where a given address is
// admin.
func (k Keeper) GroupPoliciesByAdmin(goCtx context.Context, request *group.QueryGroupPoliciesByAdminRequest) (*group.QueryGroupPoliciesByAdminResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := k.accKeeper.AddressCodec().StringToBytes(request.Admin)
	if err != nil {
		return nil, err
	}

	it, err := k.getGroupPoliciesByAdmin(ctx, addr, request.Pagination)
	if err != nil {
		return nil, err
	}
	defer it.Close()

	var res group.QueryGroupPoliciesByAdminResponse
	for it.Next() {
		policy, err := it.Value()
		if err != nil {
			return nil, err
		}

		p := group.GroupPolicyInfoFromPulsar(policy)
		res.GroupPolicies = append(res.GroupPolicies, &p)
	}

	res.Pagination, err = ormutil.PulsarPageResToGogoPageRes(it.PageResponse())
	if err != nil {
		return nil, sdkerrors.ErrLogic.Wrap(err.Error())
	}

	return &res, nil
}

// getGroupPoliciesByAdmin returns an iterator for the given admin account address and page request.
func (k Keeper) getGroupPoliciesByAdmin(ctx sdk.Context, admin sdk.AccAddress, pageRequest *query.PageRequest) (groupv1.GroupPolicyInfoIterator, error) {
	pg, err := ormutil.GogoPageReqToPulsarPageReq(pageRequest)
	if err != nil {
		return groupv1.GroupPolicyInfoIterator{}, fmt.Errorf("invalid page request: %w", err)
	}

	return k.state.GroupPolicyInfoTable().List(ctx, groupv1.GroupPolicyInfoAdminIndexKey{}.WithAdmin(admin.String()), ormlist.Paginate(pg))
}

// Proposal queries a proposal.
func (k Keeper) Proposal(goCtx context.Context, request *group.QueryProposalRequest) (*group.QueryProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	proposalID := request.ProposalId
	proposal, err := k.getProposal(ctx, proposalID)
	if err != nil {
		return nil, err
	}

	return &group.QueryProposalResponse{Proposal: &proposal}, nil
}

// ProposalsByGroupPolicy queries all proposals of a group policy.
func (k Keeper) ProposalsByGroupPolicy(goCtx context.Context, request *group.QueryProposalsByGroupPolicyRequest) (*group.QueryProposalsByGroupPolicyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := k.accKeeper.AddressCodec().StringToBytes(request.Address)
	if err != nil {
		return nil, err
	}

	it, err := k.getProposalsByGroupPolicy(ctx, addr, request.Pagination)
	if err != nil {
		return nil, err
	}
	defer it.Close()

	var res group.QueryProposalsByGroupPolicyResponse
	for it.Next() {
		proposal, err := it.Value()
		if err != nil {
			return nil, err
		}

		p := group.ProposalFromPulsar(proposal)
		res.Proposals = append(res.Proposals, &p)
	}

	res.Pagination, err = ormutil.PulsarPageResToGogoPageRes(it.PageResponse())
	if err != nil {
		return nil, sdkerrors.ErrLogic.Wrap(err.Error())
	}

	return &res, nil
}

// getProposalsByGroupPolicy returns an iterator for the given account address and page request.
func (k Keeper) getProposalsByGroupPolicy(ctx sdk.Context, account sdk.AccAddress, pageRequest *query.PageRequest) (groupv1.ProposalIterator, error) {
	pg, err := ormutil.GogoPageReqToPulsarPageReq(pageRequest)
	if err != nil {
		return groupv1.ProposalIterator{}, fmt.Errorf("invalid page request: %w", err)
	}

	return k.state.ProposalTable().List(ctx, groupv1.ProposalGroupPolicyAddressIndexKey{}.WithGroupPolicyAddress(account.String()), ormlist.Paginate(pg))
}

// getProposal gets the proposal info of the given proposal id.
func (k Keeper) getProposal(ctx sdk.Context, proposalID uint64) (group.Proposal, error) {
	proposal, err := k.state.ProposalTable().Get(ctx, proposalID)
	if err != nil {
		return group.Proposal{}, errorsmod.Wrap(err, "load proposal")
	}

	return group.ProposalFromPulsar(proposal), nil
}

// VoteByProposalVoter queries a vote given a voter and a proposal ID.
func (k Keeper) VoteByProposalVoter(goCtx context.Context, request *group.QueryVoteByProposalVoterRequest) (*group.QueryVoteByProposalVoterResponse, error) {
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
	return &group.QueryVoteByProposalVoterResponse{
		Vote: &vote,
	}, nil
}

// VotesByProposal queries all votes on a proposal.
func (k Keeper) VotesByProposal(goCtx context.Context, request *group.QueryVotesByProposalRequest) (*group.QueryVotesByProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	proposalID := request.ProposalId

	it, err := k.getVotesByProposal(ctx, proposalID, request.Pagination)
	if err != nil {
		return nil, err
	}
	defer it.Close()

	var res group.QueryVotesByProposalResponse
	for it.Next() {
		vote, err := it.Value()
		if err != nil {
			return nil, err
		}

		v := group.VoteFromPulsar(vote)
		res.Votes = append(res.Votes, &v)
	}

	res.Pagination, err = ormutil.PulsarPageResToGogoPageRes(it.PageResponse())
	if err != nil {
		return nil, sdkerrors.ErrLogic.Wrap(err.Error())
	}

	return &res, nil
}

// VotesByVoter queries all votes of a voter.
func (k Keeper) VotesByVoter(goCtx context.Context, request *group.QueryVotesByVoterRequest) (*group.QueryVotesByVoterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := k.accKeeper.AddressCodec().StringToBytes(request.Voter)
	if err != nil {
		return nil, err
	}

	it, err := k.getVotesByVoter(ctx, addr, request.Pagination)
	if err != nil {
		return nil, err
	}
	defer it.Close()

	var res group.QueryVotesByVoterResponse
	for it.Next() {
		vote, err := it.Value()
		if err != nil {
			return nil, err
		}

		v := group.VoteFromPulsar(vote)
		res.Votes = append(res.Votes, &v)
	}

	res.Pagination, err = ormutil.PulsarPageResToGogoPageRes(it.PageResponse())
	if err != nil {
		return nil, sdkerrors.ErrLogic.Wrap(err.Error())
	}

	return &res, nil
}

// GroupsByMember queries all groups where the given address is a member of.
func (k Keeper) GroupsByMember(goCtx context.Context, request *group.QueryGroupsByMemberRequest) (*group.QueryGroupsByMemberResponse, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	_, err := k.accKeeper.AddressCodec().StringToBytes(request.Address)
	if err != nil {
		return nil, err
	}

	pg, err := ormutil.GogoPageReqToPulsarPageReq(request.Pagination)
	if err != nil {
		return nil, fmt.Errorf("invalid page request: %w", err)
	}

	it, err := k.state.GroupMemberTable().List(ctx, groupv1.GroupMemberMemberAddressIndexKey{}.WithMemberAddress(request.Address), ormlist.Paginate(pg))
	if err != nil {
		return nil, err
	}
	defer it.Close()

	var res group.QueryGroupsByMemberResponse
	for it.Next() {
		member, err := it.Value()
		if err != nil {
			return nil, err
		}

		groupInfo, err := k.getGroupInfo(ctx, member.GroupId)
		if err != nil {
			return nil, err
		}

		res.Groups = append(res.Groups, &groupInfo)
	}

	res.Pagination, err = ormutil.PulsarPageResToGogoPageRes(it.PageResponse())
	if err != nil {
		return nil, sdkerrors.ErrLogic.Wrap(err.Error())
	}

	return &res, nil
}

// getVote gets the vote info for the given proposal id and voter address.
func (k Keeper) getVote(ctx sdk.Context, proposalID uint64, voter sdk.AccAddress) (group.Vote, error) {
	vote, err := k.state.VoteTable().Get(ctx, proposalID, voter.String())
	if err != nil {
		return group.Vote{}, errorsmod.Wrap(err, "load vote")
	}

	return group.VoteFromPulsar(vote), nil
}

// getVotesByProposal returns an iterator for the given proposal id and page request.
func (k Keeper) getVotesByProposal(ctx sdk.Context, proposalID uint64, pageRequest *query.PageRequest) (groupv1.VoteIterator, error) {
	pg, err := ormutil.GogoPageReqToPulsarPageReq(pageRequest)
	if err != nil {
		return groupv1.VoteIterator{}, fmt.Errorf("invalid page request: %w", err)
	}

	return k.state.VoteTable().List(ctx, groupv1.VoteProposalIdVoterIndexKey{}.WithProposalId(proposalID), ormlist.Paginate(pg))
}

// getVotesByVoter returns an iterator for the given voter address and page request.
func (k Keeper) getVotesByVoter(ctx sdk.Context, voter sdk.AccAddress, pageRequest *query.PageRequest) (groupv1.VoteIterator, error) {
	pg, err := ormutil.GogoPageReqToPulsarPageReq(pageRequest)
	if err != nil {
		return groupv1.VoteIterator{}, fmt.Errorf("invalid page request: %w", err)
	}

	return k.state.VoteTable().List(ctx, groupv1.VoteVoterIndexKey{}.WithVoter(voter.String()), ormlist.Paginate(pg))
}

// TallyResult computes the live tally result of a proposal.
func (k Keeper) TallyResult(goCtx context.Context, request *group.QueryTallyResultRequest) (*group.QueryTallyResultResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	proposalID := request.ProposalId

	proposal, err := k.getProposal(ctx, proposalID)
	if err != nil {
		return nil, err
	}

	if proposal.Status == group.PROPOSAL_STATUS_WITHDRAWN || proposal.Status == group.PROPOSAL_STATUS_ABORTED {
		return nil, errorsmod.Wrapf(errors.ErrInvalid, "can't get the tally of a proposal with status %s", proposal.Status)
	}

	var policyInfo group.GroupPolicyInfo
	if policyInfo, err = k.getGroupPolicyInfo(ctx, proposal.GroupPolicyAddress); err != nil {
		return nil, errorsmod.Wrap(err, "load group policy")
	}

	tallyResult, err := k.Tally(ctx, proposal, policyInfo.GroupId)
	if err != nil {
		return nil, err
	}

	return &group.QueryTallyResultResponse{
		Tally: tallyResult,
	}, nil
}

// Groups returns all the groups present in the state.
func (k Keeper) Groups(goCtx context.Context, request *group.QueryGroupsRequest) (*group.QueryGroupsResponse, error) {
	pg, err := ormutil.GogoPageReqToPulsarPageReq(request.Pagination)
	if err != nil {
		return nil, fmt.Errorf("invalid page in request: %w (got %v)", err, request.Pagination)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	it, err := k.state.GroupInfoTable().List(ctx, &groupv1.GroupInfoIdIndexKey{}, ormlist.Paginate(pg))
	if err != nil {
		return nil, err
	}
	defer it.Close()

	var res group.QueryGroupsResponse
	for it.Next() {
		groupInfo, err := it.Value()
		if err != nil {
			return nil, err
		}

		g := group.GroupInfoFromPulsar(groupInfo)
		res.Groups = append(res.Groups, &g)
	}

	res.Pagination, err = ormutil.PulsarPageResToGogoPageRes(it.PageResponse())
	if err != nil {
		return nil, sdkerrors.ErrLogic.Wrap(err.Error())
	}

	return &res, nil
}
