package cli_test

import (
	"fmt"
	"io"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	rpcclientmock "github.com/cometbft/cometbft/rpc/client/mock"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	testutilmod "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/client/cli"
)

var PKs = simtestutil.CreateTestPubKeys(500)

type CLITestSuite struct {
	suite.Suite

	kr        keyring.Keyring
	encCfg    testutilmod.TestEncodingConfig
	baseCtx   client.Context
	clientCtx client.Context
	addrs     []sdk.AccAddress
}

func TestCLITestSuite(t *testing.T) {
	suite.Run(t, new(CLITestSuite))
}

func (s *CLITestSuite) SetupSuite() {
	s.encCfg = testutilmod.MakeTestEncodingConfig(staking.AppModuleBasic{})
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

	s.addrs = make([]sdk.AccAddress, 0)
	for i := 0; i < 3; i++ {
		k, _, err := s.clientCtx.Keyring.NewMnemonic("NewValidator", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
		s.Require().NoError(err)

		pub, err := k.GetPubKey()
		s.Require().NoError(err)

		newAddr := sdk.AccAddress(pub.Address())
		s.addrs = append(s.addrs, newAddr)
	}
}

func (s *CLITestSuite) TestPrepareConfigForTxCreateValidator() {
	chainID := "chainID"
	ip := "1.1.1.1"
	nodeID := "nodeID"
	privKey := ed25519.GenPrivKey()
	valPubKey := privKey.PubKey()
	moniker := "DefaultMoniker"
	require := s.Require()
	mkTxValCfg := func(amount, commission, commissionMax, commissionMaxChange, minSelfDelegation string) cli.TxCreateValidatorConfig {
		return cli.TxCreateValidatorConfig{
			IP:                      ip,
			ChainID:                 chainID,
			NodeID:                  nodeID,
			P2PPort:                 26656,
			PubKey:                  valPubKey,
			Moniker:                 moniker,
			Amount:                  amount,
			CommissionRate:          commission,
			CommissionMaxRate:       commissionMax,
			CommissionMaxChangeRate: commissionMaxChange,
			MinSelfDelegation:       minSelfDelegation,
		}
	}

	tests := []struct {
		name        string
		fsModify    func(fs *pflag.FlagSet)
		expectedCfg cli.TxCreateValidatorConfig
	}{
		{
			name:        "all defaults",
			fsModify:    func(fs *pflag.FlagSet) {},
			expectedCfg: mkTxValCfg(cli.DefaultTokens.String()+sdk.DefaultBondDenom, "0.1", "0.2", "0.01", "1"),
		},
		{
			name: "Custom amount",
			fsModify: func(fs *pflag.FlagSet) {
				require.NoError(fs.Set(cli.FlagAmount, "2000stake"))
			},
			expectedCfg: mkTxValCfg("2000stake", "0.1", "0.2", "0.01", "1"),
		},
		{
			name: "Custom commission rate",
			fsModify: func(fs *pflag.FlagSet) {
				require.NoError(fs.Set(cli.FlagCommissionRate, "0.54"))
			},
			expectedCfg: mkTxValCfg(cli.DefaultTokens.String()+sdk.DefaultBondDenom, "0.54", "0.2", "0.01", "1"),
		},
		{
			name: "Custom commission max rate",
			fsModify: func(fs *pflag.FlagSet) {
				require.NoError(fs.Set(cli.FlagCommissionMaxRate, "0.89"))
			},
			expectedCfg: mkTxValCfg(cli.DefaultTokens.String()+sdk.DefaultBondDenom, "0.1", "0.89", "0.01", "1"),
		},
		{
			name: "Custom commission max change rate",
			fsModify: func(fs *pflag.FlagSet) {
				require.NoError(fs.Set(cli.FlagCommissionMaxChangeRate, "0.55"))
			},
			expectedCfg: mkTxValCfg(cli.DefaultTokens.String()+sdk.DefaultBondDenom, "0.1", "0.2", "0.55", "1"),
		},
		{
			name: "Custom min self delegations",
			fsModify: func(fs *pflag.FlagSet) {
				require.NoError(fs.Set(cli.FlagMinSelfDelegation, "0.33"))
			},
			expectedCfg: mkTxValCfg(cli.DefaultTokens.String()+sdk.DefaultBondDenom, "0.1", "0.2", "0.01", "0.33"),
		},
	}

	for _, tc := range tests {
		tc := tc
		s.Run(tc.name, func() {
			fs, _ := cli.CreateValidatorMsgFlagSet(ip)
			fs.String(flags.FlagName, "", "name of private key with which to sign the gentx")

			tc.fsModify(fs)

			cvCfg, err := cli.PrepareConfigForTxCreateValidator(fs, moniker, nodeID, chainID, valPubKey)
			require.NoError(err)

			require.Equal(tc.expectedCfg, cvCfg)
		})
	}
}

func (s *CLITestSuite) TestNewCreateValidatorCmd() {
	require := s.Require()
	cmd := cli.NewCreateValidatorCmd(addresscodec.NewBech32Codec("cosmosvaloper"))

	validJSON := fmt.Sprintf(`
	{
  		"pubkey": {"@type":"/cosmos.crypto.ed25519.PubKey","key":"oWg2ISpLF405Jcm2vXV+2v4fnjodh6aafuIdeoW+rUw="},
  		"amount": "%dstake",
  		"moniker": "NewValidator",
		"identity": "AFAF00C4",
		"website": "https://newvalidator.io",
		"security": "contact@newvalidator.io",
		"details": "'Hey, I am a new validator. Please delegate!'",
  		"commission-rate": "0.5",
  		"commission-max-rate": "1.0",
  		"commission-max-change-rate": "0.1",
  		"min-self-delegation": "1"
	}`, 100)
	validJSONFile := testutil.WriteToNewTempFile(s.T(), validJSON)
	defer validJSONFile.Close()

	validJSONWithoutOptionalFields := fmt.Sprintf(`
	{
  		"pubkey": {"@type":"/cosmos.crypto.ed25519.PubKey","key":"oWg2ISpLF405Jcm2vXV+2v4fnjodh6aafuIdeoW+rUw="},
  		"amount": "%dstake",
  		"moniker": "NewValidator",
  		"commission-rate": "0.5",
  		"commission-max-rate": "1.0",
  		"commission-max-change-rate": "0.1",
  		"min-self-delegation": "1"
	}`, 100)
	validJSONWOOptionalFile := testutil.WriteToNewTempFile(s.T(), validJSONWithoutOptionalFields)
	defer validJSONWOOptionalFile.Close()

	noAmountJSON := `
	{
  		"pubkey": {"@type":"/cosmos.crypto.ed25519.PubKey","key":"oWg2ISpLF405Jcm2vXV+2v4fnjodh6aafuIdeoW+rUw="},
  		"moniker": "NewValidator",
  		"commission-rate": "0.5",
  		"commission-max-rate": "1.0",
  		"commission-max-change-rate": "0.1",
  		"min-self-delegation": "1"
	}`
	noAmountJSONFile := testutil.WriteToNewTempFile(s.T(), noAmountJSON)
	defer noAmountJSONFile.Close()

	noPubKeyJSON := fmt.Sprintf(`
	{
  		"amount": "%dstake",
  		"moniker": "NewValidator",
  		"commission-rate": "0.5",
  		"commission-max-rate": "1.0",
  		"commission-max-change-rate": "0.1",
  		"min-self-delegation": "1"
	}`, 100)
	noPubKeyJSONFile := testutil.WriteToNewTempFile(s.T(), noPubKeyJSON)
	defer noPubKeyJSONFile.Close()

	noMonikerJSON := fmt.Sprintf(`
	{
  		"pubkey": {"@type":"/cosmos.crypto.ed25519.PubKey","key":"oWg2ISpLF405Jcm2vXV+2v4fnjodh6aafuIdeoW+rUw="},
  		"amount": "%dstake",
  		"commission-rate": "0.5",
  		"commission-max-rate": "1.0",
  		"commission-max-change-rate": "0.1",
  		"min-self-delegation": "1"
	}`, 100)
	noMonikerJSONFile := testutil.WriteToNewTempFile(s.T(), noMonikerJSON)
	defer noMonikerJSONFile.Close()

	testCases := []struct {
		name         string
		args         []string
		expectErrMsg string
	}{
		{
			"invalid transaction (missing amount)",
			[]string{
				noAmountJSONFile.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10))).String()),
			},
			"must specify amount of coins to bond",
		},
		{
			"invalid transaction (missing pubkey)",
			[]string{
				noPubKeyJSONFile.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10))).String()),
			},
			"must specify the JSON encoded pubkey",
		},
		{
			"invalid transaction (missing moniker)",
			[]string{
				noMonikerJSONFile.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10))).String()),
			},
			"must specify the moniker name",
		},
		{
			"valid transaction with all fields",
			[]string{
				validJSONFile.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10))).String()),
			},
			"",
		},
		{
			"valid transaction without optional fields",
			[]string{
				validJSONWOOptionalFile.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10))).String()),
			},
			"",
		},
	}
	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			out, err := clitestutil.ExecTestCLICmd(s.clientCtx, cmd, tc.args)
			if tc.expectErrMsg != "" {
				require.Error(err)
				require.Contains(err.Error(), tc.expectErrMsg)
			} else {
				require.NoError(err, "test: %s\noutput: %s", tc.name, out.String())
				resp := &sdk.TxResponse{}
				err = s.clientCtx.Codec.UnmarshalJSON(out.Bytes(), resp)
				require.NoError(err, out.String(), "test: %s, output\n:", tc.name, out.String())
			}
		})
	}
}

