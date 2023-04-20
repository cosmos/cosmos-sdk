package gov

import (
	"encoding/base64"
	"fmt"

	"github.com/cosmos/cosmos-sdk/testutil"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govclitestutil "github.com/cosmos/cosmos-sdk/x/gov/client/testutil"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

type E2ETestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func NewE2ETestSuite(cfg network.Config) *E2ETestSuite {
	return &E2ETestSuite{cfg: cfg}
}

func (s *E2ETestSuite) SetupSuite() {
	s.T().Log("setting up e2e test suite")

	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	val := s.network.Validators[0]
	clientCtx := val.ClientCtx
	var resp sdk.TxResponse

	// create a proposal with deposit
	out, err := govclitestutil.MsgSubmitLegacyProposal(val.ClientCtx, val.Address.String(),
		"Text Proposal 1", "Where is the title!?", v1beta1.ProposalTypeText,
		fmt.Sprintf("--%s=%s", cli.FlagDeposit, sdk.NewCoin(s.cfg.BondDenom, v1.DefaultMinDepositTokens).String()))
	s.Require().NoError(err)
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
	s.Require().NoError(clitestutil.CheckTxCode(s.network, clientCtx, resp.TxHash, 0))

	// vote for proposal
	out, err = govclitestutil.MsgVote(val.ClientCtx, val.Address.String(), "1", "yes")
	s.Require().NoError(err)
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
	s.Require().NoError(clitestutil.CheckTxCode(s.network, clientCtx, resp.TxHash, 0))

	// create a proposal without deposit
	out, err = govclitestutil.MsgSubmitLegacyProposal(val.ClientCtx, val.Address.String(),
		"Text Proposal 2", "Where is the title!?", v1beta1.ProposalTypeText)
	s.Require().NoError(err)
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
	s.Require().NoError(clitestutil.CheckTxCode(s.network, clientCtx, resp.TxHash, 0))

	// create a proposal3 with deposit
	out, err = govclitestutil.MsgSubmitLegacyProposal(val.ClientCtx, val.Address.String(),
		"Text Proposal 3", "Where is the title!?", v1beta1.ProposalTypeText,
		fmt.Sprintf("--%s=%s", cli.FlagDeposit, sdk.NewCoin(s.cfg.BondDenom, v1.DefaultMinDepositTokens).String()))
	s.Require().NoError(err)
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
	s.Require().NoError(clitestutil.CheckTxCode(s.network, clientCtx, resp.TxHash, 0))

	// create a proposal4 with deposit to check the cancel proposal cli tx
	out, err = govclitestutil.MsgSubmitLegacyProposal(val.ClientCtx, val.Address.String(),
		"Text Proposal 4", "Where is the title!?", v1beta1.ProposalTypeText,
		fmt.Sprintf("--%s=%s", cli.FlagDeposit, sdk.NewCoin(s.cfg.BondDenom, v1.DefaultMinDepositTokens).String()))
	s.Require().NoError(err)
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
	s.Require().NoError(clitestutil.CheckTxCode(s.network, clientCtx, resp.TxHash, 0))

	// vote for proposal3 as val
	out, err = govclitestutil.MsgVote(val.ClientCtx, val.Address.String(), "3", "yes=0.6,no=0.3,abstain=0.05,no_with_veto=0.05")
	s.Require().NoError(err)
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
	s.Require().NoError(clitestutil.CheckTxCode(s.network, clientCtx, resp.TxHash, 0))
}

func (s *E2ETestSuite) TearDownSuite() {
	s.T().Log("tearing down e2e test suite")
	s.network.Cleanup()
}

func (s *E2ETestSuite) TestNewCmdSubmitProposal() {
	val := s.network.Validators[0]

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
}`, authtypes.NewModuleAddress(types.ModuleName), base64.StdEncoding.EncodeToString(propMetadata), sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(5431)))
	validPropFile := testutil.WriteToNewTempFile(s.T(), validProp)
	defer validPropFile.Close()

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
		respType     proto.Message
	}{
		{
			"invalid proposal",
			[]string{
				invalidPropFile.Name(),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, 0, nil,
		},
		{
			"valid proposal",
			[]string{
				validPropFile.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, 0, &sdk.TxResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewCmdSubmitProposal()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().NoError(clitestutil.CheckTxCode(s.network, clientCtx, txResp.TxHash, tc.expectedCode))
			}
		})
	}
}

func (s *E2ETestSuite) TestNewCmdSubmitLegacyProposal() {
	val := s.network.Validators[0]
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
	}`, sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(5431)))
	validPropFile := testutil.WriteToNewTempFile(s.T(), validProp)
	defer validPropFile.Close()

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
		respType     proto.Message
	}{
		{
			"invalid proposal (file)",
			[]string{
				fmt.Sprintf("--%s=%s", cli.FlagProposal, invalidPropFile.Name()), //nolint:staticcheck // we are intentionally using a deprecated flag here.
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, 0, nil,
		},
		{
			"invalid proposal",
			[]string{
				fmt.Sprintf("--%s='Where is the title!?'", cli.FlagDescription),        //nolint:staticcheck // we are intentionally using a deprecated flag here.
				fmt.Sprintf("--%s=%s", cli.FlagProposalType, v1beta1.ProposalTypeText), //nolint:staticcheck // we are intentionally using a deprecated flag here.
				fmt.Sprintf("--%s=%s", cli.FlagDeposit, sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(5431)).String()),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, 0, nil,
		},
		{
			"valid transaction (file)",
			//nolint:staticcheck // we are intentionally using a deprecated flag here.
			[]string{
				fmt.Sprintf("--%s=%s", cli.FlagProposal, validPropFile.Name()),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, 0, &sdk.TxResponse{},
		},
		{
			"valid transaction",
			[]string{
				fmt.Sprintf("--%s='Text Proposal'", cli.FlagTitle),
				fmt.Sprintf("--%s='Where is the title!?'", cli.FlagDescription),        //nolint:staticcheck // we are intentionally using a deprecated flag here.
				fmt.Sprintf("--%s=%s", cli.FlagProposalType, v1beta1.ProposalTypeText), //nolint:staticcheck // we are intentionally using a deprecated flag here.
				fmt.Sprintf("--%s=%s", cli.FlagDeposit, sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(5431)).String()),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, 0, &sdk.TxResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewCmdSubmitLegacyProposal()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().NoError(clitestutil.CheckTxCode(s.network, clientCtx, txResp.TxHash, tc.expectedCode))
			}
		})
	}
}

