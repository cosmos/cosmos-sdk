package group

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
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
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	client "github.com/cosmos/cosmos-sdk/x/group/client/cli"
	"github.com/cosmos/cosmos-sdk/x/group/errors"
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

var tooLongMetadata = strings.Repeat("A", 256)

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

func (s *E2ETestSuite) TestTxCreateGroup() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

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
				s.commonFlags...,
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
				s.commonFlags...,
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
				s.commonFlags...,
			),
			false,
			"group metadata: limit exceeded",
			&sdk.TxResponse{},
			errors.ErrMaxLimit.ABCICode(),
		},
		{
			"invalid members address",
			append(
				[]string{
					val.Address.String(),
					"null",
					invalidMembersAddressFile.Name(),
				},
				s.commonFlags...,
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
				s.commonFlags...,
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
				s.commonFlags...,
			),
			false,
			"member metadata: limit exceeded",
			&sdk.TxResponse{},
			errors.ErrMaxLimit.ABCICode(),
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgCreateGroupCmd()

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp, err := clitestutil.GetTxResponse(s.network, clientCtx, tc.respType.(*sdk.TxResponse).TxHash)
				s.Require().NoError(err)
				s.Require().Equal(txResp.Code, tc.expectedCode)
				if tc.expectErrMsg != "" {
					s.Require().Contains(txResp.RawLog, tc.expectErrMsg)
				}
			}
		})
	}
}

func (s *E2ETestSuite) TestTxUpdateGroupAdmin() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	groupIDs := make([]string, 2)
	for i := 0; i < 2; i++ {
		validMembers := fmt.Sprintf(`{"members": [{
	  "address": "%s",
		"weight": "1",
		"metadata": "%s"
	}]}`, val.Address.String(), validMetadata)
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
		var txResp sdk.TxResponse
		s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
		txResp, err = clitestutil.GetTxResponse(s.network, val.ClientCtx, txResp.TxHash)
		s.Require().NoError(err)
		s.Require().Equal(txResp.Code, uint32(0), out.String())
		groupIDs[i] = s.getGroupIDFromTxResponse(txResp)
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
				s.commonFlags...,
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
				s.commonFlags...,
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
				s.commonFlags...,
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
				s.commonFlags...,
			),
			false,
			"not found",
			&sdk.TxResponse{},
			sdkerrors.ErrNotFound.ABCICode(),
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgUpdateGroupAdminCmd()

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp, err := clitestutil.GetTxResponse(s.network, clientCtx, tc.respType.(*sdk.TxResponse).TxHash)
				s.Require().NoError(err)
				s.Require().Equal(txResp.Code, tc.expectedCode)
				if tc.expectErrMsg != "" {
					s.Require().Contains(txResp.RawLog, tc.expectErrMsg)
				}
			}
		})
	}
}

func (s *E2ETestSuite) TestTxUpdateGroupMetadata() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

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
					"1",
					validMetadata,
				},
				s.commonFlags...,
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
					"1",
					validMetadata,
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				s.commonFlags...,
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
				s.commonFlags...,
			),
			false,
			"group metadata: limit exceeded",
			&sdk.TxResponse{},
			errors.ErrMaxLimit.ABCICode(),
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgUpdateGroupMetadataCmd()

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp, err := clitestutil.GetTxResponse(s.network, clientCtx, tc.respType.(*sdk.TxResponse).TxHash)
				s.Require().NoError(err)
				s.Require().Equal(txResp.Code, tc.expectedCode)
				if tc.expectErrMsg != "" {
					s.Require().Contains(txResp.RawLog, tc.expectErrMsg)
				}
			}
		})
	}
}

