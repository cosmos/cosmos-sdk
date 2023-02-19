//go:build e2e
// +build e2e

package server_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"testing"

	"cosmossdk.io/log"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	cmtlog "github.com/cometbft/cometbft/libs/log"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/simapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/genutil"
)

func TestExportCmd_ConsensusParams(t *testing.T) {
	tempDir := t.TempDir()

	_, ctx, _, cmd := setupApp(t, tempDir)

	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetArgs([]string{fmt.Sprintf("--%s=%s", flags.FlagHome, tempDir)})
	require.NoError(t, cmd.ExecuteContext(ctx))

	var exportedGenDoc cmttypes.GenesisDoc
	err := cmtjson.Unmarshal(output.Bytes(), &exportedGenDoc)
	if err != nil {
		t.Fatalf("error unmarshaling exported genesis doc: %s", err)
	}

	require.Equal(t, simtestutil.DefaultConsensusParams.Block.MaxBytes, exportedGenDoc.ConsensusParams.Block.MaxBytes)
	require.Equal(t, simtestutil.DefaultConsensusParams.Block.MaxGas, exportedGenDoc.ConsensusParams.Block.MaxGas)

	require.Equal(t, simtestutil.DefaultConsensusParams.Evidence.MaxAgeDuration, exportedGenDoc.ConsensusParams.Evidence.MaxAgeDuration)
	require.Equal(t, simtestutil.DefaultConsensusParams.Evidence.MaxAgeNumBlocks, exportedGenDoc.ConsensusParams.Evidence.MaxAgeNumBlocks)

	require.Equal(t, simtestutil.DefaultConsensusParams.Validator.PubKeyTypes, exportedGenDoc.ConsensusParams.Validator.PubKeyTypes)
}

func TestExportCmd_HomeDir(t *testing.T) {
	_, ctx, _, cmd := setupApp(t, t.TempDir())

	cmd.SetArgs([]string{fmt.Sprintf("--%s=%s", flags.FlagHome, "foobar")})

	err := cmd.ExecuteContext(ctx)
	require.EqualError(t, err, "stat foobar/config/genesis.json: no such file or directory")
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
				fmt.Sprintf("--%s=%d", server.FlagHeight, 3),
			},
			5, 4,
		},
		{
			"should export height 0 with --for-zero-height",
			[]string{
				fmt.Sprintf("--%s=%s", server.FlagForZeroHeight, "true"),
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
				app.BeginBlock(abci.RequestBeginBlock{Header: cmtproto.Header{Height: i}})
				app.Commit()
			}

			output := &bytes.Buffer{}
			cmd.SetOut(output)
			args := append(tc.flags, fmt.Sprintf("--%s=%s", flags.FlagHome, tempDir))
			cmd.SetArgs(args)
			require.NoError(t, cmd.ExecuteContext(ctx))

			var exportedGenDoc cmttypes.GenesisDoc
			err := cmtjson.Unmarshal(output.Bytes(), &exportedGenDoc)
			if err != nil {
				t.Fatalf("error unmarshaling exported genesis doc: %s", err)
			}

			require.Equal(t, tc.expHeight, exportedGenDoc.InitialHeight)
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
				fmt.Sprintf("--%s=%s", server.FlagOutputDocument, "foobar.json"),
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
			args := append(tc.flags, fmt.Sprintf("--%s=%s", flags.FlagHome, tempDir))
			cmd.SetArgs(args)
			require.NoError(t, cmd.ExecuteContext(ctx))

			var exportedGenDoc cmttypes.GenesisDoc
			f, err := os.ReadFile(tc.outputDocument)
			if err != nil {
				t.Fatalf("error reading exported genesis doc: %s", err)
			}
			require.NoError(t, cmtjson.Unmarshal(f, &exportedGenDoc))

			// Cleanup
			if err = os.Remove(tc.outputDocument); err != nil {
				t.Fatalf("error removing exported genesis doc: %s", err)
			}
		})
	}
}

func setupApp(t *testing.T, tempDir string) (*simapp.SimApp, context.Context, *cmttypes.GenesisDoc, *cobra.Command) {
	t.Helper()

	if err := createConfigFolder(tempDir); err != nil {
		t.Fatalf("error creating config folder: %s", err)
	}

	logger := cmtlog.NewTMLogger(cmtlog.NewSyncWriter(os.Stdout))
	db := dbm.NewMemDB()
	app := simapp.NewSimApp(logger, db, nil, true, simtestutil.NewAppOptionsWithFlagHome(tempDir))

	genesisState := simapp.GenesisStateWithSingleValidator(t, app)
	stateBytes, err := cmtjson.MarshalIndent(genesisState, "", " ")
	require.NoError(t, err)

	serverCtx := server.NewDefaultContext()
	serverCtx.Config.RootDir = tempDir

	clientCtx := client.Context{}.WithCodec(app.AppCodec())
	genDoc := &cmttypes.GenesisDoc{}
	genDoc.ChainID = "theChainId"
	genDoc.Validators = nil
	genDoc.AppState = stateBytes

	require.NoError(t, saveGenesisFile(genDoc, serverCtx.Config.GenesisFile()))
	app.InitChain(
		abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: simtestutil.DefaultConsensusParams,
			AppStateBytes:   genDoc.AppState,
		},
	)
	app.Commit()

	cmd := server.ExportCmd(
		func(_ log.Logger, _ dbm.DB, _ io.Writer, height int64, forZeroHeight bool, jailAllowedAddrs []string, appOptions types.AppOptions, modulesToExport []string) (types.ExportedApp, error) {
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
		}, tempDir)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
	ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

	return app, ctx, genDoc, cmd
}

func createConfigFolder(dir string) error {
	return os.Mkdir(path.Join(dir, "config"), 0o700)
}

func saveGenesisFile(genDoc *cmttypes.GenesisDoc, dir string) error {
	err := genutil.ExportGenesisFile(genDoc, dir)
	if err != nil {
		return errors.Wrap(err, "error creating file")
	}

	return nil
}
