package cli_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
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

func (m mockTendermintRPC) ABCIQueryWithOptions(
	_ context.Context,
	_ string, _ tmbytes.HexBytes,
	_ rpcclient.ABCIQueryOptions,
) (*coretypes.ResultABCIQuery, error) {
	return &coretypes.ResultABCIQuery{Response: m.responseQuery}, nil
}

type CLITestSuite struct {
	suite.Suite

	kr          keyring.Keyring
	encCfg      testutilmod.TestEncodingConfig
	baseCtx     client.Context
	commonFlags []string
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