func (s *E2ETestSuite) TestTxUpdateGroupMembers() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	weights := []string{"1", "1", "1"}
	accounts := s.createAccounts(3)

	groupID := s.createGroupWithMembers(weights, accounts)
	groupPolicyAddress := s.createGroupThresholdPolicyWithBalance(accounts[0], groupID, 3, 100)

	validUpdatedMembersFileName := testutil.WriteToNewTempFile(s.T(), fmt.Sprintf(`{"members": [{
		"address": "%s",
		"weight": "0",
		"metadata": "%s"
	}, {
		"address": "%s",
		"weight": "1",
		"metadata": "%s"
	}]}`, accounts[0], validMetadata, groupPolicyAddress, validMetadata)).Name()

	invalidMembersMetadata := fmt.Sprintf(`{"members": [{
	  "address": "%s",
		"weight": "1",
		"metadata": "%s"
	}]}`, accounts[0], tooLongMetadata)
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
					accounts[0],
					groupID,
					validUpdatedMembersFileName,
				},
				s.commonFlags...,
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
					accounts[0],
					groupID,
					testutil.WriteToNewTempFile(s.T(), fmt.Sprintf(`{"members": [{
		"address": "%s",
		"weight": "2",
		"metadata": "%s"
	}]}`, s.groupPolicies[0].Address, validMetadata)).Name(),
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				s.commonFlags...,
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
					accounts[0],
					groupID,
					invalidMembersMetadataFileName,
				},
				s.commonFlags...,
			),
			false,
			"group member metadata: limit exceeded",
			&sdk.TxResponse{},
			errors.ErrMaxLimit.ABCICode(),
		},
		{
			"group doesn't exist",
			append(
				[]string{
					accounts[0],
					"12345",
					validUpdatedMembersFileName,
				},
				s.commonFlags...,
			),
			false,
			"not found",
			&sdk.TxResponse{},
			sdkerrors.ErrNotFound.ABCICode(),
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgUpdateGroupMembersCmd()

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp, err := clitestutil.GetTxResponse(s.network, clientCtx, tc.respType.(*sdk.TxResponse).TxHash)
				s.Require().NoError(err)
				s.Require().Equal(txResp.Code, tc.expectedCode)
				if tc.expectErrMsg != "" {
					s.Require().Contains(txResp.RawLog, tc.expectErrMsg)
				}
			}
		})
	}
}

func (s *E2ETestSuite) TestTxCreateGroupWithPolicy() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

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

	thresholdDecisionPolicyFile := testutil.WriteToNewTempFile(s.T(), `{"@type": "/cosmos.group.v1.ThresholdDecisionPolicy","threshold": "1","windows": {"voting_period":"1s"}}`)

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
					thresholdDecisionPolicyFile.Name(),
					fmt.Sprintf("--%s=%v", client.FlagGroupPolicyAsAdmin, false),
				},
				s.commonFlags...,
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
					thresholdDecisionPolicyFile.Name(),
					fmt.Sprintf("--%s=%v", client.FlagGroupPolicyAsAdmin, true),
				},
				s.commonFlags...,
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
					thresholdDecisionPolicyFile.Name(),
					fmt.Sprintf("--%s=%v", client.FlagGroupPolicyAsAdmin, false),
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				s.commonFlags...,
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
					thresholdDecisionPolicyFile.Name(),
					fmt.Sprintf("--%s=%v", client.FlagGroupPolicyAsAdmin, false),
				},
				s.commonFlags...,
			),
			false,
			"group metadata: limit exceeded",
			&sdk.TxResponse{},
			errors.ErrMaxLimit.ABCICode(),
		},
		{
			"group policy metadata too long",
			append(
				[]string{
					val.Address.String(),
					validMetadata,
					strings.Repeat("a", 256),
					validMembersFile.Name(),
					thresholdDecisionPolicyFile.Name(),
					fmt.Sprintf("--%s=%v", client.FlagGroupPolicyAsAdmin, false),
				},
				s.commonFlags...,
			),
			false,
			"group policy metadata: limit exceeded",
			&sdk.TxResponse{},
			errors.ErrMaxLimit.ABCICode(),
		},
		{
			"invalid members address",
			append(
				[]string{
					val.Address.String(),
					validMetadata,
					validMetadata,
					invalidMembersAddressFile.Name(),
					thresholdDecisionPolicyFile.Name(),
					fmt.Sprintf("--%s=%v", client.FlagGroupPolicyAsAdmin, false),
				},
				s.commonFlags...,
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
					thresholdDecisionPolicyFile.Name(),
					fmt.Sprintf("--%s=%v", client.FlagGroupPolicyAsAdmin, false),
				},
				s.commonFlags...,
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
					thresholdDecisionPolicyFile.Name(),
					fmt.Sprintf("--%s=%v", client.FlagGroupPolicyAsAdmin, false),
				},
				s.commonFlags...,
			),
			false,
			"member metadata: limit exceeded",
			&sdk.TxResponse{},
			errors.ErrMaxLimit.ABCICode(),
		},
	}
	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgCreateGroupWithPolicyCmd()

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp, err := clitestutil.GetTxResponse(s.network, clientCtx, tc.respType.(*sdk.TxResponse).TxHash)
				s.Require().NoError(err)
				s.Require().Equal(txResp.Code, tc.expectedCode)
				if tc.expectErrMsg != "" {
					s.Require().Contains(txResp.RawLog, tc.expectErrMsg)
				}
			}
		})
	}
}

