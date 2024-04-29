//go:build e2e
// +build e2e

package genutil_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"testing"
	"time"

	abci_server "github.com/cometbft/cometbft/abci/server"
	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"

	"cosmossdk.io/log"
	"cosmossdk.io/simapp"
	"cosmossdk.io/x/staking"
	"github.com/cosmos/cosmos-sdk/testutil/network"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server"
	servercmtlog "github.com/cosmos/cosmos-sdk/server/log"
	"github.com/cosmos/cosmos-sdk/server/mock"
	"github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltest "github.com/cosmos/cosmos-sdk/x/genutil/client/testutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	simtestutil "github.com/cosmos/cosmos-sdk/x/simulation/helper"
)

func TestExportCmd_ConsensusParams(t *testing.T) {
	tempDir := t.TempDir()
	_, ctx, _, cmd := setupApp(t, tempDir)

	output := &bytes.Buffer{}
	cmd.SetOut(output)
	assert.NilError(t, cmd.ExecuteContext(ctx))

	var exportedAppGenesis genutiltypes.AppGenesis
	err := json.Unmarshal(output.Bytes(), &exportedAppGenesis)
	assert.NilError(t, err)

	assert.DeepEqual(t, simtestutil.DefaultConsensusParams.Block.MaxBytes, exportedAppGenesis.Consensus.Params.Block.MaxBytes)
	assert.DeepEqual(t, simtestutil.DefaultConsensusParams.Block.MaxGas, exportedAppGenesis.Consensus.Params.Block.MaxGas)

	assert.DeepEqual(t, simtestutil.DefaultConsensusParams.Evidence.MaxAgeDuration, exportedAppGenesis.Consensus.Params.Evidence.MaxAgeDuration)
	assert.DeepEqual(t, simtestutil.DefaultConsensusParams.Evidence.MaxAgeNumBlocks, exportedAppGenesis.Consensus.Params.Evidence.MaxAgeNumBlocks)

	assert.DeepEqual(t, simtestutil.DefaultConsensusParams.Validator.PubKeyTypes, exportedAppGenesis.Consensus.Params.Validator.PubKeyTypes)
}

func TestExportCmd_HomeDir(t *testing.T) {
	_, ctx, _, cmd := setupApp(t, t.TempDir())

	serverCtxPtr := ctx.Value(server.ServerContextKey)
	serverCtxPtr.(*server.Context).Config.SetRoot("foobar")

	err := cmd.ExecuteContext(ctx)
	assert.ErrorContains(t, err, "stat foobar/config/genesis.json: no such file or directory")
}

func TestExportCmd_Height(t *testing.T) {
	testCases := []struct {
		name        string
		flags       []string
		fastForward int64
		expHeight   int64
	}{
		{
			"should export correct height",
			[]string{},
			5, 6,
		},
		{
			"should export correct height with --height",
			[]string{
				fmt.Sprintf("--height=%d", 3),
			},
			5, 4,
		},
		{
			"should export height 0 with --for-zero-height",
			[]string{
				fmt.Sprintf("--for-zero-height=%s", "true"),
			},
			2, 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := t.TempDir()
			app, ctx, _, cmd := setupApp(t, tempDir)

			// Fast forward to block `tc.fastForward`.
			for i := int64(2); i <= tc.fastForward; i++ {
				_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
					Height: i,
				})
				assert.NilError(t, err)
				_, err = app.Commit()
				assert.NilError(t, err)
			}

			output := &bytes.Buffer{}
			cmd.SetOut(output)
			cmd.SetArgs(tc.flags)
			assert.NilError(t, cmd.ExecuteContext(ctx))

			var exportedAppGenesis genutiltypes.AppGenesis
			err := json.Unmarshal(output.Bytes(), &exportedAppGenesis)
			assert.NilError(t, err)
			assert.Equal(t, tc.expHeight, exportedAppGenesis.InitialHeight)
		})
	}
}

