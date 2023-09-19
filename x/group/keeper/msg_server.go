package keeper

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/group/errors"
	"github.com/cosmos/cosmos-sdk/x/group/internal/math"
	"github.com/cosmos/cosmos-sdk/x/group/internal/orm"
)

var _ group.MsgServer = Keeper{}

// TODO: Revisit this once we have proper gas fee framework.
// Tracking issues https://github.com/cosmos/cosmos-sdk/issues/9054, https://github.com/cosmos/cosmos-sdk/discussions/9072
const gasCostPerIteration = uint64(20)

func (k Keeper) CreateGroup(goCtx context.Context, msg *group.MsgCreateGroup) (*group.MsgCreateGroupResponse, error) {
	if _, err := k.accKeeper.AddressCodec().StringToBytes(msg.Admin); err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid admin address: %s", msg.Admin)
	}

	if err := k.validateMembers(msg.Members); err != nil {
		return nil, errorsmod.Wrap(err, "members")
	}

	if err := k.assertMetadataLength(msg.Metadata, "group metadata"); err != nil {
		return nil, err
	}

	totalWeight := math.NewDecFromInt64(0)
	for _, m := range msg.Members {
		if err := k.assertMetadataLength(m.Metadata, "member metadata"); err != nil {
			return nil, err
		}

		// Members of a group must have a positive weight.
		// NOTE: group member with zero weight are only allowed when updating group members.
		// If the member has a zero weight, it will be removed from the group.
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
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupInfo := &group.GroupInfo{
		Id:          k.groupTable.Sequence().PeekNextVal(ctx.KVStore(k.key)),
		Admin:       msg.Admin,
		Metadata:    msg.Metadata,
		Version:     1,
		TotalWeight: totalWeight.String(),
		CreatedAt:   ctx.HeaderInfo().Time,
	}
	groupID, err := k.groupTable.Create(ctx.KVStore(k.key), groupInfo)
	if err != nil {
		return nil, errorsmod.Wrap(err, "could not create group")
	}

	// Create new group members in the groupMemberTable.
	for i, m := range msg.Members {
		err := k.groupMemberTable.Create(ctx.KVStore(k.key), &group.GroupMember{
			GroupId: groupID,
			Member: &group.Member{
				Address:  m.Address,
				Weight:   m.Weight,
				Metadata: m.Metadata,
				AddedAt:  ctx.HeaderInfo().Time,
			},
		})
		if err != nil {
			return nil, errorsmod.Wrapf(err, "could not store member %d", i)
		}
	}

	if err := ctx.EventManager().EmitTypedEvent(&group.EventCreateGroup{GroupId: groupID}); err != nil {
		return nil, err
	}

	return &group.MsgCreateGroupResponse{GroupId: groupID}, nil
}

func (k Keeper) UpdateGroupMembers(goCtx context.Context, msg *group.MsgUpdateGroupMembers) (*group.MsgUpdateGroupMembersResponse, error) {
	if msg.GroupId == 0 {
		return nil, errorsmod.Wrap(errors.ErrEmpty, "group id")
	}

	if len(msg.MemberUpdates) == 0 {
		return nil, errorsmod.Wrap(errors.ErrEmpty, "member updates")
	}

	if err := k.validateMembers(msg.MemberUpdates); err != nil {
		return nil, errorsmod.Wrap(err, "members")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	action := func(g *group.GroupInfo) error {
		totalWeight, err := math.NewNonNegativeDecFromString(g.TotalWeight)
		if err != nil {
			return errorsmod.Wrap(err, "group total weight")
		}

		for _, member := range msg.MemberUpdates {
			if err := k.assertMetadataLength(member.Metadata, "group member metadata"); err != nil {
				return err
			}
			groupMember := group.GroupMember{
				GroupId: msg.GroupId,
				Member: &group.Member{
					Address:  member.Address,
					Weight:   member.Weight,
					Metadata: member.Metadata,
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
				return errorsmod.Wrap(err, "get group member")
			}

			newMemberWeight, err := math.NewNonNegativeDecFromString(groupMember.Member.Weight)
			if err != nil {
				return err
			}

			// Handle delete for members with zero weight.
			if newMemberWeight.IsZero() {
				// We can't delete a group member that doesn't already exist.
				if !found {
					return errorsmod.Wrap(sdkerrors.ErrNotFound, "unknown member")
				}

				previousMemberWeight, err := math.NewPositiveDecFromString(prevGroupMember.Member.Weight)
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
					return errorsmod.Wrap(err, "delete member")
				}
				continue
			}
			// If group member already exists, handle update
			if found {
				previousMemberWeight, err := math.NewPositiveDecFromString(prevGroupMember.Member.Weight)
				if err != nil {
					return err
				}
				// Subtract previous weight from the group total weight.
				totalWeight, err = math.SubNonNegative(totalWeight, previousMemberWeight)
				if err != nil {
					return err
				}
				// Save updated group member in the groupMemberTable.
				groupMember.Member.AddedAt = prevGroupMember.Member.AddedAt
				if err := k.groupMemberTable.Update(ctx.KVStore(k.key), &groupMember); err != nil {
					return errorsmod.Wrap(err, "add member")
				}
			} else { // else handle create.
				groupMember.Member.AddedAt = ctx.HeaderInfo().Time
				if err := k.groupMemberTable.Create(ctx.KVStore(k.key), &groupMember); err != nil {
					return errorsmod.Wrap(err, "add member")
				}
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

		if err := k.validateDecisionPolicies(ctx, *g); err != nil {
			return err
		}

		return k.groupTable.Update(ctx.KVStore(k.key), g.Id, g)
	}

	if err := k.doUpdateGroup(ctx, msg.GetGroupID(), msg.GetAdmin(), action, "members updated"); err != nil {
		return nil, err
	}

	return &group.MsgUpdateGroupMembersResponse{}, nil
}

func (k Keeper) UpdateGroupAdmin(goCtx context.Context, msg *group.MsgUpdateGroupAdmin) (*group.MsgUpdateGroupAdminResponse, error) {
	if msg.GroupId == 0 {
		return nil, errorsmod.Wrap(errors.ErrEmpty, "group id")
	}

	if strings.EqualFold(msg.Admin, msg.NewAdmin) {
		return nil, errorsmod.Wrap(errors.ErrInvalid, "new and old admin are the same")
	}

	if _, err := k.accKeeper.AddressCodec().StringToBytes(msg.Admin); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "admin address")
	}

	if _, err := k.accKeeper.AddressCodec().StringToBytes(msg.NewAdmin); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "new admin address")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	action := func(g *group.GroupInfo) error {
		g.Admin = msg.NewAdmin
		g.Version++

		return k.groupTable.Update(ctx.KVStore(k.key), g.Id, g)
	}

	if err := k.doUpdateGroup(ctx, msg.GetGroupID(), msg.GetAdmin(), action, "admin updated"); err != nil {
		return nil, err
	}

	return &group.MsgUpdateGroupAdminResponse{}, nil
}

