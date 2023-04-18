package group

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cosmos/gogoproto/proto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	client "github.com/cosmos/cosmos-sdk/x/group/client/cli"
)

type E2ETestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network

	group         *group.GroupInfo
	groupPolicies []*group.GroupPolicyInfo
	proposal      *group.Proposal
	vote          *group.Vote
	voter         *group.Member
	commonFlags   []string
}

const validMetadata = "metadata"

func NewE2ETestSuite(cfg network.Config) *E2ETestSuite {
	return &E2ETestSuite{cfg: cfg}
}

func (s *E2ETestSuite) SetupSuite() {
	s.T().Log("setting up e2e test suite")

	s.commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	val := s.network.Validators[0]

	// create a new account
	info, _, err := val.ClientCtx.Keyring.NewMnemonic("NewValidator", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)

	pk, err := info.GetPubKey()
	s.Require().NoError(err)

	account := sdk.AccAddress(pk.Address())
	_, err = clitestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		account,
		sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(2000))), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	memberWeight := "3"
	// create a group
	validMembers := fmt.Sprintf(`
	{
		"members": [
			{
				"address": "%s",
				"weight": "%s",
				"metadata": "%s"
			}
		]
	}`, val.Address.String(), memberWeight, validMetadata)
	validMembersFile := testutil.WriteToNewTempFile(s.T(), validMembers)
	out, err := clitestutil.ExecTestCLICmd(val.ClientCtx, client.MsgCreateGroupCmd(),
		append(
			[]string{
				val.Address.String(),
				validMetadata,
				validMembersFile.Name(),
			},
			s.commonFlags...,
		),
	)
	s.Require().NoError(err, out.String())
	txResp := sdk.TxResponse{}
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	s.Require().NoError(clitestutil.CheckTxCode(s.network, val.ClientCtx, txResp.TxHash, 0))

	s.group = &group.GroupInfo{Id: 1, Admin: val.Address.String(), Metadata: validMetadata, TotalWeight: "3", Version: 1}

	// create 5 group policies
	for i := 0; i < 5; i++ {
		threshold := i + 1
		if threshold > 3 {
			threshold = 3
		}

		s.createGroupThresholdPolicyWithBalance(val.Address.String(), "1", threshold, 1000)
		out, err = clitestutil.ExecTestCLICmd(val.ClientCtx, client.QueryGroupPoliciesByGroupCmd(), []string{"1", fmt.Sprintf("--%s=json", flags.FlagOutput)})
		s.Require().NoError(err, out.String())
		s.Require().NoError(s.network.WaitForNextBlock())
	}
	percentage := 0.5
	// create group policy with percentage decision policy
	out, err = clitestutil.ExecTestCLICmd(val.ClientCtx, client.MsgCreateGroupPolicyCmd(),
		append(
			[]string{
				val.Address.String(),
				"1",
				validMetadata,
				testutil.WriteToNewTempFile(s.T(), fmt.Sprintf(`{"@type":"/cosmos.group.v1.PercentageDecisionPolicy", "percentage":"%f", "windows":{"voting_period":"30000s"}}`, percentage)).Name(),
			},
			s.commonFlags...,
		),
	)
	s.Require().NoError(err, out.String())
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	s.Require().NoError(clitestutil.CheckTxCode(s.network, val.ClientCtx, txResp.TxHash, 0))

	out, err = clitestutil.ExecTestCLICmd(val.ClientCtx, client.QueryGroupPoliciesByGroupCmd(), []string{"1", fmt.Sprintf("--%s=json", flags.FlagOutput)})
	s.Require().NoError(err, out.String())

	var res group.QueryGroupPoliciesByGroupResponse
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))
	s.Require().Equal(len(res.GroupPolicies), 6)
	s.groupPolicies = res.GroupPolicies

	// create a proposal
	out, err = clitestutil.ExecTestCLICmd(val.ClientCtx, client.MsgSubmitProposalCmd(),
		append(
			[]string{
				s.createCLIProposal(
					s.groupPolicies[0].Address, val.Address.String(),
					s.groupPolicies[0].Address, val.Address.String(),
					"", "title", "summary"),
			},
			s.commonFlags...,
		),
	)
	s.Require().NoError(err, out.String())
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	s.Require().NoError(clitestutil.CheckTxCode(s.network, val.ClientCtx, txResp.TxHash, 0))

	// vote
	out, err = clitestutil.ExecTestCLICmd(val.ClientCtx, client.MsgVoteCmd(),
		append(
			[]string{
				"1",
				val.Address.String(),
				"VOTE_OPTION_YES",
				"",
			},
			s.commonFlags...,
		),
	)
	s.Require().NoError(err, out.String())
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	s.Require().NoError(clitestutil.CheckTxCode(s.network, val.ClientCtx, txResp.TxHash, 0))

	out, err = clitestutil.ExecTestCLICmd(val.ClientCtx, client.QueryProposalCmd(), []string{"1", fmt.Sprintf("--%s=json", flags.FlagOutput)})
	s.Require().NoError(err, out.String())

	var proposalRes group.QueryProposalResponse
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &proposalRes))
	s.proposal = proposalRes.Proposal

	out, err = clitestutil.ExecTestCLICmd(val.ClientCtx, client.QueryVoteByProposalVoterCmd(), []string{"1", val.Address.String(), fmt.Sprintf("--%s=json", flags.FlagOutput)})
	s.Require().NoError(err, out.String())

	var voteRes group.QueryVoteByProposalVoterResponse
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &voteRes))
	s.vote = voteRes.Vote

	s.voter = &group.Member{
		Address:  val.Address.String(),
		Weight:   memberWeight,
		Metadata: validMetadata,
	}
}

