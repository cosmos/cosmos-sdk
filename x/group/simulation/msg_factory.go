package simulation

import (
	"context"
	"github.com/cosmos/cosmos-sdk/simsx/common"
	"github.com/cosmos/cosmos-sdk/simsx/module"
	"slices"
	"strconv"
	"sync/atomic"
	"time"

	"cosmossdk.io/x/group"
	"cosmossdk.io/x/group/keeper"
)

const unsetGroupID = 100000000000000

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

func MsgCreateGroupFactory() module.SimMsgFactoryFn[*group.MsgCreateGroup] {
	return func(ctx context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter) ([]common.SimAccount, *group.MsgCreateGroup) {
		admin := testData.AnyAccount(reporter, common.WithSpendableBalance())
		members := genGroupMembersX(testData, reporter)
		msg := &group.MsgCreateGroup{Admin: admin.AddressBech32, Members: members, Metadata: testData.Rand().StringN(10)}
		return []common.SimAccount{admin}, msg
	}
}

func MsgCreateGroupPolicyFactory(k keeper.Keeper, s *SharedState) module.SimMsgFactoryFn[*group.MsgCreateGroupPolicy] {
	return func(ctx context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter) ([]common.SimAccount, *group.MsgCreateGroupPolicy) {
		groupInfo := randomGroupX(ctx, k, testData, reporter, s)
		groupAdmin := testData.GetAccount(reporter, groupInfo.Admin)
		if reporter.IsAborted() {
			return nil, nil
		}
		groupID := groupInfo.Id

		r := testData.Rand()
		msg, err := group.NewMsgCreateGroupPolicy(
			groupAdmin.AddressBech32,
			groupID,
			r.StringN(10),
			&group.ThresholdDecisionPolicy{
				Threshold: strconv.Itoa(r.IntInRange(1, 10)),
				Windows: &group.DecisionPolicyWindows{
					VotingPeriod: time.Second * time.Duration(30*24*60*60),
				},
			},
		)
		if err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}
		return []common.SimAccount{groupAdmin}, msg
	}
}

func MsgCreateGroupWithPolicyFactory() module.SimMsgFactoryFn[*group.MsgCreateGroupWithPolicy] {
	return func(ctx context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter) ([]common.SimAccount, *group.MsgCreateGroupWithPolicy) {
		admin := testData.AnyAccount(reporter, common.WithSpendableBalance())
		members := genGroupMembersX(testData, reporter)
		r := testData.Rand()
		msg := &group.MsgCreateGroupWithPolicy{
			Admin:               admin.AddressBech32,
			Members:             members,
			GroupMetadata:       r.StringN(10),
			GroupPolicyMetadata: r.StringN(10),
			GroupPolicyAsAdmin:  r.Float32() < 0.5,
		}
		decisionPolicy := &group.ThresholdDecisionPolicy{
			Threshold: strconv.Itoa(r.IntInRange(1, 10)),
			Windows: &group.DecisionPolicyWindows{
				VotingPeriod: time.Second * time.Duration(30*24*60*60),
			},
		}
		if err := msg.SetDecisionPolicy(decisionPolicy); err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}
		return []common.SimAccount{admin}, msg
	}
}

func MsgWithdrawProposalFactory(k keeper.Keeper, s *SharedState) module.SimMsgFactoryFn[*group.MsgWithdrawProposal] {
	return func(ctx context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter) ([]common.SimAccount, *group.MsgWithdrawProposal) {
		groupInfo, groupPolicy := randomGroupPolicyX(ctx, testData, reporter, k, s)
		if reporter.IsAborted() {
			return nil, nil
		}
		policy, err := groupPolicy.GetDecisionPolicy()
		if err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}
		err = policy.Validate(*groupInfo, group.DefaultConfig())
		if err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}
		proposalsResult, err := k.ProposalsByGroupPolicy(ctx,
			&group.QueryProposalsByGroupPolicyRequest{Address: groupPolicy.Address},
		)
		if err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}

		now := common.BlockTime(ctx)
		proposal := common.First(proposalsResult.GetProposals(), func(p *group.Proposal) bool {
			return p.Status == group.PROPOSAL_STATUS_SUBMITTED && p.VotingPeriodEnd.After(now)
		})
		if proposal == nil {
			reporter.Skip("no proposal found")
			return nil, nil
		}
		// select a random proposer
		r := testData.Rand()
		proposer := testData.GetAccount(reporter, common.OneOf(r, (*proposal).Proposers))

		msg := &group.MsgWithdrawProposal{
			ProposalId: (*proposal).Id,
			Address:    proposer.AddressBech32,
		}
		return []common.SimAccount{proposer}, msg
	}
}

