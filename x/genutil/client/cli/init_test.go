package cli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	abci_server "github.com/tendermint/tendermint/abci/server"
	tmcfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/mock"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
)

var testMbm = module.NewBasicManager(genutil.AppModuleBasic{})

func createDefaultTendermintConfig(rootDir string) (*tmcfg.Config, error) {
	conf := tmcfg.DefaultConfig()
	conf.SetRoot(rootDir)
	tmcfg.EnsureRoot(rootDir)

	if err := conf.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("error in config file: %v", err)
	}

	return conf, nil
}

func TestInitCmd(t *testing.T) {
	home, cleanup := tests.NewTestCaseDir(t)
	t.Cleanup(cleanup)

	logger := log.NewNopLogger()
	cfg, err := createDefaultTendermintConfig(home)
	require.NoError(t, err)

	serverCtx := server.NewContext(viper.New(), cfg, logger)
	clientCtx := client.Context{}.WithJSONMarshaler(makeCodec()).WithHomeDir(home)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
	ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

	cmd := InitCmd(testMbm, home)
	cmd.SetArgs([]string{"appnode-test"})

	require.NoError(t, cmd.ExecuteContext(ctx))
}

func setupClientHome(t *testing.T) func() {
	clientDir, cleanup := tests.NewTestCaseDir(t)
	viper.Set(cli.HomeFlag, clientDir)
	return cleanup
}

func TestEmptyState(t *testing.T) {
	t.Cleanup(setupClientHome(t))

	home, cleanup := tests.NewTestCaseDir(t)
	t.Cleanup(cleanup)

	logger := log.NewNopLogger()
	cfg, err := createDefaultTendermintConfig(home)
	require.NoError(t, err)

	serverCtx := server.NewContext(viper.New(), cfg, logger)
	clientCtx := client.Context{}.WithJSONMarshaler(makeCodec()).WithHomeDir(home)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
	ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

	cmd := InitCmd(testMbm, home)
	cmd.SetArgs([]string{"appnode-test", fmt.Sprintf("--%s=%s", cli.HomeFlag, home)})

	require.NoError(t, cmd.ExecuteContext(ctx))

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd = server.ExportCmd(nil)
	cmd.SetArgs([]string{fmt.Sprintf("--%s=%s", cli.HomeFlag, home)})
	require.NoError(t, cmd.ExecuteContext(ctx))

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	w.Close()
	os.Stdout = old
	out := <-outC

	require.Contains(t, out, "genesis_time")
	require.Contains(t, out, "chain_id")
	require.Contains(t, out, "consensus_params")
	require.Contains(t, out, "app_hash")
	require.Contains(t, out, "app_state")
}

func TestStartStandAlone(t *testing.T) {
	home, cleanup := tests.NewTestCaseDir(t)
	t.Cleanup(cleanup)
	t.Cleanup(setupClientHome(t))

	logger := log.NewNopLogger()
	cfg, err := createDefaultTendermintConfig(home)
	require.NoError(t, err)

	serverCtx := server.NewContext(viper.New(), cfg, logger)
	clientCtx := client.Context{}.WithJSONMarshaler(makeCodec()).WithHomeDir(home)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
	ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

	cmd := InitCmd(testMbm, home)
	cmd.SetArgs([]string{"appnode-test"})

	require.NoError(t, cmd.ExecuteContext(ctx))

	app, err := mock.NewApp(home, logger)
	require.NoError(t, err)

	svrAddr, _, err := server.FreeTCPAddr()
	require.NoError(t, err)

	svr, err := abci_server.NewServer(svrAddr, "socket", app)
	require.NoError(t, err, "error creating listener")

	svr.SetLogger(logger.With("module", "abci-server"))
	err = svr.Start()
	require.NoError(t, err)

	timer := time.NewTimer(time.Duration(2) * time.Second)
	for range timer.C {
		err = svr.Stop()
		require.NoError(t, err)
		break
	}
}

func TestInitNodeValidatorFiles(t *testing.T) {
	home, cleanup := tests.NewTestCaseDir(t)
	cfg, err := createDefaultTendermintConfig(home)
	t.Cleanup(cleanup)
	nodeID, valPubKey, err := genutil.InitializeNodeValidatorFiles(cfg)
	require.Nil(t, err)
	require.NotEqual(t, "", nodeID)
	require.NotEqual(t, 0, len(valPubKey.Bytes()))
}

// custom tx codec
func makeCodec() *codec.Codec {
	var cdc = codec.New()
	sdk.RegisterCodec(cdc)
	cryptocodec.RegisterCrypto(cdc)
	return cdc
}
