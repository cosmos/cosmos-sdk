package cli_test

import (
	"encoding/base64"
	"fmt"
	"io"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	rpcclientmock "github.com/cometbft/cometbft/rpc/client/mock"
	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	testutilmod "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govclitestutil "github.com/cosmos/cosmos-sdk/x/gov/client/testutil"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

type CLITestSuite struct {
	suite.Suite

	kr        keyring.Keyring
	encCfg    testutilmod.TestEncodingConfig
	baseCtx   client.Context
	clientCtx client.Context
}

func TestCLITestSuite(t *testing.T) {
	suite.Run(t, new(CLITestSuite))
}

func (s *CLITestSuite) SetupSuite() {
	s.encCfg = testutilmod.MakeTestEncodingConfig(gov.AppModuleBasic{})
	s.kr = keyring.NewInMemory(s.encCfg.Codec)
	s.baseCtx = client.Context{}.
		WithKeyring(s.kr).
		WithTxConfig(s.encCfg.TxConfig).
		WithCodec(s.encCfg.Codec).
		WithClient(clitestutil.MockCometRPC{Client: rpcclientmock.Client{}}).
		WithAccountRetriever(client.MockAccountRetriever{}).
		WithOutput(io.Discard).
		WithChainID("test-chain").
		WithAddressCodec(addresscodec.NewBech32Codec("cosmos")).
		WithValidatorAddressCodec(addresscodec.NewBech32Codec("cosmosvaloper")).
		WithConsensusAddressCodec(addresscodec.NewBech32Codec("cosmosvalcons"))

	ctxGen := func() client.Context {
		bz, _ := s.encCfg.Codec.Marshal(&sdk.TxResponse{})
		c := clitestutil.NewMockCometRPC(abci.ResponseQuery{
			Value: bz,
		})
		return s.baseCtx.WithClient(c)
	}
	s.clientCtx = ctxGen()

	val := testutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	// create a proposal with deposit
	_, err := govclitestutil.MsgSubmitLegacyProposal(s.clientCtx, val[0].Address.String(),
		"Text Proposal 1", "Where is the title!?", v1beta1.ProposalTypeText,
		fmt.Sprintf("--%s=%s", cli.FlagDeposit, sdk.NewCoin("stake", v1.DefaultMinDepositTokens).String()))
	s.Require().NoError(err)

	// vote for proposal
	_, err = govclitestutil.MsgVote(s.clientCtx, val[0].Address.String(), "1", "yes")
	s.Require().NoError(err)

	// create a proposal without deposit
	_, err = govclitestutil.MsgSubmitLegacyProposal(s.clientCtx, val[0].Address.String(),
		"Text Proposal 2", "Where is the title!?", v1beta1.ProposalTypeText)
	s.Require().NoError(err)

	// create a proposal3 with deposit
	_, err = govclitestutil.MsgSubmitLegacyProposal(s.clientCtx, val[0].Address.String(),
		"Text Proposal 3", "Where is the title!?", v1beta1.ProposalTypeText,
		fmt.Sprintf("--%s=%s", cli.FlagDeposit, sdk.NewCoin("stake", v1.DefaultMinDepositTokens).String()))
	s.Require().NoError(err)

	// vote for proposal3 as val
	_, err = govclitestutil.MsgVote(s.clientCtx, val[0].Address.String(), "3", "yes=0.6,no=0.3,abstain=0.05,no_with_veto=0.05")
	s.Require().NoError(err)
}

func (s *CLITestSuite) TestNewCmdSubmitProposal() {
	val := testutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	// Create a legacy proposal JSON, make sure it doesn't pass this new CLI
	// command.
	invalidProp := `{
		"title": "",
		"description": "Where is the title!?",
		"type": "Text",
		"deposit": "-324foocoin"
	}`
	invalidPropFile := testutil.WriteToNewTempFile(s.T(), invalidProp)
	defer invalidPropFile.Close()

	// Create a valid new proposal JSON.
	propMetadata := []byte{42}
	validProp := fmt.Sprintf(`
	{
		"messages": [
			{
				"@type": "/cosmos.gov.v1.MsgExecLegacyContent",
				"authority": "%s",
				"content": {
					"@type": "/cosmos.gov.v1beta1.TextProposal",
					"title": "My awesome title",
					"description": "My awesome description"
				}
			}
		],
		"title": "My awesome title",
		"summary": "My awesome description",
		"metadata": "%s",
		"deposit": "%s"
	}`, authtypes.NewModuleAddress(types.ModuleName), base64.StdEncoding.EncodeToString(propMetadata), sdk.NewCoin("stake", sdkmath.NewInt(5431)))
	validPropFile := testutil.WriteToNewTempFile(s.T(), validProp)
	defer validPropFile.Close()

	testCases := []struct {
		name         string
		args         []string
		expectErrMsg string
	}{
		{
			"invalid proposal",
			[]string{
				invalidPropFile.Name(),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10))).String()),
			},
			"invalid decimal coin expression",
		},
		{
			"valid proposal",
			[]string{
				validPropFile.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10))).String()),
			},
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewCmdSubmitProposal()

			out, err := clitestutil.ExecTestCLICmd(s.clientCtx, cmd, tc.args)
			if tc.expectErrMsg != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err)
				msg := &sdk.TxResponse{}
				s.Require().NoError(s.clientCtx.Codec.UnmarshalJSON(out.Bytes(), msg), out.String())
			}
		})
	}
}

