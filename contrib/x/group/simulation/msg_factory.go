package simulation

import (
	"context"
	"slices"
	"strconv"
	"sync/atomic"
	"time"

	group2 "github.com/cosmos/cosmos-sdk/contrib/x/group"
	"github.com/cosmos/cosmos-sdk/contrib/x/group/keeper"
	"github.com/cosmos/cosmos-sdk/testutil/simsx"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func MsgCreateGroupFactory() simsx.SimMsgFactoryFn[*group2.MsgCreateGroup] {
	return func(ctx context.Context, testData *simsx.ChainDataSource, reporter simsx.SimulationReporter) ([]simsx.SimAccount, *group2.MsgCreateGroup) {
		admin := testData.AnyAccount(reporter, simsx.WithSpendableBalance())
		members := genGroupMembersX(testData, reporter)
		msg := &group2.MsgCreateGroup{Admin: admin.AddressBech32, Members: members, Metadata: testData.Rand().StringN(10)}
		return []simsx.SimAccount{admin}, msg
	}
}

func MsgCreateGroupPolicyFactory(k keeper.Keeper, s *SharedState) simsx.SimMsgFactoryFn[*group2.MsgCreateGroupPolicy] {
	return func(ctx context.Context, testData *simsx.ChainDataSource, reporter simsx.SimulationReporter) ([]simsx.SimAccount, *group2.MsgCreateGroupPolicy) {
		groupInfo := randomGroupX(ctx, k, testData, reporter, s)
		groupAdmin := testData.GetAccount(reporter, groupInfo.Admin)
		if reporter.IsSkipped() {
			return nil, nil
		}
		groupID := groupInfo.Id

		r := testData.Rand()
		msg, err := group2.NewMsgCreateGroupPolicy(
			groupAdmin.Address,
			groupID,
			r.StringN(10),
			&group2.ThresholdDecisionPolicy{
				Threshold: strconv.Itoa(r.IntInRange(1, 10)),
				Windows: &group2.DecisionPolicyWindows{
					VotingPeriod: time.Second * time.Duration(30*24*60*60),
				},
			},
		)
		if err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}
		return []simsx.SimAccount{groupAdmin}, msg
	}
}

func MsgCreateGroupWithPolicyFactory() simsx.SimMsgFactoryFn[*group2.MsgCreateGroupWithPolicy] {
	return func(ctx context.Context, testData *simsx.ChainDataSource, reporter simsx.SimulationReporter) ([]simsx.SimAccount, *group2.MsgCreateGroupWithPolicy) {
		admin := testData.AnyAccount(reporter, simsx.WithSpendableBalance())
		members := genGroupMembersX(testData, reporter)
		r := testData.Rand()
		msg := &group2.MsgCreateGroupWithPolicy{
			Admin:               admin.AddressBech32,
			Members:             members,
			GroupMetadata:       r.StringN(10),
			GroupPolicyMetadata: r.StringN(10),
			GroupPolicyAsAdmin:  r.Float32() < 0.5,
		}
		decisionPolicy := &group2.ThresholdDecisionPolicy{
			Threshold: strconv.Itoa(r.IntInRange(1, 10)),
			Windows: &group2.DecisionPolicyWindows{
				VotingPeriod: time.Second * time.Duration(30*24*60*60),
			},
		}
		if err := msg.SetDecisionPolicy(decisionPolicy); err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}
		return []simsx.SimAccount{admin}, msg
	}
}

func MsgWithdrawProposalFactory(k keeper.Keeper, s *SharedState) simsx.SimMsgFactoryFn[*group2.MsgWithdrawProposal] {
	return func(ctx context.Context, testData *simsx.ChainDataSource, reporter simsx.SimulationReporter) ([]simsx.SimAccount, *group2.MsgWithdrawProposal) {
		groupInfo, groupPolicy := randomGroupPolicyX(ctx, testData, reporter, k, s)
		if reporter.IsSkipped() {
			return nil, nil
		}
		policy, err := groupPolicy.GetDecisionPolicy()
		if err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}
		err = policy.Validate(*groupInfo, group2.DefaultConfig())
		if err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}
		proposalsResult, err := k.ProposalsByGroupPolicy(ctx,
			&group2.QueryProposalsByGroupPolicyRequest{Address: groupPolicy.Address},
		)
		if err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}

		now := simsx.BlockTime(ctx)
		proposal := simsx.First(proposalsResult.GetProposals(), func(p *group2.Proposal) bool {
			return p.Status == group2.PROPOSAL_STATUS_SUBMITTED && p.VotingPeriodEnd.After(now)
		})
		if proposal == nil {
			reporter.Skip("no proposal found")
			return nil, nil
		}
		// select a random proposer
		r := testData.Rand()
		proposer := testData.GetAccount(reporter, simsx.OneOf(r, (*proposal).Proposers))

		msg := &group2.MsgWithdrawProposal{
			ProposalId: (*proposal).Id,
			Address:    proposer.AddressBech32,
		}
		return []simsx.SimAccount{proposer}, msg
	}
}

