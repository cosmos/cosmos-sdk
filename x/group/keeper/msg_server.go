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

	if err := assertMetadataLength(metadata, "group metadata"); err != nil {
		return nil, err
	}

	totalWeight := math.NewDecFromInt64(0)
	for i := range members.Members {
		m := members.Members[i]
		if err := assertMetadataLength(m.Metadata, "member metadata"); err != nil {
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
		GroupId:     k.groupTable.Sequence().PeekNextVal(ctx.KVStore(k.key)),
		Admin:       admin,
		Metadata:    metadata,
		Version:     1,
		TotalWeight: totalWeight.String(),
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
			if err := assertMetadataLength(req.MemberUpdates[i].Metadata, "group member metadata"); err != nil {
				return err
			}
			groupMember := group.GroupMember{GroupId: req.GroupId,
				Member: &group.Member{
					Address:  req.MemberUpdates[i].Address,
					Weight:   req.MemberUpdates[i].Weight,
					Metadata: req.MemberUpdates[i].Metadata,
				},
			}

			// Checking if the group member is already part of the
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
		return k.groupTable.Update(ctx.KVStore(k.key), g.GroupId, g)
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

		return k.groupTable.Update(ctx.KVStore(k.key), g.GroupId, g)
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
		return k.groupTable.Update(ctx.KVStore(k.key), g.GroupId, g)
	}

	if err := assertMetadataLength(req.Metadata, "group metadata"); err != nil {
		return nil, err
	}

	err := k.doUpdateGroup(ctx, req, action, "metadata updated")
	if err != nil {
		return nil, err
	}

	return &group.MsgUpdateGroupMetadataResponse{}, nil
}

func (k Keeper) CreateGroupAccount(goCtx context.Context, req *group.MsgCreateGroupAccount) (*group.MsgCreateGroupAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	admin, err := sdk.AccAddressFromBech32(req.GetAdmin())
	if err != nil {
		return nil, sdkerrors.Wrap(err, "request admin")
	}
	policy := req.GetDecisionPolicy()
	groupID := req.GetGroupID()
	metadata := req.GetMetadata()

	if err := assertMetadataLength(metadata, "group account metadata"); err != nil {
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
	// Only current group admin is authorized to create a group account for this
	if !groupAdmin.Equals(admin) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "not group admin")
	}

	// Generate group account address.
	var accountAddr sdk.AccAddress
	// loop here in the rare case of a collision
	for {
		nextAccVal := k.groupAccountSeq.NextVal(ctx.KVStore(k.key))
		var buf = make([]byte, 8)
		binary.BigEndian.PutUint64(buf, nextAccVal)

		parentAcc := address.Module(group.ModuleName, []byte{GroupAccountTablePrefix})
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

	groupAccount, err := group.NewGroupAccountInfo(
		accountAddr,
		groupID,
		admin,
		metadata,
		1,
		policy,
	)
	if err != nil {
		return nil, err
	}

	if err := k.groupAccountTable.Create(ctx.KVStore(k.key), &groupAccount); err != nil {
		return nil, sdkerrors.Wrap(err, "could not create group account")
	}

	err = ctx.EventManager().EmitTypedEvent(&group.EventCreateGroupAccount{Address: accountAddr.String()})
	if err != nil {
		return nil, err
	}

	return &group.MsgCreateGroupAccountResponse{Address: accountAddr.String()}, nil
}

func (k Keeper) UpdateGroupAccountAdmin(goCtx context.Context, req *group.MsgUpdateGroupAccountAdmin) (*group.MsgUpdateGroupAccountAdminResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	action := func(groupAccount *group.GroupAccountInfo) error {
		groupAccount.Admin = req.NewAdmin
		groupAccount.Version++
		return k.groupAccountTable.Update(ctx.KVStore(k.key), groupAccount)
	}

	err := k.doUpdateGroupAccount(ctx, req.Address, req.Admin, action, "group account admin updated")
	if err != nil {
		return nil, err
	}

	return &group.MsgUpdateGroupAccountAdminResponse{}, nil
}

