package group

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/stretchr/testify/suite"

	// without this import amino json encoding will fail when resolving any types
	_ "cosmossdk.io/api/cosmos/group/v1"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/group"                   //nolint:staticcheck // deprecated and to be removed
	client "github.com/cosmos/cosmos-sdk/x/group/client/cli" //nolint:staticcheck // deprecated and to be removed
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
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))).String()),
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
		sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(2000))),
		address.NewBech32Codec("cosmos"),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))).String()),
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
		s.Require().NoError(s.network.WaitForNextBlock())
		resp, err := testutil.GetRequest(fmt.Sprintf("%s/cosmos/group/v1/group_policies_by_group/1", val.APIAddress))
		s.Require().NoError(err)

		var groupPoliciesResp group.QueryGroupPoliciesByGroupResponse
		s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(resp, &groupPoliciesResp))
		s.Require().Len(groupPoliciesResp.GroupPolicies, i+1)
	}
	// create group policy with percentage decision policy
	out, err = clitestutil.ExecTestCLICmd(val.ClientCtx, client.MsgCreateGroupPolicyCmd(),
		append(
			[]string{
				val.Address.String(),
				"1",
				validMetadata,
				testutil.WriteToNewTempFile(s.T(), fmt.Sprintf(`{"@type":"/cosmos.group.v1.PercentageDecisionPolicy", "percentage":"%f", "windows":{"voting_period":"30000s"}}`, 0.5)).Name(),
			},
			s.commonFlags...,
		),
	)
	s.Require().NoError(err, out.String())
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	s.Require().NoError(clitestutil.CheckTxCode(s.network, val.ClientCtx, txResp.TxHash, 0))

	resp, err := testutil.GetRequest(fmt.Sprintf("%s/cosmos/group/v1/group_policies_by_group/1", val.APIAddress))
	s.Require().NoError(err)

	var groupPoliciesResp group.QueryGroupPoliciesByGroupResponse
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(resp, &groupPoliciesResp))
	s.Require().Equal(len(groupPoliciesResp.GroupPolicies), 6)
	s.groupPolicies = groupPoliciesResp.GroupPolicies

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

	resp, err = testutil.GetRequest(fmt.Sprintf("%s/cosmos/group/v1/proposal/1", val.APIAddress))
	s.Require().NoError(err)

	var proposalRes group.QueryProposalResponse
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(resp, &proposalRes))
	s.proposal = proposalRes.Proposal

	resp, err = testutil.GetRequest(fmt.Sprintf("%s/cosmos/group/v1/vote_by_proposal_voter/1/%s", val.APIAddress, val.Address.String()))
	s.Require().NoError(err)

	var voteRes group.QueryVoteByProposalVoterResponse
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(resp, &voteRes))
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

// createCLIProposal writes a CLI proposal with a MsgSend to a file. Returns
// the path to the JSON file.
func (s *E2ETestSuite) createCLIProposal(groupPolicyAddress, proposer, sendFrom, sendTo, metadata, title, summary string) string {
	_, err := base64.StdEncoding.DecodeString(metadata)
	s.Require().NoError(err)

	msg := banktypes.MsgSend{
		FromAddress: sendFrom,
		ToAddress:   sendTo,
		Amount:      sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(20))),
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

	resp, err := testutil.GetRequest(fmt.Sprintf("%s/cosmos/group/v1/group_policies_by_group/%s", val.APIAddress, groupID))
	s.Require().NoError(err)

	var res group.QueryGroupPoliciesByGroupResponse
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(resp, &res))
	groupPolicyAddress := res.GroupPolicies[0].Address

	addr, err := sdk.AccAddressFromBech32(groupPolicyAddress)
	s.Require().NoError(err)
	_, err = clitestutil.MsgSendExec(clientCtx, val.Address, addr,
		sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(tokens))),
		address.NewBech32Codec("cosmos"),
		s.commonFlags...,
	)
	s.Require().NoError(err)
	return groupPolicyAddress
}
