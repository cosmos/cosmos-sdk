package keeper

import (
	"context"
	"encoding/binary"
	"fmt"
	"reflect"

	gogotypes "github.com/gogo/protobuf/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/group/errors"
	"github.com/cosmos/cosmos-sdk/x/group/internal/math"
	"github.com/cosmos/cosmos-sdk/x/group/internal/orm"
)

var _ group.MsgServer = Keeper{}

// TODO: Revisit this once we have propoer gas fee framework.
// Tracking issues https://github.com/cosmos/cosmos-sdk/issues/9054, https://github.com/cosmos/cosmos-sdk/discussions/9072
const gasCostPerIteration = uint64(20)

func (k Keeper) CreateGroup(goCtx context.Context, req *group.MsgCreateGroup) (*group.MsgCreateGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	metadata := req.Metadata
	members := group.Members{Members: req.Members}
	admin := req.Admin

	if err := members.ValidateBasic(); err != nil {
		return nil, err
	}

	if err := k.assertMetadataLength(metadata, "group metadata"); err != nil {
		return nil, err
	}

	totalWeight := math.NewDecFromInt64(0)
	for i := range members.Members {
		m := members.Members[i]
		if err := k.assertMetadataLength(m.Metadata, "member metadata"); err != nil {
			return nil, err
		}

		// Members of a group must have a positive weight.
		weight, err := math.NewPositiveDecFromString(m.Weight)
		if err != nil {
			return nil, err
		}

		// Adding up members weights to compute group total weight.
		totalWeight, err = totalWeight.Add(weight)
		if err != nil {
			return nil, err
		}
	}

	// Create a new group in the groupTable.
	groupInfo := &group.GroupInfo{
		Id:          k.groupTable.Sequence().PeekNextVal(ctx.KVStore(k.key)),
		Admin:       admin,
		Metadata:    metadata,
		Version:     1,
		TotalWeight: totalWeight.String(),
		CreatedAt:   ctx.BlockTime(),
	}
	groupID, err := k.groupTable.Create(ctx.KVStore(k.key), groupInfo)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not create group")
	}

	// Create new group members in the groupMemberTable.
	for i := range members.Members {
		m := members.Members[i]
		err := k.groupMemberTable.Create(ctx.KVStore(k.key), &group.GroupMember{
			GroupId: groupID,
			Member: &group.Member{
				Address:  m.Address,
				Weight:   m.Weight,
				Metadata: m.Metadata,
				AddedAt:  ctx.BlockTime(),
			},
		})
		if err != nil {
			return nil, sdkerrors.Wrapf(err, "could not store member %d", i)
		}
	}

	err = ctx.EventManager().EmitTypedEvent(&group.EventCreateGroup{GroupId: groupID})
	if err != nil {
		return nil, err
	}

	return &group.MsgCreateGroupResponse{GroupId: groupID}, nil
}