func MsgVoteFactory(k keeper.Keeper, s *SharedState) simsx.SimMsgFactoryFn[*group2.MsgVote] {
	return func(ctx context.Context, testData *simsx.ChainDataSource, reporter simsx.SimulationReporter) ([]simsx.SimAccount, *group2.MsgVote) {
		groupInfo, groupPolicy := randomGroupPolicyX(ctx, testData, reporter, k, s)
		if reporter.IsSkipped() {
			return nil, nil
		}
		proposalsResult, err := k.ProposalsByGroupPolicy(ctx,
			&group2.QueryProposalsByGroupPolicyRequest{Address: groupPolicy.Address},
		)
		if err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}

		now := simsx.BlockTime(ctx)
		proposal := simsx.First(proposalsResult.GetProposals(), func(p *group2.Proposal) bool {
			return p.Status == group2.PROPOSAL_STATUS_SUBMITTED && p.VotingPeriodEnd.After(now)
		})
		if proposal == nil {
			reporter.Skip("no proposal found")
			return nil, nil
		}
		// select a random member
		r := testData.Rand()
		res, err := k.GroupMembers(ctx, &group2.QueryGroupMembersRequest{GroupId: groupInfo.Id})
		if err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}
		if len(res.Members) == 0 {
			reporter.Skip("group has no members")
			return nil, nil
		}
		voter := testData.GetAccount(reporter, simsx.OneOf(r, res.Members).Member.Address)
		vRes, err := k.VotesByProposal(ctx, &group2.QueryVotesByProposalRequest{
			ProposalId: (*proposal).Id,
		})
		if err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}
		if slices.ContainsFunc(vRes.Votes, func(v *group2.Vote) bool { return v.Voter == voter.AddressBech32 }) {
			reporter.Skip("voted already on proposal")
			return nil, nil
		}

		msg := &group2.MsgVote{
			ProposalId: (*proposal).Id,
			Voter:      voter.AddressBech32,
			Option:     group2.VOTE_OPTION_YES,
			Metadata:   r.StringN(10),
		}
		return []simsx.SimAccount{voter}, msg
	}
}

func MsgSubmitProposalFactory(k keeper.Keeper, s *SharedState) simsx.SimMsgFactoryFn[*group2.MsgSubmitProposal] {
	return func(ctx context.Context, testData *simsx.ChainDataSource, reporter simsx.SimulationReporter) ([]simsx.SimAccount, *group2.MsgSubmitProposal) {
		groupInfo, groupPolicy := randomGroupPolicyX(ctx, testData, reporter, k, s)
		if reporter.IsSkipped() {
			return nil, nil
		}
		// Return a no-op if we know the proposal cannot be created
		policy, err := groupPolicy.GetDecisionPolicy()
		if err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}

		if err = policy.Validate(*groupInfo, group2.DefaultConfig()); err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}

		// Pick a random member from the group
		r := testData.Rand()
		res, err := k.GroupMembers(ctx, &group2.QueryGroupMembersRequest{GroupId: groupInfo.Id})
		if err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}
		if len(res.Members) == 0 {
			reporter.Skip("group has no members")
			return nil, nil
		}
		proposer := testData.GetAccount(reporter, simsx.OneOf(r, res.Members).Member.Address)

		msg := &group2.MsgSubmitProposal{
			GroupPolicyAddress: groupPolicy.Address,
			Proposers:          []string{proposer.AddressBech32},
			Metadata:           r.StringN(10),
			Title:              "Test Proposal",
			Summary:            "Summary of the proposal",
		}
		return []simsx.SimAccount{proposer}, msg
	}
}