func (s *E2ETestSuite) TestTxCreateGroupPolicy() {
	val := s.network.Validators[0]
	wrongAdmin := s.network.Validators[1].Address
	clientCtx := val.ClientCtx

	groupID := s.group.Id

	thresholdDecisionPolicyFile := testutil.WriteToNewTempFile(s.T(), `{"@type": "/cosmos.group.v1.ThresholdDecisionPolicy","threshold": "1","windows": {"voting_period":"1s"}}`)

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
					thresholdDecisionPolicyFile.Name(),
				},
				s.commonFlags...,
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
					testutil.WriteToNewTempFile(s.T(), `{"@type":"/cosmos.group.v1.PercentageDecisionPolicy", "percentage":"0.5", "windows":{"voting_period":"1s"}}`).Name(),
				},
				s.commonFlags...,
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
					thresholdDecisionPolicyFile.Name(),
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				s.commonFlags...,
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
					thresholdDecisionPolicyFile.Name(),
				},
				s.commonFlags...,
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
					thresholdDecisionPolicyFile.Name(),
				},
				s.commonFlags...,
			),
			false,
			"group policy metadata: limit exceeded",
			&sdk.TxResponse{},
			errors.ErrMaxLimit.ABCICode(),
		},
		{
			"wrong group id",
			append(
				[]string{
					val.Address.String(),
					"10",
					validMetadata,
					thresholdDecisionPolicyFile.Name(),
				},
				s.commonFlags...,
			),
			false,
			"not found",
			&sdk.TxResponse{},
			sdkerrors.ErrNotFound.ABCICode(),
		},
		{
			"invalid percentage decision policy with negative value",
			append(
				[]string{
					val.Address.String(),
					fmt.Sprintf("%v", groupID),
					validMetadata,
					testutil.WriteToNewTempFile(s.T(), `{"@type":"/cosmos.group.v1.PercentageDecisionPolicy", "percentage":"-0.5", "windows":{"voting_period":"1s"}}`).Name(),
				},
				s.commonFlags...,
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
					testutil.WriteToNewTempFile(s.T(), `{"@type":"/cosmos.group.v1.PercentageDecisionPolicy", "percentage":"2", "windows":{"voting_period":"1s"}}`).Name(),
				},
				s.commonFlags...,
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

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp, err := clitestutil.GetTxResponse(s.network, clientCtx, tc.respType.(*sdk.TxResponse).TxHash)
				s.Require().NoError(err)
				s.Require().Equal(txResp.Code, tc.expectedCode)
				if tc.expectErrMsg != "" {
					s.Require().Contains(txResp.RawLog, tc.expectErrMsg)
				}
			}
		})
	}
}

func (s *E2ETestSuite) TestTxUpdateGroupPolicyAdmin() {
	val := s.network.Validators[0]
	newAdmin := s.network.Validators[1].Address
	clientCtx := val.ClientCtx
	groupPolicy := s.groupPolicies[3]

	commonFlags := s.commonFlags
	commonFlags = append(commonFlags, fmt.Sprintf("--%s=%d", flags.FlagGas, 300000))

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
			false,
			"load group policy: not found",
			&sdk.TxResponse{},
			sdkerrors.ErrNotFound.ABCICode(),
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgUpdateGroupPolicyAdminCmd()

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().NoError(clitestutil.CheckTxCode(s.network, clientCtx, txResp.TxHash, tc.expectedCode))
			}
		})
	}
}