func (s *E2ETestSuite) TearDownSuite() {
	s.T().Log("tearing down e2e test suite")
	s.network.Cleanup()
}

func (s *E2ETestSuite) TestExecProposalsWhenMemberLeavesOrIsUpdated() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	weights := []string{"1", "1", "2"}
	accounts := s.createAccounts(3)
	testCases := []struct {
		name         string
		votes        []string
		members      []string
		malleate     func(groupID string) error
		expectLogErr bool
		errMsg       string
		respType     proto.Message
	}{
		{
			"member leaves while all others vote yes",
			[]string{"VOTE_OPTION_YES", "VOTE_OPTION_YES", "VOTE_OPTION_YES"},
			accounts,
			func(groupID string) error {
				leavingMemberIdx := 1
				args := append(
					[]string{
						accounts[leavingMemberIdx],
						groupID,

						fmt.Sprintf("--%s=%s", flags.FlagFrom, accounts[leavingMemberIdx]),
					},
					s.commonFlags...,
				)
				out, err := clitestutil.ExecTestCLICmd(clientCtx, client.MsgLeaveGroupCmd(), args)
				s.Require().NoError(err, out.String())
				var resp sdk.TxResponse
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
				s.Require().NoError(clitestutil.CheckTxCode(s.network, clientCtx, resp.TxHash, 0))

				return err
			},
			false,
			"",
			&sdk.TxResponse{},
		},
		{
			"member leaves while all others vote yes and no",
			[]string{"VOTE_OPTION_NO", "VOTE_OPTION_YES", "VOTE_OPTION_YES"},
			accounts,
			func(groupID string) error {
				leavingMemberIdx := 1
				args := append(
					[]string{
						accounts[leavingMemberIdx],
						groupID,

						fmt.Sprintf("--%s=%s", flags.FlagFrom, accounts[leavingMemberIdx]),
					},
					s.commonFlags...,
				)
				out, err := clitestutil.ExecTestCLICmd(clientCtx, client.MsgLeaveGroupCmd(), args)
				s.Require().NoError(err, out.String())
				var resp sdk.TxResponse
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
				s.Require().NoError(clitestutil.CheckTxCode(s.network, clientCtx, resp.TxHash, 0))

				return err
			},
			true,
			"PROPOSAL_EXECUTOR_RESULT_NOT_RUN",
			&sdk.TxResponse{},
		},
		{
			"member that leaves does affect the threshold policy outcome",
			[]string{"VOTE_OPTION_YES", "VOTE_OPTION_YES"},
			accounts,
			func(groupID string) error {
				leavingMemberIdx := 2
				args := append(
					[]string{
						accounts[leavingMemberIdx],
						groupID,

						fmt.Sprintf("--%s=%s", flags.FlagFrom, accounts[leavingMemberIdx]),
					},
					s.commonFlags...,
				)
				out, err := clitestutil.ExecTestCLICmd(clientCtx, client.MsgLeaveGroupCmd(), args)
				s.Require().NoError(err, out.String())
				var resp sdk.TxResponse
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
				s.Require().NoError(clitestutil.CheckTxCode(s.network, clientCtx, resp.TxHash, 0))

				return err
			},
			false,
			"",
			&sdk.TxResponse{},
		},
		{
			"update group policy voids the proposal",
			[]string{"VOTE_OPTION_YES", "VOTE_OPTION_NO"},
			accounts,
			func(groupID string) error {
				updateGroup := s.newValidMembers(weights[0:1], accounts[0:1])

				updateGroupByte, err := json.Marshal(updateGroup)
				s.Require().NoError(err)

				validUpdateMemberFileName := testutil.WriteToNewTempFile(s.T(), string(updateGroupByte)).Name()

				args := append(
					[]string{
						accounts[0],
						groupID,
						validUpdateMemberFileName,
					},
					s.commonFlags...,
				)
				out, err := clitestutil.ExecTestCLICmd(clientCtx, client.MsgUpdateGroupMembersCmd(), args)
				s.Require().NoError(err, out.String())
				var resp sdk.TxResponse
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
				s.Require().NoError(clitestutil.CheckTxCode(s.network, clientCtx, resp.TxHash, 0))

				return err
			},
			true,
			"PROPOSAL_EXECUTOR_RESULT_NOT_RUN",
			&sdk.TxResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmdSubmitProposal := client.MsgSubmitProposalCmd()
			cmdMsgExec := client.MsgExecCmd()

			groupID := s.createGroupWithMembers(weights, accounts)
			groupPolicyAddress := s.createGroupThresholdPolicyWithBalance(accounts[0], groupID, 3, 100)

			// Submit proposal
			proposal := s.createCLIProposal(
				groupPolicyAddress, tc.members[0],
				groupPolicyAddress, tc.members[0],
				"", "title", "summary",
			)
			submitProposalArgs := append([]string{
				proposal,
			},
				s.commonFlags...,
			)
			var submitProposalResp sdk.TxResponse
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmdSubmitProposal, submitProposalArgs)
			s.Require().NoError(err, out.String())
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &submitProposalResp), out.String())
			submitProposalResp, err = clitestutil.GetTxResponse(s.network, val.ClientCtx, submitProposalResp.TxHash)
			s.Require().NoError(err)
			proposalID := s.getProposalIDFromTxResponse(submitProposalResp)

			for i, vote := range tc.votes {
				memberAddress := tc.members[i]
				out, err = clitestutil.ExecTestCLICmd(val.ClientCtx, client.MsgVoteCmd(),
					append(
						[]string{
							proposalID,
							memberAddress,
							vote,
							"",
						},
						s.commonFlags...,
					),
				)

				var txResp sdk.TxResponse
				s.Require().NoError(err, out.String())
				s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
				s.Require().NoError(clitestutil.CheckTxCode(s.network, clientCtx, txResp.TxHash, 0))

			}

			err = tc.malleate(groupID)
			s.Require().NoError(err)
			s.Require().NoError(s.network.WaitForNextBlock())

			args := append(
				[]string{
					proposalID,
					fmt.Sprintf("--%s=%s", flags.FlagFrom, tc.members[0]),
				},
				s.commonFlags...,
			)
			out, err = clitestutil.ExecTestCLICmd(clientCtx, cmdMsgExec, args)
			s.Require().NoError(err)

			var execResp sdk.TxResponse
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &execResp), out.String())
			execResp, err = clitestutil.GetTxResponse(s.network, val.ClientCtx, execResp.TxHash)
			s.Require().NoError(err)

			if tc.expectLogErr {
				s.Require().Contains(execResp.RawLog, tc.errMsg)
			}
		})
	}
}

