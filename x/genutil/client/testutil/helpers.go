package testutil

import (
	"context"
	"fmt"

	cmtcfg "github.com/cometbft/cometbft/config"
	"github.com/spf13/viper"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/types/module"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
)

func ExecInitCmd(mm *module.Manager, home string, cdc codec.Codec) error {
	logger := log.NewNopLogger()
	cfg, err := CreateDefaultCometConfig(home)
	if err != nil {
		return err
	}

	cmd := genutilcli.InitCmd(mm)
	serverCtx := server.NewContext(viper.New(), cfg, logger)
	serverCtx.Config.SetRoot(home)
	clientCtx := client.Context{}.WithCodec(cdc).WithHomeDir(home)

	_, out := testutil.ApplyMockIO(cmd)
	clientCtx = clientCtx.WithOutput(out)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
	ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

	cmd.SetArgs([]string{"appnode-test"})

	return cmd.ExecuteContext(ctx)
}

func CreateDefaultCometConfig(rootDir string) (*cmtcfg.Config, error) {
	conf := cmtcfg.DefaultConfig()
	conf.SetRoot(rootDir)
	cmtcfg.EnsureRoot(rootDir)

	if err := conf.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("error in config file: %v", err)
	}

	return conf, nil
}