func (k Keeper) UpdateGroupMetadata(goCtx context.Context, msg *group.MsgUpdateGroupMetadata) (*group.MsgUpdateGroupMetadataResponse, error) {
	if msg.GroupId == 0 {
		return nil, errorsmod.Wrap(errors.ErrEmpty, "group id")
	}

	if err := k.assertMetadataLength(msg.Metadata, "group metadata"); err != nil {
		return nil, err
	}

	if _, err := k.accKeeper.AddressCodec().StringToBytes(msg.Admin); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "admin address")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	action := func(g *group.GroupInfo) error {
		g.Metadata = msg.Metadata
		g.Version++
		return k.groupTable.Update(ctx.KVStore(k.key), g.Id, g)
	}

	if err := k.doUpdateGroup(ctx, msg.GetGroupID(), msg.GetAdmin(), action, "metadata updated"); err != nil {
		return nil, err
	}

	return &group.MsgUpdateGroupMetadataResponse{}, nil
}

func (k Keeper) CreateGroupWithPolicy(ctx context.Context, msg *group.MsgCreateGroupWithPolicy) (*group.MsgCreateGroupWithPolicyResponse, error) {
	// NOTE: admin, and group message validation is performed in the CreateGroup method
	groupRes, err := k.CreateGroup(ctx, &group.MsgCreateGroup{
		Admin:    msg.Admin,
		Members:  msg.Members,
		Metadata: msg.GroupMetadata,
	})
	if err != nil {
		return nil, errorsmod.Wrap(err, "group response")
	}
	groupID := groupRes.GroupId

	// NOTE: group policy message validation is performed in the CreateGroupPolicy method
	groupPolicyRes, err := k.CreateGroupPolicy(ctx, &group.MsgCreateGroupPolicy{
		Admin:          msg.Admin,
		GroupId:        groupID,
		Metadata:       msg.GroupPolicyMetadata,
		DecisionPolicy: msg.DecisionPolicy,
	})
	if err != nil {
		return nil, errorsmod.Wrap(err, "group policy response")
	}

	if msg.GroupPolicyAsAdmin {
		updateAdminReq := &group.MsgUpdateGroupAdmin{
			GroupId:  groupID,
			Admin:    msg.Admin,
			NewAdmin: groupPolicyRes.Address,
		}
		_, err = k.UpdateGroupAdmin(ctx, updateAdminReq)
		if err != nil {
			return nil, err
		}

		updatePolicyAddressReq := &group.MsgUpdateGroupPolicyAdmin{
			Admin:              msg.Admin,
			GroupPolicyAddress: groupPolicyRes.Address,
			NewAdmin:           groupPolicyRes.Address,
		}
		_, err = k.UpdateGroupPolicyAdmin(ctx, updatePolicyAddressReq)
		if err != nil {
			return nil, err
		}
	}

	return &group.MsgCreateGroupWithPolicyResponse{GroupId: groupID, GroupPolicyAddress: groupPolicyRes.Address}, nil
}