func (k Keeper) UpdateGroupAccountDecisionPolicy(goCtx context.Context, req *group.MsgUpdateGroupAccountDecisionPolicy) (*group.MsgUpdateGroupAccountDecisionPolicyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	policy := req.GetDecisionPolicy()

	action := func(groupAccount *group.GroupAccountInfo) error {
		err := groupAccount.SetDecisionPolicy(policy)
		if err != nil {
			return err
		}

		groupAccount.Version++
		return k.groupAccountTable.Update(ctx.KVStore(k.key), groupAccount)
	}

	err := k.doUpdateGroupAccount(ctx, req.Address, req.Admin, action, "group account decision policy updated")
	if err != nil {
		return nil, err
	}

	return &group.MsgUpdateGroupAccountDecisionPolicyResponse{}, nil
}

func (k Keeper) UpdateGroupAccountMetadata(goCtx context.Context, req *group.MsgUpdateGroupAccountMetadata) (*group.MsgUpdateGroupAccountMetadataResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	metadata := req.GetMetadata()

	action := func(groupAccount *group.GroupAccountInfo) error {
		groupAccount.Metadata = metadata
		groupAccount.Version++
		return k.groupAccountTable.Update(ctx.KVStore(k.key), groupAccount)
	}

	if err := assertMetadataLength(metadata, "group account metadata"); err != nil {
		return nil, err
	}

	err := k.doUpdateGroupAccount(ctx, req.Address, req.Admin, action, "group account metadata updated")
	if err != nil {
		return nil, err
	}

	return &group.MsgUpdateGroupAccountMetadataResponse{}, nil
}

func (k Keeper) CreateProposal(goCtx context.Context, req *group.MsgCreateProposal) (*group.MsgCreateProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	accountAddress, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "request group account")
	}
	metadata := req.Metadata
	proposers := req.Proposers
	msgs := req.GetMsgs()

	if err := assertMetadataLength(metadata, "metadata"); err != nil {
		return nil, err
	}

	account, err := k.getGroupAccountInfo(ctx, req.Address)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "load group account")
	}

	g, err := k.getGroupInfo(ctx, account.GroupId)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "get group by account")
	}

	// Only members of the group can submit a new proposal.
	for i := range proposers {
		if !k.groupMemberTable.Has(ctx.KVStore(k.key), orm.PrimaryKey(&group.GroupMember{GroupId: g.GroupId, Member: &group.Member{Address: proposers[i]}})) {
			return nil, sdkerrors.Wrapf(errors.ErrUnauthorized, "not in group: %s", proposers[i])
		}
	}

	// Check that if the messages require signers, they are all equal to the given group account.
	if err := ensureMsgAuthZ(msgs, accountAddress); err != nil {
		return nil, err
	}

	policy := account.GetDecisionPolicy()
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
		ProposalId:          k.proposalTable.Sequence().PeekNextVal(ctx.KVStore(k.key)),
		Address:             req.Address,
		Metadata:            metadata,
		Proposers:           proposers,
		SubmittedAt:         ctx.BlockTime(),
		GroupVersion:        g.Version,
		GroupAccountVersion: account.Version,
		Result:              group.ProposalResultUnfinalized,
		Status:              group.ProposalStatusSubmitted,
		ExecutorResult:      group.ProposalExecutorResultNotRun,
		Timeout:             ctx.BlockTime().Add(window),
		VoteState: group.Tally{
			YesCount:     "0",
			NoCount:      "0",
			AbstainCount: "0",
			VetoCount:    "0",
		},
	}
	if err := m.SetMsgs(msgs); err != nil {
		return nil, sdkerrors.Wrap(err, "create proposal")
	}

	id, err := k.proposalTable.Create(ctx.KVStore(k.key), m)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "create proposal")
	}

	err = ctx.EventManager().EmitTypedEvent(&group.EventCreateProposal{ProposalId: id})
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
				Choice:     group.Choice_CHOICE_YES,
			})
			if err != nil {
				return &group.MsgCreateProposalResponse{ProposalId: id}, sdkerrors.Wrap(err, "The proposal was created but failed on vote")
			}
		}
		// Then try to execute the proposal
		_, err = k.Exec(sdk.WrapSDKContext(ctx), &group.MsgExec{
			ProposalId: id,
			// We consider the first proposer as the MsgExecRequest signer
			// but that could be revisited (eg using the group account)
			Signer: proposers[0],
		})
		if err != nil {
			return &group.MsgCreateProposalResponse{ProposalId: id}, sdkerrors.Wrap(err, "The proposal was created but failed on exec")
		}
	}

	return &group.MsgCreateProposalResponse{ProposalId: id}, nil
}