func (s *CLITestSuite) TestNewEditValidatorCmd() {
	cmd := cli.NewEditValidatorCmd(addresscodec.NewBech32Codec("cosmos"))

	moniker := "testing"
	details := "bio"
	identity := "test identity"
	securityContact := "test contact"
	website := "https://test.com"

	testCases := []struct {
		name         string
		args         []string
		expectErrMsg string
	}{
		{
			"wrong from address",
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, "with wrong from address"),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10))).String()),
			},
			"key not found",
		},
		{
			"valid with no edit flag (since all are optional)",
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10))).String()),
			},
			"",
		},
		{
			"valid with edit validator details",
			[]string{
				fmt.Sprintf("--details=%s", details),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10))).String()),
			},
			"",
		},
		{
			"edit validator identity",
			[]string{
				fmt.Sprintf("--identity=%s", identity),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10))).String()),
			},
			"",
		},
		{
			"edit validator security-contact",
			[]string{
				fmt.Sprintf("--security-contact=%s", securityContact),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10))).String()),
			},
			"",
		},
		{
			"edit validator website",
			[]string{
				fmt.Sprintf("--website=%s", website),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10))).String()),
			},
			"",
		},
		{
			"edit validator moniker", // https://github.com/cosmos/cosmos-sdk/issues/10660
			[]string{
				fmt.Sprintf("--%s=%s", cli.FlagEditMoniker, moniker),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10))).String()),
			},
			"",
		},
		{
			"with all edit flags",
			[]string{
				fmt.Sprintf("--%s=%s", cli.FlagEditMoniker, moniker),
				fmt.Sprintf("--details=%s", details),
				fmt.Sprintf("--identity=%s", identity),
				fmt.Sprintf("--security-contact=%s", securityContact),
				fmt.Sprintf("--website=%s", website),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10))).String()),
			},
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			out, err := clitestutil.ExecTestCLICmd(s.clientCtx, cmd, tc.args)
			if tc.expectErrMsg != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				resp := &sdk.TxResponse{}
				s.Require().NoError(s.clientCtx.Codec.UnmarshalJSON(out.Bytes(), resp))
			}
		})
	}
}

