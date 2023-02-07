package group

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	client "github.com/cosmos/cosmos-sdk/x/group/client/cli"
)

func (s *E2ETestSuite) TestQueryGroupInfo() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		expectedCode uint32
	}{
		{
			"group not found",
			[]string{"12345", fmt.Sprintf("--%s=json", flags.FlagOutput)},
			true,
			"group: not found",
			0,
		},
		{
			"group id invalid",
			[]string{"", fmt.Sprintf("--%s=json", flags.FlagOutput)},
			true,
			"strconv.ParseUint: parsing \"\": invalid syntax",
			0,
		},
		{
			"group found",
			[]string{strconv.FormatUint(s.group.Id, 10), fmt.Sprintf("--%s=json", flags.FlagOutput)},
			false,
			"",
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryGroupInfoCmd()

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())

				var g group.GroupInfo
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &g))
				s.Require().Equal(s.group.Id, g.Id)
				s.Require().Equal(s.group.Admin, g.Admin)
				s.Require().Equal(s.group.TotalWeight, g.TotalWeight)
				s.Require().Equal(s.group.Metadata, g.Metadata)
				s.Require().Equal(s.group.Version, g.Version)
			}
		})
	}
}

func (s *E2ETestSuite) TestQueryGroupsByMembers() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx
	require := s.Require()

	cmd := client.QueryGroupsByAdminCmd()
	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, []string{val.Address.String(), fmt.Sprintf("--%s=json", flags.FlagOutput)})
	require.NoError(err)

	var groups group.QueryGroupsByAdminResponse
	val.ClientCtx.Codec.MustUnmarshalJSON(out.Bytes(), &groups)
	require.Len(groups.Groups, 1)

	cmd = client.QueryGroupMembersCmd()
	out, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, []string{fmt.Sprintf("%d", groups.Groups[0].Id), fmt.Sprintf("--%s=json", flags.FlagOutput)})
	require.NoError(err)

	var members group.QueryGroupMembersResponse
	val.ClientCtx.Codec.MustUnmarshalJSON(out.Bytes(), &members)
	require.Len(members.Members, 1)

	testAddr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		numItems     int
		expectGroups []*group.GroupInfo
	}{
		{
			"invalid address",
			[]string{"abcd", fmt.Sprintf("--%s=json", flags.FlagOutput)},
			true,
			"invalid bech32 string",
			0,
			[]*group.GroupInfo{},
		},
		{
			"not part of any group",
			[]string{testAddr.String(), fmt.Sprintf("--%s=json", flags.FlagOutput)},
			false,
			"",
			0,
			[]*group.GroupInfo{},
		},
		{
			"expect one group (request with pagination)",
			[]string{
				members.Members[0].Member.Address,
				fmt.Sprintf("--%s=json", flags.FlagOutput),
				"--limit=1",
			},
			false,
			"",
			1,
			groups.Groups,
		},
		{
			"expect one group",
			[]string{
				members.Members[0].Member.Address,
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
			"",
			1,
			groups.Groups,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := client.QueryGroupsByMemberCmd()
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				require.Contains(out.String(), tc.expectErrMsg)
			} else {
				require.NoError(err, out.String())

				var resp group.QueryGroupsByMemberResponse
				val.ClientCtx.Codec.MustUnmarshalJSON(out.Bytes(), &resp)
				require.Len(resp.Groups, tc.numItems)

				require.Equal(tc.expectGroups, resp.Groups)
			}
		})
	}
}

func (s *E2ETestSuite) TestQueryGroups() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx
	require := s.Require()

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		numItems     int
		expectGroups []*group.GroupInfo
	}{
		{
			name:      "valid req",
			args:      []string{fmt.Sprintf("--%s=json", flags.FlagOutput)},
			expectErr: false,
			numItems:  5,
		},
		{
			name: "valid req with pagination",
			args: []string{
				"--limit=2",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			expectErr: false,
			numItems:  2,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := client.QueryGroupsCmd()
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				require.Contains(out.String(), tc.expectErrMsg)
			} else {
				require.NoError(err, out.String())

				var resp group.QueryGroupsResponse
				val.ClientCtx.Codec.MustUnmarshalJSON(out.Bytes(), &resp)

				require.Len(resp.Groups, tc.numItems)
			}
		})
	}
}

