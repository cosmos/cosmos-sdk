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

	abci "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	cmtcfg "github.com/cometbft/cometbft/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"

	corectx "cosmossdk.io/core/context"
	corestore "cosmossdk.io/core/store"
	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/log"
	"cosmossdk.io/simapp"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	gentestutil "github.com/cosmos/cosmos-sdk/testutil/x/genutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
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

	v := ctx.Value(corectx.ViperContextKey)
	viper, ok := v.(*viper.Viper)
	require.True(t, ok)
	viper.Set(flags.FlagHome, "foobar")

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
				_, err := app.FinalizeBlock(&abci.FinalizeBlockRequest{
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

	db := coretesting.NewMemDB()
	app := simapp.NewSimApp(logger, db, nil, true, simtestutil.NewAppOptionsWithFlagHome(tempDir))

	genesisState := simapp.GenesisStateWithSingleValidator(t, app)
	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	assert.NilError(t, err)

	viper := viper.New()
	err = gentestutil.WriteAndTrackCometConfig(viper, tempDir, cmtcfg.DefaultConfig())
	assert.NilError(t, err)

	clientCtx := client.Context{}.WithCodec(app.AppCodec())
	appGenesis := genutiltypes.AppGenesis{
		ChainID:  "theChainId",
		AppState: stateBytes,
		Consensus: &genutiltypes.ConsensusGenesis{
			Validators: []sdk.GenesisValidator{},
		},
	}

	// save genesis file
	err = genutil.ExportGenesisFile(&appGenesis, client.GetConfigFromViper(viper).GenesisFile())
	assert.NilError(t, err)

	_, err = app.InitChain(&abci.InitChainRequest{
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: simtestutil.DefaultConsensusParams,
		AppStateBytes:   appGenesis.AppState,
	})
	assert.NilError(t, err)
	_, err = app.FinalizeBlock(&abci.FinalizeBlockRequest{
		Height: 1,
	})
	assert.NilError(t, err)
	_, err = app.Commit()
	assert.NilError(t, err)

	cmd := genutilcli.ExportCmd(func(_ log.Logger, _ corestore.KVStoreWithBatch, _ io.Writer, height int64, forZeroHeight bool, jailAllowedAddrs []string, appOptions types.AppOptions, modulesToExport []string) (types.ExportedApp, error) {
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
	ctx = context.WithValue(ctx, corectx.ViperContextKey, viper)

	return app, ctx, appGenesis, cmd
}

func createConfigFolder(dir string) error {
	return os.Mkdir(path.Join(dir, "config"), 0o700)
}