func MsgVoteFactory(k keeper.Keeper, s *SharedState) module.SimMsgFactoryFn[*group.MsgVote] {
	return func(ctx context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter) ([]common.SimAccount, *group.MsgVote) {
		groupInfo, groupPolicy := randomGroupPolicyX(ctx, testData, reporter, k, s)
		if reporter.IsAborted() {
			return nil, nil
		}
		proposalsResult, err := k.ProposalsByGroupPolicy(ctx,
			&group.QueryProposalsByGroupPolicyRequest{Address: groupPolicy.Address},
		)
		if err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}

		now := common.BlockTime(ctx)
		proposal := common.First(proposalsResult.GetProposals(), func(p *group.Proposal) bool {
			return p.Status == group.PROPOSAL_STATUS_SUBMITTED && p.VotingPeriodEnd.After(now)
		})
		if proposal == nil {
			reporter.Skip("no proposal found")
			return nil, nil
		}
		// select a random member
		r := testData.Rand()
		res, err := k.GroupMembers(ctx, &group.QueryGroupMembersRequest{GroupId: groupInfo.Id})
		if err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}
		if len(res.Members) == 0 {
			reporter.Skip("group has no members")
			return nil, nil
		}
		voter := testData.GetAccount(reporter, common.OneOf(r, res.Members).Member.Address)
		vRes, err := k.VotesByProposal(ctx, &group.QueryVotesByProposalRequest{
			ProposalId: (*proposal).Id,
		})
		if err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}
		if slices.ContainsFunc(vRes.Votes, func(v *group.Vote) bool { return v.Voter == voter.AddressBech32 }) {
			reporter.Skip("voted already on proposal")
			return nil, nil
		}

		msg := &group.MsgVote{
			ProposalId: (*proposal).Id,
			Voter:      voter.AddressBech32,
			Option:     group.VOTE_OPTION_YES,
			Metadata:   r.StringN(10),
		}
		return []common.SimAccount{voter}, msg
	}
}

func MsgSubmitProposalFactory(k keeper.Keeper, s *SharedState) module.SimMsgFactoryFn[*group.MsgSubmitProposal] {
	return func(ctx context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter) ([]common.SimAccount, *group.MsgSubmitProposal) {
		groupInfo, groupPolicy := randomGroupPolicyX(ctx, testData, reporter, k, s)
		if reporter.IsAborted() {
			return nil, nil
		}
		// Return a no-op if we know the proposal cannot be created
		policy, err := groupPolicy.GetDecisionPolicy()
		if err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}

		if err = policy.Validate(*groupInfo, group.DefaultConfig()); err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}

		// Pick a random member from the group
		r := testData.Rand()
		res, err := k.GroupMembers(ctx, &group.QueryGroupMembersRequest{GroupId: groupInfo.Id})
		if err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}
		if len(res.Members) == 0 {
			reporter.Skip("group has no members")
			return nil, nil
		}
		proposer := testData.GetAccount(reporter, common.OneOf(r, res.Members).Member.Address)

		msg := &group.MsgSubmitProposal{
			GroupPolicyAddress: groupPolicy.Address,
			Proposers:          []string{proposer.AddressBech32},
			Metadata:           r.StringN(10),
			Title:              "Test Proposal",
			Summary:            "Summary of the proposal",
		}
		return []common.SimAccount{proposer}, msg
	}
}

func MsgExecFactory(k keeper.Keeper, s *SharedState) module.SimMsgFactoryFn[*group.MsgExec] {
	return func(ctx context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter) ([]common.SimAccount, *group.MsgExec) {
		groupPolicy, policyAdmin := randomGroupPolicyWithAdmin(ctx, testData, reporter, k, s)
		if reporter.IsAborted() {
			return nil, nil
		}
		proposalsResult, err := k.ProposalsByGroupPolicy(ctx,
			&group.QueryProposalsByGroupPolicyRequest{Address: groupPolicy.Address},
		)
		if err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}

		proposal := common.First(proposalsResult.GetProposals(), func(p *group.Proposal) bool {
			return p.Status == group.PROPOSAL_STATUS_ACCEPTED
		})
		if proposal == nil {
			reporter.Skip("no proposal found")
			return nil, nil
		}

		msg := &group.MsgExec{
			ProposalId: (*proposal).Id,
			Executor:   policyAdmin.AddressBech32,
		}
		return []common.SimAccount{policyAdmin}, msg
	}
}

