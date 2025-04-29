package cli_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	rpcclientmock "github.com/cometbft/cometbft/rpc/client/mock"
	"github.com/stretchr/testify/suite"

	// without this import amino json encoding will fail when resolving any types
	_ "cosmossdk.io/api/cosmos/group/v1"
	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	testutilmod "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/group"                     //nolint:staticcheck // deprecated and to be removed
	groupcli "github.com/cosmos/cosmos-sdk/x/group/client/cli" //nolint:staticcheck // deprecated and to be removed
	groupmodule "github.com/cosmos/cosmos-sdk/x/group/module"  //nolint:staticcheck // deprecated and to be removed
)

var validMetadata = "metadata"

type CLITestSuite struct {
	suite.Suite

	kr          keyring.Keyring
	encCfg      testutilmod.TestEncodingConfig
	baseCtx     client.Context
	clientCtx   client.Context
	group       *group.GroupInfo
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
		WithClient(clitestutil.MockCometRPC{Client: rpcclientmock.Client{}}).
		WithAccountRetriever(client.MockAccountRetriever{}).
		WithOutput(io.Discard).
		WithChainID("test-chain")

	s.commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10))).String()),
	}

	accounts := testutil.CreateKeyringAccounts(s.T(), s.kr, 1)
	val := accounts[0]

	ctxGen := func() client.Context {
		bz, _ := s.encCfg.Codec.Marshal(&sdk.TxResponse{})
		c := clitestutil.NewMockCometRPC(abci.ResponseQuery{
			Value: bz,
		})
		return s.baseCtx.WithClient(c)
	}
	s.clientCtx = ctxGen()

	// create a new account
	info, _, err := s.clientCtx.Keyring.NewMnemonic("NewValidator", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)

	pk, err := info.GetPubKey()
	s.Require().NoError(err)

	account := sdk.AccAddress(pk.Address())
	_, err = clitestutil.MsgSendExec(
		s.clientCtx,
		val.Address,
		account,
		sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(2000))), address.NewBech32Codec("cosmos"), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10))).String()),
	)
	s.Require().NoError(err)

	// create a group
	validMembers := fmt.Sprintf(`
	{
		"members": [
			{
				"address": "%s",
				"weight": "3",
				"metadata": "%s"
			}
		]
	}`, val.Address.String(), validMetadata)
	validMembersFile := testutil.WriteToNewTempFile(s.T(), validMembers)
	out, err := clitestutil.ExecTestCLICmd(s.clientCtx, groupcli.MsgCreateGroupCmd(),
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
	s.Require().NoError(s.clientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	s.Require().Equal(uint32(0), txResp.Code, out.String())

	s.group = &group.GroupInfo{Id: 1, Admin: val.Address.String(), Metadata: validMetadata, TotalWeight: "3", Version: 1}
}

func (s *CLITestSuite) TestTxCreateGroup() {
	accounts := testutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	cmd := groupcli.MsgCreateGroupCmd()
	cmd.SetOut(io.Discard)

	validMembers := fmt.Sprintf(`{"members": [{
		"address": "%s",
		  "weight": "1",
		  "metadata": "%s"
	  }]}`, accounts[0].Address.String(), validMetadata)
	validMembersFile := testutil.WriteToNewTempFile(s.T(), validMembers)

	invalidMembersWeight := fmt.Sprintf(`{"members": [{
			"address": "%s",
			  "weight": "0"
		  }]}`, accounts[0].Address.String())
	invalidMembersWeightFile := testutil.WriteToNewTempFile(s.T(), invalidMembersWeight)

	testCases := []struct {
		name         string
		args         []string
		expCmdOutput string
		expectErrMsg string
	}{
		{
			name: "correct data",
			args: append(
				[]string{
					accounts[0].Address.String(),
					"",
					validMembersFile.Name(),
				},
				s.commonFlags...,
			),
			expCmdOutput: fmt.Sprintf("%s %s %s", accounts[0].Address.String(), "", validMembersFile.Name()),
			expectErrMsg: "",
		},
		{
			"with amino-json",
			append(
				[]string{
					accounts[0].Address.String(),
					"",
					validMembersFile.Name(),
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				s.commonFlags...,
			),
			fmt.Sprintf("%s %s %s", accounts[0].Address.String(), "", validMembersFile.Name()),
			"",
		},
		{
			"invalid members weight",
			append(
				[]string{
					accounts[0].Address.String(),
					"null",
					invalidMembersWeightFile.Name(),
				},
				s.commonFlags...,
			),
			"",
			"weight must be positive",
		},
		{
			"no member provided",
			append(
				[]string{
					accounts[0].Address.String(),
					"null",
					"doesnotexist.json",
				},
				s.commonFlags...,
			),
			"",
			"no such file or directory",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			ctx := svrcmd.CreateExecuteContext(context.Background())
			cmd.SetContext(ctx)
			cmd.SetArgs(tc.args)
			s.Require().NoError(client.SetCmdClientContextHandler(s.baseCtx, cmd))

			if len(tc.args) != 0 {
				s.Require().Contains(fmt.Sprint(cmd), tc.expCmdOutput)
			}

			out, err := clitestutil.ExecTestCLICmd(s.baseCtx, cmd, tc.args)
			if tc.expectErrMsg != "" {
				s.Require().Error(err)
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err)
				msg := &sdk.TxResponse{}
				s.Require().NoError(s.baseCtx.Codec.UnmarshalJSON(out.Bytes(), msg), out.String())
			}
		})
	}
}