func (k Keeper) UpdateGroupMembers(goCtx context.Context, req *group.MsgUpdateGroupMembers) (*group.MsgUpdateGroupMembersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	action := func(g *group.GroupInfo) error {
		totalWeight, err := math.NewNonNegativeDecFromString(g.TotalWeight)
		if err != nil {
			return err
		}
		for i := range req.MemberUpdates {
			if err := k.assertMetadataLength(req.MemberUpdates[i].Metadata, "group member metadata"); err != nil {
				return err
			}
			groupMember := group.GroupMember{GroupId: req.GroupId,
				Member: &group.Member{
					Address:  req.MemberUpdates[i].Address,
					Weight:   req.MemberUpdates[i].Weight,
					Metadata: req.MemberUpdates[i].Metadata,
				},
			}

			// Checking if the group member is already part of the group
			var found bool
			var prevGroupMember group.GroupMember
			switch err := k.groupMemberTable.GetOne(ctx.KVStore(k.key), orm.PrimaryKey(&groupMember), &prevGroupMember); {
			case err == nil:
				found = true
			case sdkerrors.ErrNotFound.Is(err):
				found = false
			default:
				return sdkerrors.Wrap(err, "get group member")
			}

			newMemberWeight, err := math.NewNonNegativeDecFromString(groupMember.Member.Weight)
			if err != nil {
				return err
			}

			// Handle delete for members with zero weight.
			if newMemberWeight.IsZero() {
				// We can't delete a group member that doesn't already exist.
				if !found {
					return sdkerrors.Wrap(sdkerrors.ErrNotFound, "unknown member")
				}

				previousMemberWeight, err := math.NewNonNegativeDecFromString(prevGroupMember.Member.Weight)
				if err != nil {
					return err
				}

				// Subtract the weight of the group member to delete from the group total weight.
				totalWeight, err = math.SubNonNegative(totalWeight, previousMemberWeight)
				if err != nil {
					return err
				}

				// Delete group member in the groupMemberTable.
				if err := k.groupMemberTable.Delete(ctx.KVStore(k.key), &groupMember); err != nil {
					return sdkerrors.Wrap(err, "delete member")
				}
				continue
			}
			// If group member already exists, handle update
			if found {
				previousMemberWeight, err := math.NewNonNegativeDecFromString(prevGroupMember.Member.Weight)
				if err != nil {
					return err
				}
				// Subtract previous weight from the group total weight.
				totalWeight, err = math.SubNonNegative(totalWeight, previousMemberWeight)
				if err != nil {
					return err
				}
				// Save updated group member in the groupMemberTable.
				if err := k.groupMemberTable.Update(ctx.KVStore(k.key), &groupMember); err != nil {
					return sdkerrors.Wrap(err, "add member")
				}
				// else handle create.
			} else if err := k.groupMemberTable.Create(ctx.KVStore(k.key), &groupMember); err != nil {
				return sdkerrors.Wrap(err, "add member")
			}
			// In both cases (handle + update), we need to add the new member's weight to the group total weight.
			totalWeight, err = totalWeight.Add(newMemberWeight)
			if err != nil {
				return err
			}
		}
		// Update group in the groupTable.
		g.TotalWeight = totalWeight.String()
		g.Version++
		return k.groupTable.Update(ctx.KVStore(k.key), g.Id, g)
	}

	err := k.doUpdateGroup(ctx, req, action, "members updated")
	if err != nil {
		return nil, err
	}

	return &group.MsgUpdateGroupMembersResponse{}, nil
}

func (k Keeper) UpdateGroupAdmin(goCtx context.Context, req *group.MsgUpdateGroupAdmin) (*group.MsgUpdateGroupAdminResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	action := func(g *group.GroupInfo) error {
		g.Admin = req.NewAdmin
		g.Version++

		return k.groupTable.Update(ctx.KVStore(k.key), g.Id, g)
	}

	err := k.doUpdateGroup(ctx, req, action, "admin updated")
	if err != nil {
		return nil, err
	}

	return &group.MsgUpdateGroupAdminResponse{}, nil
}

func (k Keeper) UpdateGroupMetadata(goCtx context.Context, req *group.MsgUpdateGroupMetadata) (*group.MsgUpdateGroupMetadataResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	action := func(g *group.GroupInfo) error {
		g.Metadata = req.Metadata
		g.Version++
		return k.groupTable.Update(ctx.KVStore(k.key), g.Id, g)
	}

	if err := k.assertMetadataLength(req.Metadata, "group metadata"); err != nil {
		return nil, err
	}

	err := k.doUpdateGroup(ctx, req, action, "metadata updated")
	if err != nil {
		return nil, err
	}

	return &group.MsgUpdateGroupMetadataResponse{}, nil
}