func (k Keeper) Vote(goCtx context.Context, req *group.MsgVote) (*group.MsgVoteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	id := req.ProposalId
	choice := req.Choice
	metadata := req.Metadata

	if err := assertMetadataLength(metadata, "metadata"); err != nil {
		return nil, err
	}

	proposal, err := k.getProposal(ctx, id)
	if err != nil {
		return nil, err
	}
	// Ensure that we can still accept votes for this proposal.
	if proposal.Status != group.ProposalStatusSubmitted {
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

	var accountInfo group.GroupAccountInfo

	// Ensure that group account hasn't been modified since the proposal submission.
	if accountInfo, err = k.getGroupAccountInfo(ctx, proposal.Address); err != nil {
		return nil, sdkerrors.Wrap(err, "load group account")
	}
	if proposal.GroupAccountVersion != accountInfo.Version {
		return nil, sdkerrors.Wrap(errors.ErrModified, "group account was modified")
	}

	// Ensure that group hasn't been modified since the proposal submission.
	electorate, err := k.getGroupInfo(ctx, accountInfo.GroupId)
	if err != nil {
		return nil, err
	}
	if electorate.Version != proposal.GroupVersion {
		return nil, sdkerrors.Wrap(errors.ErrModified, "group was modified")
	}

	// Count and store votes.
	voterAddr := req.Voter
	voter := group.GroupMember{GroupId: electorate.GroupId, Member: &group.Member{Address: voterAddr}}
	if err := k.groupMemberTable.GetOne(ctx.KVStore(k.key), orm.PrimaryKey(&voter), &voter); err != nil {
		return nil, sdkerrors.Wrapf(err, "address: %s", voterAddr)
	}
	newVote := group.Vote{
		ProposalId:  id,
		Voter:       voterAddr,
		Choice:      choice,
		Metadata:    metadata,
		SubmittedAt: ctx.BlockTime(),
	}
	if err := proposal.VoteState.Add(newVote, voter.Member.Weight); err != nil {
		return nil, sdkerrors.Wrap(err, "add new vote")
	}

	// The ORM will return an error if the vote already exists,
	// making sure than a voter hasn't already voted.
	if err := k.voteTable.Create(ctx.KVStore(k.key), &newVote); err != nil {
		return nil, sdkerrors.Wrap(err, "store vote")
	}

	// Run tally with new votes to close early.
	if err := doTally(ctx, &proposal, electorate, accountInfo); err != nil {
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

// doTally updates the proposal status and tally if necessary based on the group account's decision policy.
func doTally(ctx sdk.Context, p *group.Proposal, electorate group.GroupInfo, accountInfo group.GroupAccountInfo) error {
	policy := accountInfo.GetDecisionPolicy()
	pSubmittedAt, err := gogotypes.TimestampProto(p.SubmittedAt)
	if err != nil {
		return err
	}
	submittedAt, err := gogotypes.TimestampFromProto(pSubmittedAt)
	if err != nil {
		return err
	}
	switch result, err := policy.Allow(p.VoteState, electorate.TotalWeight, ctx.BlockTime().Sub(submittedAt)); {
	case err != nil:
		return sdkerrors.Wrap(err, "policy execution")
	case result.Allow && result.Final:
		p.Result = group.ProposalResultAccepted
		p.Status = group.ProposalStatusClosed
	case !result.Allow && result.Final:
		p.Result = group.ProposalResultRejected
		p.Status = group.ProposalStatusClosed
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

	if proposal.Status != group.ProposalStatusSubmitted && proposal.Status != group.ProposalStatusClosed {
		return nil, sdkerrors.Wrapf(errors.ErrInvalid, "not possible with proposal status %s", proposal.Status.String())
	}

	var accountInfo group.GroupAccountInfo
	if accountInfo, err = k.getGroupAccountInfo(ctx, proposal.Address); err != nil {
		return nil, sdkerrors.Wrap(err, "load group account")
	}

	storeUpdates := func() (*group.MsgExecResponse, error) {
		if err := k.proposalTable.Update(ctx.KVStore(k.key), id, &proposal); err != nil {
			return nil, err
		}
		return &group.MsgExecResponse{}, nil
	}

	if proposal.Status == group.ProposalStatusSubmitted {
		// Ensure that group account hasn't been modified before tally.
		if proposal.GroupAccountVersion != accountInfo.Version {
			proposal.Result = group.ProposalResultUnfinalized
			proposal.Status = group.ProposalStatusAborted
			return storeUpdates()
		}

		electorate, err := k.getGroupInfo(ctx, accountInfo.GroupId)
		if err != nil {
			return nil, sdkerrors.Wrap(err, "load group")
		}

		// Ensure that group hasn't been modified before tally.
		if electorate.Version != proposal.GroupVersion {
			proposal.Result = group.ProposalResultUnfinalized
			proposal.Status = group.ProposalStatusAborted
			return storeUpdates()
		}
		if err := doTally(ctx, &proposal, electorate, accountInfo); err != nil {
			return nil, err
		}
	}

	// Execute proposal payload.
	if proposal.Status == group.ProposalStatusClosed && proposal.Result == group.ProposalResultAccepted && proposal.ExecutorResult != group.ProposalExecutorResultSuccess {
		logger := ctx.Logger().With("module", fmt.Sprintf("x/%s", group.ModuleName))
		// Caching context so that we don't update the store in case of failure.
		ctx, flush := ctx.CacheContext()

		addr, err := sdk.AccAddressFromBech32(accountInfo.Address)
		if err != nil {
			return nil, err
		}
		_, err = k.doExecuteMsgs(ctx, k.router, proposal, addr)
		if err != nil {
			proposal.ExecutorResult = group.ProposalExecutorResultFailure
			proposalType := reflect.TypeOf(proposal).String()
			logger.Info("proposal execution failed", "cause", err, "type", proposalType, "proposalID", id)
		} else {
			proposal.ExecutorResult = group.ProposalExecutorResultSuccess
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
type groupAccountActionFn func(m *group.GroupAccountInfo) error

// doUpdateGroupAccount first makes sure that the group account admin initiated the group account update,
// before performing the group account update and emitting an event.
func (k Keeper) doUpdateGroupAccount(ctx sdk.Context, groupAccount string, admin string, action groupAccountActionFn, note string) error {
	groupAccountInfo, err := k.getGroupAccountInfo(ctx, groupAccount)
	if err != nil {
		return sdkerrors.Wrap(err, "load group account")
	}

	groupAccountAdmin, err := sdk.AccAddressFromBech32(admin)
	if err != nil {
		return sdkerrors.Wrap(err, "group account admin")
	}

	// Only current group account admin is authorized to update a group account.
	if groupAccountAdmin.String() != groupAccountInfo.Admin {
		return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "not group account admin")
	}

	if err := action(&groupAccountInfo); err != nil {
		return sdkerrors.Wrap(err, note)
	}

	err = ctx.EventManager().EmitTypedEvent(&group.EventUpdateGroupAccount{Address: admin})
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
// is greater than a fixed maxMetadataLength.
func assertMetadataLength(metadata []byte, description string) error {
	if len(metadata) > group.MaxMetadataLength {
		return sdkerrors.Wrap(errors.ErrMaxLimit, description)
	}
	return nil
}