func (s *CLITestSuite) TestTxUpdateGroupAdmin() {
	accounts := testutil.CreateKeyringAccounts(s.T(), s.kr, 2)

	cmd := groupcli.MsgUpdateGroupAdminCmd()
	cmd.SetOut(io.Discard)

	ctx := svrcmd.CreateExecuteContext(context.Background())
	cmd.SetContext(ctx)
	s.Require().NoError(client.SetCmdClientContextHandler(s.baseCtx, cmd))

	groupIDs := make([]string, 2)
	for i := range 2 {
		validMembers := fmt.Sprintf(`{"members": [{
	  "address": "%s",
		"weight": "1",
		"metadata": "%s"
	}]}`, accounts[0].Address.String(), validMetadata)
		validMembersFile := testutil.WriteToNewTempFile(s.T(), validMembers)
		out, err := clitestutil.ExecTestCLICmd(s.baseCtx, groupcli.MsgCreateGroupCmd(),
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
	}

	testCases := []struct {
		name         string
		ctxGen       func() client.Context
		args         []string
		expCmdOutput string
		expectErrMsg string
	}{
		{
			"correct data",
			func() client.Context {
				bz, _ := s.encCfg.Codec.Marshal(&sdk.TxResponse{})
				c := clitestutil.NewMockCometRPC(abci.ResponseQuery{
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
			fmt.Sprintf("%s %s %s", accounts[0].Address.String(), groupIDs[0], accounts[1].Address.String()),
			"",
		},
		{
			"with amino-json",
			func() client.Context {
				bz, _ := s.encCfg.Codec.Marshal(&sdk.TxResponse{})
				c := clitestutil.NewMockCometRPC(abci.ResponseQuery{
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
			fmt.Sprintf("%s %s %s --%s=%s", accounts[0].Address.String(), groupIDs[1], accounts[1].Address.String(), flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
			"",
		},
		{
			"group id invalid",
			func() client.Context {
				bz, _ := s.encCfg.Codec.Marshal(&sdk.TxResponse{})
				c := clitestutil.NewMockCometRPC(abci.ResponseQuery{
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
			fmt.Sprintf("%s %s %s", accounts[0].Address.String(), "", accounts[1].Address.String()),
			"strconv.ParseUint: parsing \"\": invalid syntax",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			ctx := svrcmd.CreateExecuteContext(context.Background())

			cmd.SetContext(ctx)
			cmd.SetArgs(tc.args)

			s.Require().NoError(client.SetCmdClientContextHandler(s.baseCtx, cmd))

			if len(tc.args) != 0 {
				s.Require().Contains(fmt.Sprint(cmd), tc.expCmdOutput)
			}

			out, err := clitestutil.ExecTestCLICmd(s.baseCtx, cmd, tc.args)
			if tc.expectErrMsg != "" {
				s.Require().Error(err)
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err)
				msg := &sdk.TxResponse{}
				s.Require().NoError(s.baseCtx.Codec.UnmarshalJSON(out.Bytes(), msg), out.String())
			}
		})
	}
}

func (s *CLITestSuite) TestTxUpdateGroupMetadata() {
	accounts := testutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	cmd := groupcli.MsgUpdateGroupMetadataCmd()
	cmd.SetOut(io.Discard)

	testCases := []struct {
		name         string
		args         []string
		expCmdOutput string
		expectErrMsg string
	}{
		{
			"correct data",
			append(
				[]string{
					accounts[0].Address.String(),
					"1",
					validMetadata,
				},
				s.commonFlags...,
			),
			fmt.Sprintf("%s %s %s", accounts[0].Address.String(), "1", validMetadata),
			"",
		},
		{
			"with amino-json",
			append(
				[]string{
					accounts[0].Address.String(),
					"1",
					validMetadata,
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				s.commonFlags...,
			),
			fmt.Sprintf("%s %s %s --%s=%s", accounts[0].Address.String(), "1", validMetadata, flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
			"",
		},
		{
			"invalid group id",
			append(
				[]string{
					accounts[0].Address.String(),
					"abc",
					validMetadata,
				},
				s.commonFlags...,
			),
			fmt.Sprintf("%s %s %s", accounts[0].Address.String(), "abc", validMetadata),
			"Error: strconv.ParseUint: parsing \"abc\"",
		},
		{
			"empty group id",
			append(
				[]string{
					accounts[0].Address.String(),
					"0",
					validMetadata,
				},
				s.commonFlags...,
			),
			fmt.Sprintf("%s %s %s", accounts[0].Address.String(), "0", validMetadata),
			"group id cannot be 0",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			ctx := svrcmd.CreateExecuteContext(context.Background())

			cmd.SetContext(ctx)
			cmd.SetArgs(tc.args)

			s.Require().NoError(client.SetCmdClientContextHandler(s.baseCtx, cmd))

			if len(tc.args) != 0 {
				s.Require().Contains(fmt.Sprint(cmd), tc.expCmdOutput)
			}

			out, err := clitestutil.ExecTestCLICmd(s.baseCtx, cmd, tc.args)
			if tc.expectErrMsg != "" {
				s.Require().Error(err)
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err)
				msg := &sdk.TxResponse{}
				s.Require().NoError(s.baseCtx.Codec.UnmarshalJSON(out.Bytes(), msg), out.String())
			}
		})
	}
}

func (s *CLITestSuite) TestTxUpdateGroupMembers() {
	accounts := testutil.CreateKeyringAccounts(s.T(), s.kr, 3)
	groupPolicyAddress := accounts[2]

	cmd := groupcli.MsgUpdateGroupMembersCmd()
	cmd.SetOut(io.Discard)

	groupID := "1"

	validUpdatedMembersFileName := testutil.WriteToNewTempFile(s.T(), fmt.Sprintf(`{"members": [{
		"address": "%s",
		"weight": "0",
		"metadata": "%s"
	}, {
		"address": "%s",
		"weight": "1",
		"metadata": "%s"
	}]}`, accounts[1], validMetadata, groupPolicyAddress, validMetadata)).Name()

	invalidMembersMetadata := fmt.Sprintf(`{"members": [{
		"address": "%s",
		  "weight": "-1",
		  "metadata": "foo"
	  }]}`, accounts[1])
	invalidMembersMetadataFileName := testutil.WriteToNewTempFile(s.T(), invalidMembersMetadata).Name()

	testCases := []struct {
		name         string
		args         []string
		expCmdOutput string
		expectErrMsg string
	}{
		{
			"correct data",
			append(
				[]string{
					accounts[0].Address.String(),
					groupID,
					validUpdatedMembersFileName,
				},
				s.commonFlags...,
			),
			fmt.Sprintf("%s %s %s", accounts[0].Address.String(), groupID, validUpdatedMembersFileName),
			"",
		},
		{
			"with amino-json",
			append(
				[]string{
					accounts[0].Address.String(),
					groupID,
					validUpdatedMembersFileName,
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				s.commonFlags...,
			),
			fmt.Sprintf("%s %s %s --%s=%s", accounts[0].Address.String(), groupID, validUpdatedMembersFileName, flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
			"",
		},
		{
			"group id invalid",
			append(
				[]string{
					accounts[0].Address.String(),
					"0",
					validUpdatedMembersFileName,
				},
				s.commonFlags...,
			),
			fmt.Sprintf("%s %s %s", accounts[0].Address.String(), "0", validUpdatedMembersFileName),
			"group id cannot be 0",
		},
		{
			"group member weight invalid",
			append(
				[]string{
					accounts[0].Address.String(),
					groupID,
					invalidMembersMetadataFileName,
				},
				s.commonFlags...,
			),
			fmt.Sprintf("%s %s %s", accounts[0].Address.String(), groupID, invalidMembersMetadataFileName),
			"invalid weight -1",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			ctx := svrcmd.CreateExecuteContext(context.Background())

			cmd.SetContext(ctx)
			cmd.SetArgs(tc.args)

			s.Require().NoError(client.SetCmdClientContextHandler(s.baseCtx, cmd))

			if len(tc.args) != 0 {
				s.Require().Contains(fmt.Sprint(cmd), tc.expCmdOutput)
			}

			out, err := clitestutil.ExecTestCLICmd(s.baseCtx, cmd, tc.args)
			if tc.expectErrMsg != "" {
				s.Require().Error(err)
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err)
				msg := &sdk.TxResponse{}
				s.Require().NoError(s.baseCtx.Codec.UnmarshalJSON(out.Bytes(), msg), out.String())
			}
		})
	}
}

func (s *CLITestSuite) TestTxCreateGroupWithPolicy() {
	accounts := testutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	cmd := groupcli.MsgCreateGroupWithPolicyCmd()
	cmd.SetOut(io.Discard)

	validMembers := fmt.Sprintf(`{"members": [{
		"address": "%s",
		  "weight": "1",
		  "metadata": "%s"
	}]}`, accounts[0].Address.String(), validMetadata)
	validMembersFile := testutil.WriteToNewTempFile(s.T(), validMembers)

	invalidMembersWeight := fmt.Sprintf(`{"members": [{
		"address": "%s",
		  "weight": "0"
	}]}`, accounts[0].Address.String())
	invalidMembersWeightFile := testutil.WriteToNewTempFile(s.T(), invalidMembersWeight)

	thresholdDecisionPolicyFile := testutil.WriteToNewTempFile(s.T(), `{"@type": "/cosmos.group.v1.ThresholdDecisionPolicy","threshold": "1","windows": {"voting_period":"1s"}}`)

	testCases := []struct {
		name         string
		args         []string
		expectErrMsg string
		expCmdOutput string
	}{
		{
			"correct data",
			append(
				[]string{
					accounts[0].Address.String(),
					validMetadata,
					validMetadata,
					validMembersFile.Name(),
					thresholdDecisionPolicyFile.Name(),
					fmt.Sprintf("--%s=%v", groupcli.FlagGroupPolicyAsAdmin, false),
				},
				s.commonFlags...,
			),
			"",
			fmt.Sprintf("%s %s %s %s %s --%s=%v", accounts[0].Address.String(), validMetadata, validMetadata, validMembersFile.Name(), thresholdDecisionPolicyFile.Name(), groupcli.FlagGroupPolicyAsAdmin, false),
		},
		{
			"group-policy-as-admin is true",
			append(
				[]string{
					accounts[0].Address.String(),
					validMetadata,
					validMetadata,
					validMembersFile.Name(),
					thresholdDecisionPolicyFile.Name(),
					fmt.Sprintf("--%s=%v", groupcli.FlagGroupPolicyAsAdmin, true),
				},
				s.commonFlags...,
			),
			"",
			fmt.Sprintf("%s %s %s %s %s --%s=%v", accounts[0].Address.String(), validMetadata, validMetadata, validMembersFile.Name(), thresholdDecisionPolicyFile.Name(), groupcli.FlagGroupPolicyAsAdmin, true),
		},
		{
			"with amino-json",
			append(
				[]string{
					accounts[0].Address.String(),
					validMetadata,
					validMetadata,
					validMembersFile.Name(),
					thresholdDecisionPolicyFile.Name(),
					fmt.Sprintf("--%s=%v", groupcli.FlagGroupPolicyAsAdmin, false),
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				s.commonFlags...,
			),
			"",
			fmt.Sprintf("%s %s %s %s %s --%s=%v --%s=%s", accounts[0].Address.String(), validMetadata, validMetadata, validMembersFile.Name(), thresholdDecisionPolicyFile.Name(), groupcli.FlagGroupPolicyAsAdmin, false, flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
		},
		{
			"invalid members weight",
			append(
				[]string{
					accounts[0].Address.String(),
					validMetadata,
					validMetadata,
					invalidMembersWeightFile.Name(),
					thresholdDecisionPolicyFile.Name(),
					fmt.Sprintf("--%s=%v", groupcli.FlagGroupPolicyAsAdmin, false),
				},
				s.commonFlags...,
			),
			"weight must be positive",
			fmt.Sprintf("%s %s %s %s %s --%s=%v", accounts[0].Address.String(), validMetadata, validMetadata, invalidMembersWeightFile.Name(), thresholdDecisionPolicyFile.Name(), groupcli.FlagGroupPolicyAsAdmin, false),
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			ctx := svrcmd.CreateExecuteContext(context.Background())

			cmd.SetContext(ctx)
			cmd.SetArgs(tc.args)

			s.Require().NoError(client.SetCmdClientContextHandler(s.baseCtx, cmd))

			if len(tc.args) != 0 {
				s.Require().Contains(fmt.Sprint(cmd), tc.expCmdOutput)
			}

			out, err := clitestutil.ExecTestCLICmd(s.baseCtx, cmd, tc.args)
			if tc.expectErrMsg != "" {
				s.Require().Error(err)
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				msg := &sdk.TxResponse{}
				s.Require().NoError(s.baseCtx.Codec.UnmarshalJSON(out.Bytes(), msg), out.String())
			}
		})
	}
}

func (s *CLITestSuite) TestTxCreateGroupPolicy() {
	accounts := testutil.CreateKeyringAccounts(s.T(), s.kr, 2)
	val := accounts[0]

	groupID := s.group.Id

	thresholdDecisionPolicyFile := testutil.WriteToNewTempFile(s.T(), `{"@type": "/cosmos.group.v1.ThresholdDecisionPolicy","threshold": "1","windows": {"voting_period":"1s"}}`)

	percentageDecisionPolicyFile := testutil.WriteToNewTempFile(s.T(), `{"@type":"/cosmos.group.v1.PercentageDecisionPolicy", "percentage":"0.5", "windows":{"voting_period":"1s"}}`)
	invalidNegativePercentageDecisionPolicyFile := testutil.WriteToNewTempFile(s.T(), `{"@type":"/cosmos.group.v1.PercentageDecisionPolicy", "percentage":"-0.5", "windows":{"voting_period":"1s"}}`)
	invalidPercentageDecisionPolicyFile := testutil.WriteToNewTempFile(s.T(), `{"@type":"/cosmos.group.v1.PercentageDecisionPolicy", "percentage":"2", "windows":{"voting_period":"1s"}}`)

	cmd := groupcli.MsgCreateGroupPolicyCmd()
	cmd.SetOut(io.Discard)

	testCases := []struct {
		name         string
		args         []string
		expectErrMsg string
		expCmdOutput string
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
			"",
			fmt.Sprintf("%s %s %s %s", val.Address.String(), fmt.Sprintf("%v", groupID), validMetadata, thresholdDecisionPolicyFile.Name()),
		},
		{
			"correct data with percentage decision policy",
			append(
				[]string{
					val.Address.String(),
					fmt.Sprintf("%v", groupID),
					validMetadata,
					percentageDecisionPolicyFile.Name(),
				},
				s.commonFlags...,
			),
			"",
			fmt.Sprintf("%s %s %s %s", val.Address.String(), fmt.Sprintf("%v", groupID), validMetadata, percentageDecisionPolicyFile.Name()),
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
			"",
			fmt.Sprintf("%s %s %s %s --%s=%s", val.Address.String(), fmt.Sprintf("%v", groupID), validMetadata, thresholdDecisionPolicyFile.Name(), flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
		},
		{
			"wrong admin",
			append(
				[]string{
					"wrongAdmin",
					fmt.Sprintf("%v", groupID),
					validMetadata,
					thresholdDecisionPolicyFile.Name(),
				},
				s.commonFlags...,
			),
			"key not found",
			fmt.Sprintf("%s %s %s %s", "wrongAdmin", fmt.Sprintf("%v", groupID), validMetadata, thresholdDecisionPolicyFile.Name()),
		},
		{
			"invalid percentage decision policy with negative value",
			append(
				[]string{
					val.Address.String(),
					fmt.Sprintf("%v", groupID),
					validMetadata,
					invalidNegativePercentageDecisionPolicyFile.Name(),
				},
				s.commonFlags...,
			),
			"expected a positive decimal",
			fmt.Sprintf("%s %s %s %s", val.Address.String(), fmt.Sprintf("%v", groupID), validMetadata, invalidNegativePercentageDecisionPolicyFile.Name()),
		},
		{
			"invalid percentage decision policy with value greater than 1",
			append(
				[]string{
					val.Address.String(),
					fmt.Sprintf("%v", groupID),
					validMetadata,
					invalidPercentageDecisionPolicyFile.Name(),
				},
				s.commonFlags...,
			),
			"percentage must be > 0 and <= 1",
			fmt.Sprintf("%s %s %s %s", val.Address.String(), fmt.Sprintf("%v", groupID), validMetadata, invalidPercentageDecisionPolicyFile.Name()),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			ctx := svrcmd.CreateExecuteContext(context.Background())

			cmd.SetContext(ctx)
			cmd.SetArgs(tc.args)

			s.Require().NoError(client.SetCmdClientContextHandler(s.baseCtx, cmd))

			if len(tc.args) != 0 {
				s.Require().Contains(fmt.Sprint(cmd), tc.expCmdOutput)
			}

			out, err := clitestutil.ExecTestCLICmd(s.baseCtx, cmd, tc.args)
			if tc.expectErrMsg != "" {
				s.Require().Error(err)
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				msg := &sdk.TxResponse{}
				s.Require().NoError(s.baseCtx.Codec.UnmarshalJSON(out.Bytes(), msg), out.String())
			}
		})
	}
}

func (s *CLITestSuite) TestTxUpdateGroupPolicyAdmin() {
	accounts := testutil.CreateKeyringAccounts(s.T(), s.kr, 4)
	newAdmin := accounts[0]
	groupPolicyAdmin := accounts[1]
	groupPolicyAddress := accounts[2]

	commonFlags := s.commonFlags
	commonFlags = append(commonFlags, fmt.Sprintf("--%s=%d", flags.FlagGas, 300000))

	cmd := groupcli.MsgUpdateGroupPolicyAdminCmd()
	cmd.SetOut(io.Discard)

	testCases := []struct {
		name         string
		args         []string
		expCmdOutput string
		expectErrMsg string
	}{
		{
			"correct data",
			append(
				[]string{
					groupPolicyAdmin.Address.String(),
					groupPolicyAddress.Address.String(),
					newAdmin.Address.String(),
				},
				commonFlags...,
			),
			fmt.Sprintf("%s %s %s", groupPolicyAdmin.Address.String(), groupPolicyAddress.Address.String(), newAdmin.Address.String()),
			"",
		},
		{
			"with amino-json",
			append(
				[]string{
					groupPolicyAdmin.Address.String(),
					groupPolicyAddress.Address.String(),
					newAdmin.Address.String(),
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				commonFlags...,
			),
			fmt.Sprintf("%s %s %s --%s=%s", groupPolicyAdmin.Address.String(), groupPolicyAddress.Address.String(), newAdmin.Address.String(), flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
			"",
		},
		{
			"wrong admin",
			append(
				[]string{
					"wrong admin",
					groupPolicyAddress.Address.String(),
					newAdmin.Address.String(),
				},
				commonFlags...,
			),
			fmt.Sprintf("%s %s %s", "wrong admin", groupPolicyAddress.Address.String(), newAdmin.Address.String()),
			"key not found",
		},
		{
			"identical admin and new admin",
			append(
				[]string{
					groupPolicyAdmin.Address.String(),
					groupPolicyAddress.Address.String(),
					groupPolicyAdmin.Address.String(),
				},
				commonFlags...,
			),
			fmt.Sprintf("%s %s %s", groupPolicyAdmin.Address.String(), groupPolicyAddress.Address.String(), groupPolicyAdmin.Address.String()),
			"new admin cannot be the same as the current admin",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			ctx := svrcmd.CreateExecuteContext(context.Background())

			cmd.SetContext(ctx)
			cmd.SetArgs(tc.args)

			s.Require().NoError(client.SetCmdClientContextHandler(s.baseCtx, cmd))

			if len(tc.args) != 0 {
				s.Require().Contains(fmt.Sprint(cmd), tc.expCmdOutput)
			}

			out, err := clitestutil.ExecTestCLICmd(s.baseCtx, cmd, tc.args)
			if tc.expectErrMsg != "" {
				s.Require().Error(err)
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err)
				msg := &sdk.TxResponse{}
				s.Require().NoError(s.baseCtx.Codec.UnmarshalJSON(out.Bytes(), msg), out.String())
			}
		})
	}
}

func (s *CLITestSuite) TestTxUpdateGroupPolicyDecisionPolicy() {
	accounts := testutil.CreateKeyringAccounts(s.T(), s.kr, 3)
	newAdmin := accounts[0]
	groupPolicyAdmin := accounts[1]
	groupPolicyAddress := accounts[2]

	commonFlags := s.commonFlags
	commonFlags = append(commonFlags, fmt.Sprintf("--%s=%d", flags.FlagGas, 300000))

	thresholdDecisionPolicy := testutil.WriteToNewTempFile(s.T(), `{"@type":"/cosmos.group.v1.ThresholdDecisionPolicy", "threshold":"1", "windows":{"voting_period":"40000s"}}`)
	percentageDecisionPolicy := testutil.WriteToNewTempFile(s.T(), `{"@type":"/cosmos.group.v1.PercentageDecisionPolicy", "percentage":"0.5", "windows":{"voting_period":"40000s"}}`)

	cmd := groupcli.MsgUpdateGroupPolicyDecisionPolicyCmd(address.NewBech32Codec("cosmos"))
	cmd.SetOut(io.Discard)

	testCases := []struct {
		name         string
		args         []string
		expCmdOutput string
		expectErrMsg string
	}{
		{
			"correct data",
			append(
				[]string{
					groupPolicyAdmin.Address.String(),
					groupPolicyAddress.Address.String(),
					thresholdDecisionPolicy.Name(),
				},
				commonFlags...,
			),
			fmt.Sprintf("%s %s %s", groupPolicyAdmin.Address.String(), groupPolicyAddress.Address.String(), thresholdDecisionPolicy.Name()),
			"",
		},
		{
			"correct data with percentage decision policy",
			append(
				[]string{
					groupPolicyAdmin.Address.String(),
					groupPolicyAddress.Address.String(),
					percentageDecisionPolicy.Name(),
				},
				commonFlags...,
			),
			fmt.Sprintf("%s %s %s", groupPolicyAdmin.Address.String(), groupPolicyAddress.Address.String(), percentageDecisionPolicy.Name()),
			"",
		},
		{
			"with amino-json",
			append(
				[]string{
					groupPolicyAdmin.Address.String(),
					groupPolicyAddress.Address.String(),
					thresholdDecisionPolicy.Name(),
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				commonFlags...,
			),
			fmt.Sprintf("%s %s %s --%s=%s", groupPolicyAdmin.Address.String(), groupPolicyAddress.Address.String(), thresholdDecisionPolicy.Name(), flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
			"",
		},
		{
			"wrong admin",
			append(
				[]string{
					newAdmin.Address.String(),
					"invalid",
					thresholdDecisionPolicy.Name(),
				},
				commonFlags...,
			),
			fmt.Sprintf("%s %s %s", newAdmin.Address.String(), "invalid", thresholdDecisionPolicy.Name()),
			"decoding bech32 failed",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			ctx := svrcmd.CreateExecuteContext(context.Background())

			cmd.SetContext(ctx)
			cmd.SetArgs(tc.args)

			s.Require().NoError(client.SetCmdClientContextHandler(s.baseCtx, cmd))

			if len(tc.args) != 0 {
				s.Require().Contains(fmt.Sprint(cmd), tc.expCmdOutput)
			}

			out, err := clitestutil.ExecTestCLICmd(s.baseCtx, cmd, tc.args)
			if tc.expectErrMsg != "" {
				s.Require().Error(err)
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err)
				msg := &sdk.TxResponse{}
				s.Require().NoError(s.baseCtx.Codec.UnmarshalJSON(out.Bytes(), msg), out.String())
			}
		})
	}
}

func (s *CLITestSuite) TestTxUpdateGroupPolicyMetadata() {
	accounts := testutil.CreateKeyringAccounts(s.T(), s.kr, 2)
	groupPolicyAdmin := accounts[0].Address
	groupPolicyAddress := accounts[1].Address

	commonFlags := s.commonFlags
	commonFlags = append(commonFlags, fmt.Sprintf("--%s=%d", flags.FlagGas, 300000))

	cmd := groupcli.MsgUpdateGroupPolicyMetadataCmd()
	cmd.SetOut(io.Discard)

	testCases := []struct {
		name         string
		args         []string
		expCmdOutput string
		expectErrMsg string
	}{
		{
			"correct data",
			append(
				[]string{
					groupPolicyAdmin.String(),
					groupPolicyAddress.String(),
					validMetadata,
				},
				commonFlags...,
			),
			fmt.Sprintf("%s %s %s", groupPolicyAdmin.String(), groupPolicyAddress.String(), validMetadata),
			"",
		},
		{
			"with amino-json",
			append(
				[]string{
					groupPolicyAdmin.String(),
					groupPolicyAddress.String(),
					validMetadata,
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				commonFlags...,
			),
			fmt.Sprintf("%s %s %s --%s=%s", groupPolicyAdmin.String(), groupPolicyAddress.String(), validMetadata, flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
			"",
		},
		{
			"wrong admin",
			append(
				[]string{
					"wrong admin",
					groupPolicyAddress.String(),
					validMetadata,
				},
				commonFlags...,
			),
			fmt.Sprintf("%s %s %s", "wrong admin", groupPolicyAddress.String(), validMetadata),
			"key not found",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			ctx := svrcmd.CreateExecuteContext(context.Background())

			cmd.SetContext(ctx)
			cmd.SetArgs(tc.args)

			s.Require().NoError(client.SetCmdClientContextHandler(s.baseCtx, cmd))

			if len(tc.args) != 0 {
				s.Require().Contains(fmt.Sprint(cmd), tc.expCmdOutput)
			}

			out, err := clitestutil.ExecTestCLICmd(s.baseCtx, cmd, tc.args)
			if tc.expectErrMsg != "" {
				s.Require().Error(err)
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err)
				msg := &sdk.TxResponse{}
				s.Require().NoError(s.baseCtx.Codec.UnmarshalJSON(out.Bytes(), msg), out.String())
			}
		})
	}
}

func (s *CLITestSuite) TestTxSubmitProposal() {
	accounts := testutil.CreateKeyringAccounts(s.T(), s.kr, 2)
	groupPolicyAddress := accounts[1].Address

	p := groupcli.Proposal{
		GroupPolicyAddress: groupPolicyAddress.String(),
		Messages:           []json.RawMessage{},
		Metadata:           validMetadata,
		Proposers:          []string{accounts[0].Address.String()},
	}
	bz, err := json.Marshal(&p)
	s.Require().NoError(err)
	proposalFile := testutil.WriteToNewTempFile(s.T(), string(bz))

	cmd := groupcli.MsgSubmitProposalCmd()
	cmd.SetOut(io.Discard)

	testCases := []struct {
		name         string
		args         []string
		expCmdOutput string
		expectErrMsg string
	}{
		{
			"correct data",
			append(
				[]string{
					proposalFile.Name(),
				},
				s.commonFlags...,
			),
			proposalFile.Name(),
			"",
		},
		{
			"with try exec",
			append(
				[]string{
					proposalFile.Name(),
					fmt.Sprintf("--%s=try", groupcli.FlagExec),
				},
				s.commonFlags...,
			),
			fmt.Sprintf("%s --%s=try", proposalFile.Name(), groupcli.FlagExec),
			"",
		},
		{
			"with try exec, not enough yes votes for proposal to pass",
			append(
				[]string{
					proposalFile.Name(),
					fmt.Sprintf("--%s=try", groupcli.FlagExec),
				},
				s.commonFlags...,
			),
			fmt.Sprintf("%s --%s=try", proposalFile.Name(), groupcli.FlagExec),
			"",
		},
		{
			"with amino-json",
			append(
				[]string{
					proposalFile.Name(),
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				s.commonFlags...,
			),
			fmt.Sprintf("%s --%s=%s", proposalFile.Name(), flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			ctx := svrcmd.CreateExecuteContext(context.Background())

			cmd.SetContext(ctx)
			cmd.SetArgs(tc.args)

			s.Require().NoError(client.SetCmdClientContextHandler(s.baseCtx, cmd))

			if len(tc.args) != 0 {
				s.Require().Contains(fmt.Sprint(cmd), tc.expCmdOutput)
			}

			out, err := clitestutil.ExecTestCLICmd(s.baseCtx, cmd, tc.args)
			if tc.expectErrMsg != "" {
				s.Require().Error(err)
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err)
				msg := &sdk.TxResponse{}
				s.Require().NoError(s.baseCtx.Codec.UnmarshalJSON(out.Bytes(), msg), out.String())
			}
		})
	}
}

func (s *CLITestSuite) TestTxVote() {
	accounts := testutil.CreateKeyringAccounts(s.T(), s.kr, 4)

	cmd := groupcli.MsgVoteCmd()
	cmd.SetOut(io.Discard)

	ids := make([]string, 4)
	for i := range ids {
		ids[i] = fmt.Sprint(i + 1)
	}

	testCases := []struct {
		name         string
		args         []string
		expCmdOutput string
		expectErrMsg string
	}{
		{
			"invalid vote",
			append(
				[]string{
					ids[0],
					accounts[0].Address.String(),
					"AYE",
					"",
				},
				s.commonFlags...,
			),
			fmt.Sprintf("%s %s %s", ids[0], accounts[0].Address.String(), "AYE"),
			"Error: 'AYE' is not a valid vote option",
		},
		{
			"correct data",
			append(
				[]string{
					ids[0],
					accounts[0].Address.String(),
					"VOTE_OPTION_YES",
					"",
				},
				s.commonFlags...,
			),
			fmt.Sprintf("%s %s %s", ids[0], accounts[0].Address.String(), "VOTE_OPTION_YES"),
			"",
		},
		{
			"with try exec",
			append(
				[]string{
					ids[1],
					accounts[0].Address.String(),
					"VOTE_OPTION_YES",
					"",
					fmt.Sprintf("--%s=try", groupcli.FlagExec),
				},
				s.commonFlags...,
			),
			fmt.Sprintf("%s %s %s %s --%s=try", ids[1], accounts[0].Address.String(), "VOTE_OPTION_YES", "", groupcli.FlagExec),
			"",
		},
		{
			"with amino-json",
			append(
				[]string{
					ids[3],
					accounts[0].Address.String(),
					"VOTE_OPTION_YES",
					"",
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				s.commonFlags...,
			),
			fmt.Sprintf("%s %s %s %s --%s=%s", ids[3], accounts[0].Address.String(), "VOTE_OPTION_YES", "", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			ctx := svrcmd.CreateExecuteContext(context.Background())

			cmd.SetContext(ctx)
			cmd.SetArgs(tc.args)

			s.Require().NoError(client.SetCmdClientContextHandler(s.baseCtx, cmd))

			if len(tc.args) != 0 {
				s.Require().Contains(fmt.Sprint(cmd), tc.expCmdOutput)
			}

			out, err := clitestutil.ExecTestCLICmd(s.baseCtx, cmd, tc.args)
			if tc.expectErrMsg != "" {
				s.Require().Error(err)
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err)
				msg := &sdk.TxResponse{}
				s.Require().NoError(s.baseCtx.Codec.UnmarshalJSON(out.Bytes(), msg), out.String())
			}
		})
	}
}

func (s *CLITestSuite) TestTxWithdrawProposal() {
	accounts := testutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	cmd := groupcli.MsgWithdrawProposalCmd()
	cmd.SetOut(io.Discard)

	ids := make([]string, 2)
	ids[0] = "1"
	ids[1] = "2"

	testCases := []struct {
		name         string
		args         []string
		expCmdOutput string
		expectErrMsg string
	}{
		{
			"correct data",
			append(
				[]string{
					ids[0],
					accounts[0].Address.String(),
				},
				s.commonFlags...,
			),
			fmt.Sprintf("%s %s", ids[0], accounts[0].Address.String()),
			"",
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
			fmt.Sprintf("%s %s", ids[1], "wrongAdmin"),
			"key not found",
		},
		{
			"wrong proposal id",
			append(
				[]string{
					"abc",
					accounts[0].Address.String(),
				},
				s.commonFlags...,
			),
			fmt.Sprintf("%s %s", "abc", accounts[0].Address.String()),
			"Error: strconv.ParseUint: parsing \"abc\"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			ctx := svrcmd.CreateExecuteContext(context.Background())

			cmd.SetContext(ctx)
			cmd.SetArgs(tc.args)

			s.Require().NoError(client.SetCmdClientContextHandler(s.baseCtx, cmd))

			if len(tc.args) != 0 {
				s.Require().Contains(fmt.Sprint(cmd), tc.expCmdOutput)
			}

			out, err := clitestutil.ExecTestCLICmd(s.baseCtx, cmd, tc.args)
			if tc.expectErrMsg != "" {
				s.Require().Error(err)
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err)
				msg := &sdk.TxResponse{}
				s.Require().NoError(s.baseCtx.Codec.UnmarshalJSON(out.Bytes(), msg), out.String())
			}
		})
	}
}