func (s *CLITestSuite) TestNewCmdSubmitLegacyProposal() {
	val := testutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	invalidProp := `{
	  "title": "",
		"description": "Where is the title!?",
		"type": "Text",
	  "deposit": "-324foocoin"
	}`
	invalidPropFile := testutil.WriteToNewTempFile(s.T(), invalidProp)
	defer invalidPropFile.Close()
	validProp := fmt.Sprintf(`{
	  "title": "Text Proposal",
		"description": "Hello, World!",
		"type": "Text",
	  "deposit": "%s"
	}`, sdk.NewCoin("stake", sdkmath.NewInt(5431)))
	validPropFile := testutil.WriteToNewTempFile(s.T(), validProp)
	defer validPropFile.Close()

	testCases := []struct {
		name         string
		args         []string
		expectErrMsg string
	}{
		{
			"invalid proposal (file)",
			[]string{
				fmt.Sprintf("--%s=%s", cli.FlagProposal, invalidPropFile.Name()),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10))).String()),
			},
			"failed to parse proposal: proposal title is required",
		},
		{
			"invalid proposal",
			[]string{
				fmt.Sprintf("--%s='Where is the title!?'", cli.FlagDescription),
				fmt.Sprintf("--%s=%s", cli.FlagProposalType, v1beta1.ProposalTypeText),
				fmt.Sprintf("--%s=%s", cli.FlagDeposit, sdk.NewCoin("stake", sdkmath.NewInt(5431)).String()),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10))).String()),
			},
			"failed to parse proposal: proposal title is required",
		},
		{
			"valid transaction (file)",

			[]string{
				fmt.Sprintf("--%s=%s", cli.FlagProposal, validPropFile.Name()),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10))).String()),
			},
			"",
		},
		{
			"valid transaction",
			[]string{
				fmt.Sprintf("--%s='Text Proposal'", cli.FlagTitle),
				fmt.Sprintf("--%s='Where is the title!?'", cli.FlagDescription),
				fmt.Sprintf("--%s=%s", cli.FlagProposalType, v1beta1.ProposalTypeText),
				fmt.Sprintf("--%s=%s", cli.FlagDeposit, sdk.NewCoin("stake", sdkmath.NewInt(5431)).String()),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10))).String()),
			},
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewCmdSubmitLegacyProposal()

			out, err := clitestutil.ExecTestCLICmd(s.clientCtx, cmd, tc.args)
			if tc.expectErrMsg != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err)
				msg := &sdk.TxResponse{}
				s.Require().NoError(s.clientCtx.Codec.UnmarshalJSON(out.Bytes(), msg), out.String())
			}
		})
	}
}