func (s *CLITestSuite) TestNewDelegateCmd() {
	cmd := cli.NewDelegateCmd(addresscodec.NewBech32Codec("cosmosvaloper"), addresscodec.NewBech32Codec("cosmos"))

	testCases := []struct {
		name         string
		args         []string
		expectErrMsg string
	}{
		{
			"invalid delegate amount",
			[]string{
				sdk.ValAddress(s.addrs[0]).String(),
				"fooCoin",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10))).String()),
			},
			"invalid decimal coin expression: fooCoin",
		},
		{
			"invalid validator address",
			[]string{
				"abc",
				sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(150)).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10))).String()),
			},
			"decoding bech32 failed",
		},
		{
			"valid transaction of delegate",
			[]string{
				sdk.ValAddress(s.addrs[0]).String(),
				sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(150)).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10))).String()),
			},
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			out, err := clitestutil.ExecTestCLICmd(s.clientCtx, cmd, tc.args)
			if tc.expectErrMsg != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				resp := &sdk.TxResponse{}
				s.Require().NoError(s.clientCtx.Codec.UnmarshalJSON(out.Bytes(), resp))
			}
		})
	}
}

func (s *CLITestSuite) TestNewRedelegateCmd() {
	cmd := cli.NewRedelegateCmd(addresscodec.NewBech32Codec("cosmosvaloper"), addresscodec.NewBech32Codec("cosmos"))

	testCases := []struct {
		name         string
		args         []string
		expectErrMsg string
	}{
		{
			"invalid amount",
			[]string{
				sdk.ValAddress(s.addrs[0]).String(), // src-validator-addr
				sdk.ValAddress(s.addrs[1]).String(), // dst-validator-addr
				"fooCoin",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10))).String()),
			},
			"invalid decimal coin expression: fooCoin",
		},
		{
			"wrong src validator",
			[]string{
				"invalid",                           // wrong src-validator-addr
				sdk.ValAddress(s.addrs[1]).String(), // dst-validator-addr
				sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(150)).String(), // amount
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=%d", flags.FlagGas, 300000),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10))).String()),
			},
			"invalid bech32",
		},
		{
			"wrong dst validator",
			[]string{
				sdk.ValAddress(s.addrs[0]).String(), // src-validator-addr
				"invalid",                           // wrong dst-validator-addr
				sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(150)).String(), // amount
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=%d", flags.FlagGas, 300000),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10))).String()),
			},
			"invalid bech32",
		},
		{
			"valid transaction of delegate",
			[]string{
				sdk.ValAddress(s.addrs[0]).String(),                             // src-validator-addr
				sdk.ValAddress(s.addrs[1]).String(),                             // dst-validator-addr
				sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(150)).String(), // amount
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=%d", flags.FlagGas, 300000),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10))).String()),
			},
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			out, err := clitestutil.ExecTestCLICmd(s.clientCtx, cmd, tc.args)
			if tc.expectErrMsg != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				resp := &sdk.TxResponse{}
				s.Require().NoError(s.clientCtx.Codec.UnmarshalJSON(out.Bytes(), resp))
			}
		})
	}
}