func TestExportCmd_Output(t *testing.T) {
	testCases := []struct {
		name           string
		flags          []string
		outputDocument string
	}{
		{
			"should export state to the specified file",
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagOutputDocument, "foobar.json"),
			},
			"foobar.json",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := t.TempDir()
			_, ctx, _, cmd := setupApp(t, tempDir)

			output := &bytes.Buffer{}
			cmd.SetOut(output)
			cmd.SetArgs(tc.flags)
			assert.NilError(t, cmd.ExecuteContext(ctx))

			var exportedAppGenesis genutiltypes.AppGenesis
			f, err := os.ReadFile(tc.outputDocument)
			assert.NilError(t, err)
			assert.NilError(t, json.Unmarshal(f, &exportedAppGenesis))

			// Cleanup
			assert.NilError(t, os.Remove(tc.outputDocument))
		})
	}
}

func setupApp(t *testing.T, tempDir string) (*simapp.SimApp, context.Context, genutiltypes.AppGenesis, *cobra.Command) {
	t.Helper()

	logger := log.NewTestLogger(t)
	err := createConfigFolder(tempDir)
	assert.NilError(t, err)

	db := dbm.NewMemDB()
	app := simapp.NewSimApp(logger, db, nil, true, simtestutil.NewAppOptionsWithFlagHome(tempDir))

	genesisState := simapp.GenesisStateWithSingleValidator(t, app)
	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	assert.NilError(t, err)

	serverCtx := server.NewDefaultContext()
	serverCtx.Config.RootDir = tempDir

	clientCtx := client.Context{}.WithCodec(app.AppCodec())
	appGenesis := genutiltypes.AppGenesis{
		ChainID:  "theChainId",
		AppState: stateBytes,
		Consensus: &genutiltypes.ConsensusGenesis{
			Validators: nil,
		},
	}

	// save genesis file
	err = genutil.ExportGenesisFile(&appGenesis, serverCtx.Config.GenesisFile())
	assert.NilError(t, err)

	_, err = app.InitChain(&abci.RequestInitChain{
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: simtestutil.DefaultConsensusParams,
		AppStateBytes:   appGenesis.AppState,
	})
	assert.NilError(t, err)
	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
	})
	assert.NilError(t, err)
	_, err = app.Commit()
	assert.NilError(t, err)

	cmd := genutilcli.ExportCmd(func(_ log.Logger, _ dbm.DB, _ io.Writer, height int64, forZeroHeight bool, jailAllowedAddrs []string, appOptions types.AppOptions, modulesToExport []string) (types.ExportedApp, error) {
		var simApp *simapp.SimApp
		if height != -1 {
			simApp = simapp.NewSimApp(logger, db, nil, false, appOptions)
			if err := simApp.LoadHeight(height); err != nil {
				return types.ExportedApp{}, err
			}
		} else {
			simApp = simapp.NewSimApp(logger, db, nil, true, appOptions)
		}

		return simApp.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs, modulesToExport)
	})

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
	ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

	return app, ctx, appGenesis, cmd
}

func createConfigFolder(dir string) error {
	return os.Mkdir(path.Join(dir, "config"), 0o700)
}

func TestStartStandAlone(t *testing.T) {
	home := t.TempDir()
	logger := log.NewNopLogger()
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	err := genutiltest.ExecInitCmd(module.NewManager(
		staking.NewAppModule(codec.NewProtoCodec(interfaceRegistry), nil, nil, nil),
		genutil.NewAppModule(codec.NewProtoCodec(interfaceRegistry), nil, nil, nil, nil, nil),
	), home, marshaler)
	require.NoError(t, err)

	app, err := mock.NewApp(home, logger)
	require.NoError(t, err)

	svrAddr, _, closeFn, err := network.FreeTCPAddr()
	require.NoError(t, err)
	require.NoError(t, closeFn())

	cmtApp := server.NewCometABCIWrapper(app)
	svr, err := abci_server.NewServer(svrAddr, "socket", cmtApp)
	require.NoError(t, err, "error creating listener")

	svr.SetLogger(servercmtlog.CometLoggerWrapper{Logger: logger.With("module", "abci-server")})
	err = svr.Start()
	require.NoError(t, err)

	timer := time.NewTimer(time.Duration(2) * time.Second)
	for range timer.C {
		err = svr.Stop()
		require.NoError(t, err)
		break
	}
}