func (k Keeper) CreateGroupPolicy(goCtx context.Context, msg *group.MsgCreateGroupPolicy) (*group.MsgCreateGroupPolicyResponse, error) {
	if msg.GroupId == 0 {
		return nil, errorsmod.Wrap(errors.ErrEmpty, "group id")
	}

	if err := k.assertMetadataLength(msg.GetMetadata(), "group policy metadata"); err != nil {
		return nil, err
	}

	policy, err := msg.GetDecisionPolicy()
	if err != nil {
		return nil, errorsmod.Wrap(err, "request decision policy")
	}

	if err := policy.ValidateBasic(); err != nil {
		return nil, errorsmod.Wrap(err, "decision policy")
	}

	reqGroupAdmin, err := k.accKeeper.AddressCodec().StringToBytes(msg.GetAdmin())
	if err != nil {
		return nil, errorsmod.Wrap(err, "request admin")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	groupInfo, err := k.getGroupInfo(ctx, msg.GetGroupID())
	if err != nil {
		return nil, err
	}

	groupAdmin, err := k.accKeeper.AddressCodec().StringToBytes(groupInfo.Admin)
	if err != nil {
		return nil, errorsmod.Wrap(err, "group admin")
	}

	// Only current group admin is authorized to create a group policy for this
	if !bytes.Equal(groupAdmin, reqGroupAdmin) {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "not group admin")
	}

	if err := policy.Validate(groupInfo, k.config); err != nil {
		return nil, err
	}

	// Generate account address of group policy.
	var accountAddr sdk.AccAddress
	// loop here in the rare case where a ADR-028-derived address creates a
	// collision with an existing address.
	for {
		nextAccVal := k.groupPolicySeq.NextVal(ctx.KVStore(k.key))
		derivationKey := make([]byte, 8)
		binary.BigEndian.PutUint64(derivationKey, nextAccVal)

		ac, err := authtypes.NewModuleCredential(group.ModuleName, []byte{GroupPolicyTablePrefix}, derivationKey)
		if err != nil {
			return nil, err
		}
		accountAddr = sdk.AccAddress(ac.Address())
		if k.accKeeper.GetAccount(ctx, accountAddr) != nil {
			// handle a rare collision, in which case we just go on to the
			// next sequence value and derive a new address.
			continue
		}

		// group policy accounts are unclaimable base accounts
		account, err := authtypes.NewBaseAccountWithPubKey(ac)
		if err != nil {
			return nil, errorsmod.Wrap(err, "could not create group policy account")
		}

		acc := k.accKeeper.NewAccount(ctx, account)
		k.accKeeper.SetAccount(ctx, acc)

		break
	}

	groupPolicy, err := group.NewGroupPolicyInfo(
		accountAddr,
		msg.GetGroupID(),
		reqGroupAdmin,
		msg.GetMetadata(),
		1,
		policy,
		ctx.HeaderInfo().Time,
	)
	if err != nil {
		return nil, err
	}

	if err := k.groupPolicyTable.Create(ctx.KVStore(k.key), &groupPolicy); err != nil {
		return nil, errorsmod.Wrap(err, "could not create group policy")
	}

	if err := ctx.EventManager().EmitTypedEvent(&group.EventCreateGroupPolicy{Address: accountAddr.String()}); err != nil {
		return nil, err
	}

	return &group.MsgCreateGroupPolicyResponse{Address: accountAddr.String()}, nil
}

