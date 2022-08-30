package cli_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	tmcli "github.com/tendermint/tendermint/libs/cli"

	"strings"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/cli"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	"github.com/cosmos/cosmos-sdk/x/group"
	groupcli "github.com/cosmos/cosmos-sdk/x/group/client/cli"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	rpcclientmock "github.com/tendermint/tendermint/rpc/client/mock"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil"
	testutilmod "github.com/cosmos/cosmos-sdk/types/module/testutil"
	groupmodule "github.com/cosmos/cosmos-sdk/x/group/module"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

const validMetadata = "metadata"

var tooLongMetadata = strings.Repeat("A", 256)

var _ client.TendermintRPC = (*mockTendermintRPC)(nil)

type mockTendermintRPC struct {
	rpcclientmock.Client

	responseQuery abci.ResponseQuery
}

func newMockTendermintRPC(respQuery abci.ResponseQuery) mockTendermintRPC {
	return mockTendermintRPC{responseQuery: respQuery}
}

func (_ mockTendermintRPC) BroadcastTxCommit(_ context.Context, _ tmtypes.Tx) (*coretypes.ResultBroadcastTxCommit, error) {
	return &coretypes.ResultBroadcastTxCommit{}, nil
}

// func (m mockTendermintRPC) BroadcastTxSync(context.Context, tmtypes.Tx) (*coretypes.ResultBroadcastTx, error) {
// 	return &coretypes.ResultBroadcastTx{Code: 0}, nil
// }

func (m mockTendermintRPC) ABCIQueryWithOptions(
	_ context.Context,
	_ string, _ tmbytes.HexBytes,
	_ rpcclient.ABCIQueryOptions,
) (*coretypes.ResultABCIQuery, error) {
	return &coretypes.ResultABCIQuery{Response: m.responseQuery}, nil
}

type CLITestSuite struct {
	suite.Suite

	kr      keyring.Keyring
	encCfg  testutilmod.TestEncodingConfig
	baseCtx client.Context

	group         *group.GroupInfo
	groupPolicies []*group.GroupPolicyInfo
	proposal      *group.Proposal
	vote          *group.Vote
	voter         *group.Member
	commonFlags   []string
}

func TestCLITestSuite(t *testing.T) {
	suite.Run(t, new(CLITestSuite))
}