func (s *E2ETestSuite) getProposalIDFromTxResponse(txResp sdk.TxResponse) string {
	s.Require().Greater(len(txResp.Logs), 0)
	s.Require().NotNil(txResp.Logs[0].Events)
	events := txResp.Logs[0].Events
	createProposalEvent, _ := sdk.TypedEventToEvent(&group.EventSubmitProposal{})

	for _, e := range events {
		if e.Type == createProposalEvent.Type {
			return strings.ReplaceAll(e.Attributes[0].Value, "\"", "")
		}
	}

	return ""
}

func (s *E2ETestSuite) getGroupIDFromTxResponse(txResp sdk.TxResponse) string {
	s.Require().Greater(len(txResp.Logs), 0)
	s.Require().NotNil(txResp.Logs[0].Events)
	events := txResp.Logs[0].Events
	createProposalEvent, _ := sdk.TypedEventToEvent(&group.EventCreateGroup{})

	for _, e := range events {
		if e.Type == createProposalEvent.Type {
			return strings.ReplaceAll(e.Attributes[0].Value, "\"", "")
		}
	}

	return ""
}

// createCLIProposal writes a CLI proposal with a MsgSend to a file. Returns
// the path to the JSON file.
func (s *E2ETestSuite) createCLIProposal(groupPolicyAddress, proposer, sendFrom, sendTo, metadata, title, summary string) string {
	_, err := base64.StdEncoding.DecodeString(metadata)
	s.Require().NoError(err)

	msg := banktypes.MsgSend{
		FromAddress: sendFrom,
		ToAddress:   sendTo,
		Amount:      sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(20))),
	}
	msgJSON, err := s.cfg.Codec.MarshalInterfaceJSON(&msg)
	s.Require().NoError(err)

	p := client.Proposal{
		GroupPolicyAddress: groupPolicyAddress,
		Messages:           []json.RawMessage{msgJSON},
		Metadata:           metadata,
		Proposers:          []string{proposer},
		Title:              title,
		Summary:            summary,
	}

	bz, err := json.Marshal(&p)
	s.Require().NoError(err)

	return testutil.WriteToNewTempFile(s.T(), string(bz)).Name()
}