func randomGroupPolicyWithAdmin(ctx context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter, k keeper.Keeper, s *SharedState) (*group.GroupPolicyInfo, common.SimAccount) {
	for i := 0; i < 5; i++ {
		_, groupPolicy := randomGroupPolicyX(ctx, testData, reporter, k, s)
		if groupPolicy != nil && testData.HasAccount(groupPolicy.Admin) {
			return groupPolicy, testData.GetAccount(reporter, groupPolicy.Admin)
		}
	}
	reporter.Skip("no group policy found with a sims account")
	return nil, common.SimAccount{}
}

func MsgUpdateGroupMetadataFactory(k keeper.Keeper, s *SharedState) module.SimMsgFactoryFn[*group.MsgUpdateGroupMetadata] {
	return func(ctx context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter) ([]common.SimAccount, *group.MsgUpdateGroupMetadata) {
		groupInfo := randomGroupX(ctx, k, testData, reporter, s)
		groupAdmin := testData.GetAccount(reporter, groupInfo.Admin)
		if reporter.IsAborted() {
			return nil, nil
		}
		msg := &group.MsgUpdateGroupMetadata{
			GroupId:  groupInfo.Id,
			Admin:    groupAdmin.AddressBech32,
			Metadata: testData.Rand().StringN(10),
		}
		return []common.SimAccount{groupAdmin}, msg
	}
}

func MsgUpdateGroupAdminFactory(k keeper.Keeper, s *SharedState) module.SimMsgFactoryFn[*group.MsgUpdateGroupAdmin] {
	return func(ctx context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter) ([]common.SimAccount, *group.MsgUpdateGroupAdmin) {
		groupInfo := randomGroupX(ctx, k, testData, reporter, s)
		groupAdmin := testData.GetAccount(reporter, groupInfo.Admin)
		if reporter.IsAborted() {
			return nil, nil
		}
		newAdmin := testData.AnyAccount(reporter, common.ExcludeAccounts(groupAdmin))
		msg := &group.MsgUpdateGroupAdmin{
			GroupId:  groupInfo.Id,
			Admin:    groupAdmin.AddressBech32,
			NewAdmin: newAdmin.AddressBech32,
		}
		return []common.SimAccount{groupAdmin}, msg
	}
}

func MsgUpdateGroupMembersFactory(k keeper.Keeper, s *SharedState) module.SimMsgFactoryFn[*group.MsgUpdateGroupMembers] {
	return func(ctx context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter) ([]common.SimAccount, *group.MsgUpdateGroupMembers) {
		groupInfo := randomGroupX(ctx, k, testData, reporter, s)
		groupAdmin := testData.GetAccount(reporter, groupInfo.Admin)
		if reporter.IsAborted() {
			return nil, nil
		}
		res, err := k.GroupMembers(ctx, &group.QueryGroupMembersRequest{GroupId: groupInfo.Id})
		if err != nil {
			reporter.Skip("group members not found")
			return nil, nil
		}
		oldMemberAddrs := common.Collect(res.Members, func(a *group.GroupMember) string { return a.Member.Address })
		members := genGroupMembersX(testData, reporter, common.ExcludeAddresses(oldMemberAddrs...))
		if len(res.Members) != 0 {
			// set existing random group member weight to zero to remove from the group
			obsoleteMember := common.OneOf(testData.Rand(), res.Members)
			obsoleteMember.Member.Weight = "0"
			members = append(members, group.MemberToMemberRequest(obsoleteMember.Member))
		}
		msg := &group.MsgUpdateGroupMembers{
			GroupId:       groupInfo.Id,
			Admin:         groupAdmin.AddressBech32,
			MemberUpdates: members,
		}
		return []common.SimAccount{groupAdmin}, msg
	}
}

func MsgUpdateGroupPolicyAdminFactory(k keeper.Keeper, s *SharedState) module.SimMsgFactoryFn[*group.MsgUpdateGroupPolicyAdmin] {
	return func(ctx context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter) ([]common.SimAccount, *group.MsgUpdateGroupPolicyAdmin) {
		groupPolicy, policyAdmin := randomGroupPolicyWithAdmin(ctx, testData, reporter, k, s)
		if reporter.IsAborted() {
			return nil, nil
		}
		newAdmin := testData.AnyAccount(reporter, common.ExcludeAccounts(policyAdmin))
		msg := &group.MsgUpdateGroupPolicyAdmin{
			Admin:              policyAdmin.AddressBech32,
			GroupPolicyAddress: groupPolicy.Address,
			NewAdmin:           newAdmin.AddressBech32,
		}
		return []common.SimAccount{policyAdmin}, msg
	}
}