func (k Keeper) UpdateGroupPolicyAdmin(goCtx context.Context, msg *group.MsgUpdateGroupPolicyAdmin) (*group.MsgUpdateGroupPolicyAdminResponse, error) {
	if strings.EqualFold(msg.Admin, msg.NewAdmin) {
		return nil, errorsmod.Wrap(errors.ErrInvalid, "new and old admin are same")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	action := func(groupPolicy *group.GroupPolicyInfo) error {
		groupPolicy.Admin = msg.NewAdmin
		groupPolicy.Version++
		return k.groupPolicyTable.Update(ctx.KVStore(k.key), groupPolicy)
	}

	if err := k.doUpdateGroupPolicy(ctx, msg.GroupPolicyAddress, msg.Admin, action, "group policy admin updated"); err != nil {
		return nil, err
	}

	return &group.MsgUpdateGroupPolicyAdminResponse{}, nil
}

func (k Keeper) UpdateGroupPolicyDecisionPolicy(goCtx context.Context, msg *group.MsgUpdateGroupPolicyDecisionPolicy) (*group.MsgUpdateGroupPolicyDecisionPolicyResponse, error) {
	policy, err := msg.GetDecisionPolicy()
	if err != nil {
		return nil, errorsmod.Wrap(err, "decision policy")
	}

	if err := policy.ValidateBasic(); err != nil {
		return nil, errorsmod.Wrap(err, "decision policy")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	action := func(groupPolicy *group.GroupPolicyInfo) error {
		groupInfo, err := k.getGroupInfo(ctx, groupPolicy.GroupId)
		if err != nil {
			return err
		}

		err = policy.Validate(groupInfo, k.config)
		if err != nil {
			return err
		}

		err = groupPolicy.SetDecisionPolicy(policy)
		if err != nil {
			return err
		}

		groupPolicy.Version++
		return k.groupPolicyTable.Update(ctx.KVStore(k.key), groupPolicy)
	}

	if err = k.doUpdateGroupPolicy(ctx, msg.GroupPolicyAddress, msg.Admin, action, "group policy's decision policy updated"); err != nil {
		return nil, err
	}

	return &group.MsgUpdateGroupPolicyDecisionPolicyResponse{}, nil
}

func (k Keeper) UpdateGroupPolicyMetadata(goCtx context.Context, msg *group.MsgUpdateGroupPolicyMetadata) (*group.MsgUpdateGroupPolicyMetadataResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	metadata := msg.GetMetadata()

	action := func(groupPolicy *group.GroupPolicyInfo) error {
		groupPolicy.Metadata = metadata
		groupPolicy.Version++
		return k.groupPolicyTable.Update(ctx.KVStore(k.key), groupPolicy)
	}

	if err := k.assertMetadataLength(metadata, "group policy metadata"); err != nil {
		return nil, err
	}

	err := k.doUpdateGroupPolicy(ctx, msg.GroupPolicyAddress, msg.Admin, action, "group policy metadata updated")
	if err != nil {
		return nil, err
	}

	return &group.MsgUpdateGroupPolicyMetadataResponse{}, nil
}

func (k Keeper) SubmitProposal(goCtx context.Context, msg *group.MsgSubmitProposal) (*group.MsgSubmitProposalResponse, error) {
	if len(msg.Proposers) == 0 {
		return nil, errorsmod.Wrap(errors.ErrEmpty, "proposers")
	}

	if err := k.validateProposers(msg.Proposers); err != nil {
		return nil, err
	}

	groupPolicyAddr, err := k.accKeeper.AddressCodec().StringToBytes(msg.GroupPolicyAddress)
	if err != nil {
		return nil, errorsmod.Wrap(err, "request account address of group policy")
	}

	if err := k.assertMetadataLength(msg.Title, "proposal Title"); err != nil {
		return nil, err
	}

	if err := k.assertSummaryLength(msg.Summary); err != nil {
		return nil, err
	}

	if err := k.assertMetadataLength(msg.Metadata, "metadata"); err != nil {
		return nil, err
	}

	// verify that if present, the metadata title and summary equals the proposal title and summary
	if len(msg.Metadata) != 0 {
		proposalMetadata := govtypes.ProposalMetadata{}
		if err := json.Unmarshal([]byte(msg.Metadata), &proposalMetadata); err == nil {
			if proposalMetadata.Title != msg.Title {
				return nil, fmt.Errorf("metadata title '%s' must equal proposal title '%s'", proposalMetadata.Title, msg.Title)
			}

			if proposalMetadata.Summary != msg.Summary {
				return nil, fmt.Errorf("metadata summary '%s' must equal proposal summary '%s'", proposalMetadata.Summary, msg.Summary)
			}
		}

		// if we can't unmarshal the metadata, this means the client didn't use the recommended metadata format
		// nothing can be done here, and this is still a valid case, so we ignore the error
	}

	msgs, err := msg.GetMsgs()
	if err != nil {
		return nil, errorsmod.Wrap(err, "request msgs")
	}

	if err := validateMsgs(msgs); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	policyAcc, err := k.getGroupPolicyInfo(ctx, msg.GroupPolicyAddress)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "load group policy: %s", msg.GroupPolicyAddress)
	}

	groupInfo, err := k.getGroupInfo(ctx, policyAcc.GroupId)
	if err != nil {
		return nil, errorsmod.Wrap(err, "get group by groupId of group policy")
	}

	// Only members of the group can submit a new proposal.
	for _, proposer := range msg.Proposers {
		if !k.groupMemberTable.Has(ctx.KVStore(k.key), orm.PrimaryKey(&group.GroupMember{GroupId: groupInfo.Id, Member: &group.Member{Address: proposer}})) {
			return nil, errorsmod.Wrapf(errors.ErrUnauthorized, "not in group: %s", proposer)
		}
	}

	// Check that if the messages require signers, they are all equal to the given account address of group policy.
	if err := ensureMsgAuthZ(msgs, groupPolicyAddr, k.cdc); err != nil {
		return nil, err
	}

	policy, err := policyAcc.GetDecisionPolicy()
	if err != nil {
		return nil, errorsmod.Wrap(err, "proposal group policy decision policy")
	}

	// Prevent proposal that cannot succeed.
	if err = policy.Validate(groupInfo, k.config); err != nil {
		return nil, err
	}

	m := &group.Proposal{
		Id:                 k.proposalTable.Sequence().PeekNextVal(ctx.KVStore(k.key)),
		GroupPolicyAddress: msg.GroupPolicyAddress,
		Metadata:           msg.Metadata,
		Proposers:          msg.Proposers,
		SubmitTime:         ctx.HeaderInfo().Time,
		GroupVersion:       groupInfo.Version,
		GroupPolicyVersion: policyAcc.Version,
		Status:             group.PROPOSAL_STATUS_SUBMITTED,
		ExecutorResult:     group.PROPOSAL_EXECUTOR_RESULT_NOT_RUN,
		VotingPeriodEnd:    ctx.HeaderInfo().Time.Add(policy.GetVotingPeriod()), // The voting window begins as soon as the proposal is submitted.
		FinalTallyResult:   group.DefaultTallyResult(),
		Title:              msg.Title,
		Summary:            msg.Summary,
	}

	if err := m.SetMsgs(msgs); err != nil {
		return nil, errorsmod.Wrap(err, "create proposal")
	}

	id, err := k.proposalTable.Create(ctx.KVStore(k.key), m)
	if err != nil {
		return nil, errorsmod.Wrap(err, "create proposal")
	}

	if err := ctx.EventManager().EmitTypedEvent(&group.EventSubmitProposal{ProposalId: id}); err != nil {
		return nil, err
	}

	// Try to execute proposal immediately
	if msg.Exec == group.Exec_EXEC_TRY {
		// Consider proposers as Yes votes
		for _, proposer := range msg.Proposers {
			ctx.GasMeter().ConsumeGas(gasCostPerIteration, "vote on proposal")
			_, err = k.Vote(ctx, &group.MsgVote{
				ProposalId: id,
				Voter:      proposer,
				Option:     group.VOTE_OPTION_YES,
			})
			if err != nil {
				return &group.MsgSubmitProposalResponse{ProposalId: id}, errorsmod.Wrapf(err, "the proposal was created but failed on vote for voter %s", proposer)
			}
		}

		// Then try to execute the proposal
		_, err = k.Exec(ctx, &group.MsgExec{
			ProposalId: id,
			// We consider the first proposer as the MsgExecRequest signer
			// but that could be revisited (eg using the group policy)
			Executor: msg.Proposers[0],
		})
		if err != nil {
			return &group.MsgSubmitProposalResponse{ProposalId: id}, errorsmod.Wrap(err, "the proposal was created but failed on exec")
		}
	}

	return &group.MsgSubmitProposalResponse{ProposalId: id}, nil
}

