package cli_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     testutil.Config
	network *testutil.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	cfg := testutil.DefaultConfig()
	cfg.NumValidators = 1

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

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		respType  fmt.Stringer
		expected  fmt.Stringer
	}{
		{"no address provided", []string{}, true, nil, nil},
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

			cmd := cli.GetBalancesCmd()
			cmd.SetErr(buf)
			cmd.SetOut(buf)
			cmd.SetArgs(tc.args)

			err := cmd.ExecuteContext(ctx)
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

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)

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

			cmd := cli.GetCmdQueryTotalSupply()
			cmd.SetErr(buf)
			cmd.SetOut(buf)
			cmd.SetArgs(tc.args)

			err := cmd.ExecuteContext(ctx)
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

	testCases := []struct {
		name         string
		from, to     sdk.AccAddress
		amount       sdk.Coins
		args         []string
		expectErr    bool
		respType     fmt.Stringer
		expectedCode uint32
	}{
		{
			"valid transaction (gen-only)",
			val.Address,
			val.Address,
			sdk.NewCoins(
				sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), sdk.NewInt(10)),
				sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10)),
			),
			[]string{
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
			},
			false,
			&sdk.TxResponse{},
			0,
		},
		{
			"valid transaction",
			val.Address,
			val.Address,
			sdk.NewCoins(
				sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), sdk.NewInt(10)),
				sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10)),
			),
			[]string{
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false,
			&sdk.TxResponse{},
			0,
		},
		{
			"not enough fees",
			val.Address,
			val.Address,
			sdk.NewCoins(
				sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), sdk.NewInt(10)),
				sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10)),
			),
			[]string{
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(1))).String()),
			},
			false,
			&sdk.TxResponse{},
			sdkerrors.ErrInsufficientFee.ABCICode(),
		},
		{
			"not enough gas",
			val.Address,
			val.Address,
			sdk.NewCoins(
				sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), sdk.NewInt(10)),
				sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10)),
			),
			[]string{
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
				"--gas=10",
			},
			false,
			&sdk.TxResponse{},
			sdkerrors.ErrOutOfGas.ABCICode(),
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			buf.Reset()

			cmd := cli.NewSendTxCmd()
			cmd.SetErr(buf)
			cmd.SetOut(buf)
			cmd.SetArgs(tc.args)

			out, err := banktestutil.MsgSendExec(clientCtx, tc.from, tc.to, tc.amount, tc.args...)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out, tc.respType), string(out))

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code)
			}
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
