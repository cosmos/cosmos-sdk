package cli_test

import (
	"context"
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/gogoproto/proto"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec/address"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/client/cli"
)

func (s *CLITestSuite) TestTxWithdrawRewardsCmd() {
	val := testutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	testCases := []struct {
		name         string
		valAddr      fmt.Stringer
		args         []string
		expectErrMsg string
	}{
		{
			"invalid validator address",
			val[0].Address,
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10))).String()),
			},
			"invalid Bech32 prefix",
		},
		{
			"valid transaction",
			sdk.ValAddress(val[0].Address),
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10))).String()),
			},
			"",
		},
		{
			"valid transaction (with commission)",
			sdk.ValAddress(val[0].Address),
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=true", cli.FlagCommission),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10))).String()),
			},
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			args := append([]string{tc.valAddr.String()}, tc.args...)

			ctx := svrcmd.CreateExecuteContext(context.Background())
			cmd := cli.NewWithdrawRewardsCmd()
			cmd.SetContext(ctx)
			cmd.SetArgs(args)
			s.Require().NoError(client.SetCmdClientContextHandler(s.clientCtx, cmd))

			out, err := clitestutil.ExecTestCLICmd(s.clientCtx, cmd, args)
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

func (s *CLITestSuite) TestTxWithdrawAllRewardsCmd() {
	val := testutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	testCases := []struct {
		name         string
		args         []string
		expectErrMsg string
	}{
		{
			"invalid transaction (offline)",
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagOffline),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10))).String()),
			},
			"cannot generate tx in offline mode",
		},
		{
			"valid transaction",
			[]string{
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
			cmd := cli.NewWithdrawAllRewardsCmd()

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

func (s *CLITestSuite) TestTxSetWithdrawAddrCmd() {
	val := testutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		respType  proto.Message
	}{
		{
			"invalid withdraw address",
			[]string{
				"foo",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10))).String()),
			},
			true, nil,
		},
		{
			"valid transaction",
			[]string{
				val[0].Address.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewSetWithdrawAddrCmd(address.NewBech32Codec("cosmos"))

			out, err := clitestutil.ExecTestCLICmd(s.clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(s.clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}

func (s *CLITestSuite) TestTxFundCommunityPoolCmd() {
	val := testutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		respType  proto.Message
	}{
		{
			"invalid funding amount",
			[]string{
				"-43foocoin",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10))).String()),
			},
			true, nil,
		},
		{
			"valid transaction",
			[]string{
				sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(5431))).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewFundCommunityPoolCmd()

			out, err := clitestutil.ExecTestCLICmd(s.clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(s.clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}
