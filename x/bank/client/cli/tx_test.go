package cli_test

import (
	"context"
	"fmt"
	"io"
	"testing"

	rpcclientmock "github.com/cometbft/cometbft/rpc/client/mock"
	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	testutilmod "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/bank/client/cli"
)

type CLITestSuite struct {
	suite.Suite

	kr      keyring.Keyring
	encCfg  testutilmod.TestEncodingConfig
	baseCtx client.Context
}

func TestCLITestSuite(t *testing.T) {
	suite.Run(t, new(CLITestSuite))
}

func (s *CLITestSuite) SetupSuite() {
	s.encCfg = testutilmod.MakeTestEncodingConfig(bank.AppModuleBasic{})
	s.kr = keyring.NewInMemory(s.encCfg.Codec)
	s.baseCtx = client.Context{}.
		WithKeyring(s.kr).
		WithTxConfig(s.encCfg.TxConfig).
		WithCodec(s.encCfg.Codec).
		WithClient(clitestutil.MockCometRPC{Client: rpcclientmock.Client{}}).
		WithAccountRetriever(client.MockAccountRetriever{}).
		WithOutput(io.Discard).
		WithAddressCodec(addresscodec.NewBech32Codec("cosmos")).
		WithValidatorAddressCodec(addresscodec.NewBech32Codec("cosmosvaloper")).
		WithConsensusAddressCodec(addresscodec.NewBech32Codec("cosmosvalcons"))
}

func (s *CLITestSuite) TestSendTxCmd() {
	accounts := testutil.CreateKeyringAccounts(s.T(), s.kr, 1)
	cmd := cli.NewSendTxCmd(addresscodec.NewBech32Codec("cosmos"))
	cmd.SetOutput(io.Discard)

	extraArgs := []string{
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("photon", sdkmath.NewInt(10))).String()),
		fmt.Sprintf("--%s=test-chain", flags.FlagChainID),
	}

	testCases := []struct {
		name         string
		ctxGen       func() client.Context
		from, to     sdk.AccAddress
		amount       sdk.Coins
		extraArgs    []string
		expectErrMsg string
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
			extraArgs,
			"",
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
			extraArgs,
			"empty address string is not allowed",
		},
		{
			"invalid coins",
			func() client.Context {
				return s.baseCtx
			},
			accounts[0].Address,
			accounts[0].Address,
			nil,
			extraArgs,
			"invalid coins",
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			args := append([]string{tc.from.String(), tc.to.String(), tc.amount.String()}, tc.extraArgs...)

			ctx := svrcmd.CreateExecuteContext(context.Background())
			cmd.SetContext(ctx)
			cmd.SetArgs(args)
			s.Require().NoError(client.SetCmdClientContextHandler(tc.ctxGen(), cmd))

			out, err := clitestutil.ExecTestCLICmd(tc.ctxGen(), cmd, args)
			if tc.expectErrMsg != "" {
				s.Require().Error(err)
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err)
				msg := &sdk.TxResponse{}
				s.Require().NoError(tc.ctxGen().Codec.UnmarshalJSON(out.Bytes(), msg), out.String())
			}
		})
	}
}

func (s *CLITestSuite) TestMultiSendTxCmd() {
	accounts := testutil.CreateKeyringAccounts(s.T(), s.kr, 3)

	cmd := cli.NewMultiSendTxCmd(addresscodec.NewBech32Codec("cosmos"))
	cmd.SetOutput(io.Discard)

	extraArgs := []string{
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("photon", sdkmath.NewInt(10))).String()),
		fmt.Sprintf("--%s=test-chain", flags.FlagChainID),
	}

	testCases := []struct {
		name         string
		ctxGen       func() client.Context
		from         string
		to           []string
		amount       sdk.Coins
		extraArgs    []string
		expectErrMsg string
	}{
		{
			"valid transaction",
			func() client.Context {
				return s.baseCtx
			},
			accounts[0].Address.String(),
			[]string{
				accounts[1].Address.String(),
				accounts[2].Address.String(),
			},
			sdk.NewCoins(
				sdk.NewCoin("stake", sdkmath.NewInt(10)),
				sdk.NewCoin("photon", sdkmath.NewInt(40)),
			),
			extraArgs,
			"",
		},
		{
			"invalid from Address",
			func() client.Context {
				return s.baseCtx
			},
			"foo",
			[]string{
				accounts[1].Address.String(),
				accounts[2].Address.String(),
			},
			sdk.NewCoins(
				sdk.NewCoin("stake", sdkmath.NewInt(10)),
				sdk.NewCoin("photon", sdkmath.NewInt(40)),
			),
			extraArgs,
			"key not found",
		},
		{
			"invalid recipients",
			func() client.Context {
				return s.baseCtx
			},
			accounts[0].Address.String(),
			[]string{
				accounts[1].Address.String(),
				"bar",
			},
			sdk.NewCoins(
				sdk.NewCoin("stake", sdkmath.NewInt(10)),
				sdk.NewCoin("photon", sdkmath.NewInt(40)),
			),
			extraArgs,
			"invalid bech32 string",
		},
		{
			"invalid amount",
			func() client.Context {
				return s.baseCtx
			},
			accounts[0].Address.String(),
			[]string{
				accounts[1].Address.String(),
				accounts[2].Address.String(),
			},
			nil,
			extraArgs,
			"must send positive amount",
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			ctx := svrcmd.CreateExecuteContext(context.Background())

			var args []string
			args = append(args, tc.from)
			args = append(args, tc.to...)
			args = append(args, tc.amount.String())
			args = append(args, tc.extraArgs...)

			cmd.SetContext(ctx)
			cmd.SetArgs(args)

			s.Require().NoError(client.SetCmdClientContextHandler(tc.ctxGen(), cmd))

			out, err := clitestutil.ExecTestCLICmd(tc.ctxGen(), cmd, args)
			if tc.expectErrMsg != "" {
				s.Require().Error(err)
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err)
				msg := &sdk.TxResponse{}
				s.Require().NoError(tc.ctxGen().Codec.UnmarshalJSON(out.Bytes(), msg), out.String())
			}
		})
	}
}