func (s *E2ETestSuite) TestQueryGroupMembers() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name          string
		args          []string
		expectErr     bool
		expectErrMsg  string
		expectedCode  uint32
		expectMembers []*group.GroupMember
	}{
		{
			"no group",
			[]string{"12345", fmt.Sprintf("--%s=json", flags.FlagOutput)},
			false,
			"",
			0,
			[]*group.GroupMember{},
		},
		{
			"members found",
			[]string{strconv.FormatUint(s.group.Id, 10), fmt.Sprintf("--%s=json", flags.FlagOutput)},
			false,
			"",
			0,
			[]*group.GroupMember{
				{
					GroupId: s.group.Id,
					Member: &group.Member{
						Address:  val.Address.String(),
						Weight:   "3",
						Metadata: validMetadata,
					},
				},
			},
		},
		{
			"members found (request with pagination)",
			[]string{
				strconv.FormatUint(s.group.Id, 10),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
				"--limit=1",
			},
			false,
			"",
			0,
			[]*group.GroupMember{
				{
					GroupId: s.group.Id,
					Member: &group.Member{
						Address:  val.Address.String(),
						Weight:   "3",
						Metadata: validMetadata,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryGroupMembersCmd()

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())

				var res group.QueryGroupMembersResponse
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))
				s.Require().Equal(len(res.Members), len(tc.expectMembers))
				for i := range res.Members {
					s.Require().Equal(res.Members[i].GroupId, tc.expectMembers[i].GroupId)
					s.Require().Equal(res.Members[i].Member.Address, tc.expectMembers[i].Member.Address)
					s.Require().Equal(res.Members[i].Member.Metadata, tc.expectMembers[i].Member.Metadata)
					s.Require().Equal(res.Members[i].Member.Weight, tc.expectMembers[i].Member.Weight)
				}
			}
		})
	}
}

func (s *E2ETestSuite) TestQueryGroupsByAdmin() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		expectedCode uint32
		expectGroups []*group.GroupInfo
	}{
		{
			"invalid admin address",
			[]string{"invalid"},
			true,
			"decoding bech32 failed: invalid bech32 string",
			0,
			[]*group.GroupInfo{},
		},
		{
			"no group",
			[]string{s.network.Validators[1].Address.String(), fmt.Sprintf("--%s=json", flags.FlagOutput)},
			false,
			"",
			0,
			[]*group.GroupInfo{},
		},
		{
			"found groups",
			[]string{val.Address.String(), fmt.Sprintf("--%s=json", flags.FlagOutput)},
			false,
			"",
			0,
			[]*group.GroupInfo{
				s.group,
			},
		},
		{
			"found groups (request with pagination)",
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
				"--limit=2",
			},
			false,
			"",
			0,
			[]*group.GroupInfo{
				s.group,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryGroupsByAdminCmd()

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())

				var res group.QueryGroupsByAdminResponse
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))
				s.Require().Equal(len(res.Groups), len(tc.expectGroups))
				for i := range res.Groups {
					s.Require().Equal(res.Groups[i].Id, tc.expectGroups[i].Id)
					s.Require().Equal(res.Groups[i].Metadata, tc.expectGroups[i].Metadata)
					s.Require().Equal(res.Groups[i].Version, tc.expectGroups[i].Version)
					s.Require().Equal(res.Groups[i].TotalWeight, tc.expectGroups[i].TotalWeight)
					s.Require().Equal(res.Groups[i].Admin, tc.expectGroups[i].Admin)
				}
			}
		})
	}
}