func (k Keeper) CreateGroupWithPolicy(goCtx context.Context, req *group.MsgCreateGroupWithPolicy) (*group.MsgCreateGroupWithPolicyResponse, error) {
	groupRes, err := k.CreateGroup(goCtx, &group.MsgCreateGroup{
		Admin:    req.Admin,
		Members:  req.Members,
		Metadata: req.GroupMetadata,
	})
	if err != nil {
		return nil, sdkerrors.Wrap(err, "group response")
	}
	groupId := groupRes.GroupId

	var groupPolicyAddr sdk.AccAddress
	groupPolicyRes, err := k.CreateGroupPolicy(goCtx, &group.MsgCreateGroupPolicy{
		Admin:          req.Admin,
		GroupId:        groupId,
		Metadata:       req.GroupPolicyMetadata,
		DecisionPolicy: req.DecisionPolicy,
	})
	if err != nil {
		return nil, sdkerrors.Wrap(err, "group policy response")
	}
	policyAddr := groupPolicyRes.Address

	groupPolicyAddr, err = sdk.AccAddressFromBech32(policyAddr)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "group policy address")
	}
	groupPolicyAddress := groupPolicyAddr.String()

	if req.GroupPolicyAsAdmin {
		updateAdminReq := &group.MsgUpdateGroupAdmin{
			GroupId:  groupId,
			Admin:    req.Admin,
			NewAdmin: groupPolicyAddress,
		}
		_, err = k.UpdateGroupAdmin(goCtx, updateAdminReq)
		if err != nil {
			return nil, err
		}

		updatePolicyAddressReq := &group.MsgUpdateGroupPolicyAdmin{
			Admin:    req.Admin,
			Address:  groupPolicyAddress,
			NewAdmin: groupPolicyAddress,
		}
		_, err = k.UpdateGroupPolicyAdmin(goCtx, updatePolicyAddressReq)
		if err != nil {
			return nil, err
		}
	}

	return &group.MsgCreateGroupWithPolicyResponse{GroupId: groupId, GroupPolicyAddress: groupPolicyAddress}, nil
}

func (k Keeper) CreateGroupPolicy(goCtx context.Context, req *group.MsgCreateGroupPolicy) (*group.MsgCreateGroupPolicyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	admin, err := sdk.AccAddressFromBech32(req.GetAdmin())
	if err != nil {
		return nil, sdkerrors.Wrap(err, "request admin")
	}
	policy := req.GetDecisionPolicy()
	groupID := req.GetGroupID()
	metadata := req.GetMetadata()

	if err := k.assertMetadataLength(metadata, "group policy metadata"); err != nil {
		return nil, err
	}

	g, err := k.getGroupInfo(ctx, groupID)
	if err != nil {
		return nil, err
	}
	groupAdmin, err := sdk.AccAddressFromBech32(g.Admin)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "group admin")
	}
	// Only current group admin is authorized to create a group policy for this
	if !groupAdmin.Equals(admin) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "not group admin")
	}

	// Generate account address of group policy.
	var accountAddr sdk.AccAddress
	// loop here in the rare case of a collision
	for {
		nextAccVal := k.groupPolicySeq.NextVal(ctx.KVStore(k.key))
		var buf = make([]byte, 8)
		binary.BigEndian.PutUint64(buf, nextAccVal)

		parentAcc := address.Module(group.ModuleName, []byte{GroupPolicyTablePrefix})
		accountAddr = address.Derive(parentAcc, buf)

		if k.accKeeper.GetAccount(ctx, accountAddr) != nil {
			// handle a rare collision
			continue
		}
		acc := k.accKeeper.NewAccount(ctx, &authtypes.ModuleAccount{
			BaseAccount: &authtypes.BaseAccount{
				Address: accountAddr.String(),
			},
			Name: accountAddr.String(),
		})
		k.accKeeper.SetAccount(ctx, acc)

		break
	}

	groupPolicy, err := group.NewGroupPolicyInfo(
		accountAddr,
		groupID,
		admin,
		metadata,
		1,
		policy,
		ctx.BlockTime(),
	)
	if err != nil {
		return nil, err
	}

	if err := k.groupPolicyTable.Create(ctx.KVStore(k.key), &groupPolicy); err != nil {
		return nil, sdkerrors.Wrap(err, "could not create group policy")
	}

	err = ctx.EventManager().EmitTypedEvent(&group.EventCreateGroupPolicy{Address: accountAddr.String()})
	if err != nil {
		return nil, err
	}

	return &group.MsgCreateGroupPolicyResponse{Address: accountAddr.String()}, nil
}