func (s *E2ETestSuite) TestTxUpdateGroupPolicyDecisionPolicy() {
	val := s.network.Validators[0]
	newAdmin := s.network.Validators[1].Address
	clientCtx := val.ClientCtx
	groupPolicy := s.groupPolicies[2]

	commonFlags := s.commonFlags
	commonFlags = append(commonFlags, fmt.Sprintf("--%s=%d", flags.FlagGas, 300000))

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
					testutil.WriteToNewTempFile(s.T(), `{"@type":"/cosmos.group.v1.ThresholdDecisionPolicy", "threshold":"1", "windows":{"voting_period":"40000s"}}`).Name(),
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
					testutil.WriteToNewTempFile(s.T(), `{"@type":"/cosmos.group.v1.PercentageDecisionPolicy", "percentage":"0.5", "windows":{"voting_period":"40000s"}}`).Name(),
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
					testutil.WriteToNewTempFile(s.T(), `{"@type":"/cosmos.group.v1.ThresholdDecisionPolicy", "threshold":"1", "windows":{"voting_period":"50000s"}}`).Name(),
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
					testutil.WriteToNewTempFile(s.T(), `{"@type":"/cosmos.group.v1.ThresholdDecisionPolicy", "threshold":"1", "windows":{"voting_period":"1s"}}`).Name(),
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
					testutil.WriteToNewTempFile(s.T(), `{"@type":"/cosmos.group.v1.ThresholdDecisionPolicy", "threshold":"1", "windows":{"voting_period":"1s"}}`).Name(),
				},
				commonFlags...,
			),
			false,
			"load group policy: not found",
			&sdk.TxResponse{},
			sdkerrors.ErrNotFound.ABCICode(),
		},
		{
			"invalid percentage decision policy with negative value",
			append(
				[]string{
					groupPolicy.Admin,
					groupPolicy.Address,
					testutil.WriteToNewTempFile(s.T(), `{"@type":"/cosmos.group.v1.PercentageDecisionPolicy", "percentage":"-0.5", "windows":{"voting_period":"1s"}}`).Name(),
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
					testutil.WriteToNewTempFile(s.T(), `{"@type":"/cosmos.group.v1.PercentageDecisionPolicy", "percentage":"2", "windows":{"voting_period":"40000s"}}`).Name(),
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

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().NoError(clitestutil.CheckTxCode(s.network, clientCtx, txResp.TxHash, tc.expectedCode))
			}
		})
	}
}

func (s *E2ETestSuite) TestTxUpdateGroupPolicyMetadata() {
	val := s.network.Validators[0]
	newAdmin := s.network.Validators[1].Address
	clientCtx := val.ClientCtx
	groupPolicy := s.groupPolicies[2]

	commonFlags := s.commonFlags
	commonFlags = append(commonFlags, fmt.Sprintf("--%s=%d", flags.FlagGas, 300000))

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
			false,
			"group policy metadata: limit exceeded",
			&sdk.TxResponse{},
			errors.ErrMaxLimit.ABCICode(),
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
			false,
			"load group policy: not found",
			&sdk.TxResponse{},
			sdkerrors.ErrNotFound.ABCICode(),
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgUpdateGroupPolicyMetadataCmd()

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp, err := clitestutil.GetTxResponse(s.network, clientCtx, tc.respType.(*sdk.TxResponse).TxHash)
				s.Require().NoError(err)
				s.Require().Equal(txResp.Code, tc.expectedCode)
				if tc.expectErrMsg != "" {
					s.Require().Contains(txResp.RawLog, tc.expectErrMsg)
				}
			}
		})
	}
}

