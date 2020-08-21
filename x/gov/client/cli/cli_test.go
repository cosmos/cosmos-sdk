package cli_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	tmcli "github.com/tendermint/tendermint/libs/cli"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	cfg := network.DefaultConfig()
	cfg.NumValidators = 1

	s.cfg = cfg
	s.network = network.New(s.T(), cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestNewCmdSubmitProposal() {
	val := s.network.Validators[0]

	invalidPropFile, err := ioutil.TempFile(os.TempDir(), "invalid_text_proposal.*.json")
	s.Require().NoError(err)
	defer os.Remove(invalidPropFile.Name())

	invalidProp := `{
  "title": "",
	"description": "Where is the title!?",
	"type": "Text",
  "deposit": "-324foocoin"
}`

	_, err = invalidPropFile.WriteString(invalidProp)
	s.Require().NoError(err)

	validPropFile, err := ioutil.TempFile(os.TempDir(), "valid_text_proposal.*.json")
	s.Require().NoError(err)
	defer os.Remove(validPropFile.Name())

	validProp := fmt.Sprintf(`{
  "title": "Text Proposal",
	"description": "Hello, World!",
	"type": "Text",
  "deposit": "%s"
}`, sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(5431)))

	_, err = validPropFile.WriteString(validProp)
	s.Require().NoError(err)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		respType     fmt.Stringer
		expectedCode uint32
	}{
		{
			"invalid proposal (file)",
			[]string{
				fmt.Sprintf("--%s=%s", cli.FlagProposal, invalidPropFile.Name()),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, nil, 0,
		},
		{
			"invalid proposal",
			[]string{
				fmt.Sprintf("--%s='Where is the title!?'", cli.FlagDescription),
				fmt.Sprintf("--%s=%s", cli.FlagProposalType, types.ProposalTypeText),
				fmt.Sprintf("--%s=%s", cli.FlagDeposit, sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(5431)).String()),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, nil, 0,
		},
		{
			"valid transaction (file)",
			[]string{
				fmt.Sprintf("--%s=%s", cli.FlagProposal, validPropFile.Name()),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"valid transaction",
			[]string{
				fmt.Sprintf("--%s='Text Proposal'", cli.FlagTitle),
				fmt.Sprintf("--%s='Where is the title!?'", cli.FlagDescription),
				fmt.Sprintf("--%s=%s", cli.FlagProposalType, types.ProposalTypeText),
				fmt.Sprintf("--%s=%s", cli.FlagDeposit, sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(5431)).String()),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
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
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestCmdGetProposal() {
	val := s.network.Validators[0]

	title := "Text Proposal1"

	testCases := []struct {
		name      string
		args      []string
		malleate  func()
		expectErr bool
	}{
		{
			"get proposal with json response",
			[]string{
				"1",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			func() {},
			true,
		},
		{
			"get proposal without proposal id",
			[]string{},
			func() {
				args := []string{
					fmt.Sprintf("--%s=%s", cli.FlagTitle, title),
					fmt.Sprintf("--%s='Where is the title!?'", cli.FlagDescription),
					fmt.Sprintf("--%s=%s", cli.FlagProposalType, types.ProposalTypeText),
					fmt.Sprintf("--%s=%s", cli.FlagDeposit, sdk.NewCoin(s.cfg.BondDenom, types.DefaultMinDepositTokens).String()),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
					fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
					fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
					fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
				}

				_, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cli.NewCmdSubmitProposal(), args)
				s.Require().NoError(err)
				_, err = s.network.WaitForHeight(1)
				s.Require().NoError(err)
			},
			true,
		},
		{
			"get proposal with json response",
			[]string{
				"1",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			func() {},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryProposal()
			clientCtx := val.ClientCtx

			tc.malleate()

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var proposal types.Proposal
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &proposal), out.String())
				s.Require().Equal(title, proposal.GetTitle())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestCmdGetProposals() {
	val := s.network.Validators[0]

	args := []string{
		fmt.Sprintf("--%s='Text Proposal 2'", cli.FlagTitle),
		fmt.Sprintf("--%s='Where is the title!?'", cli.FlagDescription),
		fmt.Sprintf("--%s=%s", cli.FlagProposalType, types.ProposalTypeText),
		fmt.Sprintf("--%s=%s", cli.FlagDeposit, sdk.NewCoin(s.cfg.BondDenom, types.DefaultMinDepositTokens).String()),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	_, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cli.NewCmdSubmitProposal(), args)
	s.Require().NoError(err)
	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"get proposals as json repsonse",
			[]string{
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryProposals()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var proposals types.Proposals
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &proposals), out.String())
				s.Require().Len(proposals, 1)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestNewCmdVoteProposal() {
	val := s.network.Validators[0]

	args := []string{
		fmt.Sprintf("--%s='Text Proposal 3'", cli.FlagTitle),
		fmt.Sprintf("--%s='Where is the title!?'", cli.FlagDescription),
		fmt.Sprintf("--%s=%s", cli.FlagProposalType, types.ProposalTypeText),
		fmt.Sprintf("--%s=%s", cli.FlagDeposit, sdk.NewCoin(s.cfg.BondDenom, types.DefaultMinDepositTokens).String()),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	_, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cli.NewCmdSubmitProposal(), args)
	s.Require().NoError(err)
	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		respType     fmt.Stringer
		expectedCode uint32
	}{
		{
			"invalid vote",
			[]string{},
			true, nil, 0,
		},
		{
			"vote for invalid proposal",
			[]string{
				"10",
				fmt.Sprintf("%s", "yes"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 2,
		},
		{
			"valid vote",
			[]string{
				"1",
				fmt.Sprintf("%s", "yes"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.NewCmdVote()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