func (s *CLITestSuite) TestNewCmdDeposit() {
	val := testutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	testCases := []struct {
		name         string
		args         []string
		expectErrMsg string
	}{
		{
			"invalid proposal id",
			[]string{
				"abc",
				sdk.NewCoin("stake", sdkmath.NewInt(10)).String(), // 10stake
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10))).String()),
			},
			"proposal-id abc not a valid uint, please input a valid proposal-id",
		},
		{
			"without deposit amount",
			[]string{
				"1",
				"invalidCoin",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10))).String()),
			},
			"invalid decimal coin expression: invalidCoin",
		},
		{
			"deposit on a proposal",
			[]string{
				"10",
				sdk.NewCoin("stake", sdkmath.NewInt(10)).String(), // 10stake
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10))).String()),
			},
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewCmdDeposit()

			out, err := clitestutil.ExecTestCLICmd(s.clientCtx, cmd, tc.args)
			if tc.expectErrMsg != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err)
				resp := &sdk.TxResponse{}
				s.Require().NoError(s.clientCtx.Codec.UnmarshalJSON(out.Bytes(), resp), out.String())
			}
		})
	}
}

func (s *CLITestSuite) TestNewCmdVote() {
	val := testutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	testCases := []struct {
		name         string
		args         []string
		expectErrMsg string
	}{
		{
			"vote for invalid proposal",
			[]string{
				"abc",
				"yes",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--metadata=%s", "AQ=="),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10))).String()),
			},
			"proposal-id abc not a valid int, please input a valid proposal-id",
		},
		{
			"invalid vote",
			[]string{
				"1",
				"AYE",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10))).String()),
			},
			"'AYE' is not a valid vote option",
		},
		{
			"valid vote",
			[]string{
				"1",
				"yes",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10))).String()),
			},
			"",
		},
		{
			"valid vote with metadata",
			[]string{
				"1",
				"yes",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--metadata=%s", "AQ=="),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10))).String()),
			},
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.NewCmdVote()
			out, err := clitestutil.ExecTestCLICmd(s.clientCtx, cmd, tc.args)

			if tc.expectErrMsg != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err)
				resp := &sdk.TxResponse{}
				s.Require().NoError(s.clientCtx.Codec.UnmarshalJSON(out.Bytes(), resp), out.String())
			}
		})
	}
}

func (s *CLITestSuite) TestNewCmdWeightedVote() {
	val := testutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	testCases := []struct {
		name         string
		args         []string
		expectErrMsg string
	}{
		{
			"vote for invalid proposal",
			[]string{
				"abc",
				"yes",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10))).String()),
			},
			"proposal-id abc not a valid int, please input a valid proposal-id",
		},
		{
			"invalid vote",
			[]string{
				"1",
				"AYE",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10))).String()),
			},
			"'AYE' is not a valid vote option",
		},
		{
			"valid vote",
			[]string{
				"1",
				"yes",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10))).String()),
			},
			"",
		},
		{
			"valid vote with metadata",
			[]string{
				"1",
				"yes",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--metadata=%s", "AQ=="),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10))).String()),
			},
			"",
		},
		{
			"invalid valid split vote string",
			[]string{
				"1",
				"yes/0.6,no/0.3,abstain/0.05,no_with_veto/0.05",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10))).String()),
			},
			"'yes/0.6' is not a valid vote option",
		},
		{
			"valid split vote",
			[]string{
				"1",
				"yes=0.6,no=0.3,abstain=0.05,no_with_veto=0.05",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10))).String()),
			},
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.NewCmdWeightedVote()
			out, err := clitestutil.ExecTestCLICmd(s.clientCtx, cmd, tc.args)
			if tc.expectErrMsg != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err)
				resp := &sdk.TxResponse{}
				s.Require().NoError(s.clientCtx.Codec.UnmarshalJSON(out.Bytes(), resp), out.String())
			}
		})
	}
}