func (k Keeper) UpdateGroupPolicyAdmin(goCtx context.Context, req *group.MsgUpdateGroupPolicyAdmin) (*group.MsgUpdateGroupPolicyAdminResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	action := func(groupPolicy *group.GroupPolicyInfo) error {
		groupPolicy.Admin = req.NewAdmin
		groupPolicy.Version++
		return k.groupPolicyTable.Update(ctx.KVStore(k.key), groupPolicy)
	}

	err := k.doUpdateGroupPolicy(ctx, req.Address, req.Admin, action, "group policy admin updated")
	if err != nil {
		return nil, err
	}

	return &group.MsgUpdateGroupPolicyAdminResponse{}, nil
}

func (k Keeper) UpdateGroupPolicyDecisionPolicy(goCtx context.Context, req *group.MsgUpdateGroupPolicyDecisionPolicy) (*group.MsgUpdateGroupPolicyDecisionPolicyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	policy := req.GetDecisionPolicy()

	action := func(groupPolicy *group.GroupPolicyInfo) error {
		err := groupPolicy.SetDecisionPolicy(policy)
		if err != nil {
			return err
		}

		groupPolicy.Version++
		return k.groupPolicyTable.Update(ctx.KVStore(k.key), groupPolicy)
	}

	err := k.doUpdateGroupPolicy(ctx, req.Address, req.Admin, action, "group policy's decision policy updated")
	if err != nil {
		return nil, err
	}

	return &group.MsgUpdateGroupPolicyDecisionPolicyResponse{}, nil
}

func (k Keeper) UpdateGroupPolicyMetadata(goCtx context.Context, req *group.MsgUpdateGroupPolicyMetadata) (*group.MsgUpdateGroupPolicyMetadataResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	metadata := req.GetMetadata()

	action := func(groupPolicy *group.GroupPolicyInfo) error {
		groupPolicy.Metadata = metadata
		groupPolicy.Version++
		return k.groupPolicyTable.Update(ctx.KVStore(k.key), groupPolicy)
	}

	if err := k.assertMetadataLength(metadata, "group policy metadata"); err != nil {
		return nil, err
	}

	err := k.doUpdateGroupPolicy(ctx, req.Address, req.Admin, action, "group policy metadata updated")
	if err != nil {
		return nil, err
	}

	return &group.MsgUpdateGroupPolicyMetadataResponse{}, nil
}

