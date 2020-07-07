package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/tendermint/tendermint/config"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	abciServer "github.com/tendermint/tendermint/abci/server"
	"github.com/tendermint/tendermint/libs/log"

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

func TestInitCmd(t *testing.T) {
	home, cleanup := tests.NewTestCaseDir(t)
	t.Cleanup(cleanup)
	tests.CreateConfigFolder(t, home)

	logger := log.NewNopLogger()
	cfg := config.TestConfig()
	cfg.RootDir = home

	ctx := server.NewContext(viper.New(), cfg, logger)
	cdc := makeCodec()
	cmd := InitCmd(ctx, cdc, testMbm, home)
	cmd.SetArgs([]string{
		"appnode-test",
	})

	require.NoError(t, cmd.Execute())
}

func TestEmptyState(t *testing.T) {
	home, cleanup := tests.NewTestCaseDir(t)
	t.Cleanup(cleanup)
	tests.CreateConfigFolder(t, home)

	logger := log.NewNopLogger()
	cfg := config.TestConfig()

	ctx := server.NewContext(viper.New(), cfg, logger)
	cdc := makeCodec()

	cmd := InitCmd(ctx, cdc, testMbm, home)
	cmd.SetArgs([]string{
		"appnode-test",
	})
	require.NoError(t, cmd.Execute())

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd = server.ExportCmd(ctx, cdc, nil)
	cmd.SetArgs([]string{
		fmt.Sprintf("--home=%s", home),
	})
	require.NoError(t, cmd.Execute())

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
	tests.CreateConfigFolder(t, home)

	logger := log.NewNopLogger()
	cfg := config.TestConfig()
	ctx := server.NewContext(viper.New(), cfg, logger)
	cdc := makeCodec()
	initCmd := InitCmd(ctx, cdc, testMbm, home)
	require.NoError(t, initCmd.RunE(initCmd, []string{"appnode-test"}))

	app, err := mock.NewApp(home, logger)
	require.Nil(t, err)
	svrAddr, _, err := server.FreeTCPAddr()
	require.Nil(t, err)
	svr, err := abciServer.NewServer(svrAddr, "socket", app)
	require.Nil(t, err, "error creating listener")
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
	t.Cleanup(cleanup)
	tests.CreateConfigFolder(t, home)

	cfg := config.TestConfig()
	cfg.RootDir = home
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
