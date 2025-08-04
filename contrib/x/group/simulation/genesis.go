package simulation

import (
	"math/rand"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	group2 "github.com/cosmos/cosmos-sdk/contrib/x/group"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

const (
	GroupInfo       = "group-info"
	GroupMembers    = "group-members"
	GroupPolicyInfo = "group-policy-info"
	GroupProposals  = "group-proposals"
	GroupVote       = "group-vote"
)

func checkAccExists(acc sdk.AccAddress, g []*group2.GroupMember, lastIndex int) bool {
	s := acc.String()
	for i := range lastIndex {
		if g[i].Member.Address == s {
			return true
		}
	}
	return false
}

func getGroups(r *rand.Rand, accounts []simtypes.Account) []*group2.GroupInfo {
	groups := make([]*group2.GroupInfo, 3)
	for i := range 3 {
		acc, _ := simtypes.RandomAcc(r, accounts)
		groups[i] = &group2.GroupInfo{
			Id:          uint64(i + 1),
			Admin:       acc.Address.String(),
			Metadata:    simtypes.RandStringOfLength(r, 10),
			Version:     1,
			TotalWeight: "10",
		}
	}
	return groups
}

func getGroupMembers(r *rand.Rand, accounts []simtypes.Account) []*group2.GroupMember {
	groupMembers := make([]*group2.GroupMember, 3)
	for i := range 3 {
		acc, _ := simtypes.RandomAcc(r, accounts)
		for checkAccExists(acc.Address, groupMembers, i) {
			acc, _ = simtypes.RandomAcc(r, accounts)
		}
		groupMembers[i] = &group2.GroupMember{
			GroupId: uint64(i + 1),
			Member: &group2.Member{
				Address:  acc.Address.String(),
				Weight:   "10",
				Metadata: simtypes.RandStringOfLength(r, 10),
			},
		}
	}
	return groupMembers
}

func getGroupPolicies(r *rand.Rand, simState *module.SimulationState) []*group2.GroupPolicyInfo {
	var groupPolicies []*group2.GroupPolicyInfo

	usedAccs := make(map[string]bool)
	for i := 0; i < 3; i++ {
		acc, _ := simtypes.RandomAcc(r, simState.Accounts)

		if usedAccs[acc.Address.String()] {
			if len(usedAccs) != len(simState.Accounts) {
				// Go again if the account is used and there are more to take from
				i--
			}

			continue
		}
		usedAccs[acc.Address.String()] = true

		any, err := codectypes.NewAnyWithValue(group2.NewThresholdDecisionPolicy("10", time.Second, 0))
		if err != nil {
			panic(err)
		}
		groupPolicies = append(groupPolicies, &group2.GroupPolicyInfo{
			GroupId:        uint64(i + 1),
			Admin:          acc.Address.String(),
			Address:        acc.Address.String(),
			Version:        1,
			DecisionPolicy: any,
			Metadata:       simtypes.RandStringOfLength(r, 10),
		})
	}
	return groupPolicies
}

func getProposals(r *rand.Rand, simState *module.SimulationState, groupPolicies []*group2.GroupPolicyInfo) []*group2.Proposal {
	proposals := make([]*group2.Proposal, 3)
	proposers := []string{simState.Accounts[0].Address.String(), simState.Accounts[1].Address.String()}
	for i := range 3 {
		idx := r.Intn(len(groupPolicies))
		groupPolicyAddress := groupPolicies[idx].Address
		to, _ := simtypes.RandomAcc(r, simState.Accounts)

		submittedAt := time.Unix(0, 0)
		timeout := submittedAt.Add(time.Second * 1000).UTC()

		proposal := &group2.Proposal{
			Id:                 uint64(i + 1),
			Proposers:          proposers,
			GroupPolicyAddress: groupPolicyAddress,
			GroupVersion:       uint64(i + 1),
			GroupPolicyVersion: uint64(i + 1),
			Status:             group2.PROPOSAL_STATUS_SUBMITTED,
			FinalTallyResult: group2.TallyResult{
				YesCount:        "1",
				NoCount:         "1",
				AbstainCount:    "1",
				NoWithVetoCount: "0",
			},
			ExecutorResult:  group2.PROPOSAL_EXECUTOR_RESULT_NOT_RUN,
			Metadata:        simtypes.RandStringOfLength(r, 50),
			SubmitTime:      submittedAt,
			VotingPeriodEnd: timeout,
		}
		err := proposal.SetMsgs([]sdk.Msg{&banktypes.MsgSend{
			FromAddress: groupPolicyAddress,
			ToAddress:   to.Address.String(),
			Amount:      sdk.NewCoins(sdk.NewInt64Coin("test", 10)),
		}})
		if err != nil {
			panic(err)
		}

		proposals[i] = proposal
	}

	return proposals
}

func getVotes(r *rand.Rand, simState *module.SimulationState) []*group2.Vote {
	votes := make([]*group2.Vote, 3)

	for i := range 3 {
		votes[i] = &group2.Vote{
			ProposalId: uint64(i + 1),
			Voter:      simState.Accounts[i].Address.String(),
			Option:     getVoteOption(i),
			Metadata:   simtypes.RandStringOfLength(r, 50),
			SubmitTime: time.Unix(0, 0),
		}
	}

	return votes
}

func getVoteOption(index int) group2.VoteOption {
	switch index {
	case 0:
		return group2.VOTE_OPTION_YES
	case 1:
		return group2.VOTE_OPTION_NO
	case 2:
		return group2.VOTE_OPTION_ABSTAIN
	default:
		return group2.VOTE_OPTION_NO_WITH_VETO
	}
}

// RandomizedGenState generates a random GenesisState for the group module.
func RandomizedGenState(simState *module.SimulationState) {
	// The test requires we have at least 3 accounts.
	if len(simState.Accounts) < 3 {
		return
	}

	// groups
	var groups []*group2.GroupInfo
	simState.AppParams.GetOrGenerate(GroupInfo, &groups, simState.Rand, func(r *rand.Rand) { groups = getGroups(r, simState.Accounts) })

	// group members
	var members []*group2.GroupMember
	simState.AppParams.GetOrGenerate(GroupMembers, &members, simState.Rand, func(r *rand.Rand) { members = getGroupMembers(r, simState.Accounts) })

	// group policies
	var groupPolicies []*group2.GroupPolicyInfo
	simState.AppParams.GetOrGenerate(GroupPolicyInfo, &groupPolicies, simState.Rand, func(r *rand.Rand) { groupPolicies = getGroupPolicies(r, simState) })

	// proposals
	var proposals []*group2.Proposal
	simState.AppParams.GetOrGenerate(GroupProposals, &proposals, simState.Rand, func(r *rand.Rand) { proposals = getProposals(r, simState, groupPolicies) })

	// votes
	var votes []*group2.Vote
	simState.AppParams.GetOrGenerate(GroupVote, &votes, simState.Rand, func(r *rand.Rand) { votes = getVotes(r, simState) })

	groupGenesis := group2.GenesisState{
		GroupSeq:       3,
		Groups:         groups,
		GroupMembers:   members,
		GroupPolicySeq: 3,
		GroupPolicies:  groupPolicies,
		ProposalSeq:    3,
		Proposals:      proposals,
		Votes:          votes,
	}

	simState.GenState[group2.ModuleName] = simState.Cdc.MustMarshalJSON(&groupGenesis)
}