func (k Keeper) SubmitProposal(goCtx context.Context, req *group.MsgSubmitProposal) (*group.MsgSubmitProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	accountAddress, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "request account address of group policy")
	}
	metadata := req.Metadata
	proposers := req.Proposers
	msgs := req.GetMsgs()

	if err := k.assertMetadataLength(metadata, "metadata"); err != nil {
		return nil, err
	}

	policyAcc, err := k.getGroupPolicyInfo(ctx, req.Address)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "load group policy")
	}

	g, err := k.getGroupInfo(ctx, policyAcc.GroupId)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "get group by groupId of group policy")
	}

	// Only members of the group can submit a new proposal.
	for i := range proposers {
		if !k.groupMemberTable.Has(ctx.KVStore(k.key), orm.PrimaryKey(&group.GroupMember{GroupId: g.Id, Member: &group.Member{Address: proposers[i]}})) {
			return nil, sdkerrors.Wrapf(errors.ErrUnauthorized, "not in group: %s", proposers[i])
		}
	}

	// Check that if the messages require signers, they are all equal to the given account address of group policy.
	if err := ensureMsgAuthZ(msgs, accountAddress); err != nil {
		return nil, err
	}

	policy := policyAcc.GetDecisionPolicy()
	if policy == nil {
		return nil, sdkerrors.Wrap(errors.ErrEmpty, "nil policy")
	}

	// Prevent proposal that can not succeed.
	err = policy.Validate(g)
	if err != nil {
		return nil, err
	}

	// Define proposal timout.
	// The voting window begins as soon as the proposal is submitted.
	timeout := policy.GetTimeout()
	window := timeout

	m := &group.Proposal{
		Id:                 k.proposalTable.Sequence().PeekNextVal(ctx.KVStore(k.key)),
		Address:            req.Address,
		Metadata:           metadata,
		Proposers:          proposers,
		SubmitTime:         ctx.BlockTime(),
		GroupVersion:       g.Version,
		GroupPolicyVersion: policyAcc.Version,
		Result:             group.PROPOSAL_RESULT_UNFINALIZED,
		Status:             group.PROPOSAL_STATUS_SUBMITTED,
		ExecutorResult:     group.PROPOSAL_EXECUTOR_RESULT_NOT_RUN,
		Timeout:            ctx.BlockTime().Add(window),
		FinalTallyResult: group.TallyResult{
			YesCount:        "0",
			NoCount:         "0",
			AbstainCount:    "0",
			NoWithVetoCount: "0",
		},
	}
	if err := m.SetMsgs(msgs); err != nil {
		return nil, sdkerrors.Wrap(err, "create proposal")
	}

	id, err := k.proposalTable.Create(ctx.KVStore(k.key), m)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "create proposal")
	}

	err = ctx.EventManager().EmitTypedEvent(&group.EventSubmitProposal{ProposalId: id})
	if err != nil {
		return nil, err
	}

	// Try to execute proposal immediately
	if req.Exec == group.Exec_EXEC_TRY {
		// Consider proposers as Yes votes
		for i := range proposers {
			ctx.GasMeter().ConsumeGas(gasCostPerIteration, "vote on proposal")
			_, err = k.Vote(sdk.WrapSDKContext(ctx), &group.MsgVote{
				ProposalId: id,
				Voter:      proposers[i],
				Option:     group.VOTE_OPTION_YES,
			})
			if err != nil {
				return &group.MsgSubmitProposalResponse{ProposalId: id}, sdkerrors.Wrap(err, "The proposal was created but failed on vote")
			}
		}
		// Then try to execute the proposal
		_, err = k.Exec(sdk.WrapSDKContext(ctx), &group.MsgExec{
			ProposalId: id,
			// We consider the first proposer as the MsgExecRequest signer
			// but that could be revisited (eg using the group policy)
			Signer: proposers[0],
		})
		if err != nil {
			return &group.MsgSubmitProposalResponse{ProposalId: id}, sdkerrors.Wrap(err, "The proposal was created but failed on exec")
		}
	}

	return &group.MsgSubmitProposalResponse{ProposalId: id}, nil
}

func (k Keeper) WithdrawProposal(goCtx context.Context, req *group.MsgWithdrawProposal) (*group.MsgWithdrawProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	id := req.ProposalId
	address := req.Address

	proposal, err := k.getProposal(ctx, id)
	if err != nil {
		return nil, err
	}

	// Ensure the proposal can be withdrawn.
	if proposal.Status != group.PROPOSAL_STATUS_SUBMITTED {
		return nil, sdkerrors.Wrapf(errors.ErrInvalid, "cannot withdraw a proposal with the status of %s", proposal.Status.String())
	}

	var policyInfo group.GroupPolicyInfo
	if policyInfo, err = k.getGroupPolicyInfo(ctx, proposal.Address); err != nil {
		return nil, sdkerrors.Wrap(err, "load group policy")
	}

	storeUpdates := func() (*group.MsgWithdrawProposalResponse, error) {
		if err := k.proposalTable.Update(ctx.KVStore(k.key), id, &proposal); err != nil {
			return nil, err
		}
		return &group.MsgWithdrawProposalResponse{}, nil
	}

	// check address is the group policy admin.
	if address == policyInfo.Address {
		err = ctx.EventManager().EmitTypedEvent(&group.EventWithdrawProposal{ProposalId: id})
		if err != nil {
			return nil, err
		}

		proposal.Result = group.PROPOSAL_RESULT_UNFINALIZED
		proposal.Status = group.PROPOSAL_STATUS_WITHDRAWN
		return storeUpdates()
	}

	// if address is not group policy admin then check whether he is in proposers list.
	validProposer := false
	for _, proposer := range proposal.Proposers {
		if proposer == address {
			validProposer = true
			break
		}
	}

	if !validProposer {
		return nil, sdkerrors.Wrapf(errors.ErrUnauthorized, "given address is neither group policy admin nor in proposers: %s", address)
	}

	err = ctx.EventManager().EmitTypedEvent(&group.EventWithdrawProposal{ProposalId: id})
	if err != nil {
		return nil, err
	}

	proposal.Result = group.PROPOSAL_RESULT_UNFINALIZED
	proposal.Status = group.PROPOSAL_STATUS_WITHDRAWN
	return storeUpdates()
}

