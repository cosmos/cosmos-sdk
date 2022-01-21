package server_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	dbm "github.com/cosmos/cosmos-sdk/db"
	"github.com/cosmos/cosmos-sdk/db/memdb"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/simapp"
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

	var exportedGenDoc tmtypes.GenesisDoc
	err := tmjson.Unmarshal(output.Bytes(), &exportedGenDoc)
	if err != nil {
		t.Fatalf("error unmarshaling exported genesis doc: %s", err)
	}

	require.Equal(t, simapp.DefaultConsensusParams.Block.MaxBytes, exportedGenDoc.ConsensusParams.Block.MaxBytes)
	require.Equal(t, simapp.DefaultConsensusParams.Block.MaxGas, exportedGenDoc.ConsensusParams.Block.MaxGas)

	require.Equal(t, simapp.DefaultConsensusParams.Evidence.MaxAgeDuration, exportedGenDoc.ConsensusParams.Evidence.MaxAgeDuration)
	require.Equal(t, simapp.DefaultConsensusParams.Evidence.MaxAgeNumBlocks, exportedGenDoc.ConsensusParams.Evidence.MaxAgeNumBlocks)

	require.Equal(t, simapp.DefaultConsensusParams.Validator.PubKeyTypes, exportedGenDoc.ConsensusParams.Validator.PubKeyTypes)
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
				app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: i}})
				app.Commit()
			}

			output := &bytes.Buffer{}
			cmd.SetOut(output)
			args := append(tc.flags, fmt.Sprintf("--%s=%s", flags.FlagHome, tempDir))
			cmd.SetArgs(args)
			require.NoError(t, cmd.ExecuteContext(ctx), tc.name)

			var exportedGenDoc tmtypes.GenesisDoc
			err := tmjson.Unmarshal(output.Bytes(), &exportedGenDoc)
			if err != nil {
				t.Fatalf("error unmarshaling exported genesis doc: %s", err)
			}

			require.Equal(t, tc.expHeight, exportedGenDoc.InitialHeight)
		})
	}

}

func setupApp(t *testing.T, tempDir string) (*simapp.SimApp, context.Context, *tmtypes.GenesisDoc, *cobra.Command) {
	t.Helper()

	if err := createConfigFolder(tempDir); err != nil {
		t.Fatalf("error creating config folder: %s", err)
	}

	logger, _ := log.NewDefaultLogger("plain", "info", false)
	db := memdb.NewDB()
	encCfg := simapp.MakeTestEncodingConfig()
	app := simapp.NewSimApp(logger, db, nil, true, map[int64]bool{}, tempDir, 0, encCfg, simapp.EmptyAppOptions{})

	genesisState := simapp.GenesisStateWithSingleValidator(t, app)
	stateBytes, err := tmjson.MarshalIndent(genesisState, "", " ")
	require.NoError(t, err)

	serverCtx := server.NewDefaultContext()
	serverCtx.Config.RootDir = tempDir

	clientCtx := client.Context{}.WithCodec(app.AppCodec())
	genDoc := &tmtypes.GenesisDoc{}
	genDoc.ChainID = "theChainId"
	genDoc.Validators = nil
	genDoc.AppState = stateBytes

	require.NoError(t, saveGenesisFile(genDoc, serverCtx.Config.GenesisFile()))
	app.InitChain(
		abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: simapp.DefaultConsensusParams,
			AppStateBytes:   genDoc.AppState,
		},
	)
	app.Commit()

	cmd := server.ExportCmd(
		func(_ log.Logger, _ dbm.DBConnection, _ io.Writer, height int64, forZeroHeight bool, jailAllowedAddrs []string, appOptons types.AppOptions) (types.ExportedApp, error) {
			require.NoError(t, app.CloseStore())
			encCfg := simapp.MakeTestEncodingConfig()
			simApp := simapp.NewSimApp(logger, db, nil, true, map[int64]bool{}, "", 0, encCfg, appOptons)
			require.NoError(t, simApp.Init())

			return simApp.ExportAppStateAndValidatorsAt(forZeroHeight, jailAllowedAddrs, height)
		}, tempDir)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
	ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

	return app, ctx, genDoc, cmd
}

func createConfigFolder(dir string) error {
	return os.Mkdir(path.Join(dir, "config"), 0700)
}

func saveGenesisFile(genDoc *tmtypes.GenesisDoc, dir string) error {
	err := genutil.ExportGenesisFile(genDoc, dir)
	if err != nil {
		return errors.Wrap(err, "error creating file")
	}

	return nil
}
