package cli_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	rpcclientmock "github.com/cometbft/cometbft/rpc/client/mock"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/log/v2"
	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/server"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	testutilmod "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltest "github.com/cosmos/cosmos-sdk/x/genutil/client/testutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingcli "github.com/cosmos/cosmos-sdk/x/staking/client/cli"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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
	s.encCfg = testutilmod.MakeTestEncodingConfig(genutil.AppModuleBasic{})
	s.kr = keyring.NewInMemory(s.encCfg.Codec)
	s.baseCtx = client.Context{}.
		WithKeyring(s.kr).
		WithTxConfig(s.encCfg.TxConfig).
		WithCodec(s.encCfg.Codec).
		WithClient(clitestutil.MockCometRPC{Client: rpcclientmock.Client{}}).
		WithAccountRetriever(client.MockAccountRetriever{}).
		WithOutput(io.Discard).
		WithChainID("test-chain")

	ctxGen := func() client.Context {
		bz, _ := s.encCfg.Codec.Marshal(&sdk.TxResponse{})
		c := clitestutil.NewMockCometRPC(abci.ResponseQuery{
			Value: bz,
		})
		return s.baseCtx.WithClient(c)
	}
	s.clientCtx = ctxGen()
}

func (s *CLITestSuite) TestGenTxCmd() {
	amount := sdk.NewCoin("stake", sdkmath.NewInt(12))

	tests := []struct {
		name         string
		args         []string
		expCmdOutput string
	}{
		{
			name: "invalid commission rate returns error",
			args: []string{
				fmt.Sprintf("--%s=%s", flags.FlagChainID, s.baseCtx.ChainID),
				fmt.Sprintf("--%s=1", stakingcli.FlagCommissionRate),
				"node0",
				amount.String(),
			},
			expCmdOutput: fmt.Sprintf("--%s=%s --%s=1 %s %s", flags.FlagChainID, s.baseCtx.ChainID, stakingcli.FlagCommissionRate, "node0", amount.String()),
		},
		{
			name: "valid gentx",
			args: []string{
				fmt.Sprintf("--%s=%s", flags.FlagChainID, s.baseCtx.ChainID),
				"node0",
				amount.String(),
			},
			expCmdOutput: fmt.Sprintf("--%s=%s %s %s", flags.FlagChainID, s.baseCtx.ChainID, "node0", amount.String()),
		},
		{
			name: "invalid pubkey",
			args: []string{
				fmt.Sprintf("--%s=%s", flags.FlagChainID, "test-chain-1"),
				fmt.Sprintf("--%s={\"key\":\"BOIkjkFruMpfOFC9oNPhiJGfmY2pHF/gwHdLDLnrnS0=\"}", stakingcli.FlagPubKey),
				"node0",
				amount.String(),
			},
			expCmdOutput: fmt.Sprintf("--%s=test-chain-1 --%s={\"key\":\"BOIkjkFruMpfOFC9oNPhiJGfmY2pHF/gwHdLDLnrnS0=\"} %s %s ", flags.FlagChainID, stakingcli.FlagPubKey, "node0", amount.String()),
		},
		{
			name: "valid pubkey flag",
			args: []string{
				fmt.Sprintf("--%s=%s", flags.FlagChainID, "test-chain-1"),
				fmt.Sprintf("--%s={\"@type\":\"/cosmos.crypto.ed25519.PubKey\",\"key\":\"BOIkjkFruMpfOFC9oNPhiJGfmY2pHF/gwHdLDLnrnS0=\"}", stakingcli.FlagPubKey),
				"node0",
				amount.String(),
			},
			expCmdOutput: fmt.Sprintf("--%s=test-chain-1 --%s={\"@type\":\"/cosmos.crypto.ed25519.PubKey\",\"key\":\"BOIkjkFruMpfOFC9oNPhiJGfmY2pHF/gwHdLDLnrnS0=\"} %s %s ", flags.FlagChainID, stakingcli.FlagPubKey, "node0", amount.String()),
		},
	}

	for _, tc := range tests {

		dir := s.T().TempDir()
		genTxFile := filepath.Join(dir, "myTx")
		tc.args = append(tc.args, fmt.Sprintf("--%s=%s", flags.FlagOutputDocument, genTxFile))

		s.Run(tc.name, func() {
			clientCtx := s.clientCtx
			ctx := svrcmd.CreateExecuteContext(context.Background())

			cmd := cli.GenTxCmd(
				module.NewBasicManager(),
				clientCtx.TxConfig,
				banktypes.GenesisBalancesIterator{},
				clientCtx.HomeDir,
				address.NewBech32Codec("cosmosvaloper"),
			)
			cmd.SetContext(ctx)
			cmd.SetArgs(tc.args)

			s.Require().NoError(client.SetCmdClientContextHandler(clientCtx, cmd))

			if len(tc.args) != 0 {
				s.Require().Contains(fmt.Sprint(cmd), tc.expCmdOutput)
			}
		})
	}
}