func (k Keeper) Vote(goCtx context.Context, req *group.MsgVote) (*group.MsgVoteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	id := req.ProposalId
	voteOption := req.Option
	metadata := req.Metadata

	if err := k.assertMetadataLength(metadata, "metadata"); err != nil {
		return nil, err
	}

	proposal, err := k.getProposal(ctx, id)
	if err != nil {
		return nil, err
	}
	// Ensure that we can still accept votes for this proposal.
	if proposal.Status != group.PROPOSAL_STATUS_SUBMITTED {
		return nil, sdkerrors.Wrap(errors.ErrInvalid, "proposal not open for voting")
	}
	proposalTimeout, err := gogotypes.TimestampProto(proposal.Timeout)
	if err != nil {
		return nil, err
	}
	votingPeriodEnd, err := gogotypes.TimestampFromProto(proposalTimeout)
	if err != nil {
		return nil, err
	}
	if votingPeriodEnd.Before(ctx.BlockTime()) || votingPeriodEnd.Equal(ctx.BlockTime()) {
		return nil, sdkerrors.Wrap(errors.ErrExpired, "voting period has ended already")
	}

	var policyInfo group.GroupPolicyInfo

	// Ensure that group policy hasn't been modified since the proposal submission.
	if policyInfo, err = k.getGroupPolicyInfo(ctx, proposal.Address); err != nil {
		return nil, sdkerrors.Wrap(err, "load group policy")
	}
	if proposal.GroupPolicyVersion != policyInfo.Version {
		return nil, sdkerrors.Wrap(errors.ErrModified, "group policy was modified")
	}

	// Ensure that group hasn't been modified since the proposal submission.
	electorate, err := k.getGroupInfo(ctx, policyInfo.GroupId)
	if err != nil {
		return nil, err
	}
	if electorate.Version != proposal.GroupVersion {
		return nil, sdkerrors.Wrap(errors.ErrModified, "group was modified")
	}

	// Count and store votes.
	voterAddr := req.Voter
	voter := group.GroupMember{GroupId: electorate.Id, Member: &group.Member{Address: voterAddr}}
	if err := k.groupMemberTable.GetOne(ctx.KVStore(k.key), orm.PrimaryKey(&voter), &voter); err != nil {
		return nil, sdkerrors.Wrapf(err, "address: %s", voterAddr)
	}
	newVote := group.Vote{
		ProposalId: id,
		Voter:      voterAddr,
		Option:     voteOption,
		Metadata:   metadata,
		SubmitTime: ctx.BlockTime(),
	}
	if err := proposal.FinalTallyResult.Add(newVote, voter.Member.Weight); err != nil {
		return nil, sdkerrors.Wrap(err, "add new vote")
	}

	// The ORM will return an error if the vote already exists,
	// making sure than a voter hasn't already voted.
	if err := k.voteTable.Create(ctx.KVStore(k.key), &newVote); err != nil {
		return nil, sdkerrors.Wrap(err, "store vote")
	}

	// Run tally with new votes to close early.
	if err := doTally(ctx, &proposal, electorate, policyInfo); err != nil {
		return nil, err
	}

	if err = k.proposalTable.Update(ctx.KVStore(k.key), id, &proposal); err != nil {
		return nil, err
	}

	err = ctx.EventManager().EmitTypedEvent(&group.EventVote{ProposalId: id})
	if err != nil {
		return nil, err
	}

	// Try to execute proposal immediately
	if req.Exec == group.Exec_EXEC_TRY {
		_, err = k.Exec(sdk.WrapSDKContext(ctx), &group.MsgExec{
			ProposalId: id,
			Signer:     voterAddr,
		})
		if err != nil {
			return nil, err
		}
	}

	return &group.MsgVoteResponse{}, nil
}