func MsgExecFactory(k keeper.Keeper, s *SharedState) simsx.SimMsgFactoryFn[*group2.MsgExec] {
	return func(ctx context.Context, testData *simsx.ChainDataSource, reporter simsx.SimulationReporter) ([]simsx.SimAccount, *group2.MsgExec) {
		groupPolicy, policyAdmin := randomGroupPolicyWithAdmin(ctx, testData, reporter, k, s)
		if reporter.IsSkipped() {
			return nil, nil
		}
		proposalsResult, err := k.ProposalsByGroupPolicy(ctx,
			&group2.QueryProposalsByGroupPolicyRequest{Address: groupPolicy.Address},
		)
		if err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}

		proposal := simsx.First(proposalsResult.GetProposals(), func(p *group2.Proposal) bool {
			return p.Status == group2.PROPOSAL_STATUS_ACCEPTED
		})
		if proposal == nil {
			reporter.Skip("no proposal found")
			return nil, nil
		}

		msg := &group2.MsgExec{
			ProposalId: (*proposal).Id,
			Executor:   policyAdmin.AddressBech32,
		}
		return []simsx.SimAccount{policyAdmin}, msg
	}
}

func randomGroupPolicyWithAdmin(ctx context.Context, testData *simsx.ChainDataSource, reporter simsx.SimulationReporter, k keeper.Keeper, s *SharedState) (*group2.GroupPolicyInfo, simsx.SimAccount) {
	for range 5 {
		_, groupPolicy := randomGroupPolicyX(ctx, testData, reporter, k, s)
		if groupPolicy != nil && testData.HasAccount(groupPolicy.Admin) {
			return groupPolicy, testData.GetAccount(reporter, groupPolicy.Admin)
		}
	}
	reporter.Skip("no group policy found with a sims account")
	return nil, simsx.SimAccount{}
}

func MsgUpdateGroupMetadataFactory(k keeper.Keeper, s *SharedState) simsx.SimMsgFactoryFn[*group2.MsgUpdateGroupMetadata] {
	return func(ctx context.Context, testData *simsx.ChainDataSource, reporter simsx.SimulationReporter) ([]simsx.SimAccount, *group2.MsgUpdateGroupMetadata) {
		groupInfo := randomGroupX(ctx, k, testData, reporter, s)
		groupAdmin := testData.GetAccount(reporter, groupInfo.Admin)
		if reporter.IsSkipped() {
			return nil, nil
		}
		msg := &group2.MsgUpdateGroupMetadata{
			GroupId:  groupInfo.Id,
			Admin:    groupAdmin.AddressBech32,
			Metadata: testData.Rand().StringN(10),
		}
		return []simsx.SimAccount{groupAdmin}, msg
	}
}

func MsgUpdateGroupAdminFactory(k keeper.Keeper, s *SharedState) simsx.SimMsgFactoryFn[*group2.MsgUpdateGroupAdmin] {
	return func(ctx context.Context, testData *simsx.ChainDataSource, reporter simsx.SimulationReporter) ([]simsx.SimAccount, *group2.MsgUpdateGroupAdmin) {
		groupInfo := randomGroupX(ctx, k, testData, reporter, s)
		groupAdmin := testData.GetAccount(reporter, groupInfo.Admin)
		if reporter.IsSkipped() {
			return nil, nil
		}
		newAdmin := testData.AnyAccount(reporter, simsx.ExcludeAccounts(groupAdmin))
		msg := &group2.MsgUpdateGroupAdmin{
			GroupId:  groupInfo.Id,
			Admin:    groupAdmin.AddressBech32,
			NewAdmin: newAdmin.AddressBech32,
		}
		return []simsx.SimAccount{groupAdmin}, msg
	}
}

