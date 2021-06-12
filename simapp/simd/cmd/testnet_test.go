package cmd

import (
	"context"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/simapp"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

var testMbm = module.NewBasicManager(genutil.AppModuleBasic{})

func Test_TestnetCmd(t *testing.T) {
	home := t.TempDir()
	encodingConfig := simapp.MakeTestEncodingConfig()
	//logger := log.NewNopLogger()
	//cfg, err := genutiltest.CreateDefaultTendermintConfig(home)
	//require.NoError(t, err)

	//err = genutiltest.ExecInitCmd(testMbm, home, encodingConfig.Marshaler)
	//require.NoError(t, err)

	//serverCtx := server.NewContext(viper.New(), cfg, logger)
	clientCtx := client.Context{}.
		WithJSONCodec(encodingConfig.Marshaler).
		WithHomeDir(home).
		WithTxConfig(encodingConfig.TxConfig)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
	//ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

	cmd := testnetCmd(testMbm, banktypes.GenesisBalancesIterator{})
	cmd.SetArgs([]string{fmt.Sprintf("--%s=test", flags.FlagKeyringBackend)})
	err := cmd.ExecuteContext(ctx)
	require.NoError(t, err)
}