// doTally updates the proposal status and tally if necessary based on the group policy's decision policy.
func doTally(ctx sdk.Context, p *group.Proposal, electorate group.GroupInfo, policyInfo group.GroupPolicyInfo) error {
	policy := policyInfo.GetDecisionPolicy()
	pSubmittedAt, err := gogotypes.TimestampProto(p.SubmitTime)
	if err != nil {
		return err
	}
	submittedAt, err := gogotypes.TimestampFromProto(pSubmittedAt)
	if err != nil {
		return err
	}
	switch result, err := policy.Allow(p.FinalTallyResult, electorate.TotalWeight, ctx.BlockTime().Sub(submittedAt)); {
	case err != nil:
		return sdkerrors.Wrap(err, "policy execution")
	case result.Allow && result.Final:
		p.Result = group.PROPOSAL_RESULT_ACCEPTED
		p.Status = group.PROPOSAL_STATUS_CLOSED
	case !result.Allow && result.Final:
		p.Result = group.PROPOSAL_RESULT_REJECTED
		p.Status = group.PROPOSAL_STATUS_CLOSED
	}
	return nil
}

// Exec executes the messages from a proposal.
func (k Keeper) Exec(goCtx context.Context, req *group.MsgExec) (*group.MsgExecResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	id := req.ProposalId

	proposal, err := k.getProposal(ctx, id)
	if err != nil {
		return nil, err
	}

	if proposal.Status != group.PROPOSAL_STATUS_SUBMITTED && proposal.Status != group.PROPOSAL_STATUS_CLOSED {
		return nil, sdkerrors.Wrapf(errors.ErrInvalid, "not possible with proposal status %s", proposal.Status.String())
	}

	var policyInfo group.GroupPolicyInfo
	if policyInfo, err = k.getGroupPolicyInfo(ctx, proposal.Address); err != nil {
		return nil, sdkerrors.Wrap(err, "load group policy")
	}

	storeUpdates := func() (*group.MsgExecResponse, error) {
		if err := k.proposalTable.Update(ctx.KVStore(k.key), id, &proposal); err != nil {
			return nil, err
		}
		return &group.MsgExecResponse{}, nil
	}

	if proposal.Status == group.PROPOSAL_STATUS_SUBMITTED {
		// Ensure that group policy hasn't been modified before tally.
		if proposal.GroupPolicyVersion != policyInfo.Version {
			proposal.Result = group.PROPOSAL_RESULT_UNFINALIZED
			proposal.Status = group.PROPOSAL_STATUS_ABORTED
			return storeUpdates()
		}

		electorate, err := k.getGroupInfo(ctx, policyInfo.GroupId)
		if err != nil {
			return nil, sdkerrors.Wrap(err, "load group")
		}

		// Ensure that group hasn't been modified before tally.
		if electorate.Version != proposal.GroupVersion {
			proposal.Result = group.PROPOSAL_RESULT_UNFINALIZED
			proposal.Status = group.PROPOSAL_STATUS_ABORTED
			return storeUpdates()
		}
		if err := doTally(ctx, &proposal, electorate, policyInfo); err != nil {
			return nil, err
		}
	}

	// Execute proposal payload.
	if proposal.Status == group.PROPOSAL_STATUS_CLOSED && proposal.Result == group.PROPOSAL_RESULT_ACCEPTED && proposal.ExecutorResult != group.PROPOSAL_EXECUTOR_RESULT_SUCCESS {
		logger := ctx.Logger().With("module", fmt.Sprintf("x/%s", group.ModuleName))
		// Caching context so that we don't update the store in case of failure.
		ctx, flush := ctx.CacheContext()

		addr, err := sdk.AccAddressFromBech32(policyInfo.Address)
		if err != nil {
			return nil, err
		}
		_, err = k.doExecuteMsgs(ctx, k.router, proposal, addr)
		if err != nil {
			proposal.ExecutorResult = group.PROPOSAL_EXECUTOR_RESULT_FAILURE
			proposalType := reflect.TypeOf(proposal).String()
			logger.Info("proposal execution failed", "cause", err, "type", proposalType, "proposalID", id)
		} else {
			proposal.ExecutorResult = group.PROPOSAL_EXECUTOR_RESULT_SUCCESS
			flush()
		}
	}

	// Update proposal in proposalTable
	res, err := storeUpdates()
	if err != nil {
		return nil, err
	}

	err = ctx.EventManager().EmitTypedEvent(&group.EventExec{ProposalId: id})
	if err != nil {
		return nil, err
	}

	return res, nil
}