func (s *E2ETestSuite) TestQueryGroupPolicyInfo() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		expectedCode uint32
	}{
		{
			"group policy not found",
			[]string{val.Address.String(), fmt.Sprintf("--%s=json", flags.FlagOutput)},
			true,
			"group policy: not found",
			0,
		},
		{
			"group policy found",
			[]string{s.groupPolicies[0].Address, fmt.Sprintf("--%s=json", flags.FlagOutput)},
			false,
			"",
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryGroupPolicyInfoCmd()

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())

				var g group.GroupPolicyInfo
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &g))
				s.Require().Equal(s.groupPolicies[0].GroupId, g.GroupId)
				s.Require().Equal(s.groupPolicies[0].Address, g.Address)
				s.Require().Equal(s.groupPolicies[0].Admin, g.Admin)
				s.Require().Equal(s.groupPolicies[0].Metadata, g.Metadata)
				s.Require().Equal(s.groupPolicies[0].Version, g.Version)
				dp1, err := s.groupPolicies[0].GetDecisionPolicy()
				s.Require().NoError(err)
				dp2, err := g.GetDecisionPolicy()
				s.Require().NoError(err)
				s.Require().Equal(dp1, dp2)
			}
		})
	}
}

func (s *E2ETestSuite) TestQueryGroupPoliciesByGroup() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name                string
		args                []string
		expectErr           bool
		expectErrMsg        string
		expectedCode        uint32
		expectGroupPolicies []*group.GroupPolicyInfo
	}{
		{
			"invalid group id",
			[]string{""},
			true,
			"strconv.ParseUint: parsing \"\": invalid syntax",
			0,
			[]*group.GroupPolicyInfo{},
		},
		{
			"no group policy",
			[]string{"12345", fmt.Sprintf("--%s=json", flags.FlagOutput)},
			false,
			"",
			0,
			[]*group.GroupPolicyInfo{},
		},
		{
			"found group policies",
			[]string{strconv.FormatUint(s.group.Id, 10), fmt.Sprintf("--%s=json", flags.FlagOutput)},
			false,
			"",
			0,
			[]*group.GroupPolicyInfo{
				s.groupPolicies[0],
				s.groupPolicies[1],
				s.groupPolicies[2],
				s.groupPolicies[3],
				s.groupPolicies[4],
				s.groupPolicies[5],
			},
		},
		{
			"found group policies (request with pagination)",
			[]string{
				strconv.FormatUint(s.group.Id, 10),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
				"--limit=2",
			},
			false,
			"",
			0,
			[]*group.GroupPolicyInfo{
				s.groupPolicies[0],
				s.groupPolicies[1],
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryGroupPoliciesByGroupCmd()

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())

				var res group.QueryGroupPoliciesByGroupResponse
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))
				s.Require().Equal(len(res.GroupPolicies), len(tc.expectGroupPolicies))
				for i := range res.GroupPolicies {
					s.Require().Equal(res.GroupPolicies[i].GroupId, tc.expectGroupPolicies[i].GroupId)
					s.Require().Equal(res.GroupPolicies[i].Metadata, tc.expectGroupPolicies[i].Metadata)
					s.Require().Equal(res.GroupPolicies[i].Version, tc.expectGroupPolicies[i].Version)
					s.Require().Equal(res.GroupPolicies[i].Admin, tc.expectGroupPolicies[i].Admin)
					dp1, err := s.groupPolicies[i].GetDecisionPolicy()
					s.Require().NoError(err)
					dp2, err := tc.expectGroupPolicies[i].GetDecisionPolicy()
					s.Require().NoError(err)
					s.Require().Equal(dp1, dp2)
				}
			}
		})
	}
}