func (k Keeper) WithdrawProposal(goCtx context.Context, msg *group.MsgWithdrawProposal) (*group.MsgWithdrawProposalResponse, error) {
	if msg.ProposalId == 0 {
		return nil, errorsmod.Wrap(errors.ErrEmpty, "proposal id")
	}

	if _, err := k.accKeeper.AddressCodec().StringToBytes(msg.Address); err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid group policy admin / proposer address: %s", msg.Address)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	proposal, err := k.getProposal(ctx, msg.ProposalId)
	if err != nil {
		return nil, err
	}

	// Ensure the proposal can be withdrawn.
	if proposal.Status != group.PROPOSAL_STATUS_SUBMITTED {
		return nil, errorsmod.Wrapf(errors.ErrInvalid, "cannot withdraw a proposal with the status of %s", proposal.Status.String())
	}

	var policyInfo group.GroupPolicyInfo
	if policyInfo, err = k.getGroupPolicyInfo(ctx, proposal.GroupPolicyAddress); err != nil {
		return nil, errorsmod.Wrap(err, "load group policy")
	}

	// check address is the group policy admin he is in proposers list..
	if msg.Address != policyInfo.Admin && !isProposer(proposal, msg.Address) {
		return nil, errorsmod.Wrapf(errors.ErrUnauthorized, "given address is neither group policy admin nor in proposers: %s", msg.Address)
	}

	proposal.Status = group.PROPOSAL_STATUS_WITHDRAWN
	if err := k.proposalTable.Update(ctx.KVStore(k.key), msg.ProposalId, &proposal); err != nil {
		return nil, err
	}

	if err := ctx.EventManager().EmitTypedEvent(&group.EventWithdrawProposal{ProposalId: msg.ProposalId}); err != nil {
		return nil, err
	}

	return &group.MsgWithdrawProposalResponse{}, nil
}

func (k Keeper) Vote(goCtx context.Context, msg *group.MsgVote) (*group.MsgVoteResponse, error) {
	if msg.ProposalId == 0 {
		return nil, errorsmod.Wrap(errors.ErrEmpty, "proposal id")
	}

	// verify vote options
	if msg.Option == group.VOTE_OPTION_UNSPECIFIED {
		return nil, errorsmod.Wrap(errors.ErrEmpty, "vote option")
	}

	if _, ok := group.VoteOption_name[int32(msg.Option)]; !ok {
		return nil, errorsmod.Wrap(errors.ErrInvalid, "vote option")
	}

	if err := k.assertMetadataLength(msg.Metadata, "metadata"); err != nil {
		return nil, err
	}

	if _, err := k.accKeeper.AddressCodec().StringToBytes(msg.Voter); err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid voter address: %s", msg.Voter)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	proposal, err := k.getProposal(ctx, msg.ProposalId)
	if err != nil {
		return nil, err
	}

	// Ensure that we can still accept votes for this proposal.
	if proposal.Status != group.PROPOSAL_STATUS_SUBMITTED {
		return nil, errorsmod.Wrap(errors.ErrInvalid, "proposal not open for voting")
	}

	if ctx.HeaderInfo().Time.After(proposal.VotingPeriodEnd) {
		return nil, errorsmod.Wrap(errors.ErrExpired, "voting period has ended already")
	}

	policyInfo, err := k.getGroupPolicyInfo(ctx, proposal.GroupPolicyAddress)
	if err != nil {
		return nil, errorsmod.Wrap(err, "load group policy")
	}

	groupInfo, err := k.getGroupInfo(ctx, policyInfo.GroupId)
	if err != nil {
		return nil, err
	}

	// Count and store votes.
	voter := group.GroupMember{GroupId: groupInfo.Id, Member: &group.Member{Address: msg.Voter}}
	if err := k.groupMemberTable.GetOne(ctx.KVStore(k.key), orm.PrimaryKey(&voter), &voter); err != nil {
		return nil, errorsmod.Wrapf(err, "voter address: %s", msg.Voter)
	}
	newVote := group.Vote{
		ProposalId: msg.ProposalId,
		Voter:      msg.Voter,
		Option:     msg.Option,
		Metadata:   msg.Metadata,
		SubmitTime: ctx.HeaderInfo().Time,
	}

	// The ORM will return an error if the vote already exists,
	// making sure than a voter hasn't already voted.
	if err := k.voteTable.Create(ctx.KVStore(k.key), &newVote); err != nil {
		return nil, errorsmod.Wrap(err, "store vote")
	}

	if err := ctx.EventManager().EmitTypedEvent(&group.EventVote{ProposalId: msg.ProposalId}); err != nil {
		return nil, err
	}

	// Try to execute proposal immediately
	if msg.Exec == group.Exec_EXEC_TRY {
		_, err = k.Exec(ctx, &group.MsgExec{ProposalId: msg.ProposalId, Executor: msg.Voter})
		if err != nil {
			return nil, err
		}
	}

	return &group.MsgVoteResponse{}, nil
}