type authNGroupReq interface {
	GetGroupID() uint64
	GetAdmin() string
}

type actionFn func(m *group.GroupInfo) error
type groupPolicyActionFn func(m *group.GroupPolicyInfo) error

// doUpdateGroupPolicy first makes sure that the group policy admin initiated the group policy update,
// before performing the group policy update and emitting an event.
func (k Keeper) doUpdateGroupPolicy(ctx sdk.Context, groupPolicy string, admin string, action groupPolicyActionFn, note string) error {
	groupPolicyInfo, err := k.getGroupPolicyInfo(ctx, groupPolicy)
	if err != nil {
		return sdkerrors.Wrap(err, "load group policy")
	}

	groupPolicyAdmin, err := sdk.AccAddressFromBech32(admin)
	if err != nil {
		return sdkerrors.Wrap(err, "group policy admin")
	}

	// Only current group policy admin is authorized to update a group policy.
	if groupPolicyAdmin.String() != groupPolicyInfo.Admin {
		return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "not group policy admin")
	}

	if err := action(&groupPolicyInfo); err != nil {
		return sdkerrors.Wrap(err, note)
	}

	err = ctx.EventManager().EmitTypedEvent(&group.EventUpdateGroupPolicy{Address: admin})
	if err != nil {
		return err
	}

	return nil
}

// doUpdateGroup first makes sure that the group admin initiated the group update,
// before performing the group update and emitting an event.
func (k Keeper) doUpdateGroup(ctx sdk.Context, req authNGroupReq, action actionFn, note string) error {
	err := k.doAuthenticated(ctx, req, action, note)
	if err != nil {
		return err
	}

	err = ctx.EventManager().EmitTypedEvent(&group.EventUpdateGroup{GroupId: req.GetGroupID()})
	if err != nil {
		return err
	}

	return nil
}

// doAuthenticated makes sure that the group admin initiated the request,
// and perform the provided action on the
func (k Keeper) doAuthenticated(ctx sdk.Context, req authNGroupReq, action actionFn, note string) error {
	group, err := k.getGroupInfo(ctx, req.GetGroupID())
	if err != nil {
		return err
	}
	admin, err := sdk.AccAddressFromBech32(group.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "group admin")
	}
	reqAdmin, err := sdk.AccAddressFromBech32(req.GetAdmin())
	if err != nil {
		return sdkerrors.Wrap(err, "request admin")
	}
	if !admin.Equals(reqAdmin) {
		return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "not group admin")
	}
	if err := action(&group); err != nil {
		return sdkerrors.Wrap(err, note)
	}
	return nil
}

// assertMetadataLength returns an error if given metadata length
// is greater than a pre-defined maxMetadataLen.
func (k Keeper) assertMetadataLength(metadata []byte, description string) error {
	if metadata != nil && uint64(len(metadata)) > k.config.MaxMetadataLen {
		return sdkerrors.Wrapf(errors.ErrMaxLimit, description)
	}
	return nil
}