func (s *CLITestSuite) TestNewUnbondCmd() {
	cmd := cli.NewUnbondCmd(addresscodec.NewBech32Codec("cosmosvaloper"), addresscodec.NewBech32Codec("cosmos"))

	testCases := []struct {
		name         string
		args         []string
		expectErrMsg string
	}{
		{
			"invalid unbond amount",
			[]string{
				sdk.ValAddress(s.addrs[0]).String(),
				"foo",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10))).String()),
			},
			"invalid decimal coin expression: foo",
		},
		{
			"invalid validator address",
			[]string{
				"foo",
				sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(150)).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10))).String()),
			},
			"decoding bech32 failed",
		},
		{
			"valid transaction of unbond",
			[]string{
				sdk.ValAddress(s.addrs[0]).String(),
				sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(150)).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10))).String()),
			},
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			out, err := clitestutil.ExecTestCLICmd(s.clientCtx, cmd, tc.args)
			if tc.expectErrMsg != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				resp := &sdk.TxResponse{}
				s.Require().NoError(s.clientCtx.Codec.UnmarshalJSON(out.Bytes(), resp))
			}
		})
	}
}

func (s *CLITestSuite) TestNewCancelUnbondingDelegationCmd() {
	cmd := cli.NewCancelUnbondingDelegation(addresscodec.NewBech32Codec("cosmosvaloper"), addresscodec.NewBech32Codec("cosmos"))

	testCases := []struct {
		name         string
		args         []string
		expectErrMsg string
	}{
		{
			"invalid validator address",
			[]string{
				"foo",
				sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(150)).String(),
				sdkmath.NewInt(10000).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10))).String()),
			},
			"decoding bech32 failed",
		},
		{
			"invalid canceling unbond delegation amount",
			[]string{
				sdk.ValAddress(s.addrs[0]).String(),
				"fooCoin",
				sdkmath.NewInt(10000).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10))).String()),
			},
			"invalid decimal coin expression",
		},
		{
			"without unbond creation height",
			[]string{
				sdk.ValAddress(s.addrs[0]).String(),
				sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(150)).String(),
				"abc",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10))).String()),
			},
			"invalid height: invalid height: 0",
		},
		{
			"valid transaction of canceling unbonding delegation",
			[]string{
				sdk.ValAddress(s.addrs[0]).String(),
				sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(5)).String(),
				sdkmath.NewInt(10000).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10))).String()),
			},
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			out, err := clitestutil.ExecTestCLICmd(s.clientCtx, cmd, tc.args)
			if tc.expectErrMsg != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				resp := &sdk.TxResponse{}
				s.Require().NoError(s.clientCtx.Codec.UnmarshalJSON(out.Bytes(), resp))
			}
		})
	}
}