func MsgUpdateGroupPolicyDecisionPolicyFactory(k keeper.Keeper, s *SharedState) module.SimMsgFactoryFn[*group.MsgUpdateGroupPolicyDecisionPolicy] {
	return func(ctx context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter) ([]common.SimAccount, *group.MsgUpdateGroupPolicyDecisionPolicy) {
		groupPolicy, policyAdmin := randomGroupPolicyWithAdmin(ctx, testData, reporter, k, s)
		if reporter.IsAborted() {
			return nil, nil
		}
		r := testData.Rand()
		msg, err := group.NewMsgUpdateGroupPolicyDecisionPolicy(policyAdmin.AddressBech32, groupPolicy.Address, &group.ThresholdDecisionPolicy{
			Threshold: strconv.Itoa(r.IntInRange(1, 10)),
			Windows: &group.DecisionPolicyWindows{
				VotingPeriod: time.Second * time.Duration(r.IntInRange(100, 1000)),
			},
		})
		if err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}
		return []common.SimAccount{policyAdmin}, msg
	}
}

func MsgUpdateGroupPolicyMetadataFactory(k keeper.Keeper, s *SharedState) module.SimMsgFactoryFn[*group.MsgUpdateGroupPolicyMetadata] {
	return func(ctx context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter) ([]common.SimAccount, *group.MsgUpdateGroupPolicyMetadata) {
		groupPolicy, policyAdmin := randomGroupPolicyWithAdmin(ctx, testData, reporter, k, s)
		if reporter.IsAborted() {
			return nil, nil
		}
		msg := &group.MsgUpdateGroupPolicyMetadata{
			Admin:              policyAdmin.AddressBech32,
			GroupPolicyAddress: groupPolicy.Address,
			Metadata:           testData.Rand().StringN(10),
		}
		return []common.SimAccount{policyAdmin}, msg
	}
}

func MsgLeaveGroupFactory(k keeper.Keeper, s *SharedState) module.SimMsgFactoryFn[*group.MsgLeaveGroup] {
	return func(ctx context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter) ([]common.SimAccount, *group.MsgLeaveGroup) {
		groupInfo := randomGroupX(ctx, k, testData, reporter, s)
		if reporter.IsAborted() {
			return nil, nil
		}
		res, err := k.GroupMembers(ctx, &group.QueryGroupMembersRequest{GroupId: groupInfo.Id})
		if err != nil {
			reporter.Skip("group members not found")
			return nil, nil
		}
		if len(res.Members) == 0 {
			reporter.Skip("group has no members")
			return nil, nil
		}
		anyMember := common.OneOf(testData.Rand(), res.Members)
		leaver := testData.GetAccount(reporter, anyMember.Member.Address)
		msg := &group.MsgLeaveGroup{
			GroupId: groupInfo.Id,
			Address: leaver.AddressBech32,
		}
		return []common.SimAccount{leaver}, msg
	}
}

func genGroupMembersX(testData *common.ChainDataSource, reporter common.SimulationReporter, filters ...common.SimAccountFilter) []group.MemberRequest {
	r := testData.Rand()
	membersCount := r.Intn(5) + 1
	members := make([]group.MemberRequest, membersCount)
	uniqueAccountsFilter := common.UniqueAccounts()
	for i := 0; i < membersCount && !reporter.IsAborted(); i++ {
		m := testData.AnyAccount(reporter, append(filters, uniqueAccountsFilter)...)
		members[i] = group.MemberRequest{
			Address:  m.AddressBech32,
			Weight:   strconv.Itoa(r.IntInRange(1, 10)),
			Metadata: r.StringN(10),
		}
	}
	return members
}

func randomGroupX(ctx context.Context, k keeper.Keeper, testdata *common.ChainDataSource, reporter common.SimulationReporter, s *SharedState) *group.GroupInfo {
	r := testdata.Rand()
	groupID := k.GetGroupSequence(ctx)
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

	res, err := k.GroupInfo(ctx, &group.QueryGroupInfoRequest{GroupId: groupID})
	if err != nil {
		reporter.Skip(err.Error())
		return nil
	}
	return res.Info
}

func randomGroupPolicyX(
	ctx context.Context,
	testdata *common.ChainDataSource,
	reporter common.SimulationReporter,
	k keeper.Keeper,
	s *SharedState,
) (*group.GroupInfo, *group.GroupPolicyInfo) {
	for i := 0; i < 5; i++ {
		groupInfo := randomGroupX(ctx, k, testdata, reporter, s)
		if reporter.IsAborted() {
			return nil, nil
		}
		groupID := groupInfo.Id
		result, err := k.GroupPoliciesByGroup(ctx, &group.QueryGroupPoliciesByGroupRequest{GroupId: groupID})
		if err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}
		if len(result.GroupPolicies) != 0 {
			return groupInfo, common.OneOf(testdata.Rand(), result.GroupPolicies)
		}
	}
	reporter.Skip("no group policies")
	return nil, nil
}