func (s *E2ETestSuite) TestTxSubmitProposal() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

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
						"title", "summary",
					),
				},
				s.commonFlags...,
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
						"title", "summary",
					),
					fmt.Sprintf("--%s=try", client.FlagExec),
				},
				s.commonFlags...,
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
						"", "title", "summary"),
					fmt.Sprintf("--%s=try", client.FlagExec),
				},
				s.commonFlags...,
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
						"", "title", "summary",
					),
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				s.commonFlags...,
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
						tooLongMetadata, "title", "summary",
					),
				},
				s.commonFlags...,
			),
			false,
			"metadata: limit exceeded",
			&sdk.TxResponse{},
			errors.ErrMaxLimit.ABCICode(),
		},
		{
			"unauthorized msg",
			append(
				[]string{
					s.createCLIProposal(
						s.groupPolicies[0].Address, val.Address.String(),
						val.Address.String(), s.groupPolicies[0].Address,
						"", "title", "summary"),
				},
				s.commonFlags...,
			),
			false,
			"msg does not have group policy authorization",
			&sdk.TxResponse{},
			sdkerrors.ErrUnauthorized.ABCICode(),
		},
		{
			"invalid proposers",
			append(
				[]string{
					s.createCLIProposal(
						s.groupPolicies[0].Address, "invalid",
						s.groupPolicies[0].Address, val.Address.String(),
						"", "title", "summary",
					),
				},
				s.commonFlags...,
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
						"", "title", "summary",
					),
				},
				s.commonFlags...,
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
						"", "title", "summary",
					),
				},
				s.commonFlags...,
			),
			false,
			"group policy: not found",
			&sdk.TxResponse{},
			sdkerrors.ErrNotFound.ABCICode(),
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgSubmitProposalCmd()

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp, err := clitestutil.GetTxResponse(s.network, clientCtx, tc.respType.(*sdk.TxResponse).TxHash)
				s.Require().NoError(err)
				s.Require().Equal(txResp.Code, tc.expectedCode)
				if tc.expectErrMsg != "" {
					s.Require().Contains(txResp.RawLog, tc.expectErrMsg)
				}
			}
		})
	}
}

func (s *E2ETestSuite) TestTxVote() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	ids := make([]string, 4)
	weights := []string{"1", "1", "1"}
	accounts := s.createAccounts(3)

	groupID := s.createGroupWithMembers(weights, accounts)
	groupPolicyAddress := s.createGroupThresholdPolicyWithBalance(accounts[0], groupID, 3, 100)

	for i := 0; i < 4; i++ {
		out, err := clitestutil.ExecTestCLICmd(val.ClientCtx, client.MsgSubmitProposalCmd(),
			append(
				[]string{
					s.createCLIProposal(
						groupPolicyAddress, accounts[0],
						groupPolicyAddress, accounts[0],
						"", "title", "summary",
					),
				},
				s.commonFlags...,
			),
		)
		s.Require().NoError(err, out.String())

		var txResp sdk.TxResponse
		s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
		txResp, err = clitestutil.GetTxResponse(s.network, val.ClientCtx, txResp.TxHash)
		s.Require().NoError(err)
		s.Require().Equal(txResp.Code, uint32(0), out.String())
		ids[i] = s.getProposalIDFromTxResponse(txResp)
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
					accounts[0],
					"VOTE_OPTION_YES",
					"",
				},
				s.commonFlags...,
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
					accounts[0],
					"VOTE_OPTION_YES",
					"",
					fmt.Sprintf("--%s=try", client.FlagExec),
				},
				s.commonFlags...,
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
					accounts[0],
					"VOTE_OPTION_NO",
					"",
					fmt.Sprintf("--%s=try", client.FlagExec),
				},
				s.commonFlags...,
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
					accounts[0],
					"VOTE_OPTION_YES",
					"",
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				s.commonFlags...,
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
					accounts[0],
					"VOTE_OPTION_YES",
					"",
				},
				s.commonFlags...,
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
					accounts[0],
					"VOTE_OPTION_YES",
					"",
				},
				s.commonFlags...,
			),
			false,
			"proposal: not found",
			&sdk.TxResponse{},
			sdkerrors.ErrNotFound.ABCICode(),
		},
		{
			"metadata too long",
			append(
				[]string{
					"2",
					accounts[0],
					"VOTE_OPTION_YES",
					tooLongMetadata,
				},
				s.commonFlags...,
			),
			false,
			"metadata: limit exceeded",
			&sdk.TxResponse{},
			errors.ErrMaxLimit.ABCICode(),
		},
		{
			"invalid vote option",
			append(
				[]string{
					"2",
					accounts[0],
					"INVALID_VOTE_OPTION",
					"",
				},
				s.commonFlags...,
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

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp, err := clitestutil.GetTxResponse(s.network, clientCtx, tc.respType.(*sdk.TxResponse).TxHash)
				s.Require().NoError(err)
				s.Require().Equal(txResp.Code, tc.expectedCode)
				if tc.expectErrMsg != "" {
					s.Require().Contains(txResp.RawLog, tc.expectErrMsg)
				}
			}
		})
	}
}