func (s *E2ETestSuite) TestNewCmdCancelProposal() {
	val := s.network.Validators[0]
	val2 := sdk.AccAddress("invalid_acc_addr")

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
	}{
		{
			"without proposal id",
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, 0,
		},
		{
			"invalid proposal id",
			[]string{
				"asdasd",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, 0,
		},
		{
			"valid proposal-id but invalid proposer",
			[]string{
				"4",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val2),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, 0,
		},
		{
			"valid proposer",
			[]string{
				"4",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, 0,
		},
		{
			"proposal not exists after cancel",
			[]string{
				"4",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, 17,
		},
	}

	for _, tc := range testCases {
		var resp sdk.TxResponse

		s.Run(tc.name, func() {
			cmd := cli.NewCmdCancelProposal()
			clientCtx := val.ClientCtx
			var balRes banktypes.QueryAllBalancesResponse
			var newBalance banktypes.QueryAllBalancesResponse
			if !tc.expectErr && tc.expectedCode == 0 {
				resp, err := clitestutil.QueryBalancesExec(clientCtx, val.Address)
				s.Require().NoError(err)
				err = val.ClientCtx.Codec.UnmarshalJSON(resp.Bytes(), &balRes)
				s.Require().NoError(err)
			}

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
				s.Require().NoError(clitestutil.CheckTxCode(s.network, clientCtx, resp.TxHash, tc.expectedCode))

				if !tc.expectErr && tc.expectedCode == 0 {
					resp, err := clitestutil.QueryBalancesExec(clientCtx, val.Address)
					s.Require().NoError(err)
					err = val.ClientCtx.Codec.UnmarshalJSON(resp.Bytes(), &newBalance)
					s.Require().NoError(err)
					remainingAmount := v1.DefaultMinDepositTokens.Mul(
						v1.DefaultProposalCancelRatio.Mul(sdk.MustNewDecFromStr("100")).TruncateInt(),
					).Quo(sdk.NewIntFromUint64(100))

					// new balance = old balance + remaining amount from proposal deposit - txFee (cancel proposal)
					txFee := sdk.NewInt(10)
					s.Require().True(
						newBalance.Balances.AmountOf(s.network.Config.BondDenom).Equal(
							balRes.Balances.AmountOf(s.network.Config.BondDenom).Add(remainingAmount).Sub(txFee),
						),
					)
				}
			}
		})
	}
}

func (s *E2ETestSuite) TestNewCmdDeposit() {
	val := s.network.Validators[0]

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
	}{
		{
			"without proposal id",
			[]string{
				sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10)).String(), // 10stake
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, 0,
		},
		{
			"without deposit amount",
			[]string{
				"1",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, 0,
		},
		{
			"deposit on non existing proposal",
			[]string{
				"10",
				sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10)).String(), // 10stake
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, 2,
		},
		{
			"deposit on existing proposal",
			[]string{
				"1",
				sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10)).String(), // 10stake
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, 0,
		},
	}

	for _, tc := range testCases {
		tc := tc
		var resp sdk.TxResponse

		s.Run(tc.name, func() {
			cmd := cli.NewCmdDeposit()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
				s.Require().NoError(clitestutil.CheckTxCode(s.network, clientCtx, resp.TxHash, tc.expectedCode))
			}
		})
	}
}