func MsgUpdateGroupMembersFactory(k keeper.Keeper, s *SharedState) simsx.SimMsgFactoryFn[*group2.MsgUpdateGroupMembers] {
	return func(ctx context.Context, testData *simsx.ChainDataSource, reporter simsx.SimulationReporter) ([]simsx.SimAccount, *group2.MsgUpdateGroupMembers) {
		groupInfo := randomGroupX(ctx, k, testData, reporter, s)
		groupAdmin := testData.GetAccount(reporter, groupInfo.Admin)
		if reporter.IsSkipped() {
			return nil, nil
		}
		res, err := k.GroupMembers(ctx, &group2.QueryGroupMembersRequest{GroupId: groupInfo.Id})
		if err != nil {
			reporter.Skip("group members not found")
			return nil, nil
		}
		oldMemberAddrs := simsx.Collect(res.Members, func(a *group2.GroupMember) string { return a.Member.Address })
		members := genGroupMembersX(testData, reporter, simsx.ExcludeAddresses(oldMemberAddrs...))
		if len(res.Members) > 1 {
			// set existing random group member weight to zero to remove from the group
			obsoleteMember := simsx.OneOf(testData.Rand(), res.Members)
			obsoleteMember.Member.Weight = "0"
			members = append(members, group2.MemberToMemberRequest(obsoleteMember.Member))
		}
		msg := &group2.MsgUpdateGroupMembers{
			GroupId:       groupInfo.Id,
			Admin:         groupAdmin.AddressBech32,
			MemberUpdates: members,
		}
		return []simsx.SimAccount{groupAdmin}, msg
	}
}

func MsgUpdateGroupPolicyAdminFactory(k keeper.Keeper, s *SharedState) simsx.SimMsgFactoryFn[*group2.MsgUpdateGroupPolicyAdmin] {
	return func(ctx context.Context, testData *simsx.ChainDataSource, reporter simsx.SimulationReporter) ([]simsx.SimAccount, *group2.MsgUpdateGroupPolicyAdmin) {
		groupPolicy, policyAdmin := randomGroupPolicyWithAdmin(ctx, testData, reporter, k, s)
		if reporter.IsSkipped() {
			return nil, nil
		}
		newAdmin := testData.AnyAccount(reporter, simsx.ExcludeAccounts(policyAdmin))
		msg := &group2.MsgUpdateGroupPolicyAdmin{
			Admin:              policyAdmin.AddressBech32,
			GroupPolicyAddress: groupPolicy.Address,
			NewAdmin:           newAdmin.AddressBech32,
		}
		return []simsx.SimAccount{policyAdmin}, msg
	}
}

func MsgUpdateGroupPolicyDecisionPolicyFactory(k keeper.Keeper, s *SharedState) simsx.SimMsgFactoryFn[*group2.MsgUpdateGroupPolicyDecisionPolicy] {
	return func(ctx context.Context, testData *simsx.ChainDataSource, reporter simsx.SimulationReporter) ([]simsx.SimAccount, *group2.MsgUpdateGroupPolicyDecisionPolicy) {
		groupPolicy, policyAdmin := randomGroupPolicyWithAdmin(ctx, testData, reporter, k, s)
		if reporter.IsSkipped() {
			return nil, nil
		}
		r := testData.Rand()
		policyAddr, err := k.AddressCodec().StringToBytes(groupPolicy.Address)
		if err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}
		msg, err := group2.NewMsgUpdateGroupPolicyDecisionPolicy(policyAdmin.Address, policyAddr, &group2.ThresholdDecisionPolicy{
			Threshold: strconv.Itoa(r.IntInRange(1, 10)),
			Windows: &group2.DecisionPolicyWindows{
				VotingPeriod: time.Second * time.Duration(r.IntInRange(100, 1000)),
			},
		})
		if err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}
		return []simsx.SimAccount{policyAdmin}, msg
	}
}

func MsgUpdateGroupPolicyMetadataFactory(k keeper.Keeper, s *SharedState) simsx.SimMsgFactoryFn[*group2.MsgUpdateGroupPolicyMetadata] {
	return func(ctx context.Context, testData *simsx.ChainDataSource, reporter simsx.SimulationReporter) ([]simsx.SimAccount, *group2.MsgUpdateGroupPolicyMetadata) {
		groupPolicy, policyAdmin := randomGroupPolicyWithAdmin(ctx, testData, reporter, k, s)
		if reporter.IsSkipped() {
			return nil, nil
		}
		msg := &group2.MsgUpdateGroupPolicyMetadata{
			Admin:              policyAdmin.AddressBech32,
			GroupPolicyAddress: groupPolicy.Address,
			Metadata:           testData.Rand().StringN(10),
		}
		return []simsx.SimAccount{policyAdmin}, msg
	}
}

