package testutil

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	client "github.com/cosmos/cosmos-sdk/x/group/client/cli"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network

	group         *group.GroupInfo
	groupPolicies []*group.GroupPolicyInfo
	proposal      *group.Proposal
	vote          *group.Vote
	voter         *group.Member
}

const validMetadata = "metadata"

var tooLongMetadata = strings.Repeat("A", 256)

func NewIntegrationTestSuite(cfg network.Config) *IntegrationTestSuite {
	return &IntegrationTestSuite{cfg: cfg}
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)

	val := s.network.Validators[0]

	// create a new account
	info, _, err := val.ClientCtx.Keyring.NewMnemonic("NewValidator", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)

	pk, err := info.GetPubKey()
	s.Require().NoError(err)

	account := sdk.AccAddress(pk.Address())
	_, err = banktestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		account,
		sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(2000))), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	)
	s.Require().NoError(err)

	var commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

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
	out, err := cli.ExecTestCLICmd(val.ClientCtx, client.MsgCreateGroupCmd(),
		append(
			[]string{
				val.Address.String(),
				validMetadata,
				validMembersFile.Name(),
			},
			commonFlags...,
		),
	)

	s.Require().NoError(err, out.String())
	var txResp = sdk.TxResponse{}
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	s.Require().Equal(uint32(0), txResp.Code, out.String())

	s.group = &group.GroupInfo{Id: 1, Admin: val.Address.String(), Metadata: validMetadata, TotalWeight: "3", Version: 1}

	// create 5 group policies
	for i := 0; i < 5; i++ {
		threshold := i + 1
		if threshold > 3 {
			threshold = 3
		}
		out, err = cli.ExecTestCLICmd(val.ClientCtx, client.MsgCreateGroupPolicyCmd(),
			append(
				[]string{
					val.Address.String(),
					"1",
					validMetadata,
					fmt.Sprintf("{\"@type\":\"/cosmos.group.v1.ThresholdDecisionPolicy\", \"threshold\":\"%d\", \"windows\":{\"voting_period\":\"30000s\"}}", threshold),
				},
				commonFlags...,
			),
		)
		s.Require().NoError(err, out.String())
		s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
		s.Require().Equal(uint32(0), txResp.Code, out.String())

		out, err = cli.ExecTestCLICmd(val.ClientCtx, client.QueryGroupPoliciesByGroupCmd(), []string{"1", fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
		s.Require().NoError(err, out.String())
	}
	percentage := 0.5
	// create group policy with percentage decision policy
	out, err = cli.ExecTestCLICmd(val.ClientCtx, client.MsgCreateGroupPolicyCmd(),
		append(
			[]string{
				val.Address.String(),
				"1",
				validMetadata,
				fmt.Sprintf("{\"@type\":\"/cosmos.group.v1.PercentageDecisionPolicy\", \"percentage\":\"%f\", \"windows\":{\"voting_period\":\"30000s\"}}", percentage),
			},
			commonFlags...,
		),
	)
	s.Require().NoError(err, out.String())
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	s.Require().Equal(uint32(0), txResp.Code, out.String())

	out, err = cli.ExecTestCLICmd(val.ClientCtx, client.QueryGroupPoliciesByGroupCmd(), []string{"1", fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
	s.Require().NoError(err, out.String())

	var res group.QueryGroupPoliciesByGroupResponse
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))
	s.Require().Equal(len(res.GroupPolicies), 6)
	s.groupPolicies = res.GroupPolicies

	// create a proposal
	out, err = cli.ExecTestCLICmd(val.ClientCtx, client.MsgSubmitProposalCmd(),
		append(
			[]string{
				s.createCLIProposal(
					s.groupPolicies[0].Address, val.Address.String(),
					s.groupPolicies[0].Address, val.Address.String(),
					""),
			},
			commonFlags...,
		),
	)
	s.Require().NoError(err, out.String())
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	s.Require().Equal(uint32(0), txResp.Code, out.String())

	// vote
	out, err = cli.ExecTestCLICmd(val.ClientCtx, client.MsgVoteCmd(),
		append(
			[]string{
				"1",
				val.Address.String(),
				"VOTE_OPTION_YES",
				"",
			},
			commonFlags...,
		),
	)
	s.Require().NoError(err, out.String())
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	s.Require().Equal(uint32(0), txResp.Code, out.String())

	out, err = cli.ExecTestCLICmd(val.ClientCtx, client.QueryProposalCmd(), []string{"1", fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
	s.Require().NoError(err, out.String())

	var proposalRes group.QueryProposalResponse
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &proposalRes))
	s.proposal = proposalRes.Proposal

	out, err = cli.ExecTestCLICmd(val.ClientCtx, client.QueryVoteByProposalVoterCmd(), []string{"1", val.Address.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
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

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestTxCreateGroup() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	var commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	validMembers := fmt.Sprintf(`{"members": [{
	  "address": "%s",
		"weight": "1",
		"metadata": "%s"
	}]}`, val.Address.String(), validMetadata)
	validMembersFile := testutil.WriteToNewTempFile(s.T(), validMembers)

	invalidMembersAddress := `{"members": [{
	"address": "",
	"weight": "1"
}]}`
	invalidMembersAddressFile := testutil.WriteToNewTempFile(s.T(), invalidMembersAddress)

	invalidMembersWeight := fmt.Sprintf(`{"members": [{
	  "address": "%s",
		"weight": "0"
	}]}`, val.Address.String())
	invalidMembersWeightFile := testutil.WriteToNewTempFile(s.T(), invalidMembersWeight)

	invalidMembersMetadata := fmt.Sprintf(`{"members": [{
	  "address": "%s",
		"weight": "1",
		"metadata": "%s"
	}]}`, val.Address.String(), tooLongMetadata)
	invalidMembersMetadataFile := testutil.WriteToNewTempFile(s.T(), invalidMembersMetadata)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"correct data",
			append(
				[]string{
					val.Address.String(),
					"",
					validMembersFile.Name(),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"with amino-json",
			append(
				[]string{
					val.Address.String(),
					"",
					validMembersFile.Name(),
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"group metadata too long",
			append(
				[]string{
					val.Address.String(),
					strings.Repeat("a", 256),
					"",
				},
				commonFlags...,
			),
			true,
			"group metadata: limit exceeded",
			nil,
			0,
		},
		{
			"invalid members address",
			append(
				[]string{
					val.Address.String(),
					"null",
					invalidMembersAddressFile.Name(),
				},
				commonFlags...,
			),
			true,
			"message validation failed: address: empty address string is not allowed",
			nil,
			0,
		},
		{
			"invalid members weight",
			append(
				[]string{
					val.Address.String(),
					"null",
					invalidMembersWeightFile.Name(),
				},
				commonFlags...,
			),
			true,
			"expected a positive decimal, got 0: invalid decimal string",
			nil,
			0,
		},
		{
			"members metadata too long",
			append(
				[]string{
					val.Address.String(),
					"null",
					invalidMembersMetadataFile.Name(),
				},
				commonFlags...,
			),
			true,
			"member metadata: limit exceeded",
			nil,
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgCreateGroupCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestTxUpdateGroupAdmin() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx
	require := s.Require()

	var commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	groupIDs := make([]string, 2)
	for i := 0; i < 2; i++ {
		validMembers := fmt.Sprintf(`{"members": [{
	  "address": "%s",
		"weight": "1",
		"metadata": "%s"
	}]}`, val.Address.String(), validMetadata)
		validMembersFile := testutil.WriteToNewTempFile(s.T(), validMembers)
		out, err := cli.ExecTestCLICmd(val.ClientCtx, client.MsgCreateGroupCmd(),
			append(
				[]string{
					val.Address.String(),
					validMetadata,
					validMembersFile.Name(),
				},
				commonFlags...,
			),
		)
		require.NoError(err, out.String())
		var txResp sdk.TxResponse
		s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
		s.Require().Equal(uint32(0), txResp.Code, out.String())
		groupIDs[i] = s.getGroupIdFromTxResponse(txResp)
	}

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"correct data",
			append(
				[]string{
					val.Address.String(),
					groupIDs[0],
					s.network.Validators[1].Address.String(),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"with amino-json",
			append(
				[]string{
					val.Address.String(),
					groupIDs[1],
					s.network.Validators[1].Address.String(),
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"group id invalid",
			append(
				[]string{
					val.Address.String(),
					"",
					s.network.Validators[1].Address.String(),
				},
				commonFlags...,
			),
			true,
			"strconv.ParseUint: parsing \"\": invalid syntax",
			nil,
			0,
		},
		{
			"group doesn't exist",
			append(
				[]string{
					val.Address.String(),
					"12345",
					s.network.Validators[1].Address.String(),
				},
				commonFlags...,
			),
			true,
			"not found",
			nil,
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgUpdateGroupAdminCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				require.Contains(out.String(), tc.expectErrMsg)
			} else {
				require.NoError(err, out.String())
				require.NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				require.Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestTxUpdateGroupMetadata() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	var commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"correct data",
			append(
				[]string{
					val.Address.String(),
					"2",
					validMetadata,
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"with amino-json",
			append(
				[]string{
					val.Address.String(),
					"2",
					validMetadata,
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"group metadata too long",
			append(
				[]string{
					val.Address.String(),
					strconv.FormatUint(s.group.Id, 10),
					strings.Repeat("a", 256),
				},
				commonFlags...,
			),
			true,
			"group metadata: limit exceeded",
			nil,
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgUpdateGroupMetadataCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestTxUpdateGroupMembers() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	var commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	validUpdatedMembersFileName := testutil.WriteToNewTempFile(s.T(), fmt.Sprintf(`{"members": [{
		"address": "%s",
		"weight": "0",
		"metadata": "%s"
	}, {
		"address": "%s",
		"weight": "1",
		"metadata": "%s"
	}]}`, val.Address.String(), validMetadata, s.groupPolicies[0].Address, validMetadata)).Name()

	invalidMembersMetadata := fmt.Sprintf(`{"members": [{
	  "address": "%s",
		"weight": "1",
		"metadata": "%s"
	}]}`, val.Address.String(), tooLongMetadata)
	invalidMembersMetadataFileName := testutil.WriteToNewTempFile(s.T(), invalidMembersMetadata).Name()

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"correct data",
			append(
				[]string{
					val.Address.String(),
					"2",
					validUpdatedMembersFileName,
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"with amino-json",
			append(
				[]string{
					val.Address.String(),
					"2",
					testutil.WriteToNewTempFile(s.T(), fmt.Sprintf(`{"members": [{
		"address": "%s",
		"weight": "2",
		"metadata": "%s"
	}]}`, s.groupPolicies[0].Address, validMetadata)).Name(),
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"group member metadata too long",
			append(
				[]string{
					val.Address.String(),
					strconv.FormatUint(s.group.Id, 10),
					invalidMembersMetadataFileName,
				},
				commonFlags...,
			),
			true,
			"group member metadata: limit exceeded",
			nil,
			0,
		},
		{
			"group doesn't exist",
			append(
				[]string{
					val.Address.String(),
					"12345",
					validUpdatedMembersFileName,
				},
				commonFlags...,
			),
			true,
			"not found",
			nil,
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgUpdateGroupMembersCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestTxCreateGroupWithPolicy() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	var commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	validMembers := fmt.Sprintf(`{"members": [{
		"address": "%s",
		  "weight": "1",
		  "metadata": "%s"
	}]}`, val.Address.String(), validMetadata)
	validMembersFile := testutil.WriteToNewTempFile(s.T(), validMembers)

	invalidMembersAddress := `{"members": [{
	  "address": "",
	  "weight": "1"
	}]}`
	invalidMembersAddressFile := testutil.WriteToNewTempFile(s.T(), invalidMembersAddress)

	invalidMembersWeight := fmt.Sprintf(`{"members": [{
		"address": "%s",
		  "weight": "0"
	}]}`, val.Address.String())
	invalidMembersWeightFile := testutil.WriteToNewTempFile(s.T(), invalidMembersWeight)

	invalidMembersMetadata := fmt.Sprintf(`{"members": [{
		"address": "%s",
		  "weight": "1",
		  "metadata": "%s"
	}]}`, val.Address.String(), tooLongMetadata)
	invalidMembersMetadataFile := testutil.WriteToNewTempFile(s.T(), invalidMembersMetadata)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"correct data",
			append(
				[]string{
					val.Address.String(),
					validMetadata,
					validMetadata,
					validMembersFile.Name(),
					"{\"@type\":\"/cosmos.group.v1.ThresholdDecisionPolicy\", \"threshold\":\"1\", \"windows\":{\"voting_period\":\"1s\"}}",
					fmt.Sprintf("--%s=%v", client.FlagGroupPolicyAsAdmin, false),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"group-policy-as-admin is true",
			append(
				[]string{
					val.Address.String(),
					validMetadata,
					validMetadata,
					validMembersFile.Name(),
					"{\"@type\":\"/cosmos.group.v1.ThresholdDecisionPolicy\", \"threshold\":\"1\", \"windows\":{\"voting_period\":\"1s\"}}",
					fmt.Sprintf("--%s=%v", client.FlagGroupPolicyAsAdmin, true),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"with amino-json",
			append(
				[]string{
					val.Address.String(),
					validMetadata,
					validMetadata,
					validMembersFile.Name(),
					"{\"@type\":\"/cosmos.group.v1.ThresholdDecisionPolicy\", \"threshold\":\"1\", \"windows\":{\"voting_period\":\"1s\"}}",
					fmt.Sprintf("--%s=%v", client.FlagGroupPolicyAsAdmin, false),
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"group metadata too long",
			append(
				[]string{
					val.Address.String(),
					strings.Repeat("a", 256),
					validMetadata,
					validMembersFile.Name(),
					"{\"@type\":\"/cosmos.group.v1.ThresholdDecisionPolicy\", \"threshold\":\"1\", \"windows\":{\"voting_period\":\"1s\"}}",
					fmt.Sprintf("--%s=%v", client.FlagGroupPolicyAsAdmin, false),
				},
				commonFlags...,
			),
			true,
			"group metadata: limit exceeded",
			nil,
			0,
		},
		{
			"group policy metadata too long",
			append(
				[]string{
					val.Address.String(),
					validMetadata,
					strings.Repeat("a", 256),
					validMembersFile.Name(),
					"{\"@type\":\"/cosmos.group.v1.ThresholdDecisionPolicy\", \"threshold\":\"1\", \"windows\":{\"voting_period\":\"1s\"}}",
					fmt.Sprintf("--%s=%v", client.FlagGroupPolicyAsAdmin, false),
				},
				commonFlags...,
			),
			true,
			"group policy metadata: limit exceeded",
			nil,
			0,
		},
		{
			"invalid members address",
			append(
				[]string{
					val.Address.String(),
					validMetadata,
					validMetadata,
					invalidMembersAddressFile.Name(),
					"{\"@type\":\"/cosmos.group.v1.ThresholdDecisionPolicy\", \"threshold\":\"1\", \"windows\":{\"voting_period\":\"1s\"}}",
					fmt.Sprintf("--%s=%v", client.FlagGroupPolicyAsAdmin, false),
				},
				commonFlags...,
			),
			true,
			"message validation failed: address: empty address string is not allowed",
			nil,
			0,
		},
		{
			"invalid members weight",
			append(
				[]string{
					val.Address.String(),
					validMetadata,
					validMetadata,
					invalidMembersWeightFile.Name(),
					"{\"@type\":\"/cosmos.group.v1.ThresholdDecisionPolicy\", \"threshold\":\"1\", \"windows\":{\"voting_period\":\"1s\"}}",
					fmt.Sprintf("--%s=%v", client.FlagGroupPolicyAsAdmin, false),
				},
				commonFlags...,
			),
			true,
			"expected a positive decimal, got 0: invalid decimal string",
			nil,
			0,
		},
		{
			"members metadata too long",
			append(
				[]string{
					val.Address.String(),
					validMetadata,
					validMetadata,
					invalidMembersMetadataFile.Name(),
					"{\"@type\":\"/cosmos.group.v1.ThresholdDecisionPolicy\", \"threshold\":\"1\", \"windows\":{\"voting_period\":\"1s\"}}",
					fmt.Sprintf("--%s=%v", client.FlagGroupPolicyAsAdmin, false),
				},
				commonFlags...,
			),
			true,
			"member metadata: limit exceeded",
			nil,
			0,
		},
	}
	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgCreateGroupWithPolicyCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestTxCreateGroupPolicy() {
	val := s.network.Validators[0]
	wrongAdmin := s.network.Validators[1].Address
	clientCtx := val.ClientCtx

	var commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	groupID := s.group.Id

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"correct data",
			append(
				[]string{
					val.Address.String(),
					fmt.Sprintf("%v", groupID),
					validMetadata,
					"{\"@type\":\"/cosmos.group.v1.ThresholdDecisionPolicy\", \"threshold\":\"1\", \"windows\":{\"voting_period\":\"1s\"}}",
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"correct data with percentage decision policy",
			append(
				[]string{
					val.Address.String(),
					fmt.Sprintf("%v", groupID),
					validMetadata,
					"{\"@type\":\"/cosmos.group.v1.PercentageDecisionPolicy\", \"percentage\":\"0.5\", \"windows\":{\"voting_period\":\"1s\"}}",
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"with amino-json",
			append(
				[]string{
					val.Address.String(),
					fmt.Sprintf("%v", groupID),
					validMetadata,
					"{\"@type\":\"/cosmos.group.v1.ThresholdDecisionPolicy\", \"threshold\":\"1\", \"windows\":{\"voting_period\":\"1s\"}}",
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"wrong admin",
			append(
				[]string{
					wrongAdmin.String(),
					fmt.Sprintf("%v", groupID),
					validMetadata,
					"{\"@type\":\"/cosmos.group.v1.ThresholdDecisionPolicy\", \"threshold\":\"1\", \"windows\":{\"voting_period\":\"1s\"}}",
				},
				commonFlags...,
			),
			true,
			"key not found",
			&sdk.TxResponse{},
			0,
		},
		{
			"metadata too long",
			append(
				[]string{
					val.Address.String(),
					fmt.Sprintf("%v", groupID),
					strings.Repeat("a", 500),
					"{\"@type\":\"/cosmos.group.v1.ThresholdDecisionPolicy\", \"threshold\":\"1\", \"windows\":{\"voting_period\":\"1s\"}}",
				},
				commonFlags...,
			),
			true,
			"group policy metadata: limit exceeded",
			&sdk.TxResponse{},
			0,
		},
		{
			"wrong group id",
			append(
				[]string{
					val.Address.String(),
					"10",
					validMetadata,
					"{\"@type\":\"/cosmos.group.v1.ThresholdDecisionPolicy\", \"threshold\":\"1\", \"windows\":{\"voting_period\":\"1s\"}}",
				},
				commonFlags...,
			),
			true,
			"not found",
			&sdk.TxResponse{},
			0,
		},
		{
			"invalid percentage decision policy with negative value",
			append(
				[]string{
					val.Address.String(),
					fmt.Sprintf("%v", groupID),
					validMetadata,
					"{\"@type\":\"/cosmos.group.v1.PercentageDecisionPolicy\", \"percentage\":\"-0.5\", \"windows\":{\"voting_period\":\"1s\"}}",
				},
				commonFlags...,
			),
			true,
			"expected a positive decimal",
			&sdk.TxResponse{},
			0,
		},
		{
			"invalid percentage decision policy with value greater than 1",
			append(
				[]string{
					val.Address.String(),
					fmt.Sprintf("%v", groupID),
					validMetadata,
					"{\"@type\":\"/cosmos.group.v1.PercentageDecisionPolicy\", \"percentage\":\"2\", \"windows\":{\"voting_period\":\"1s\"}}",
				},
				commonFlags...,
			),
			true,
			"percentage must be > 0 and <= 1",
			&sdk.TxResponse{},
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgCreateGroupPolicyCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestTxUpdateGroupPolicyAdmin() {
	val := s.network.Validators[0]
	newAdmin := s.network.Validators[1].Address
	clientCtx := val.ClientCtx
	groupPolicy := s.groupPolicies[3]

	var commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=%d", flags.FlagGas, 300000),
	}

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"correct data",
			append(
				[]string{
					groupPolicy.Admin,
					groupPolicy.Address,
					newAdmin.String(),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"with amino-json",
			append(
				[]string{
					groupPolicy.Admin,
					s.groupPolicies[4].Address,
					newAdmin.String(),
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"wrong admin",
			append(
				[]string{
					newAdmin.String(),
					groupPolicy.Address,
					newAdmin.String(),
				},
				commonFlags...,
			),
			true,
			"key not found",
			&sdk.TxResponse{},
			0,
		},
		{
			"wrong group policy",
			append(
				[]string{
					groupPolicy.Admin,
					newAdmin.String(),
					newAdmin.String(),
				},
				commonFlags...,
			),
			true,
			"load group policy: not found",
			&sdk.TxResponse{},
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgUpdateGroupPolicyAdminCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestTxUpdateGroupPolicyDecisionPolicy() {
	val := s.network.Validators[0]
	newAdmin := s.network.Validators[1].Address
	clientCtx := val.ClientCtx
	groupPolicy := s.groupPolicies[2]

	var commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=%d", flags.FlagGas, 300000),
	}

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"correct data",
			append(
				[]string{
					groupPolicy.Admin,
					groupPolicy.Address,
					"{\"@type\":\"/cosmos.group.v1.ThresholdDecisionPolicy\", \"threshold\":\"1\", \"windows\":{\"voting_period\":\"40000s\"}}",
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"correct data with percentage decision policy",
			append(
				[]string{
					groupPolicy.Admin,
					groupPolicy.Address,
					"{\"@type\":\"/cosmos.group.v1.PercentageDecisionPolicy\", \"percentage\":\"0.5\", \"windows\":{\"voting_period\":\"40000s\"}}",
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"with amino-json",
			append(
				[]string{
					groupPolicy.Admin,
					groupPolicy.Address,
					"{\"@type\":\"/cosmos.group.v1.ThresholdDecisionPolicy\", \"threshold\":\"1\", \"windows\":{\"voting_period\":\"50000s\"}}",
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"wrong admin",
			append(
				[]string{
					newAdmin.String(),
					groupPolicy.Address,
					"{\"@type\":\"/cosmos.group.v1.ThresholdDecisionPolicy\", \"threshold\":\"1\", \"windows\":{\"voting_period\":\"1s\"}}",
				},
				commonFlags...,
			),
			true,
			"key not found",
			&sdk.TxResponse{},
			0,
		},
		{
			"wrong group policy",
			append(
				[]string{
					groupPolicy.Admin,
					newAdmin.String(),
					"{\"@type\":\"/cosmos.group.v1.ThresholdDecisionPolicy\", \"threshold\":\"1\", \"windows\":{\"voting_period\":\"1s\"}}",
				},
				commonFlags...,
			),
			true,
			"load group policy: not found",
			&sdk.TxResponse{},
			0,
		},
		{
			"invalid percentage decision policy with negative value",
			append(
				[]string{
					groupPolicy.Admin,
					groupPolicy.Address,
					"{\"@type\":\"/cosmos.group.v1.PercentageDecisionPolicy\", \"percentage\":\"-0.5\", \"windows\":{\"voting_period\":\"1s\"}}",
				},
				commonFlags...,
			),
			true,
			"expected a positive decimal",
			&sdk.TxResponse{},
			0,
		},
		{
			"invalid percentage decision policy with value greater than 1",
			append(
				[]string{
					groupPolicy.Admin,
					groupPolicy.Address,
					"{\"@type\":\"/cosmos.group.v1.PercentageDecisionPolicy\", \"percentage\":\"2\", \"windows\":{\"voting_period\":\"40000s\"}}",
				},
				commonFlags...,
			),
			true,
			"percentage must be > 0 and <= 1",
			&sdk.TxResponse{},
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgUpdateGroupPolicyDecisionPolicyCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestTxUpdateGroupPolicyMetadata() {
	val := s.network.Validators[0]
	newAdmin := s.network.Validators[1].Address
	clientCtx := val.ClientCtx
	groupPolicy := s.groupPolicies[2]

	var commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=%d", flags.FlagGas, 300000),
	}

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"correct data",
			append(
				[]string{
					groupPolicy.Admin,
					groupPolicy.Address,
					validMetadata,
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"with amino-json",
			append(
				[]string{
					groupPolicy.Admin,
					groupPolicy.Address,
					validMetadata,
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"long metadata",
			append(
				[]string{
					groupPolicy.Admin,
					groupPolicy.Address,
					strings.Repeat("a", 500),
				},
				commonFlags...,
			),
			true,
			"group policy metadata: limit exceeded",
			&sdk.TxResponse{},
			0,
		},
		{
			"wrong admin",
			append(
				[]string{
					newAdmin.String(),
					groupPolicy.Address,
					validMetadata,
				},
				commonFlags...,
			),
			true,
			"key not found",
			&sdk.TxResponse{},
			0,
		},
		{
			"wrong group policy",
			append(
				[]string{
					groupPolicy.Admin,
					newAdmin.String(),
					validMetadata,
				},
				commonFlags...,
			),
			true,
			"load group policy: not found",
			&sdk.TxResponse{},
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgUpdateGroupPolicyMetadataCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestTxSubmitProposal() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	var commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"correct data",
			append(
				[]string{
					s.createCLIProposal(
						s.groupPolicies[0].Address, val.Address.String(),
						s.groupPolicies[0].Address, val.Address.String(),
						"",
					),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"with try exec",
			append(
				[]string{
					s.createCLIProposal(
						s.groupPolicies[0].Address, val.Address.String(),
						s.groupPolicies[0].Address, val.Address.String(),
						"",
					),
					fmt.Sprintf("--%s=try", client.FlagExec),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"with try exec, not enough yes votes for proposal to pass",
			append(
				[]string{
					s.createCLIProposal(
						s.groupPolicies[3].Address, val.Address.String(),
						s.groupPolicies[3].Address, val.Address.String(),
						""),
					fmt.Sprintf("--%s=try", client.FlagExec),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"with amino-json",
			append(
				[]string{
					s.createCLIProposal(
						s.groupPolicies[0].Address, val.Address.String(),
						s.groupPolicies[0].Address, val.Address.String(),
						"",
					),
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"metadata too long",
			append(
				[]string{
					s.createCLIProposal(
						s.groupPolicies[0].Address, val.Address.String(),
						s.groupPolicies[0].Address, val.Address.String(),
						tooLongMetadata,
					),
				},
				commonFlags...,
			),
			true,
			"metadata: limit exceeded",
			nil,
			0,
		},
		{
			"unauthorized msg",
			append(
				[]string{
					s.createCLIProposal(
						s.groupPolicies[0].Address, val.Address.String(),
						val.Address.String(), s.groupPolicies[0].Address,
						""),
				},
				commonFlags...,
			),
			true,
			"msg does not have group policy authorization",
			nil,
			0,
		},
		{
			"invalid proposers",
			append(
				[]string{
					s.createCLIProposal(
						s.groupPolicies[0].Address, "invalid",
						s.groupPolicies[0].Address, val.Address.String(),
						"",
					),
				},
				commonFlags...,
			),
			true,
			"invalid.info: key not found",
			nil,
			0,
		},
		{
			"invalid group policy",
			append(
				[]string{
					s.createCLIProposal(
						"invalid", val.Address.String(),
						s.groupPolicies[0].Address, val.Address.String(),
						"",
					),
				},
				commonFlags...,
			),
			true,
			"group policy: decoding bech32 failed",
			nil,
			0,
		},
		{
			"no group policy",
			append(
				[]string{
					s.createCLIProposal(
						val.Address.String(), val.Address.String(),
						s.groupPolicies[0].Address, val.Address.String(),
						"",
					),
				},
				commonFlags...,
			),
			true,
			"group policy: not found",
			nil,
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgSubmitProposalCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestTxVote() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	var commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	ids := make([]string, 4)

	for i := 0; i < 4; i++ {
		out, err := cli.ExecTestCLICmd(val.ClientCtx, client.MsgSubmitProposalCmd(),
			append(
				[]string{
					s.createCLIProposal(
						s.groupPolicies[1].Address, val.Address.String(),
						s.groupPolicies[1].Address, val.Address.String(),
						""),
				},
				commonFlags...,
			),
		)
		s.Require().NoError(err, out.String())

		var txResp sdk.TxResponse
		s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
		s.Require().Equal(uint32(0), txResp.Code, out.String())
		ids[i] = s.getProposalIdFromTxResponse(txResp)
	}

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"correct data",
			append(
				[]string{
					ids[0],
					val.Address.String(),
					"VOTE_OPTION_YES",
					"",
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"with try exec",
			append(
				[]string{
					ids[1],
					val.Address.String(),
					"VOTE_OPTION_YES",
					"",
					fmt.Sprintf("--%s=try", client.FlagExec),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"with try exec, not enough yes votes for proposal to pass",
			append(
				[]string{
					ids[2],
					val.Address.String(),
					"VOTE_OPTION_NO",
					"",
					fmt.Sprintf("--%s=try", client.FlagExec),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"with amino-json",
			append(
				[]string{
					ids[3],
					val.Address.String(),
					"VOTE_OPTION_YES",
					"",
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"invalid proposal id",
			append(
				[]string{
					"abcd",
					val.Address.String(),
					"VOTE_OPTION_YES",
					"",
				},
				commonFlags...,
			),
			true,
			"invalid syntax",
			nil,
			0,
		},
		{
			"proposal not found",
			append(
				[]string{
					"1234",
					val.Address.String(),
					"VOTE_OPTION_YES",
					"",
				},
				commonFlags...,
			),
			true,
			"proposal: not found",
			nil,
			0,
		},
		{
			"metadata too long",
			append(
				[]string{
					"2",
					val.Address.String(),
					"VOTE_OPTION_YES",
					tooLongMetadata,
				},
				commonFlags...,
			),
			true,
			"metadata: limit exceeded",
			nil,
			0,
		},
		{
			"invalid vote option",
			append(
				[]string{
					"2",
					val.Address.String(),
					"INVALID_VOTE_OPTION",
					"",
				},
				commonFlags...,
			),
			true,
			"not a valid vote option",
			nil,
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgVoteCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestTxWithdrawProposal() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	var commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	ids := make([]string, 2)

	for i := 0; i < 2; i++ {
		out, err := cli.ExecTestCLICmd(val.ClientCtx, client.MsgSubmitProposalCmd(),
			append(
				[]string{
					s.createCLIProposal(
						s.groupPolicies[1].Address, val.Address.String(),
						s.groupPolicies[1].Address, val.Address.String(),
						""),
				},
				commonFlags...,
			),
		)
		s.Require().NoError(err, out.String())

		var txResp sdk.TxResponse
		s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
		s.Require().Equal(uint32(0), txResp.Code, out.String())
		ids[i] = s.getProposalIdFromTxResponse(txResp)
	}

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"correct data",
			append(
				[]string{
					ids[0],
					val.Address.String(),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"already withdrawn proposal",
			append(
				[]string{
					ids[0],
					val.Address.String(),
				},
				commonFlags...,
			),
			true,
			"cannot withdraw a proposal with the status of PROPOSAL_STATUS_WITHDRAWN",
			&sdk.TxResponse{},
			0,
		},
		{
			"proposal not found",
			append(
				[]string{
					"222",
					"wrongAdmin",
				},
				commonFlags...,
			),
			true,
			"not found",
			&sdk.TxResponse{},
			0,
		},
		{
			"invalid proposal",
			append(
				[]string{
					"abc",
					val.Address.String(),
				},
				commonFlags...,
			),
			true,
			"invalid syntax",
			&sdk.TxResponse{},
			0,
		},
		{
			"wrong admin",
			append(
				[]string{
					ids[1],
					"wrongAdmin",
				},
				commonFlags...,
			),
			true,
			"key not found",
			&sdk.TxResponse{},
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgWithdrawProposalCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) getProposalIdFromTxResponse(txResp sdk.TxResponse) string {
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

func (s *IntegrationTestSuite) TestTxExec() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx
	require := s.Require()

	var commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	var proposalIDs []string
	// create proposals and vote
	for i := 0; i < 2; i++ {
		out, err := cli.ExecTestCLICmd(val.ClientCtx, client.MsgSubmitProposalCmd(),
			append(
				[]string{
					s.createCLIProposal(
						s.groupPolicies[0].Address, val.Address.String(),
						s.groupPolicies[0].Address, val.Address.String(),
						"",
					),
				},
				commonFlags...,
			),
		)
		require.NoError(err, out.String())

		var txResp sdk.TxResponse
		require.NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
		require.Equal(uint32(0), txResp.Code, out.String())
		proposalID := s.getProposalIdFromTxResponse(txResp)
		proposalIDs = append(proposalIDs, proposalID)

		out, err = cli.ExecTestCLICmd(val.ClientCtx, client.MsgVoteCmd(),
			append(
				[]string{
					proposalID,
					val.Address.String(),
					"VOTE_OPTION_YES",
					"",
				},
				commonFlags...,
			),
		)
		require.NoError(err, out.String())
	}

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"correct data",
			append(
				[]string{
					proposalIDs[0],
					fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"with amino-json",
			append(
				[]string{
					proposalIDs[1],
					fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"invalid proposal id",
			append(
				[]string{
					"abcd",
					fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				},
				commonFlags...,
			),
			true,
			"invalid syntax",
			nil,
			0,
		},
		{
			"proposal not found",
			append(
				[]string{
					"1234",
					fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				},
				commonFlags...,
			),
			true,
			"proposal: not found",
			nil,
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgExecCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				require.Contains(out.String(), tc.expectErrMsg)
			} else {
				require.NoError(err, out.String())
				require.NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				require.Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestTxLeaveGroup() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx
	require := s.Require()

	commonFlags := []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	// create 3 accounts with some tokens
	members := make([]string, 3)
	for i := 1; i <= 3; i++ {
		info, _, err := clientCtx.Keyring.NewMnemonic(fmt.Sprintf("member%d", i), keyring.English, sdk.FullFundraiserPath,
			keyring.DefaultBIP39Passphrase, hd.Secp256k1)
		require.NoError(err)

		pk, err := info.GetPubKey()
		require.NoError(err)

		account := sdk.AccAddress(pk.Address())
		members[i-1] = account.String()

		_, err = banktestutil.MsgSendExec(clientCtx, val.Address, account,
			sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(100))),
			commonFlags...,
		)
		require.NoError(err)
	}

	// create a group with three members
	validMembers := fmt.Sprintf(`{"members": [{
		"address": "%s",
		  "weight": "1",
		  "metadata": "AQ=="
	  },{
		"address": "%s",
		  "weight": "2",
		  "metadata": "AQ=="
	  },{
		"address": "%s",
		  "weight": "2",
		  "metadata": "AQ=="
	  }]}`, members[0], members[1], members[2])
	validMembersFile := testutil.WriteToNewTempFile(s.T(), validMembers)
	out, err := cli.ExecTestCLICmd(clientCtx, client.MsgCreateGroupCmd(),
		append(
			[]string{
				val.Address.String(),
				validMetadata,
				validMembersFile.Name(),
			},
			commonFlags...,
		),
	)
	require.NoError(err, out.String())
	var txResp sdk.TxResponse
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	groupID := s.getGroupIdFromTxResponse(txResp)

	// create group policy
	out, err = cli.ExecTestCLICmd(clientCtx, client.MsgCreateGroupPolicyCmd(),
		append(
			[]string{
				val.Address.String(),
				groupID,
				"AQ==",
				"{\"@type\":\"/cosmos.group.v1.ThresholdDecisionPolicy\", \"threshold\":\"3\", \"windows\":{\"voting_period\":\"1s\"}}",
			},
			commonFlags...,
		),
	)
	require.NoError(err, out.String())

	out, err = cli.ExecTestCLICmd(clientCtx, client.QueryGroupPoliciesByGroupCmd(), []string{groupID, fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
	require.NoError(err, out.String())
	require.NotNil(out)
	var resp group.QueryGroupPoliciesByGroupResponse
	require.NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp))
	require.Len(resp.GroupPolicies, 1)

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		errMsg    string
	}{
		{
			"invalid member address",
			append(
				[]string{
					"address",
					groupID,
					fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				},
				commonFlags...,
			),
			true,
			"key not found",
		},
		{
			"group not found",
			append(
				[]string{
					members[0],
					"40",
					fmt.Sprintf("--%s=%s", flags.FlagFrom, members[0]),
				},
				commonFlags...,
			),
			true,
			"group: not found",
		},
		{
			"valid case",
			append(
				[]string{
					members[2],
					groupID,
					fmt.Sprintf("--%s=%s", flags.FlagFrom, members[2]),
				},
				commonFlags...,
			),
			false,
			"",
		},
		{
			"not part of group",
			append(
				[]string{
					members[2],
					groupID,
					fmt.Sprintf("--%s=%s", flags.FlagFrom, members[2]),
				},
				commonFlags...,
			),
			true,
			"is not part of group",
		},
		{
			"can leave group policy threshold is more than group weight",
			append(
				[]string{
					members[1],
					groupID,
					fmt.Sprintf("--%s=%s", flags.FlagFrom, members[1]),
				},
				commonFlags...,
			),
			false,
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgLeaveGroupCmd()
			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				require.Contains(out.String(), tc.errMsg)
			} else {
				require.NoError(err, out.String())
				var resp sdk.TxResponse
				require.NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) getGroupIdFromTxResponse(txResp sdk.TxResponse) string {
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
func (s *IntegrationTestSuite) createCLIProposal(groupPolicyAddress, proposer, sendFrom, sendTo, metadata string) string {
	_, err := base64.StdEncoding.DecodeString(metadata)
	s.Require().NoError(err)

	msg := banktypes.MsgSend{
		FromAddress: sendFrom,
		ToAddress:   sendTo,
		Amount:      sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))),
	}
	msgJSON, err := s.cfg.Codec.MarshalInterfaceJSON(&msg)
	s.Require().NoError(err)

	p := client.CLIProposal{
		GroupPolicyAddress: groupPolicyAddress,
		Messages:           []json.RawMessage{msgJSON},
		Metadata:           metadata,
		Proposers:          []string{proposer},
	}

	bz, err := json.Marshal(&p)
	s.Require().NoError(err)

	return testutil.WriteToNewTempFile(s.T(), string(bz)).Name()
}