func (s *E2ETestSuite) TestTxWithdrawProposal() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	ids := make([]string, 2)

	for i := 0; i < 2; i++ {
		out, err := clitestutil.ExecTestCLICmd(val.ClientCtx, client.MsgSubmitProposalCmd(),
			append(
				[]string{
					s.createCLIProposal(
						s.groupPolicies[1].Address, val.Address.String(),
						s.groupPolicies[1].Address, val.Address.String(),
						"", "title", "summary"),
				},
				s.commonFlags...,
			),
		)
		s.Require().NoError(err, out.String())

		var txResp sdk.TxResponse
		s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
		txResp, err = clitestutil.GetTxResponse(s.network, val.ClientCtx, txResp.TxHash)
		s.Require().NoError(err)
		s.Require().Equal(txResp.Code, uint32(0), out.String())
		ids[i] = s.getProposalIDFromTxResponse(txResp)
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
				s.commonFlags...,
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
				s.commonFlags...,
			),
			false,
			"cannot withdraw a proposal with the status of PROPOSAL_STATUS_WITHDRAWN",
			&sdk.TxResponse{},
			errors.ErrInvalid.ABCICode(),
		},
		{
			"proposal not found",
			append(
				[]string{
					"222",
					"wrongAdmin",
				},
				s.commonFlags...,
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
				s.commonFlags...,
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
				s.commonFlags...,
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

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp, err := clitestutil.GetTxResponse(s.network, clientCtx, tc.respType.(*sdk.TxResponse).TxHash)
				s.Require().NoError(err)
				s.Require().Equal(txResp.Code, tc.expectedCode)
				if tc.expectErrMsg != "" {
					s.Require().Contains(txResp.RawLog, tc.expectErrMsg)
				}
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

func (s *E2ETestSuite) TestTxExec() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	var proposalIDs []string
	// create proposals and vote
	for i := 0; i < 2; i++ {
		out, err := clitestutil.ExecTestCLICmd(val.ClientCtx, client.MsgSubmitProposalCmd(),
			append(
				[]string{
					s.createCLIProposal(
						s.groupPolicies[0].Address, val.Address.String(),
						s.groupPolicies[0].Address, val.Address.String(),
						"", "title", "summary",
					),
				},
				s.commonFlags...,
			),
		)
		s.Require().NoError(err, out.String())

		var txResp sdk.TxResponse
		s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
		txResp, err = clitestutil.GetTxResponse(s.network, clientCtx, txResp.TxHash)
		s.Require().NoError(err)
		s.Require().Equal(txResp.Code, uint32(0), out.String())
		proposalID := s.getProposalIDFromTxResponse(txResp)
		proposalIDs = append(proposalIDs, proposalID)

		out, err = clitestutil.ExecTestCLICmd(val.ClientCtx, client.MsgVoteCmd(),
			append(
				[]string{
					proposalID,
					val.Address.String(),
					"VOTE_OPTION_YES",
					"",
				},
				s.commonFlags...,
			),
		)
		s.Require().NoError(err, out.String())
		s.Require().NoError(s.network.WaitForNextBlock())
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
				s.commonFlags...,
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
				s.commonFlags...,
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
				s.commonFlags...,
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
				s.commonFlags...,
			),
			false,
			"proposal: not found",
			&sdk.TxResponse{},
			sdkerrors.ErrNotFound.ABCICode(),
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgExecCmd()

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp, err := clitestutil.GetTxResponse(s.network, clientCtx, tc.respType.(*sdk.TxResponse).TxHash)
				s.Require().NoError(err)
				s.Require().Equal(txResp.Code, tc.expectedCode)
				if tc.expectErrMsg != "" {
					s.Require().Contains(txResp.RawLog, tc.expectErrMsg)
				}
			}
		})
	}
}