func (s *E2ETestSuite) createAccounts(quantity int) []string {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx
	accounts := make([]string, quantity)

	for i := 1; i <= quantity; i++ {
		memberNumber := uuid.New().String()

		info, _, err := clientCtx.Keyring.NewMnemonic(fmt.Sprintf("member%s", memberNumber), keyring.English, sdk.FullFundraiserPath,
			keyring.DefaultBIP39Passphrase, hd.Secp256k1)
		s.Require().NoError(err)

		pk, err := info.GetPubKey()
		s.Require().NoError(err)

		account := sdk.AccAddress(pk.Address())
		accounts[i-1] = account.String()

		_, err = clitestutil.MsgSendExec(
			val.ClientCtx,
			val.Address,
			account,
			sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(2000))), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		)
		s.Require().NoError(err)
		s.Require().NoError(s.network.WaitForNextBlock())
	}

	return accounts
}

func (s *E2ETestSuite) createGroupWithMembers(membersWeight, membersAddress []string) string {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	s.Require().Equal(len(membersWeight), len(membersAddress))

	membersValid := s.newValidMembers(membersWeight, membersAddress)
	membersByte, err := json.Marshal(membersValid)

	s.Require().NoError(err)

	validMembersFile := testutil.WriteToNewTempFile(s.T(), string(membersByte))
	out, err := clitestutil.ExecTestCLICmd(clientCtx, client.MsgCreateGroupCmd(),
		append(
			[]string{
				membersAddress[0],
				validMetadata,
				validMembersFile.Name(),
			},
			s.commonFlags...,
		),
	)
	s.Require().NoError(err, out.String())
	var txResp sdk.TxResponse
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	txResp, err = clitestutil.GetTxResponse(s.network, val.ClientCtx, txResp.TxHash)
	s.Require().NoError(err)
	return s.getGroupIDFromTxResponse(txResp)
}

func (s *E2ETestSuite) createGroupThresholdPolicyWithBalance(adminAddress, groupID string, threshold int, tokens int64) string {
	s.Require().NoError(s.network.WaitForNextBlock())

	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	out, err := clitestutil.ExecTestCLICmd(clientCtx, client.MsgCreateGroupPolicyCmd(),
		append(
			[]string{
				adminAddress,
				groupID,
				validMetadata,
				testutil.WriteToNewTempFile(s.T(), fmt.Sprintf(`{"@type":"/cosmos.group.v1.ThresholdDecisionPolicy", "threshold":"%d", "windows":{"voting_period":"30000s"}}`, threshold)).Name(),
			},
			s.commonFlags...,
		),
	)
	txResp := sdk.TxResponse{}
	s.Require().NoError(err, out.String())
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	s.Require().NoError(clitestutil.CheckTxCode(s.network, val.ClientCtx, txResp.TxHash, 0))

	out, err = clitestutil.ExecTestCLICmd(val.ClientCtx, client.QueryGroupPoliciesByGroupCmd(), []string{groupID, fmt.Sprintf("--%s=json", flags.FlagOutput)})
	s.Require().NoError(err, out.String())

	var res group.QueryGroupPoliciesByGroupResponse
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))
	groupPolicyAddress := res.GroupPolicies[0].Address

	addr, err := sdk.AccAddressFromBech32(groupPolicyAddress)
	s.Require().NoError(err)
	_, err = clitestutil.MsgSendExec(clientCtx, val.Address, addr,
		sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(tokens))),
		s.commonFlags...,
	)
	s.Require().NoError(err)
	return groupPolicyAddress
}

func (s *E2ETestSuite) newValidMembers(weights, membersAddress []string) struct{ Members []group.MemberRequest } {
	s.Require().Equal(len(weights), len(membersAddress))
	membersValid := []group.MemberRequest{}
	for i, address := range membersAddress {
		membersValid = append(membersValid, group.MemberRequest{
			Address:  address,
			Weight:   weights[i],
			Metadata: validMetadata,
		})
	}

	return struct {
		Members []group.MemberRequest
	}{
		Members: membersValid,
	}
}