func (s *E2ETestSuite) TestNewCmdVote() {
	val := s.network.Validators[0]

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
	}{
		{
			"invalid vote",
			[]string{},
			true, 0,
		},
		{
			"vote for invalid proposal",
			[]string{
				"10",
				"yes",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--metadata=%s", "AQ=="),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, 3,
		},
		{
			"valid vote",
			[]string{
				"1",
				"yes",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, 0,
		},
		{
			"valid vote with metadata",
			[]string{
				"1",
				"yes",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--metadata=%s", "AQ=="),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, 0,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.NewCmdVote()
			clientCtx := val.ClientCtx
			var txResp sdk.TxResponse

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
				s.Require().NoError(clitestutil.CheckTxCode(s.network, clientCtx, txResp.TxHash, tc.expectedCode))
			}
		})
	}
}

func (s *E2ETestSuite) TestNewCmdWeightedVote() {
	val := s.network.Validators[0]

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
	}{
		{
			"invalid vote",
			[]string{},
			true, 0,
		},
		{
			"vote for invalid proposal",
			[]string{
				"10",
				"yes",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, 3,
		},
		{
			"valid vote",
			[]string{
				"1",
				"yes",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, 0,
		},
		{
			"valid vote with metadata",
			[]string{
				"1",
				"yes",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--metadata=%s", "AQ=="),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, 0,
		},
		{
			"invalid valid split vote string",
			[]string{
				"1",
				"yes/0.6,no/0.3,abstain/0.05,no_with_veto/0.05",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, 0,
		},
		{
			"valid split vote",
			[]string{
				"1",
				"yes=0.6,no=0.3,abstain=0.05,no_with_veto=0.05",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, 0,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.NewCmdWeightedVote()
			clientCtx := val.ClientCtx
			var txResp sdk.TxResponse

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
				s.Require().NoError(clitestutil.CheckTxCode(s.network, clientCtx, txResp.TxHash, tc.expectedCode))
			}
		})
	}
}