func MsgLeaveGroupFactory(k keeper.Keeper, s *SharedState) simsx.SimMsgFactoryFn[*group2.MsgLeaveGroup] {
	return func(ctx context.Context, testData *simsx.ChainDataSource, reporter simsx.SimulationReporter) ([]simsx.SimAccount, *group2.MsgLeaveGroup) {
		groupInfo := randomGroupX(ctx, k, testData, reporter, s)
		if reporter.IsSkipped() {
			return nil, nil
		}
		res, err := k.GroupMembers(ctx, &group2.QueryGroupMembersRequest{GroupId: groupInfo.Id})
		if err != nil {
			reporter.Skip("group members not found")
			return nil, nil
		}
		if len(res.Members) == 0 {
			reporter.Skip("group has no members")
			return nil, nil
		}
		anyMember := simsx.OneOf(testData.Rand(), res.Members)
		leaver := testData.GetAccount(reporter, anyMember.Member.Address)
		msg := &group2.MsgLeaveGroup{
			GroupId: groupInfo.Id,
			Address: leaver.AddressBech32,
		}
		return []simsx.SimAccount{leaver}, msg
	}
}

func genGroupMembersX(testData *simsx.ChainDataSource, reporter simsx.SimulationReporter, filters ...simsx.SimAccountFilter) []group2.MemberRequest {
	r := testData.Rand()
	membersCount := r.Intn(5) + 1
	members := make([]group2.MemberRequest, membersCount)
	uniqueAccountsFilter := simsx.UniqueAccounts()
	for i := 0; i < membersCount && !reporter.IsSkipped(); i++ {
		m := testData.AnyAccount(reporter, append(filters, uniqueAccountsFilter)...)
		members[i] = group2.MemberRequest{
			Address:  m.AddressBech32,
			Weight:   strconv.Itoa(r.IntInRange(1, 10)),
			Metadata: r.StringN(10),
		}
	}
	return members
}

func randomGroupX(ctx context.Context, k keeper.Keeper, testdata *simsx.ChainDataSource, reporter simsx.SimulationReporter, s *SharedState) *group2.GroupInfo {
	r := testdata.Rand()
	groupID := k.GetGroupSequence(sdk.UnwrapSDKContext(ctx))
	if initialGroupID := s.getMinGroupID(); initialGroupID == unsetGroupID {
		s.setMinGroupID(groupID)
	} else if initialGroupID < groupID {
		groupID = r.Uint64InRange(initialGroupID+1, groupID+1)
	}

	// when groupID is 0, it proves that SimulateMsgCreateGroup has never been called. that is, no group exists in the chain
	if groupID == 0 {
		reporter.Skip("no group exists")
		return nil
	}

	res, err := k.GroupInfo(ctx, &group2.QueryGroupInfoRequest{GroupId: groupID})
	if err != nil {
		reporter.Skip(err.Error())
		return nil
	}
	return res.Info
}

func randomGroupPolicyX(
	ctx context.Context,
	testdata *simsx.ChainDataSource,
	reporter simsx.SimulationReporter,
	k keeper.Keeper,
	s *SharedState,
) (*group2.GroupInfo, *group2.GroupPolicyInfo) {
	for range 5 {
		groupInfo := randomGroupX(ctx, k, testdata, reporter, s)
		if reporter.IsSkipped() {
			return nil, nil
		}
		groupID := groupInfo.Id
		result, err := k.GroupPoliciesByGroup(ctx, &group2.QueryGroupPoliciesByGroupRequest{GroupId: groupID})
		if err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}
		if len(result.GroupPolicies) != 0 {
			return groupInfo, simsx.OneOf(testdata.Rand(), result.GroupPolicies)
		}
	}
	reporter.Skip("no group policies")
	return nil, nil
}

// SharedState shared state between message invocations
type SharedState struct {
	minGroupID atomic.Uint64
}

// NewSharedState constructor
func NewSharedState() *SharedState {
	r := &SharedState{}
	r.setMinGroupID(unsetGroupID)
	return r
}

func (s *SharedState) getMinGroupID() uint64 {
	return s.minGroupID.Load()
}

func (s *SharedState) setMinGroupID(id uint64) {
	s.minGroupID.Store(id)
}
