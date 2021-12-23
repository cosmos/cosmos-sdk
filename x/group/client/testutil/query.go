package testutil

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	client "github.com/cosmos/cosmos-sdk/x/group/client/cli"

	tmcli "github.com/tendermint/tendermint/libs/cli"
)

func (s *IntegrationTestSuite) TestQueryGroupInfo() {
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
			[]string{"12345", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
			"not found: invalid request",
			0,
		},
		{
			"group id invalid",
			[]string{"", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
			"strconv.ParseUint: parsing \"\": invalid syntax",
			0,
		},
		{
			"group found",
			[]string{strconv.FormatUint(s.group.GroupId, 10), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			"",
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryGroupInfoCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())

				var g group.GroupInfo
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &g))
				s.Require().Equal(s.group.GroupId, g.GroupId)
				s.Require().Equal(s.group.Admin, g.Admin)
				s.Require().Equal(s.group.TotalWeight, g.TotalWeight)
				s.Require().Equal(s.group.Metadata, g.Metadata)
				s.Require().Equal(s.group.Version, g.Version)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryGroupsByMembers() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx
	require := s.Require()

	cmd := client.QueryGroupsByAdminCmd()
	out, err := cli.ExecTestCLICmd(clientCtx, cmd, []string{val.Address.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
	require.NoError(err)

	var groups group.QueryGroupsByAdminResponse
	val.ClientCtx.Codec.MustUnmarshalJSON(out.Bytes(), &groups)
	require.Len(groups.Groups, 1)

	cmd = client.QueryGroupMembersCmd()
	out, err = cli.ExecTestCLICmd(clientCtx, cmd, []string{fmt.Sprintf("%d", groups.Groups[0].GroupId), fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
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
			[]string{"abcd", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
			"invalid bech32 string",
			0,
			[]*group.GroupInfo{},
		},
		{
			"not part of any group",
			[]string{testAddr.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			"",
			0,
			[]*group.GroupInfo{},
		},
		{
			"expect one group",
			[]string{members.Members[0].Member.Address, fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
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
			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
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

func (s *IntegrationTestSuite) TestQueryGroupMembers() {
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
			[]string{"12345", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			"",
			0,
			[]*group.GroupMember{},
		},
		{
			"members found",
			[]string{strconv.FormatUint(s.group.GroupId, 10), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			"",
			0,
			[]*group.GroupMember{
				{
					GroupId: s.group.GroupId,
					Member: &group.Member{
						Address:  val.Address.String(),
						Weight:   "3",
						Metadata: []byte{1},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryGroupMembersCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
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

func (s *IntegrationTestSuite) TestQueryGroupsByAdmin() {
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
			[]string{s.network.Validators[1].Address.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			"",
			0,
			[]*group.GroupInfo{},
		},
		{
			"found groups",
			[]string{val.Address.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
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

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())

				var res group.QueryGroupsByAdminResponse
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))
				s.Require().Equal(len(res.Groups), len(tc.expectGroups))
				for i := range res.Groups {
					s.Require().Equal(res.Groups[i].GroupId, tc.expectGroups[i].GroupId)
					s.Require().Equal(res.Groups[i].Metadata, tc.expectGroups[i].Metadata)
					s.Require().Equal(res.Groups[i].Version, tc.expectGroups[i].Version)
					s.Require().Equal(res.Groups[i].TotalWeight, tc.expectGroups[i].TotalWeight)
					s.Require().Equal(res.Groups[i].Admin, tc.expectGroups[i].Admin)
				}
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryGroupAccountInfo() {
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
			"group account not found",
			[]string{val.Address.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
			"not found: invalid request",
			0,
		},
		{
			"group account found",
			[]string{s.groupAccounts[0].Address, fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			"",
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryGroupAccountInfoCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())

				var g group.GroupAccountInfo
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &g))
				s.Require().Equal(s.groupAccounts[0].GroupId, g.GroupId)
				s.Require().Equal(s.groupAccounts[0].Address, g.Address)
				s.Require().Equal(s.groupAccounts[0].Admin, g.Admin)
				s.Require().Equal(s.groupAccounts[0].Metadata, g.Metadata)
				s.Require().Equal(s.groupAccounts[0].Version, g.Version)
				s.Require().Equal(s.groupAccounts[0].GetDecisionPolicy(), g.GetDecisionPolicy())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryGroupAccountsByGroup() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name                string
		args                []string
		expectErr           bool
		expectErrMsg        string
		expectedCode        uint32
		expectGroupAccounts []*group.GroupAccountInfo
	}{
		{
			"invalid group id",
			[]string{""},
			true,
			"strconv.ParseUint: parsing \"\": invalid syntax",
			0,
			[]*group.GroupAccountInfo{},
		},
		{
			"no group account",
			[]string{"12345", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			"",
			0,
			[]*group.GroupAccountInfo{},
		},
		{
			"found group accounts",
			[]string{strconv.FormatUint(s.group.GroupId, 10), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			"",
			0,
			[]*group.GroupAccountInfo{
				s.groupAccounts[0],
				s.groupAccounts[1],
				s.groupAccounts[2],
				s.groupAccounts[3],
				s.groupAccounts[4],
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryGroupAccountsByGroupCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())

				var res group.QueryGroupAccountsByGroupResponse
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))
				s.Require().Equal(len(res.GroupAccounts), len(tc.expectGroupAccounts))
				for i := range res.GroupAccounts {
					s.Require().Equal(res.GroupAccounts[i].GroupId, tc.expectGroupAccounts[i].GroupId)
					s.Require().Equal(res.GroupAccounts[i].Metadata, tc.expectGroupAccounts[i].Metadata)
					s.Require().Equal(res.GroupAccounts[i].Version, tc.expectGroupAccounts[i].Version)
					s.Require().Equal(res.GroupAccounts[i].Admin, tc.expectGroupAccounts[i].Admin)
					s.Require().Equal(res.GroupAccounts[i].GetDecisionPolicy(), tc.expectGroupAccounts[i].GetDecisionPolicy())
				}
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryGroupAccountsByAdmin() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name                string
		args                []string
		expectErr           bool
		expectErrMsg        string
		expectedCode        uint32
		expectGroupAccounts []*group.GroupAccountInfo
	}{
		{
			"invalid admin address",
			[]string{"invalid"},
			true,
			"decoding bech32 failed: invalid bech32 string",
			0,
			[]*group.GroupAccountInfo{},
		},
		{
			"no group account",
			[]string{s.network.Validators[1].Address.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			"",
			0,
			[]*group.GroupAccountInfo{},
		},
		{
			"found group accounts",
			[]string{val.Address.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			"",
			0,
			[]*group.GroupAccountInfo{
				s.groupAccounts[0],
				s.groupAccounts[1],
				s.groupAccounts[2],
				s.groupAccounts[3],
				s.groupAccounts[4],
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryGroupAccountsByAdminCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())

				var res group.QueryGroupAccountsByAdminResponse
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))
				s.Require().Equal(len(res.GroupAccounts), len(tc.expectGroupAccounts))
				for i := range res.GroupAccounts {
					s.Require().Equal(res.GroupAccounts[i].GroupId, tc.expectGroupAccounts[i].GroupId)
					s.Require().Equal(res.GroupAccounts[i].Metadata, tc.expectGroupAccounts[i].Metadata)
					s.Require().Equal(res.GroupAccounts[i].Version, tc.expectGroupAccounts[i].Version)
					s.Require().Equal(res.GroupAccounts[i].Admin, tc.expectGroupAccounts[i].Admin)
					s.Require().Equal(res.GroupAccounts[i].GetDecisionPolicy(), tc.expectGroupAccounts[i].GetDecisionPolicy())
				}
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryProposal() {
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
			[]string{"12345", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
			"not found",
			0,
		},
		{
			"invalid proposal id",
			[]string{"", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
			"strconv.ParseUint: parsing \"\": invalid syntax",
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryProposalCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryProposalsByGroupAccount() {
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
			"invalid group account address",
			[]string{"invalid"},
			true,
			"decoding bech32 failed: invalid bech32 string",
			0,
			[]*group.Proposal{},
		},
		{
			"no group account",
			[]string{s.network.Validators[1].Address.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			"",
			0,
			[]*group.Proposal{},
		},
		{
			"found proposals",
			[]string{s.groupAccounts[0].Address, fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
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
			cmd := client.QueryProposalsByGroupAccountCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())

				var res group.QueryProposalsByGroupAccountResponse
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))
				s.Require().Equal(len(res.Proposals), len(tc.expectProposals))
				for i := range res.Proposals {
					s.Require().Equal(res.Proposals[i], tc.expectProposals[i])
				}
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryVoteByProposalVoter() {
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
			[]string{"1", "invalid", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
			"decoding bech32 failed: invalid bech32",
			0,
		},
		{
			"invalid proposal id",
			[]string{"", val.Address.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
			"strconv.ParseUint: parsing \"\": invalid syntax",
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryVoteByProposalVoterCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryVotesByProposal() {
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
			[]string{"", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
			"strconv.ParseUint: parsing \"\": invalid syntax",
			0,
			[]*group.Vote{},
		},
		{
			"no votes",
			[]string{"12345", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			"",
			0,
			[]*group.Vote{},
		},
		{
			"found votes",
			[]string{"1", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
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

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
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

func (s *IntegrationTestSuite) TestQueryVotesByVoter() {
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
			[]string{"abcd", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
			"decoding bech32 failed: invalid bech32",
			0,
			[]*group.Vote{},
		},
		{
			"no votes",
			[]string{s.groupAccounts[0].Address, fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
			"",
			0,
			[]*group.Vote{},
		},
		{
			"found votes",
			[]string{val.Address.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
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

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
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
