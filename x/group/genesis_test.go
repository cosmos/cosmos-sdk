package group

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
)

var (
	memberPub  = secp256k1.GenPrivKey().PubKey()
	accPub     = secp256k1.GenPrivKey().PubKey()
	accAddr    = sdk.AccAddress(accPub.Address())
	memberAddr = sdk.AccAddress(memberPub.Address())
)

func TestGenesisStateValidate(t *testing.T) {

	submittedAt := time.Now().UTC()
	timeout := submittedAt.Add(time.Second * 1).UTC()

	groupPolicy := &GroupPolicyInfo{
		Address:  accAddr.String(),
		GroupId:  1,
		Admin:    accAddr.String(),
		Version:  1,
		Metadata: []byte("policy metadata"),
	}
	err := groupPolicy.SetDecisionPolicy(&ThresholdDecisionPolicy{
		Threshold: "1",
		Timeout:   time.Second,
	})
	require.NoError(t, err)

	// create another group policy to set invalid decision policy for testing
	groupPolicy2 := &GroupPolicyInfo{
		Address:  accAddr.String(),
		GroupId:  1,
		Admin:    accAddr.String(),
		Version:  1,
		Metadata: []byte("policy metadata"),
	}
	err = groupPolicy2.SetDecisionPolicy(&ThresholdDecisionPolicy{
		Threshold: "1",
		Timeout:   0,
	})
	require.NoError(t, err)

	proposal := &Proposal{
		ProposalId:         1,
		Address:            accAddr.String(),
		Metadata:           []byte("proposal metadata"),
		GroupVersion:       1,
		GroupPolicyVersion: 1,
		Proposers: []string{
			memberAddr.String(),
		},
		SubmittedAt: submittedAt,
		Status:      ProposalStatusClosed,
		Result:      ProposalResultAccepted,
		VoteState: Tally{
			YesCount:     "1",
			NoCount:      "0",
			AbstainCount: "0",
			VetoCount:    "0",
		},
		Timeout:        timeout,
		ExecutorResult: ProposalExecutorResultSuccess,
	}
	err = proposal.SetMsgs([]sdk.Msg{&banktypes.MsgSend{
		FromAddress: accAddr.String(),
		ToAddress:   memberAddr.String(),
		Amount:      sdk.Coins{sdk.NewInt64Coin("test", 100)},
	}})
	require.NoError(t, err)

	testCases := []struct {
		name         string
		genesisState GenesisState
		expErr       bool
	}{
		{
			"valid genesisState",
			GenesisState{
				GroupSeq:              2,
				Groups:                []*GroupInfo{{GroupId: 1, Admin: accAddr.String(), Metadata: []byte("1"), Version: 1, TotalWeight: "1"}, {GroupId: 2, Admin: accAddr.String(), Metadata: []byte("2"), Version: 2, TotalWeight: "2"}},
				GroupMembers:          []*GroupMember{{GroupId: 1, Member: &Member{Address: memberAddr.String(), Weight: "1", Metadata: []byte("member metadata")}}, {GroupId: 2, Member: &Member{Address: memberAddr.String(), Weight: "2", Metadata: []byte("member metadata")}}},
				GroupPolicyAccountSeq: 1,
				GroupPolicies:         []*GroupPolicyInfo{groupPolicy},
				ProposalSeq:           1,
				Proposals:             []*Proposal{proposal},
				Votes:                 []*Vote{{ProposalId: proposal.ProposalId, Voter: memberAddr.String(), SubmittedAt: submittedAt, Choice: Choice_CHOICE_YES}},
			},
			false,
		},
		{
			"empty genesisState",
			GenesisState{},
			false,
		},
		{
			"empty group id",
			GenesisState{
				Groups: []*GroupInfo{
					{
						GroupId:     0,
						Admin:       accAddr.String(),
						Metadata:    []byte("1"),
						Version:     1,
						TotalWeight: "1",
					},
				},
			},
			true,
		},
		{
			"invalid group admin",
			GenesisState{
				Groups: []*GroupInfo{
					{
						GroupId:     1,
						Admin:       "invalid admin",
						Metadata:    []byte("1"),
						Version:     1,
						TotalWeight: "1",
					},
				},
			},
			true,
		},
		{
			"invalid group version",
			GenesisState{
				Groups: []*GroupInfo{
					{
						GroupId:     1,
						Admin:       accAddr.String(),
						Metadata:    []byte("1"),
						Version:     0,
						TotalWeight: "1",
					},
				},
			},
			true,
		},
		{
			"invalid group TotalWeight",
			GenesisState{
				Groups: []*GroupInfo{
					{
						GroupId:     1,
						Admin:       accAddr.String(),
						Metadata:    []byte("1"),
						Version:     1,
						TotalWeight: "-1",
					},
				},
			},
			true,
		},
		{
			"invalid group policy address",
			GenesisState{
				Groups: []*GroupInfo{
					{
						GroupId:     1,
						Admin:       accAddr.String(),
						Metadata:    []byte("1"),
						Version:     1,
						TotalWeight: "1",
					},
				},
				GroupPolicies: []*GroupPolicyInfo{
					{
						Address:  "invalid address",
						GroupId:  1,
						Admin:    accAddr.String(),
						Version:  1,
						Metadata: []byte("policy metadata"),
					},
				},
			},
			true,
		},
		{
			"invalid group policy admin",
			GenesisState{
				Groups: []*GroupInfo{
					{
						GroupId:     1,
						Admin:       accAddr.String(),
						Metadata:    []byte("1"),
						Version:     1,
						TotalWeight: "1",
					},
				},
				GroupPolicies: []*GroupPolicyInfo{
					{
						Address:  accAddr.String(),
						GroupId:  1,
						Admin:    "invalid admin",
						Version:  1,
						Metadata: []byte("policy metadata"),
					},
				},
			},
			true,
		},
		{
			"invalid group policy's group id",
			GenesisState{
				Groups: []*GroupInfo{
					{
						GroupId:     1,
						Admin:       accAddr.String(),
						Metadata:    []byte("1"),
						Version:     1,
						TotalWeight: "1",
					},
				},
				GroupPolicies: []*GroupPolicyInfo{
					{
						Address:  accAddr.String(),
						GroupId:  0,
						Admin:    accAddr.String(),
						Version:  1,
						Metadata: []byte("policy metadata"),
					},
				},
			},
			true,
		},
		{
			"invalid group policy version",
			GenesisState{
				Groups: []*GroupInfo{
					{
						GroupId:     1,
						Admin:       accAddr.String(),
						Metadata:    []byte("1"),
						Version:     1,
						TotalWeight: "1",
					},
				},
				GroupPolicies: []*GroupPolicyInfo{
					{
						Address:  accAddr.String(),
						GroupId:  1,
						Admin:    accAddr.String(),
						Version:  0,
						Metadata: []byte("policy metadata"),
					},
				},
			},
			true,
		},
		{
			"invalid group policy's decision policy",
			GenesisState{
				Groups: []*GroupInfo{
					{
						GroupId:     1,
						Admin:       accAddr.String(),
						Metadata:    []byte("1"),
						Version:     1,
						TotalWeight: "1",
					},
				},
				GroupPolicies: []*GroupPolicyInfo{
					{
						Address:        accAddr.String(),
						GroupId:        1,
						Admin:          accAddr.String(),
						Version:        1,
						Metadata:       []byte("policy metadata"),
						DecisionPolicy: groupPolicy2.DecisionPolicy,
					},
				},
			},
			true,
		},
		{
			"invalid group member's group id",
			GenesisState{
				Groups: []*GroupInfo{
					{
						GroupId:     1,
						Admin:       accAddr.String(),
						Metadata:    []byte("1"),
						Version:     1,
						TotalWeight: "1",
					},
				},
				GroupMembers: []*GroupMember{
					{
						GroupId: 0,
						Member: &Member{
							Address: memberAddr.String(),
							Weight:  "1", Metadata: []byte("member metadata"),
						},
					},
				},
			},
			true,
		},
		{
			"invalid group member address",
			GenesisState{
				Groups: []*GroupInfo{
					{
						GroupId:     1,
						Admin:       accAddr.String(),
						Metadata:    []byte("1"),
						Version:     1,
						TotalWeight: "1",
					},
				},
				GroupMembers: []*GroupMember{
					{
						GroupId: 1,
						Member: &Member{
							Address: "invalid address",
							Weight:  "1", Metadata: []byte("member metadata"),
						},
					},
				},
			},
			true,
		},
		{
			"invalid group member weight",
			GenesisState{
				Groups: []*GroupInfo{
					{
						GroupId:     1,
						Admin:       accAddr.String(),
						Metadata:    []byte("1"),
						Version:     1,
						TotalWeight: "1",
					},
				},
				GroupMembers: []*GroupMember{
					{
						GroupId: 1,
						Member: &Member{
							Address: memberAddr.String(),
							Weight:  "-1", Metadata: []byte("member metadata"),
						},
					},
				},
			},
			true,
		},
		{
			"invalid proposal id",
			GenesisState{
				Groups: []*GroupInfo{
					{
						GroupId:     1,
						Admin:       accAddr.String(),
						Metadata:    []byte("1"),
						Version:     1,
						TotalWeight: "1",
					},
				},
				GroupPolicies: []*GroupPolicyInfo{
					groupPolicy,
				},
				Proposals: []*Proposal{
					{
						ProposalId:         0,
						Address:            accAddr.String(),
						Metadata:           []byte("proposal metadata"),
						GroupVersion:       1,
						GroupPolicyVersion: 1,
					},
				},
			},
			true,
		},
		{
			"invalid group policy address of proposal",
			GenesisState{
				Groups: []*GroupInfo{
					{
						GroupId:     1,
						Admin:       accAddr.String(),
						Metadata:    []byte("1"),
						Version:     1,
						TotalWeight: "1",
					},
				},
				GroupPolicies: []*GroupPolicyInfo{
					groupPolicy,
				},
				Proposals: []*Proposal{
					{
						ProposalId:         1,
						Address:            "invalid address",
						Metadata:           []byte("proposal metadata"),
						GroupVersion:       1,
						GroupPolicyVersion: 1,
					},
				},
			},
			true,
		},
		{
			"invalid group version of proposal",
			GenesisState{
				Groups: []*GroupInfo{
					{
						GroupId:     1,
						Admin:       accAddr.String(),
						Metadata:    []byte("1"),
						Version:     1,
						TotalWeight: "1",
					},
				},
				GroupPolicies: []*GroupPolicyInfo{
					groupPolicy,
				},
				Proposals: []*Proposal{
					{
						ProposalId:         1,
						Address:            accAddr.String(),
						Metadata:           []byte("proposal metadata"),
						GroupVersion:       0,
						GroupPolicyVersion: 1,
					},
				},
			},
			true,
		},
		{
			"invalid group policy version",
			GenesisState{
				Groups: []*GroupInfo{
					{
						GroupId:     1,
						Admin:       accAddr.String(),
						Metadata:    []byte("1"),
						Version:     1,
						TotalWeight: "1",
					},
				},
				GroupPolicies: []*GroupPolicyInfo{
					groupPolicy,
				},
				Proposals: []*Proposal{
					{
						ProposalId:         1,
						Address:            accAddr.String(),
						Metadata:           []byte("proposal metadata"),
						GroupVersion:       1,
						GroupPolicyVersion: 0,
					},
				},
			},
			true,
		},
		{
			"invalid VoteState with negative YesCount",
			GenesisState{
				Groups: []*GroupInfo{
					{
						GroupId:     1,
						Admin:       accAddr.String(),
						Metadata:    []byte("1"),
						Version:     1,
						TotalWeight: "1",
					},
				},
				GroupPolicies: []*GroupPolicyInfo{
					groupPolicy,
				},
				Proposals: []*Proposal{
					{
						ProposalId:         1,
						Address:            accAddr.String(),
						Metadata:           []byte("proposal metadata"),
						GroupVersion:       1,
						GroupPolicyVersion: 1,
						Proposers: []string{
							memberAddr.String(),
						},
						SubmittedAt: submittedAt,
						Status:      ProposalStatusClosed,
						Result:      ProposalResultAccepted,
						VoteState: Tally{
							YesCount:     "-1",
							NoCount:      "0",
							AbstainCount: "0",
							VetoCount:    "0",
						},
					},
				},
			},
			true,
		},
		{
			"invalid VoteState with negative NoCount",
			GenesisState{
				Groups: []*GroupInfo{
					{
						GroupId:     1,
						Admin:       accAddr.String(),
						Metadata:    []byte("1"),
						Version:     1,
						TotalWeight: "1",
					},
				},
				GroupPolicies: []*GroupPolicyInfo{
					groupPolicy,
				},
				Proposals: []*Proposal{
					{
						ProposalId:         1,
						Address:            accAddr.String(),
						Metadata:           []byte("proposal metadata"),
						GroupVersion:       1,
						GroupPolicyVersion: 1,
						Proposers: []string{
							memberAddr.String(),
						},
						SubmittedAt: submittedAt,
						Status:      ProposalStatusClosed,
						Result:      ProposalResultAccepted,
						VoteState: Tally{
							YesCount:     "0",
							NoCount:      "-1",
							AbstainCount: "0",
							VetoCount:    "0",
						},
					},
				},
			},
			true,
		},
		{
			"invalid VoteState with negative AbstainCount",
			GenesisState{
				Groups: []*GroupInfo{
					{
						GroupId:     1,
						Admin:       accAddr.String(),
						Metadata:    []byte("1"),
						Version:     1,
						TotalWeight: "1",
					},
				},
				GroupPolicies: []*GroupPolicyInfo{
					groupPolicy,
				},
				Proposals: []*Proposal{
					{
						ProposalId:         1,
						Address:            accAddr.String(),
						Metadata:           []byte("proposal metadata"),
						GroupVersion:       1,
						GroupPolicyVersion: 1,
						Proposers: []string{
							memberAddr.String(),
						},
						SubmittedAt: submittedAt,
						Status:      ProposalStatusClosed,
						Result:      ProposalResultAccepted,
						VoteState: Tally{
							YesCount:     "0",
							NoCount:      "0",
							AbstainCount: "-1",
							VetoCount:    "0",
						},
					},
				},
			},
			true,
		},
		{
			"invalid VoteState with negative VetoCount",
			GenesisState{
				Groups: []*GroupInfo{
					{
						GroupId:     1,
						Admin:       accAddr.String(),
						Metadata:    []byte("1"),
						Version:     1,
						TotalWeight: "1",
					},
				},
				GroupPolicies: []*GroupPolicyInfo{
					groupPolicy,
				},
				Proposals: []*Proposal{
					{
						ProposalId:         1,
						Address:            accAddr.String(),
						Metadata:           []byte("proposal metadata"),
						GroupVersion:       1,
						GroupPolicyVersion: 1,
						Proposers: []string{
							memberAddr.String(),
						},
						SubmittedAt: submittedAt,
						Status:      ProposalStatusClosed,
						Result:      ProposalResultAccepted,
						VoteState: Tally{
							YesCount:     "0",
							NoCount:      "0",
							AbstainCount: "0",
							VetoCount:    "-1",
						},
					},
				},
			},
			true,
		},
		{
			"invalid voter",
			GenesisState{
				Groups: []*GroupInfo{
					{
						GroupId:     1,
						Admin:       accAddr.String(),
						Metadata:    []byte("1"),
						Version:     1,
						TotalWeight: "1",
					},
				},
				GroupPolicies: []*GroupPolicyInfo{
					groupPolicy,
				},
				Proposals: []*Proposal{
					proposal,
				},
				Votes: []*Vote{
					{
						ProposalId:  proposal.ProposalId,
						Voter:       "invalid voter",
						SubmittedAt: submittedAt,
						Choice:      Choice_CHOICE_YES,
					},
				},
			},
			true,
		},
		{
			"invalid proposal id",
			GenesisState{
				Groups: []*GroupInfo{
					{
						GroupId:     1,
						Admin:       accAddr.String(),
						Metadata:    []byte("1"),
						Version:     1,
						TotalWeight: "1",
					},
				},
				GroupPolicies: []*GroupPolicyInfo{
					groupPolicy,
				},
				Proposals: []*Proposal{
					proposal,
				},
				Votes: []*Vote{
					{
						ProposalId:  0,
						Voter:       memberAddr.String(),
						SubmittedAt: submittedAt,
						Choice:      Choice_CHOICE_YES,
					},
				},
			},
			true,
		},
		{
			"vote on proposal that doesn't exist",
			GenesisState{
				Groups: []*GroupInfo{
					{
						GroupId:     1,
						Admin:       accAddr.String(),
						Metadata:    []byte("1"),
						Version:     1,
						TotalWeight: "1",
					},
				},
				GroupPolicies: []*GroupPolicyInfo{
					groupPolicy,
				},
				Proposals: []*Proposal{
					proposal,
				},
				Votes: []*Vote{
					{
						ProposalId:  2,
						Voter:       memberAddr.String(),
						SubmittedAt: submittedAt,
						Choice:      Choice_CHOICE_YES,
					},
				},
			},
			true,
		},
		{
			"invalid choice",
			GenesisState{
				Groups: []*GroupInfo{
					{
						GroupId:     1,
						Admin:       accAddr.String(),
						Metadata:    []byte("1"),
						Version:     1,
						TotalWeight: "1",
					},
				},
				GroupPolicies: []*GroupPolicyInfo{
					groupPolicy,
				},
				Proposals: []*Proposal{
					proposal,
				},
				Votes: []*Vote{
					{
						ProposalId:  proposal.ProposalId,
						Voter:       memberAddr.String(),
						SubmittedAt: submittedAt,
						Choice:      Choice_CHOICE_UNSPECIFIED,
					},
				},
			},
			true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			err := tc.genesisState.Validate()
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