// doTallyAndUpdate performs a tally, and, if the tally result is final, then:
// - updates the proposal's `Status` and `FinalTallyResult` fields,
// - prune all the votes.
func (k Keeper) doTallyAndUpdate(ctx sdk.Context, p *group.Proposal, groupInfo group.GroupInfo, policyInfo group.GroupPolicyInfo) error {
	policy, err := policyInfo.GetDecisionPolicy()
	if err != nil {
		return err
	}

	tallyResult, err := k.Tally(ctx, *p, policyInfo.GroupId)
	if err != nil {
		return err
	}

	result, err := policy.Allow(tallyResult, groupInfo.TotalWeight)
	if err != nil {
		return errorsmod.Wrap(err, "policy allow")
	}

	// If the result was final (i.e. enough votes to pass) or if the voting
	// period ended, then we consider the proposal as final.
	if isFinal := result.Final || ctx.HeaderInfo().Time.After(p.VotingPeriodEnd); isFinal {
		if err := k.pruneVotes(ctx, p.Id); err != nil {
			return err
		}
		p.FinalTallyResult = tallyResult
		if result.Allow {
			p.Status = group.PROPOSAL_STATUS_ACCEPTED
		} else {
			p.Status = group.PROPOSAL_STATUS_REJECTED
		}

	}

	return nil
}

