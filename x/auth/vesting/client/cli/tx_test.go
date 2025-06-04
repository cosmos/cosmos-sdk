package cli_test

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	rpcclientmock "github.com/cometbft/cometbft/v2/rpc/client/mock"
	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	testutilmod "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/client/cli"
)

type CLITestSuite struct {
	suite.Suite

	kr      keyring.Keyring
	encCfg  testutilmod.TestEncodingConfig
	baseCtx client.Context
}

func TestMigrateTestSuite(t *testing.T) {
	suite.Run(t, new(CLITestSuite))
}

func (s *CLITestSuite) SetupSuite() {
	s.encCfg = testutilmod.MakeTestEncodingConfig(vesting.AppModuleBasic{})
	s.kr = keyring.NewInMemory(s.encCfg.Codec)
	s.baseCtx = client.Context{}.
		WithKeyring(s.kr).
		WithTxConfig(s.encCfg.TxConfig).
		WithCodec(s.encCfg.Codec).
		WithClient(clitestutil.MockCometRPC{Client: rpcclientmock.Client{}}).
		WithAccountRetriever(client.MockAccountRetriever{}).
		WithOutput(io.Discard)
}

func (s *CLITestSuite) TestNewMsgCreateVestingAccountCmd() {
	accounts := testutil.CreateKeyringAccounts(s.T(), s.kr, 1)
	cmd := cli.NewMsgCreateVestingAccountCmd(address.NewBech32Codec("cosmos"))
	cmd.SetOut(io.Discard)

	extraArgs := []string{
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("photon", sdkmath.NewInt(10))).String()),
		fmt.Sprintf("--%s=test-chain", flags.FlagChainID),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, accounts[0].Address),
	}

	t := time.Date(2033, time.April, 1, 12, 34, 56, 789, time.UTC).Unix()

	testCases := []struct {
		name      string
		ctxGen    func() client.Context
		from, to  sdk.AccAddress
		amount    sdk.Coins
		endTime   int64
		extraArgs []string
		expectErr bool
	}{
		{
			"valid transaction",
			func() client.Context {
				return s.baseCtx
			},
			accounts[0].Address,
			accounts[0].Address,
			sdk.NewCoins(
				sdk.NewCoin("stake", sdkmath.NewInt(10)),
				sdk.NewCoin("photon", sdkmath.NewInt(40)),
			),
			t,
			extraArgs,
			false,
		},
		{
			"invalid to Address",
			func() client.Context {
				return s.baseCtx
			},
			accounts[0].Address,
			sdk.AccAddress{},
			sdk.NewCoins(
				sdk.NewCoin("stake", sdkmath.NewInt(10)),
				sdk.NewCoin("photon", sdkmath.NewInt(40)),
			),
			t,
			extraArgs,
			true,
		},
		{
			"invalid coins",
			func() client.Context {
				return s.baseCtx
			},
			accounts[0].Address,
			accounts[0].Address,
			nil,
			t,
			extraArgs,
			true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			ctx := svrcmd.CreateExecuteContext(context.Background())

			cmd.SetContext(ctx)
			fmt.Println(tc.amount.String())
			cmd.SetArgs(append([]string{tc.to.String(), tc.amount.String(), fmt.Sprint(tc.endTime)}, tc.extraArgs...))

			s.Require().NoError(client.SetCmdClientContextHandler(tc.ctxGen(), cmd))

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *CLITestSuite) TestNewMsgCreatePermanentLockedAccountCmd() {
	accounts := testutil.CreateKeyringAccounts(s.T(), s.kr, 1)
	cmd := cli.NewMsgCreatePermanentLockedAccountCmd(address.NewBech32Codec("cosmos"))
	cmd.SetOut(io.Discard)

	extraArgs := []string{
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("photon", sdkmath.NewInt(10))).String()),
		fmt.Sprintf("--%s=test-chain", flags.FlagChainID),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, accounts[0].Address),
	}

	testCases := []struct {
		name      string
		ctxGen    func() client.Context
		to        sdk.AccAddress
		amount    sdk.Coins
		extraArgs []string
		expectErr bool
	}{
		{
			"valid transaction",
			func() client.Context {
				return s.baseCtx
			},
			accounts[0].Address,
			sdk.NewCoins(
				sdk.NewCoin("stake", sdkmath.NewInt(10)),
				sdk.NewCoin("photon", sdkmath.NewInt(40)),
			),
			extraArgs,
			false,
		},
		{
			"invalid to Address",
			func() client.Context {
				return s.baseCtx
			},
			sdk.AccAddress{},
			sdk.NewCoins(
				sdk.NewCoin("stake", sdkmath.NewInt(10)),
				sdk.NewCoin("photon", sdkmath.NewInt(40)),
			),
			extraArgs,
			true,
		},
		{
			"invalid coins",
			func() client.Context {
				return s.baseCtx
			},
			accounts[0].Address,
			nil,
			extraArgs,
			true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			ctx := svrcmd.CreateExecuteContext(context.Background())

			cmd.SetContext(ctx)
			cmd.SetArgs(append([]string{tc.to.String(), tc.amount.String()}, tc.extraArgs...))

			s.Require().NoError(client.SetCmdClientContextHandler(tc.ctxGen(), cmd))

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *CLITestSuite) TestNewMsgCreatePeriodicVestingAccountCmd() {
	accounts := testutil.CreateKeyringAccounts(s.T(), s.kr, 1)
	cmd := cli.NewMsgCreatePeriodicVestingAccountCmd(address.NewBech32Codec("cosmos"))
	cmd.SetOut(io.Discard)

	extraArgs := []string{
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("photon", sdkmath.NewInt(10))).String()),
		fmt.Sprintf("--%s=test-chain", flags.FlagChainID),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, accounts[0].Address),
	}

	testCases := []struct {
		name      string
		ctxGen    func() client.Context
		to        sdk.AccAddress
		extraArgs []string
		expectErr bool
	}{
		{
			"valid transaction",
			func() client.Context {
				return s.baseCtx
			},
			accounts[0].Address,
			extraArgs,
			false,
		},
		{
			"invalid to Address",
			func() client.Context {
				return s.baseCtx
			},
			sdk.AccAddress{},
			extraArgs,
			true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			ctx := svrcmd.CreateExecuteContext(context.Background())

			cmd.SetContext(ctx)
			cmd.SetArgs(append([]string{tc.to.String(), "./test.json"}, tc.extraArgs...))

			s.Require().NoError(client.SetCmdClientContextHandler(tc.ctxGen(), cmd))

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}