func TestGenTxCmdHonorsSignerAddress(t *testing.T) {
	home := t.TempDir()
	encCfg := testutilmod.MakeTestEncodingConfig(
		bank.AppModuleBasic{},
		staking.AppModuleBasic{},
		genutil.AppModuleBasic{},
	)
	require.NoError(t, genutiltest.ExecInitCmd(testMbm, home, encCfg.Codec))

	kr := keyring.NewInMemory(encCfg.Codec)
	record, _, err := kr.NewMnemonic(
		"rotated-key", keyring.English, hd.CreateHDPath(118, 0, 0).String(),
		keyring.DefaultBIP39Passphrase, hd.Secp256k1,
	)
	require.NoError(t, err)
	keyAddr, err := record.GetAddress()
	require.NoError(t, err)

	originalKey := secp256k1.GenPrivKey()
	signerAddr := sdk.AccAddress(originalKey.PubKey().Address())
	require.NotEqual(t, signerAddr, keyAddr)
	amount := sdk.NewInt64Coin(sdk.DefaultBondDenom, 12)

	genesisPath := filepath.Join(home, "config", "genesis.json")
	appGenesis, err := genutiltypes.AppGenesisFromFile(genesisPath)
	require.NoError(t, err)
	var appState map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(appGenesis.AppState, &appState))
	bankGenesis := banktypes.DefaultGenesisState()
	bankGenesis.Balances = []banktypes.Balance{{Address: signerAddr.String(), Coins: sdk.NewCoins(amount)}}
	bankGenesis.Supply = sdk.NewCoins(amount)
	appState[banktypes.ModuleName] = encCfg.Codec.MustMarshalJSON(bankGenesis)
	appGenesis.AppState, err = json.Marshal(appState)
	require.NoError(t, err)
	require.NoError(t, appGenesis.SaveAs(genesisPath))

	cfg, err := genutiltest.CreateDefaultCometConfig(home)
	require.NoError(t, err)
	serverCtx := server.NewContext(viper.New(), cfg, log.NewNopLogger())
	clientCtx := client.Context{}.
		WithCodec(encCfg.Codec).
		WithLegacyAmino(encCfg.Amino).
		WithTxConfig(encCfg.TxConfig).
		WithKeyring(kr).
		WithChainID(appGenesis.ChainID).
		WithHomeDir(home)

	output := filepath.Join(home, "gentx.json")
	valCodec := address.NewBech32Codec("cosmosvaloper")
	cmd := cli.GenTxCmd(testMbm, encCfg.TxConfig, banktypes.GenesisBalancesIterator{}, home, valCodec)
	cmd.PreRunE = func(cmd *cobra.Command, _ []string) error {
		return client.SetCmdClientContextHandler(clientCtx, cmd)
	}
	cmd.SetArgs([]string{
		"rotated-key",
		amount.String(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, "rotated-key"),
		fmt.Sprintf("--%s=%s", flags.FlagSignerAddress, signerAddr.String()),
		fmt.Sprintf("--%s=%s", flags.FlagOutputDocument, output),
	})
	ctx := context.WithValue(context.Background(), server.ServerContextKey, serverCtx)
	require.NoError(t, cmd.ExecuteContext(ctx))

	bz, err := os.ReadFile(output)
	require.NoError(t, err)
	tx, err := encCfg.TxConfig.TxJSONDecoder()(bz)
	require.NoError(t, err)
	msg, ok := tx.GetMsgs()[0].(*stakingtypes.MsgCreateValidator)
	require.True(t, ok)
	wantValAddr, err := valCodec.BytesToString(sdk.ValAddress(signerAddr))
	require.NoError(t, err)
	require.Equal(t, wantValAddr, msg.ValidatorAddress)
}
