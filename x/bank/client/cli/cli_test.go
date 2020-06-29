package cli_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/client/cli"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     testutil.Config
	network *testutil.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	cfg := testutil.DefaultConfig()
	cfg.NumValidators = 2

	s.cfg = cfg
	s.network = testutil.NewTestNetwork(s.T(), cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestGetBalancesCmd() {
	buf := new(bytes.Buffer)
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx.WithOutput(buf)

	cmd := cli.GetBalancesCmd(clientCtx)
	cmd.SetErr(buf)
	cmd.SetOut(buf)

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		respType  fmt.Stringer
		expected  fmt.Stringer
	}{
		{"no address provided", nil, true, nil, nil},
		{
			"total account balance",
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=1", flags.FlagHeight),
			},
			false,
			&sdk.Coins{},
			sdk.NewCoins(
				sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), s.cfg.AccountTokens),
				sdk.NewCoin(s.cfg.BondDenom, s.cfg.StakingTokens.Sub(s.cfg.BondedTokens)),
			),
		},
		{
			"total account balance of a specific denom",
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=%s", cli.FlagDenom, s.cfg.BondDenom),
				fmt.Sprintf("--%s=1", flags.FlagHeight),
			},
			false,
			&sdk.Coin{},
			sdk.NewCoin(s.cfg.BondDenom, s.cfg.StakingTokens.Sub(s.cfg.BondedTokens)),
		},
		{
			"total account balance of a bogus denom",
			[]string{val.Address.String(), fmt.Sprintf("--%s=foobar", cli.FlagDenom)},
			false,
			&sdk.Coin{},
			sdk.NewCoin("foobar", sdk.ZeroInt()),
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			buf.Reset()
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(buf.Bytes(), tc.respType))
				s.Require().Equal(tc.expected.String(), tc.respType.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdQueryTotalSupply() {
	buf := new(bytes.Buffer)
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx.WithOutput(buf)

	cmd := cli.GetCmdQueryTotalSupply(clientCtx)
	cmd.SetErr(buf)
	cmd.SetOut(buf)

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		respType  fmt.Stringer
		expected  fmt.Stringer
	}{
		{
			"total supply",
			[]string{fmt.Sprintf("--%s=1", flags.FlagHeight)},
			false,
			&sdk.Coins{},
			sdk.NewCoins(
				sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), s.cfg.AccountTokens),
				sdk.NewCoin(s.cfg.BondDenom, s.cfg.StakingTokens.Add(sdk.NewInt(10))),
			),
		},
		{
			"total supply of a specific denomination",
			[]string{
				fmt.Sprintf("--%s=1", flags.FlagHeight),
				fmt.Sprintf("--%s=%s", cli.FlagDenom, s.cfg.BondDenom),
			},
			false,
			&sdk.Coin{},
			sdk.NewCoin(s.cfg.BondDenom, s.cfg.StakingTokens.Add(sdk.NewInt(10))),
		},
		{
			"total supply of a bogus denom",
			[]string{
				fmt.Sprintf("--%s=1", flags.FlagHeight),
				fmt.Sprintf("--%s=foobar", cli.FlagDenom),
			},
			false,
			&sdk.Coin{},
			sdk.NewCoin("foobar", sdk.ZeroInt()),
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			buf.Reset()
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(buf.Bytes(), tc.respType), buf.String())
				s.Require().Equal(tc.expected.String(), tc.respType.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestNewSendTxCmd() {
	buf := new(bytes.Buffer)
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx.WithOutput(buf)

	cmd := cli.NewSendTxCmd(clientCtx)
	cmd.SetErr(buf)
	cmd.SetOut(buf)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		respType     fmt.Stringer
		expectedCode uint32
	}{
		// TODO: send gen only
		// TODO: send invalid fees
		// TODO: send invalid gas
		{
			"valid transaction",
			[]string{
				val.Address.String(),
				s.network.Validators[1].Address.String(),
				sdk.NewCoins(
					sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), sdk.NewInt(10)),
					sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10)),
				).String(),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false,
			&sdk.TxResponse{},
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			buf.Reset()
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(buf.Bytes(), tc.respType), buf.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code)
			}
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
