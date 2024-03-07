package cmd

import (
	"context"
	"fmt"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/x/auth"
	banktypes "cosmossdk.io/x/bank/types"
	"cosmossdk.io/x/staking"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	"github.com/cosmos/cosmos-sdk/types/module"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	genutiltest "github.com/cosmos/cosmos-sdk/x/genutil/client/testutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

func Test_TestnetCmd(t *testing.T) {
	config := configurator.NewAppConfig(
		configurator.AuthModule(),
		configurator.BankModule(),
		configurator.GenutilModule(),
		configurator.StakingModule(),
		configurator.ConsensusModule(),
		configurator.TxModule(),
		configurator.MintModule(),
	)
	var moduleManager *module.Manager
	err := depinject.Inject(
		depinject.Configs(config,
			depinject.Supply(log.NewNopLogger()),
		),
		&moduleManager,
	)
	require.NoError(t, err)
	require.NotNil(t, moduleManager)
	require.Len(t, moduleManager.Modules, 7)

	home := t.TempDir()
	encodingConfig := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{}, auth.AppModule{}, staking.AppModule{})
	logger := log.NewNopLogger()
	cfg, err := genutiltest.CreateDefaultCometConfig(home)
	require.NoError(t, err)

	err = genutiltest.ExecInitCmd(moduleManager, home, encodingConfig.Codec)
	require.NoError(t, err)

	serverCtx := server.NewContext(viper.New(), cfg, logger)
	clientCtx := client.Context{}.
		WithCodec(encodingConfig.Codec).
		WithHomeDir(home).
		WithTxConfig(encodingConfig.TxConfig)

	ctx := context.Background()
	ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
	cmd := testnetInitFilesCmd(moduleManager, banktypes.GenesisBalancesIterator{})
	cmd.SetArgs([]string{fmt.Sprintf("--%s=test", flags.FlagKeyringBackend), fmt.Sprintf("--output-dir=%s", home)})
	err = cmd.ExecuteContext(ctx)
	require.NoError(t, err)

	genFile := cfg.GenesisFile()
	appState, _, err := genutiltypes.GenesisStateFromGenFile(genFile)
	require.NoError(t, err)

	bankGenState := banktypes.GetGenesisStateFromAppState(encodingConfig.Codec, appState)
	require.NotEmpty(t, bankGenState.Supply.String())
}