func (s *E2ETestSuite) TestTxLeaveGroup() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	// create 3 accounts with some tokens
	members := s.createAccounts(3)

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
	out, err := clitestutil.ExecTestCLICmd(clientCtx, client.MsgCreateGroupCmd(),
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
	s.Require().NoError(s.network.WaitForNextBlock())

	var txResp sdk.TxResponse
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	txResp, err = clitestutil.GetTxResponse(s.network, val.ClientCtx, txResp.TxHash)
	s.Require().NoError(err)
	groupID := s.getGroupIDFromTxResponse(txResp)

	// create group policy
	out, err = clitestutil.ExecTestCLICmd(clientCtx, client.MsgCreateGroupPolicyCmd(),
		append(
			[]string{
				val.Address.String(),
				groupID,
				"AQ==",
				testutil.WriteToNewTempFile(s.T(), `{"@type":"/cosmos.group.v1.ThresholdDecisionPolicy", "threshold":"3", "windows":{"voting_period":"1s"}}`).Name(),
			},
			s.commonFlags...,
		),
	)
	s.Require().NoError(err, out.String())
	s.Require().NoError(s.network.WaitForNextBlock())

	out, err = clitestutil.ExecTestCLICmd(clientCtx, client.QueryGroupPoliciesByGroupCmd(), []string{groupID, fmt.Sprintf("--%s=json", flags.FlagOutput)})
	s.Require().NoError(err, out.String())
	s.Require().NotNil(out)
	var resp group.QueryGroupPoliciesByGroupResponse
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp))
	s.Require().Len(resp.GroupPolicies, 1)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		errMsg       string
		expectedCode uint32
	}{
		{
			"invalid member address",
			append(
				[]string{
					"address",
					groupID,
					fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				},
				s.commonFlags...,
			),
			true,
			"key not found",
			0,
		},
		{
			"group not found",
			append(
				[]string{
					members[0],
					"40",
					fmt.Sprintf("--%s=%s", flags.FlagFrom, members[0]),
				},
				s.commonFlags...,
			),
			false,
			"group: not found",
			sdkerrors.ErrNotFound.ABCICode(),
		},
		{
			"valid case",
			append(
				[]string{
					members[2],
					groupID,
					fmt.Sprintf("--%s=%s", flags.FlagFrom, members[2]),
				},
				s.commonFlags...,
			),
			false,
			"",
			0,
		},
		{
			"not part of group",
			append(
				[]string{
					members[2],
					groupID,
					fmt.Sprintf("--%s=%s", flags.FlagFrom, members[2]),
				},
				s.commonFlags...,
			),
			false,
			"is not part of group",
			sdkerrors.ErrNotFound.ABCICode(),
		},
		{
			"can leave group policy threshold is more than group weight",
			append(
				[]string{
					members[1],
					groupID,
					fmt.Sprintf("--%s=%s", flags.FlagFrom, members[1]),
				},
				s.commonFlags...,
			),
			false,
			"",
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgLeaveGroupCmd()
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.errMsg)
			} else {
				s.Require().NoError(err, out.String())
				var resp sdk.TxResponse
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
				s.Require().NoError(clitestutil.CheckTxCode(s.network, clientCtx, resp.TxHash, tc.expectedCode))
			}
		})
	}
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

func (s *E2ETestSuite) newValidMembers(weights, membersAddress []string) group.MemberRequests {
	s.Require().Equal(len(weights), len(membersAddress))
	membersValid := group.MemberRequests{}
	for i, address := range membersAddress {
		membersValid.Members = append(membersValid.Members, group.MemberRequest{
			Address:  address,
			Weight:   weights[i],
			Metadata: validMetadata,
		})
	}
	return membersValid
}