func (s *E2ETestSuite) TestQueryGroupPoliciesByAdmin() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name                string
		args                []string
		expectErr           bool
		expectErrMsg        string
		expectedCode        uint32
		expectGroupPolicies []*group.GroupPolicyInfo
	}{
		{
			"invalid admin address",
			[]string{"invalid"},
			true,
			"decoding bech32 failed: invalid bech32 string",
			0,
			[]*group.GroupPolicyInfo{},
		},
		{
			"no group policy",
			[]string{s.network.Validators[1].Address.String(), fmt.Sprintf("--%s=json", flags.FlagOutput)},
			false,
			"",
			0,
			[]*group.GroupPolicyInfo{},
		},
		{
			"found group policies",
			[]string{val.Address.String(), fmt.Sprintf("--%s=json", flags.FlagOutput)},
			false,
			"",
			0,
			[]*group.GroupPolicyInfo{
				s.groupPolicies[0],
				s.groupPolicies[1],
				s.groupPolicies[2],
				s.groupPolicies[3],
				s.groupPolicies[4],
				s.groupPolicies[5],
			},
		},
		{
			"found group policies (request with pagination)",
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
				"--limit=2",
			},
			false,
			"",
			0,
			[]*group.GroupPolicyInfo{
				s.groupPolicies[0],
				s.groupPolicies[1],
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryGroupPoliciesByAdminCmd()

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())

				var res group.QueryGroupPoliciesByAdminResponse
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))
				s.Require().Equal(len(res.GroupPolicies), len(tc.expectGroupPolicies))
				for i := range res.GroupPolicies {
					s.Require().Equal(res.GroupPolicies[i].GroupId, tc.expectGroupPolicies[i].GroupId)
					s.Require().Equal(res.GroupPolicies[i].Metadata, tc.expectGroupPolicies[i].Metadata)
					s.Require().Equal(res.GroupPolicies[i].Version, tc.expectGroupPolicies[i].Version)
					s.Require().Equal(res.GroupPolicies[i].Admin, tc.expectGroupPolicies[i].Admin)
					dp1, err := s.groupPolicies[i].GetDecisionPolicy()
					s.Require().NoError(err)
					dp2, err := tc.expectGroupPolicies[i].GetDecisionPolicy()
					s.Require().NoError(err)
					s.Require().Equal(dp1, dp2)
				}
			}
		})
	}
}

func (s *E2ETestSuite) TestQueryProposal() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		expectedCode uint32
	}{
		{
			"not found",
			[]string{"12345", fmt.Sprintf("--%s=json", flags.FlagOutput)},
			true,
			"not found",
			0,
		},
		{
			"invalid proposal id",
			[]string{"", fmt.Sprintf("--%s=json", flags.FlagOutput)},
			true,
			"strconv.ParseUint: parsing \"\": invalid syntax",
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryProposalCmd()

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
			}
		})
	}
}

func (s *E2ETestSuite) TestQueryProposalsByGroupPolicy() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name            string
		args            []string
		expectErr       bool
		expectErrMsg    string
		expectedCode    uint32
		expectProposals []*group.Proposal
	}{
		{
			"invalid group policy address",
			[]string{"invalid"},
			true,
			"decoding bech32 failed: invalid bech32 string",
			0,
			[]*group.Proposal{},
		},
		{
			"no group policy",
			[]string{s.network.Validators[1].Address.String(), fmt.Sprintf("--%s=json", flags.FlagOutput)},
			false,
			"",
			0,
			[]*group.Proposal{},
		},
		{
			"found proposals",
			[]string{s.groupPolicies[0].Address, fmt.Sprintf("--%s=json", flags.FlagOutput)},
			false,
			"",
			0,
			[]*group.Proposal{
				s.proposal,
			},
		},
		{
			"found proposals (request with pagination)",
			[]string{
				s.groupPolicies[0].Address,
				fmt.Sprintf("--%s=json", flags.FlagOutput),
				"--limit=2",
			},
			false,
			"",
			0,
			[]*group.Proposal{
				s.proposal,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryProposalsByGroupPolicyCmd()

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())

				var res group.QueryProposalsByGroupPolicyResponse
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))
				s.Require().Equal(len(res.Proposals), len(tc.expectProposals))
				for i := range res.Proposals {
					s.Require().Equal(res.Proposals[i], tc.expectProposals[i])
				}
			}
		})
	}
}

func (s *E2ETestSuite) TestQueryVoteByProposalVoter() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		expectedCode uint32
	}{
		{
			"invalid voter address",
			[]string{"1", "invalid", fmt.Sprintf("--%s=json", flags.FlagOutput)},
			true,
			"decoding bech32 failed: invalid bech32",
			0,
		},
		{
			"invalid proposal id",
			[]string{"", val.Address.String(), fmt.Sprintf("--%s=json", flags.FlagOutput)},
			true,
			"strconv.ParseUint: parsing \"\": invalid syntax",
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryVoteByProposalVoterCmd()

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
			}
		})
	}
}

