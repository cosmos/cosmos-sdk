package cli_test

import (
	"context"
	"fmt"
	"io"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	rpcclientmock "github.com/cometbft/cometbft/rpc/client/mock"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	testutilmod "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/distribution/client/cli"
	distrtestutil "github.com/cosmos/cosmos-sdk/x/distribution/testutil"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
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
	s.encCfg = testutilmod.MakeTestEncodingConfig(bank.AppModuleBasic{})
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

	cfg, err := network.DefaultConfigWithAppConfig(distrtestutil.AppConfig)
	s.Require().NoError(err)

	genesisState := cfg.GenesisState
	var mintData minttypes.GenesisState
	s.Require().NoError(cfg.Codec.UnmarshalJSON(genesisState[minttypes.ModuleName], &mintData))

	inflation := sdkmath.LegacyMustNewDecFromStr("1.0")
	mintData.Minter.Inflation = inflation
	mintData.Params.InflationMin = inflation
	mintData.Params.InflationMax = inflation

	mintDataBz, err := cfg.Codec.MarshalJSON(&mintData)
	s.Require().NoError(err)
	genesisState[minttypes.ModuleName] = mintDataBz
	cfg.GenesisState = genesisState
}

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
			"hrp does not match bech32 prefix",
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
			cmd := cli.NewWithdrawRewardsCmd(addresscodec.NewBech32Codec("cosmosvaloper"), addresscodec.NewBech32Codec("cosmos"))
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
			cmd := cli.NewWithdrawAllRewardsCmd(addresscodec.NewBech32Codec("cosmosvaloper"), addresscodec.NewBech32Codec("cosmos"))

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
			cmd := cli.NewSetWithdrawAddrCmd(addresscodec.NewBech32Codec("cosmos"))

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
			cmd := cli.NewFundCommunityPoolCmd(addresscodec.NewBech32Codec("cosmos"))

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
