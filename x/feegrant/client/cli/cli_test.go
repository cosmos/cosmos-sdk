// +build norace

package cli_test

import (
	"fmt"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant/client/cli"
	"github.com/cosmos/cosmos-sdk/x/feegrant/types"
	govtestutil "github.com/cosmos/cosmos-sdk/x/gov/client/testutil"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg          network.Config
	network      *network.Network
	addedGranter sdk.AccAddress
	addedGrantee sdk.AccAddress
	addedGrant   types.FeeAllowanceGrant
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	if testing.Short() {
		s.T().Skip("skipping test in unit-tests mode.")
	}

	cfg := network.DefaultConfig()
	cfg.NumValidators = 2

	s.cfg = cfg
	s.network = network.New(s.T(), cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)

	val := s.network.Validators[0]
	granter := val.Address
	grantee := s.network.Validators[1].Address

	clientCtx := val.ClientCtx
	commonFlags := []string{
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	fee := sdk.NewCoin("stake", sdk.NewInt(100))
	duration := 365 * 24 * 60 * 60

	args := append(
		[]string{
			granter.String(),
			grantee.String(),
			fmt.Sprintf("--%s=%s", cli.FlagSpendLimit, fee.String()),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
			fmt.Sprintf("--%s=%v", cli.FlagExpiration, duration),
		},
		commonFlags...,
	)

	cmd := cli.NewCmdFeeGrant()

	_, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	s.Require().NoError(err)
	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)

	s.addedGranter = granter
	s.addedGrantee = grantee

	grant, err := types.NewFeeAllowanceGrant(granter, grantee, &types.BasicFeeAllowance{
		SpendLimit: sdk.NewCoins(fee),
	})
	s.Require().NoError(err)

	s.addedGrant = grant
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestCmdGetFeeGrant() {
	val := s.network.Validators[0]
	granter := val.Address
	grantee := s.addedGrantee
	clientCtx := val.ClientCtx

	testCases := []struct {
		name         string
		args         []string
		expectErrMsg string
		expectErr    bool
		respType     *types.FeeAllowanceGrant
		resp         *types.FeeAllowanceGrant
	}{
		{
			"wrong granter",
			[]string{
				"wrong_granter",
				grantee.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			"decoding bech32 failed",
			true, nil, nil,
		},
		{
			"wrong grantee",
			[]string{
				granter.String(),
				"wrong_grantee",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			"decoding bech32 failed",
			true, nil, nil,
		},
		{
			"non existed grant",
			[]string{
				"cosmos1nph3cfzk6trsmfxkeu943nvach5qw4vwstnvkl",
				grantee.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			"no fee allowance found",
			true, nil, nil,
		},
		{
			"valid req",
			[]string{
				granter.String(),
				grantee.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			"",
			false,
			&types.FeeAllowanceGrant{},
			&s.addedGrant,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryFeeGrant()
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				s.Require().Equal(tc.respType.Grantee, tc.respType.Grantee)
				s.Require().Equal(tc.respType.Granter, tc.respType.Granter)
				s.Require().Equal(
					tc.respType.GetFeeGrant().(*types.BasicFeeAllowance).SpendLimit,
					tc.resp.GetFeeGrant().(*types.BasicFeeAllowance).SpendLimit,
				)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestCmdGetFeeGrants() {
	val := s.network.Validators[0]
	grantee := s.addedGrantee
	clientCtx := val.ClientCtx

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		resp         *types.QueryFeeAllowancesResponse
		expectLength int
	}{
		{
			"wrong grantee",
			[]string{
				"wrong_grantee",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true, nil, 0,
		},
		{
			"non existed grantee",
			[]string{
				"cosmos1nph3cfzk6trsmfxkeu943nvach5qw4vwstnvkl",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false, &types.QueryFeeAllowancesResponse{}, 0,
		},
		{
			"valid req",
			[]string{
				grantee.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false, &types.QueryFeeAllowancesResponse{}, 1,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryFeeGrants()
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.resp), out.String())
				s.Require().Len(tc.resp.FeeAllowances, tc.expectLength)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestNewCmdFeeGrant() {
	val := s.network.Validators[0]
	granter := val.Address
	alreadyExistedGrantee := s.addedGrantee
	clientCtx := val.ClientCtx

	commonFlags := []string{
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"wrong granter address",
			append(
				[]string{
					"wrong_granter",
					"cosmos1nph3cfzk6trsmfxkeu943nvach5qw4vwstnvkl",
					fmt.Sprintf("--%s=%s", cli.FlagSpendLimit, "100stake"),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
				},
				commonFlags...,
			),
			true, nil, 0,
		},
		{
			"wrong grantee address",
			append(
				[]string{
					granter.String(),
					"wrong_grantee",
					fmt.Sprintf("--%s=%s", cli.FlagSpendLimit, "100stake"),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
				},
				commonFlags...,
			),
			true, nil, 0,
		},
		{
			"valid basic fee grant",
			append(
				[]string{
					granter.String(),
					"cosmos1nph3cfzk6trsmfxkeu943nvach5qw4vwstnvkl",
					fmt.Sprintf("--%s=%s", cli.FlagSpendLimit, "100stake"),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
				},
				commonFlags...,
			),
			false, &sdk.TxResponse{}, 0,
		},
		{
			"valid basic fee grant without spend limit",
			append(
				[]string{
					granter.String(),
					"cosmos17h5lzptx3ghvsuhk7wx4c4hnl7rsswxjer97em",
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
				},
				commonFlags...,
			),
			false, &sdk.TxResponse{}, 0,
		},
		{
			"valid basic fee grant without expiration",
			append(
				[]string{
					granter.String(),
					"cosmos16dlc38dcqt0uralyd8hksxyrny6kaeqfjvjwp5",
					fmt.Sprintf("--%s=%s", cli.FlagSpendLimit, "100stake"),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
				},
				commonFlags...,
			),
			false, &sdk.TxResponse{}, 0,
		},
		{
			"valid basic fee grant without spend-limit and expiration",
			append(
				[]string{
					granter.String(),
					"cosmos1ku40qup9vwag4wtf8cls9mkszxfthaklxkp3c8",
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
				},
				commonFlags...,
			),
			false, &sdk.TxResponse{}, 0,
		},
		{
			"try to add existed grant",
			append(
				[]string{
					granter.String(),
					alreadyExistedGrantee.String(),
					fmt.Sprintf("--%s=%s", cli.FlagSpendLimit, "100stake"),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
				},
				commonFlags...,
			),
			false, &sdk.TxResponse{}, 18,
		},
		{
			"invalid number of args(periodic fee grant)",
			append(
				[]string{
					granter.String(),
					"cosmos1nph3cfzk6trsmfxkeu943nvach5qw4vwstnvkl",
					fmt.Sprintf("--%s=%s", cli.FlagSpendLimit, "100stake"),
					fmt.Sprintf("--%s=%s", cli.FlagPeriodLimit, "10stake"),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
					fmt.Sprintf("--%s=%d", cli.FlagExpiration, 10*60*60),
				},
				commonFlags...,
			),
			true, nil, 0,
		},
		{
			"period mentioned and period limit omitted, invalid periodic grant",
			append(
				[]string{
					granter.String(),
					"cosmos1nph3cfzk6trsmfxkeu943nvach5qw4vwstnvkl",
					fmt.Sprintf("--%s=%s", cli.FlagSpendLimit, "100stake"),
					fmt.Sprintf("--%s=%d", cli.FlagPeriod, 10*60*60),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
					fmt.Sprintf("--%s=%d", cli.FlagExpiration, 60*60),
				},
				commonFlags...,
			),
			true, nil, 0,
		},
		{
			"period cannot be greater than the actual expiration(periodic fee grant)",
			append(
				[]string{
					granter.String(),
					"cosmos1nph3cfzk6trsmfxkeu943nvach5qw4vwstnvkl",
					fmt.Sprintf("--%s=%s", cli.FlagSpendLimit, "100stake"),
					fmt.Sprintf("--%s=%d", cli.FlagPeriod, 10*60*60),
					fmt.Sprintf("--%s=%s", cli.FlagPeriodLimit, "10stake"),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
					fmt.Sprintf("--%s=%d", cli.FlagExpiration, 60*60),
				},
				commonFlags...,
			),
			true, nil, 0,
		},
		{
			"valid periodic fee grant",
			append(
				[]string{
					granter.String(),
					"cosmos1w55kgcf3ltaqdy4ww49nge3klxmrdavrr6frmp",
					fmt.Sprintf("--%s=%s", cli.FlagSpendLimit, "100stake"),
					fmt.Sprintf("--%s=%d", cli.FlagPeriod, 60*60),
					fmt.Sprintf("--%s=%s", cli.FlagPeriodLimit, "10stake"),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
					fmt.Sprintf("--%s=%d", cli.FlagExpiration, 10*60*60),
				},
				commonFlags...,
			),
			false, &sdk.TxResponse{}, 0,
		},
		{
			"valid periodic fee grant without spend-limit",
			append(
				[]string{
					granter.String(),
					"cosmos1vevyks8pthkscvgazc97qyfjt40m6g9xe85ry8",
					fmt.Sprintf("--%s=%d", cli.FlagPeriod, 60*60),
					fmt.Sprintf("--%s=%s", cli.FlagPeriodLimit, "10stake"),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
					fmt.Sprintf("--%s=%d", cli.FlagExpiration, 10*60*60),
				},
				commonFlags...,
			),
			false, &sdk.TxResponse{}, 0,
		},
		{
			"valid periodic fee grant without expiration",
			append(
				[]string{
					granter.String(),
					"cosmos14cm33pvnrv2497tyt8sp9yavhmw83nwej3m0e8",
					fmt.Sprintf("--%s=%s", cli.FlagSpendLimit, "100stake"),
					fmt.Sprintf("--%s=%d", cli.FlagPeriod, 60*60),
					fmt.Sprintf("--%s=%s", cli.FlagPeriodLimit, "10stake"),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
				},
				commonFlags...,
			),
			false, &sdk.TxResponse{}, 0,
		},
		{
			"valid periodic fee grant without spend-limit and expiration",
			append(
				[]string{
					granter.String(),
					"cosmos12nyk4pcf4arshznkpz882e4l4ts0lt0ap8ce54",
					fmt.Sprintf("--%s=%d", cli.FlagPeriod, 60*60),
					fmt.Sprintf("--%s=%s", cli.FlagPeriodLimit, "10stake"),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
				},
				commonFlags...,
			),
			false, &sdk.TxResponse{}, 0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewCmdFeeGrant()
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

func (s *IntegrationTestSuite) TestNewCmdRevokeFeegrant() {
	val := s.network.Validators[0]
	granter := s.addedGranter
	grantee := s.addedGrantee
	clientCtx := val.ClientCtx

	commonFlags := []string{
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"invalid grantee",
			append(
				[]string{
					"wrong_granter",
					grantee.String(),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
				},
				commonFlags...,
			),
			true,
			nil,
			0,
		},
		{
			"invalid grantee",
			append(
				[]string{
					granter.String(),
					"wrong_grantee",
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
				},
				commonFlags...,
			),
			true,
			nil,
			0,
		},
		{
			"Non existed grant",
			append(
				[]string{
					granter.String(),
					"cosmos1aeuqja06474dfrj7uqsvukm6rael982kk89mqr",
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
				},
				commonFlags...,
			),
			false,
			&sdk.TxResponse{},
			4,
		},
		{
			"Valid revoke",
			append(
				[]string{
					granter.String(),
					grantee.String(),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
				},
				commonFlags...,
			),
			false,
			&sdk.TxResponse{},
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewCmdRevokeFeegrant()
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

func (s *IntegrationTestSuite) TestTxWithFeeGrant() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx
	granter := val.Address

	// creating an account manually (This account won't be exist in state)
	info, _, err := val.ClientCtx.Keyring.NewMnemonic("grantee", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)
	grantee := sdk.AccAddress(info.GetPubKey().Address())

	commonFlags := []string{
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	fee := sdk.NewCoin("stake", sdk.NewInt(100))
	duration := 365 * 24 * 60 * 60

	args := append(
		[]string{
			granter.String(),
			grantee.String(),
			fmt.Sprintf("--%s=%s", cli.FlagSpendLimit, fee.String()),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
			fmt.Sprintf("--%s=%v", cli.FlagExpiration, duration),
		},
		commonFlags...,
	)

	cmd := cli.NewCmdFeeGrant()

	_, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	s.Require().NoError(err)
	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)

	// granted fee allowance for an account which is not in state and creating
	// any tx with it by using --fee-account shouldn't fail
	out, err := govtestutil.MsgSubmitProposal(val.ClientCtx, grantee.String(),
		"Text Proposal", "No desc", govtypes.ProposalTypeText,
		fmt.Sprintf("--%s=%s", flags.FlagFeeAccount, granter.String()),
	)

	s.Require().NoError(err)
	var resp sdk.TxResponse
	s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &resp), out.String())
	s.Require().Equal(uint32(0), resp.Code)
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