func (s *E2ETestSuite) TestQueryVotesByProposal() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		expectedCode uint32
		expectVotes  []*group.Vote
	}{
		{
			"invalid proposal id",
			[]string{"", fmt.Sprintf("--%s=json", flags.FlagOutput)},
			true,
			"strconv.ParseUint: parsing \"\": invalid syntax",
			0,
			[]*group.Vote{},
		},
		{
			"no votes",
			[]string{"12345", fmt.Sprintf("--%s=json", flags.FlagOutput)},
			false,
			"",
			0,
			[]*group.Vote{},
		},
		{
			"found votes",
			[]string{"1", fmt.Sprintf("--%s=json", flags.FlagOutput)},
			false,
			"",
			0,
			[]*group.Vote{
				s.vote,
			},
		},
		{
			"found votes (request with pagination)",
			[]string{
				"1",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
				"--limit=2",
			},
			false,
			"",
			0,
			[]*group.Vote{
				s.vote,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryVotesByProposalCmd()

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())

				var res group.QueryVotesByProposalResponse
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))
				s.Require().Equal(len(res.Votes), len(tc.expectVotes))
				for i := range res.Votes {
					s.Require().Equal(res.Votes[i], tc.expectVotes[i])
				}
			}
		})
	}
}

func (s *E2ETestSuite) TestQueryVotesByVoter() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		expectedCode uint32
		expectVotes  []*group.Vote
	}{
		{
			"invalid voter address",
			[]string{"abcd", fmt.Sprintf("--%s=json", flags.FlagOutput)},
			true,
			"decoding bech32 failed: invalid bech32",
			0,
			[]*group.Vote{},
		},
		{
			"no votes",
			[]string{s.groupPolicies[0].Address, fmt.Sprintf("--%s=json", flags.FlagOutput)},
			true,
			"",
			0,
			[]*group.Vote{},
		},
		{
			"found votes",
			[]string{val.Address.String(), fmt.Sprintf("--%s=json", flags.FlagOutput)},
			false,
			"",
			0,
			[]*group.Vote{
				s.vote,
			},
		},
		{
			"found votes (request with pagination)",
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
				"--limit=2",
			},
			false,
			"",
			0,
			[]*group.Vote{
				s.vote,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryVotesByVoterCmd()

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())

				var res group.QueryVotesByVoterResponse
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))
				s.Require().Equal(len(res.Votes), len(tc.expectVotes))
				for i := range res.Votes {
					s.Require().Equal(res.Votes[i], tc.expectVotes[i])
				}
			}
		})
	}
}

func (s *E2ETestSuite) TestTallyResult() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	member := s.voter

	commonFlags := []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	// create a proposal
	out, err := clitestutil.ExecTestCLICmd(val.ClientCtx, client.MsgSubmitProposalCmd(),
		append(
			[]string{
				s.createCLIProposal(
					s.groupPolicies[0].Address, val.Address.String(),
					s.groupPolicies[0].Address, val.Address.String(),
					"", "title", "summary"),
			},
			commonFlags...,
		),
	)
	s.Require().NoError(err, out.String())

	var txResp sdk.TxResponse
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	txResp, err = clitestutil.GetTxResponse(s.network, clientCtx, txResp.TxHash)
	s.Require().NoError(err)
	s.Require().Equal(txResp.Code, uint32(0), out.String())

	proposalID := s.getProposalIDFromTxResponse(txResp)

	testCases := []struct {
		name           string
		args           []string
		expectErr      bool
		expTallyResult group.TallyResult
		expectErrMsg   string
		expectedCode   uint32
	}{
		{
			"not found",
			[]string{
				"12345",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true,
			group.TallyResult{},
			"not found",
			0,
		},
		{
			"invalid proposal id",
			[]string{
				"",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true,
			group.TallyResult{},
			"strconv.ParseUint: parsing \"\": invalid syntax",
			0,
		},
		{
			"valid proposal id with no votes",
			[]string{
				proposalID,
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
			group.DefaultTallyResult(),
			"",
			0,
		},
		{
			"valid proposal id",
			[]string{
				"1",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
			group.TallyResult{
				YesCount:        member.Weight,
				AbstainCount:    "0",
				NoCount:         "0",
				NoWithVetoCount: "0",
			},
			"",
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryTallyResultCmd()

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				var tallyResultRes group.QueryTallyResultResponse
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &tallyResultRes))
				s.Require().NotNil(tallyResultRes)
				s.Require().Equal(tc.expTallyResult, tallyResultRes.Tally)
			}
		})
	}
}
