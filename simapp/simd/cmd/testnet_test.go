package cmd

import (
	"context"
	"fmt"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	corectx "cosmossdk.io/core/context"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	banktypes "cosmossdk.io/x/bank/types"
	"cosmossdk.io/x/staking"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	genutiltest "github.com/cosmos/cosmos-sdk/testutil/x/genutil"
	"github.com/cosmos/cosmos-sdk/types/module"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

func Test_TestnetCmd(t *testing.T) {
	config := configurator.NewAppConfig(
		configurator.AccountsModule(),
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
	require.Len(t, moduleManager.Modules, 9) // the registered above + runtime

	home := t.TempDir()
	cdcOpts := codectestutil.CodecOptions{}
	encodingConfig := moduletestutil.MakeTestEncodingConfig(cdcOpts, auth.AppModule{}, staking.AppModule{})
	logger := log.NewNopLogger()
	viper := viper.New()
	cfg, err := genutiltest.CreateDefaultCometConfig(home)
	require.NoError(t, err)

	err = genutiltest.ExecInitCmd(moduleManager, home, encodingConfig.Codec)
	require.NoError(t, err)

	err = genutiltest.WriteAndTrackCometConfig(viper, home, cfg)
	require.NoError(t, err)
	clientCtx := client.Context{}.
		WithCodec(encodingConfig.Codec).
		WithHomeDir(home).
		WithTxConfig(encodingConfig.TxConfig).
		WithAddressCodec(cdcOpts.GetAddressCodec()).
		WithValidatorAddressCodec(cdcOpts.GetValidatorCodec())

	ctx := context.Background()
	ctx = context.WithValue(ctx, corectx.ViperContextKey, viper)
	ctx = context.WithValue(ctx, corectx.LoggerContextKey, logger)
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
	cmd := testnetInitFilesCmd(moduleManager)
	cmd.SetArgs(
		[]string{fmt.Sprintf("--%s=test", flags.FlagKeyringBackend), fmt.Sprintf("--output-dir=%s", home)},
	)
	err = cmd.ExecuteContext(ctx)
	require.NoError(t, err)

	genFile := client.GetConfigFromCmd(cmd).GenesisFile()
	appState, _, err := genutiltypes.GenesisStateFromGenFile(genFile)
	require.NoError(t, err)

	bankGenState := banktypes.GetGenesisStateFromAppState(encodingConfig.Codec, appState)
	require.NotEmpty(t, bankGenState.Supply.String())
}