func (s *CLITestSuite) SetupSuite() {
	s.encCfg = testutilmod.MakeTestEncodingConfig(groupmodule.AppModuleBasic{})
	s.kr = keyring.NewInMemory(s.encCfg.Codec)
	s.baseCtx = client.Context{}.
		WithKeyring(s.kr).
		WithTxConfig(s.encCfg.TxConfig).
		WithCodec(s.encCfg.Codec).
		WithClient(mockTendermintRPC{Client: rpcclientmock.Client{}}).
		WithAccountRetriever(client.MockAccountRetriever{}).
		WithOutput(io.Discard).
		WithChainID("test-chain")

	s.commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
	}

	accounts := testutil.CreateKeyringAccounts(s.T(), s.kr, 1)
	val := accounts[0]

	var outBuf bytes.Buffer
	ctxGen := func() client.Context {
		bz, _ := s.encCfg.Codec.Marshal(&sdk.TxResponse{})
		c := newMockTendermintRPC(abci.ResponseQuery{
			Value: bz,
		})
		return s.baseCtx.WithClient(c)
	}
	clientCtx := ctxGen().WithOutput(&outBuf)
	//

	// ctx := svrcmd.CreateExecuteContext(context.Background())

	// create a new account
	info, _, err := clientCtx.Keyring.NewMnemonic("NewValidator", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)

	pk, err := info.GetPubKey()
	s.Require().NoError(err)

	account := sdk.AccAddress(pk.Address())
	_, err = cli.MsgSendExec(
		clientCtx,
		val.Address,
		account,
		sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(2000))), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(10))).String()),
	)
	s.Require().NoError(err)

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
	out, err := cli.ExecTestCLICmd(clientCtx, groupcli.MsgCreateGroupCmd(),
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
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	s.Require().Equal(uint32(0), txResp.Code, out.String())

	s.group = &group.GroupInfo{Id: 1, Admin: val.Address.String(), Metadata: validMetadata, TotalWeight: "3", Version: 1}

	fmt.Printf("s.group: %v\n", s.group)
	// create 5 group policies
	for i := 0; i < 5; i++ {
		threshold := i + 1
		if threshold > 3 {
			threshold = 3
		}

		s.createGroupThresholdPolicyWithBalance(val.Address.String(), "1", threshold, 1000)

		out, err = cli.ExecTestCLICmd(clientCtx, groupcli.QueryGroupPoliciesByGroupCmd(), []string{"1", fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
		s.Require().NoError(err, out.String())
	}
	percentage := 0.5
	// create group policy with percentage decision policy
	out, err = cli.ExecTestCLICmd(clientCtx, groupcli.MsgCreateGroupPolicyCmd(),
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
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	s.Require().Equal(uint32(0), txResp.Code, out.String())

	out, err = cli.ExecTestCLICmd(clientCtx, groupcli.QueryGroupPoliciesByGroupCmd(), []string{"1", fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
	s.Require().NoError(err, out.String())

	var res group.QueryGroupPoliciesByGroupResponse
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))
	s.Require().Equal(len(res.GroupPolicies), 6)
	s.groupPolicies = res.GroupPolicies

	// create a proposal
	out, err = cli.ExecTestCLICmd(clientCtx, groupcli.MsgSubmitProposalCmd(),
		append(
			[]string{
				s.createCLIProposal(
					s.groupPolicies[0].Address, val.Address.String(),
					s.groupPolicies[0].Address, val.Address.String(),
					""),
			},
			s.commonFlags...,
		),
	)
	s.Require().NoError(err, out.String())
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	s.Require().Equal(uint32(0), txResp.Code, out.String())

	// vote
	out, err = cli.ExecTestCLICmd(clientCtx, groupcli.MsgVoteCmd(),
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
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	s.Require().Equal(uint32(0), txResp.Code, out.String())

	out, err = cli.ExecTestCLICmd(clientCtx, groupcli.QueryProposalCmd(), []string{"1", fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
	s.Require().NoError(err, out.String())

	var proposalRes group.QueryProposalResponse
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &proposalRes))
	s.proposal = proposalRes.Proposal

	out, err = cli.ExecTestCLICmd(clientCtx, groupcli.QueryVoteByProposalVoterCmd(), []string{"1", val.Address.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
	s.Require().NoError(err, out.String())

	var voteRes group.QueryVoteByProposalVoterResponse
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &voteRes))
	s.vote = voteRes.Vote

	s.voter = &group.Member{
		Address:  val.Address.String(),
		Weight:   memberWeight,
		Metadata: validMetadata,
	}
}

func (s *CLITestSuite) TestTxCreateGroup() {
	accounts := testutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	cmd := groupcli.MsgCreateGroupCmd()
	cmd.SetOutput(io.Discard)

	validMembers := fmt.Sprintf(`{"members": [{
		"address": "%s",
		  "weight": "1",
		  "metadata": "%s"
	  }]}`, accounts[0].Address.String(), validMetadata)
	validMembersFile := testutil.WriteToNewTempFile(s.T(), validMembers)

	invalidMembersAddress := `{"members": [{
		  "address": "",
		  "weight": "1"
	  }]}`
	invalidMembersAddressFile := testutil.WriteToNewTempFile(s.T(), invalidMembersAddress)

	invalidMembersWeight := fmt.Sprintf(`{"members": [{
			"address": "%s",
			  "weight": "0"
		  }]}`, accounts[0].Address.String())
	invalidMembersWeightFile := testutil.WriteToNewTempFile(s.T(), invalidMembersWeight)

	testCases := []struct {
		name         string
		ctxGen       func() client.Context
		args         []string
		respType     proto.Message
		expectErr    bool
		expectErrMsg string
	}{
		{
			"correct data",
			func() client.Context {
				bz, _ := s.encCfg.Codec.Marshal(&sdk.TxResponse{})
				c := newMockTendermintRPC(abci.ResponseQuery{
					Value: bz,
				})
				return s.baseCtx.WithClient(c)
			},
			append(
				[]string{
					accounts[0].Address.String(),
					"",
					validMembersFile.Name(),
				},
				s.commonFlags...,
			),
			&sdk.TxResponse{},
			false,
			"",
		},
		{
			"with amino-json",
			func() client.Context {
				bz, _ := s.encCfg.Codec.Marshal(&sdk.TxResponse{})
				c := newMockTendermintRPC(abci.ResponseQuery{
					Value: bz,
				})
				return s.baseCtx.WithClient(c)
			},
			append(
				[]string{
					accounts[0].Address.String(),
					"",
					validMembersFile.Name(),
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				s.commonFlags...,
			),
			&sdk.TxResponse{},
			false,
			"",
		},
		{
			"invalid members address",
			func() client.Context {
				bz, _ := s.encCfg.Codec.Marshal(&sdk.TxResponse{})
				c := newMockTendermintRPC(abci.ResponseQuery{
					Value: bz,
				})
				return s.baseCtx.WithClient(c)
			},
			append(
				[]string{
					accounts[0].Address.String(),
					"null",
					invalidMembersAddressFile.Name(),
				},
				s.commonFlags...,
			),
			&sdk.TxResponse{},
			true,
			"message validation failed: address: empty address string is not allowed",
		},
		{
			"invalid members weight",
			func() client.Context {
				bz, _ := s.encCfg.Codec.Marshal(&sdk.TxResponse{})
				c := newMockTendermintRPC(abci.ResponseQuery{
					Value: bz,
				})
				return s.baseCtx.WithClient(c)
			},
			append(
				[]string{
					accounts[0].Address.String(),
					"null",
					invalidMembersWeightFile.Name(),
				},
				s.commonFlags...,
			),
			&sdk.TxResponse{},
			true,
			"expected a positive decimal, got 0: invalid decimal string",
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {

			var outBuf bytes.Buffer

			clientCtx := tc.ctxGen().WithOutput(&outBuf)
			ctx := svrcmd.CreateExecuteContext(context.Background())

			cmd.SetContext(ctx)
			cmd.SetArgs(tc.args)

			s.Require().NoError(client.SetCmdClientContextHandler(clientCtx, cmd))

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}

func (s *CLITestSuite) TestTxUpdateGroupAdmin() {
	accounts := testutil.CreateKeyringAccounts(s.T(), s.kr, 2)

	cmd := groupcli.MsgUpdateGroupAdminCmd()
	cmd.SetOutput(io.Discard)

	var outBuf bytes.Buffer

	ctxGen := func() client.Context {
		bz, _ := s.encCfg.Codec.Marshal(&sdk.TxResponse{})
		c := newMockTendermintRPC(abci.ResponseQuery{
			Value: bz,
		})
		return s.baseCtx.WithClient(c)
	}
	clientCtx := ctxGen().WithOutput(&outBuf)
	ctx := svrcmd.CreateExecuteContext(context.Background())

	cmd.SetContext(ctx)

	s.Require().NoError(client.SetCmdClientContextHandler(clientCtx, cmd))

	groupIDs := make([]string, 2)
	for i := 0; i < 2; i++ {
		validMembers := fmt.Sprintf(`{"members": [{
	  "address": "%s",
		"weight": "1",
		"metadata": "%s"
	}]}`, accounts[0].Address.String(), validMetadata)
		validMembersFile := testutil.WriteToNewTempFile(s.T(), validMembers)
		out, err := cli.ExecTestCLICmd(clientCtx, groupcli.MsgCreateGroupCmd(),
			append(
				[]string{
					accounts[0].Address.String(),
					validMetadata,
					validMembersFile.Name(),
				},
				s.commonFlags...,
			),
		)
		s.Require().NoError(err, out.String())
		groupIDs[i] = fmt.Sprintf("%d", i+1)
		// var txResp sdk.TxResponse
		// s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
		// s.Require().Equal(uint32(0), txResp.Code, out.String())
		// groupIDs[i] = s.getGroupIDFromTxResponse(txResp)
	}

	testCases := []struct {
		name         string
		ctxGen       func() client.Context
		args         []string
		respType     proto.Message
		expectErr    bool
		expectErrMsg string
	}{
		{
			"correct data",
			func() client.Context {
				bz, _ := s.encCfg.Codec.Marshal(&sdk.TxResponse{})
				c := newMockTendermintRPC(abci.ResponseQuery{
					Value: bz,
				})
				return s.baseCtx.WithClient(c)
			},
			append(
				[]string{
					accounts[0].Address.String(),
					groupIDs[0],
					accounts[1].Address.String(),
				},
				s.commonFlags...,
			),
			&sdk.TxResponse{},
			false,
			"",
		},
		{
			"with amino-json",
			func() client.Context {
				bz, _ := s.encCfg.Codec.Marshal(&sdk.TxResponse{})
				c := newMockTendermintRPC(abci.ResponseQuery{
					Value: bz,
				})
				return s.baseCtx.WithClient(c)
			},
			append(
				[]string{
					accounts[0].Address.String(),
					groupIDs[1],
					accounts[1].Address.String(),
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				s.commonFlags...,
			),
			&sdk.TxResponse{},
			false,
			"",
		},
		{
			"group id invalid",
			func() client.Context {
				bz, _ := s.encCfg.Codec.Marshal(&sdk.TxResponse{})
				c := newMockTendermintRPC(abci.ResponseQuery{
					Value: bz,
				})
				return s.baseCtx.WithClient(c)
			},
			append(
				[]string{
					accounts[0].Address.String(),
					"",
					accounts[1].Address.String(),
				},
				s.commonFlags...,
			),
			&sdk.TxResponse{},
			true,
			"strconv.ParseUint: parsing \"\": invalid syntax",
		},
		// {
		// 	"group doesn't exist",
		// 	func() client.Context {
		// 		bz, _ := s.encCfg.Codec.Marshal(&sdk.TxResponse{})
		// 		c := newMockTendermintRPC(abci.ResponseQuery{
		// 			Value: bz,
		// 		})
		// 		return s.baseCtx.WithClient(c)
		// 	},
		// 	append(
		// 		[]string{
		// 			accounts[0].Address.String(),
		// 			"12345",
		// 			accounts[1].Address.String(),
		// 		},
		// 		s.commonFlags...,
		// 	),
		// 	&sdk.TxResponse{},
		// 	true,
		// 	"not found",
		// },
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			var outBuf bytes.Buffer

			clientCtx := tc.ctxGen().WithOutput(&outBuf)
			ctx := svrcmd.CreateExecuteContext(context.Background())

			cmd.SetContext(ctx)
			cmd.SetArgs(tc.args)

			s.Require().NoError(client.SetCmdClientContextHandler(clientCtx, cmd))

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}

func (s *CLITestSuite) getGroupIDFromTxResponse(txResp sdk.TxResponse) string {
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
func (s *CLITestSuite) createCLIProposal(groupPolicyAddress, proposer, sendFrom, sendTo, metadata string) string {
	_, err := base64.StdEncoding.DecodeString(metadata)
	s.Require().NoError(err)

	msg := banktypes.MsgSend{
		FromAddress: sendFrom,
		ToAddress:   sendTo,
		Amount:      sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(20))),
	}
	msgJSON, err := s.encCfg.Codec.MarshalInterfaceJSON(&msg)
	s.Require().NoError(err)

	p := groupcli.Proposal{
		GroupPolicyAddress: groupPolicyAddress,
		Messages:           []json.RawMessage{msgJSON},
		Metadata:           metadata,
		Proposers:          []string{proposer},
	}

	bz, err := json.Marshal(&p)
	s.Require().NoError(err)

	return testutil.WriteToNewTempFile(s.T(), string(bz)).Name()
}

func (s *CLITestSuite) createGroupThresholdPolicyWithBalance(adminAddress, groupID string, threshold int, tokens int64) string {
	accounts := testutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	var outBuf bytes.Buffer

	ctxGen := func() client.Context {
		bz, _ := s.encCfg.Codec.Marshal(&sdk.TxResponse{})
		c := newMockTendermintRPC(abci.ResponseQuery{
			Value: bz,
		})
		return s.baseCtx.WithClient(c)
	}
	clientCtx := ctxGen().WithOutput(&outBuf)

	out, err := cli.ExecTestCLICmd(clientCtx, groupcli.MsgCreateGroupPolicyCmd(),
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
	fmt.Println("---------------->", out.String())
	txResp := sdk.TxResponse{}
	s.Require().NoError(err, out.String())
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	s.Require().Equal(uint32(0), txResp.Code, out.String())

	out, err = cli.ExecTestCLICmd(clientCtx, groupcli.QueryGroupPoliciesByGroupCmd(), []string{groupID, fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
	s.Require().NoError(err, out.String())

	var res group.QueryGroupPoliciesByGroupResponse
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))
	fmt.Println("-------------")
	fmt.Println(res.String())
	fmt.Println("=============")
	groupPolicyAddress := res.GroupPolicies[0].Address

	addr, err := sdk.AccAddressFromBech32(groupPolicyAddress)
	s.Require().NoError(err)
	_, err = cli.MsgSendExec(clientCtx, accounts[0].Address, addr,
		sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(tokens))),
		s.commonFlags...,
	)
	s.Require().NoError(err)
	return groupPolicyAddress
}