// Exec executes the messages from a proposal.
func (k Keeper) Exec(goCtx context.Context, msg *group.MsgExec) (*group.MsgExecResponse, error) {
	if msg.ProposalId == 0 {
		return nil, errorsmod.Wrap(errors.ErrEmpty, "proposal id")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	proposal, err := k.getProposal(ctx, msg.ProposalId)
	if err != nil {
		return nil, err
	}

	if proposal.Status != group.PROPOSAL_STATUS_SUBMITTED && proposal.Status != group.PROPOSAL_STATUS_ACCEPTED {
		return nil, errorsmod.Wrapf(errors.ErrInvalid, "not possible to exec with proposal status %s", proposal.Status.String())
	}

	policyInfo, err := k.getGroupPolicyInfo(ctx, proposal.GroupPolicyAddress)
	if err != nil {
		return nil, errorsmod.Wrap(err, "load group policy")
	}

	// If proposal is still in SUBMITTED phase, it means that the voting period
	// didn't end yet, and tallying hasn't been done. In this case, we need to
	// tally first.
	if proposal.Status == group.PROPOSAL_STATUS_SUBMITTED {
		groupInfo, err := k.getGroupInfo(ctx, policyInfo.GroupId)
		if err != nil {
			return nil, errorsmod.Wrap(err, "load group")
		}

		if err = k.doTallyAndUpdate(ctx, &proposal, groupInfo, policyInfo); err != nil {
			return nil, err
		}
	}

	// Execute proposal payload.
	var logs string
	if proposal.Status == group.PROPOSAL_STATUS_ACCEPTED && proposal.ExecutorResult != group.PROPOSAL_EXECUTOR_RESULT_SUCCESS {
		// Caching context so that we don't update the store in case of failure.
		cacheCtx, flush := ctx.CacheContext()

		addr, err := k.accKeeper.AddressCodec().StringToBytes(policyInfo.Address)
		if err != nil {
			return nil, err
		}

		decisionPolicy := policyInfo.DecisionPolicy.GetCachedValue().(group.DecisionPolicy)
		if results, err := k.doExecuteMsgs(cacheCtx, k.router, proposal, addr, decisionPolicy); err != nil {
			proposal.ExecutorResult = group.PROPOSAL_EXECUTOR_RESULT_FAILURE
			logs = fmt.Sprintf("proposal execution failed on proposal %d, because of error %s", proposal.Id, err.Error())
			k.Logger(ctx).Info("proposal execution failed", "cause", err, "proposalID", proposal.Id)
		} else {
			proposal.ExecutorResult = group.PROPOSAL_EXECUTOR_RESULT_SUCCESS
			flush()

			for _, res := range results {
				// NOTE: The sdk msg handler creates a new EventManager, so events must be correctly propagated back to the current context
				ctx.EventManager().EmitEvents(res.GetEvents())
			}
		}
	}

	// Update proposal in proposalTable
	// If proposal has successfully run, delete it from state.
	if proposal.ExecutorResult == group.PROPOSAL_EXECUTOR_RESULT_SUCCESS {
		if err := k.pruneProposal(ctx, proposal.Id); err != nil {
			return nil, err
		}

		// Emit event for proposal finalized with its result
		if err := ctx.EventManager().EmitTypedEvent(
			&group.EventProposalPruned{
				ProposalId:  proposal.Id,
				Status:      proposal.Status,
				TallyResult: &proposal.FinalTallyResult,
			}); err != nil {
			return nil, err
		}
	} else {
		store := ctx.KVStore(k.key)
		if err := k.proposalTable.Update(store, proposal.Id, &proposal); err != nil {
			return nil, err
		}
	}

	if err := ctx.EventManager().EmitTypedEvent(&group.EventExec{
		ProposalId: proposal.Id,
		Logs:       logs,
		Result:     proposal.ExecutorResult,
	}); err != nil {
		return nil, err
	}

	return &group.MsgExecResponse{
		Result: proposal.ExecutorResult,
	}, nil
}

// LeaveGroup implements the MsgServer/LeaveGroup method.
func (k Keeper) LeaveGroup(goCtx context.Context, msg *group.MsgLeaveGroup) (*group.MsgLeaveGroupResponse, error) {
	if msg.GroupId == 0 {
		return nil, errorsmod.Wrap(errors.ErrEmpty, "group-id")
	}

	_, err := k.accKeeper.AddressCodec().StringToBytes(msg.Address)
	if err != nil {
		return nil, errorsmod.Wrap(err, "group member")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	groupInfo, err := k.getGroupInfo(ctx, msg.GroupId)
	if err != nil {
		return nil, errorsmod.Wrap(err, "group")
	}

	groupWeight, err := math.NewNonNegativeDecFromString(groupInfo.TotalWeight)
	if err != nil {
		return nil, err
	}

	gm, err := k.getGroupMember(ctx, &group.GroupMember{
		GroupId: msg.GroupId,
		Member:  &group.Member{Address: msg.Address},
	})
	if err != nil {
		return nil, err
	}

	memberWeight, err := math.NewPositiveDecFromString(gm.Member.Weight)
	if err != nil {
		return nil, err
	}

	updatedWeight, err := math.SubNonNegative(groupWeight, memberWeight)
	if err != nil {
		return nil, err
	}

	// delete group member in the groupMemberTable.
	if err := k.groupMemberTable.Delete(ctx.KVStore(k.key), gm); err != nil {
		return nil, errorsmod.Wrap(err, "group member")
	}

	// update group weight
	groupInfo.TotalWeight = updatedWeight.String()
	groupInfo.Version++

	if err := k.validateDecisionPolicies(ctx, groupInfo); err != nil {
		return nil, err
	}

	if err := k.groupTable.Update(ctx.KVStore(k.key), groupInfo.Id, &groupInfo); err != nil {
		return nil, err
	}

	if err := ctx.EventManager().EmitTypedEvent(&group.EventLeaveGroup{
		GroupId: msg.GroupId,
		Address: msg.Address,
	}); err != nil {
		return nil, err
	}

	return &group.MsgLeaveGroupResponse{}, nil
}

func (k Keeper) getGroupMember(ctx sdk.Context, member *group.GroupMember) (*group.GroupMember, error) {
	var groupMember group.GroupMember
	switch err := k.groupMemberTable.GetOne(ctx.KVStore(k.key),
		orm.PrimaryKey(member), &groupMember); {
	case err == nil:
		break
	case sdkerrors.ErrNotFound.Is(err):
		return nil, sdkerrors.ErrNotFound.Wrapf("%s is not part of group %d", member.Member.Address, member.GroupId)
	default:
		return nil, err
	}

	return &groupMember, nil
}

type (
	actionFn            func(m *group.GroupInfo) error
	groupPolicyActionFn func(m *group.GroupPolicyInfo) error
)

// doUpdateGroupPolicy first makes sure that the group policy admin initiated the group policy update,
// before performing the group policy update and emitting an event.
func (k Keeper) doUpdateGroupPolicy(ctx sdk.Context, reqGroupPolicy, reqAdmin string, action groupPolicyActionFn, note string) error {
	groupPolicyAddr, err := k.accKeeper.AddressCodec().StringToBytes(reqGroupPolicy)
	if err != nil {
		return errorsmod.Wrap(err, "group policy address")
	}

	_, err = k.accKeeper.AddressCodec().StringToBytes(reqAdmin)
	if err != nil {
		return errorsmod.Wrap(err, "group policy admin")
	}

	groupPolicyInfo, err := k.getGroupPolicyInfo(ctx, reqGroupPolicy)
	if err != nil {
		return errorsmod.Wrap(err, "load group policy")
	}

	// Only current group policy admin is authorized to update a group policy.
	if reqAdmin != groupPolicyInfo.Admin {
		return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "not group policy admin")
	}

	if err := action(&groupPolicyInfo); err != nil {
		return errorsmod.Wrap(err, note)
	}

	if err = k.abortProposals(ctx, groupPolicyAddr); err != nil {
		return err
	}

	if err = ctx.EventManager().EmitTypedEvent(&group.EventUpdateGroupPolicy{Address: groupPolicyInfo.Address}); err != nil {
		return err
	}

	return nil
}

// doUpdateGroup first makes sure that the group admin initiated the group update,
// before performing the group update and emitting an event.
func (k Keeper) doUpdateGroup(ctx sdk.Context, groupID uint64, reqGroupAdmin string, action actionFn, errNote string) error {
	groupInfo, err := k.getGroupInfo(ctx, groupID)
	if err != nil {
		return err
	}

	if !strings.EqualFold(groupInfo.Admin, reqGroupAdmin) {
		return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "not group admin; got %s, expected %s", reqGroupAdmin, groupInfo.Admin)
	}

	if err := action(&groupInfo); err != nil {
		return errorsmod.Wrap(err, errNote)
	}

	if err := ctx.EventManager().EmitTypedEvent(&group.EventUpdateGroup{GroupId: groupID}); err != nil {
		return err
	}

	return nil
}

// assertMetadataLength returns an error if given metadata length
// is greater than a pre-defined maxMetadataLen.
func (k Keeper) assertMetadataLength(metadata, description string) error {
	if metadata != "" && uint64(len(metadata)) > k.config.MaxMetadataLen {
		return errorsmod.Wrapf(errors.ErrMaxLimit, description)
	}
	return nil
}

// assertSummaryLength returns an error if given summary length
// is greater than a pre-defined 40*MaxMetadataLen.
func (k Keeper) assertSummaryLength(summary string) error {
	if summary != "" && uint64(len(summary)) > 40*k.config.MaxMetadataLen {
		return errorsmod.Wrapf(errors.ErrMaxLimit, "proposal summary is too long")
	}
	return nil
}

// validateDecisionPolicies loops through all decision policies from the group,
// and calls each of their Validate() method.
func (k Keeper) validateDecisionPolicies(ctx sdk.Context, g group.GroupInfo) error {
	it, err := k.groupPolicyByGroupIndex.Get(ctx.KVStore(k.key), g.Id)
	if err != nil {
		return err
	}
	defer it.Close()

	for {
		var groupPolicy group.GroupPolicyInfo
		_, err = it.LoadNext(&groupPolicy)
		if errors.ErrORMIteratorDone.Is(err) {
			break
		}
		if err != nil {
			return err
		}

		err = groupPolicy.DecisionPolicy.GetCachedValue().(group.DecisionPolicy).Validate(g, k.config)
		if err != nil {
			return err
		}
	}

	return nil
}

// validateProposers checks that all proposers addresses are valid.
// It as well verifies that there is no duplicate address.
func (k Keeper) validateProposers(proposers []string) error {
	index := make(map[string]struct{}, len(proposers))
	for _, proposer := range proposers {
		if _, exists := index[proposer]; exists {
			return errorsmod.Wrapf(errors.ErrDuplicate, "address: %s", proposer)
		}

		_, err := k.accKeeper.AddressCodec().StringToBytes(proposer)
		if err != nil {
			return errorsmod.Wrapf(err, "proposer address %s", proposer)
		}

		index[proposer] = struct{}{}
	}

	return nil
}

// validateMembers checks that all members addresses are valid.
// additionally it verifies that there is no duplicate address
// and the member weight is non-negative.
// Note: in state, a member's weight MUST be positive. However, in some Msgs,
// it's possible to set a zero member weight, for example in
// MsgUpdateGroupMembers to denote that we're removing a member.
// It returns an error if any of the above conditions is not met.
func (k Keeper) validateMembers(members []group.MemberRequest) error {
	index := make(map[string]struct{}, len(members))
	for _, member := range members {
		if _, exists := index[member.Address]; exists {
			return errorsmod.Wrapf(errors.ErrDuplicate, "address: %s", member.Address)
		}

		_, err := k.accKeeper.AddressCodec().StringToBytes(member.Address)
		if err != nil {
			return errorsmod.Wrapf(err, "member address %s", member.Address)
		}

		if _, err := math.NewNonNegativeDecFromString(member.Weight); err != nil {
			return errorsmod.Wrap(err, "weight must be non negative")
		}

		index[member.Address] = struct{}{}
	}

	return nil
}

// isProposer checks that an address is a proposer of a given proposal.
func isProposer(proposal group.Proposal, address string) bool {
	for _, proposer := range proposal.Proposers {
		if proposer == address {
			return true
		}
	}

	return false
}

func validateMsgs(msgs []sdk.Msg) error {
	for i, msg := range msgs {
		m, ok := msg.(sdk.HasValidateBasic)
		if !ok {
			continue
		}

		if err := m.ValidateBasic(); err != nil {
			return errorsmod.Wrapf(err, "msg %d", i)
		}
	}

	return nil
}
