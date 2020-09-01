package server

import (
	"bytes"
	"context"
	"encoding/json"
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
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/genutil"
)

func TestExportCmd_ConsensusParams(t *testing.T) {
	tempDir, clean := testutil.NewTestCaseDir(t)
	defer clean()

	_, ctx, genDoc, cmd := setupApp(t, tempDir)

	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetArgs([]string{fmt.Sprintf("--%s=%s", flags.FlagHome, tempDir)})
	require.NoError(t, cmd.ExecuteContext(ctx))

	var exportedGenDoc tmtypes.GenesisDoc
	err := tmjson.Unmarshal(output.Bytes(), &exportedGenDoc)
	if err != nil {
		t.Fatalf("error unmarshaling exported genesis doc: %s", err)
	}

	require.Equal(t, genDoc.ConsensusParams.Block.TimeIotaMs, exportedGenDoc.ConsensusParams.Block.TimeIotaMs)
	require.Equal(t, simapp.DefaultConsensusParams.Block.MaxBytes, exportedGenDoc.ConsensusParams.Block.MaxBytes)
	require.Equal(t, simapp.DefaultConsensusParams.Block.MaxGas, exportedGenDoc.ConsensusParams.Block.MaxGas)

	require.Equal(t, simapp.DefaultConsensusParams.Evidence.MaxAgeDuration, exportedGenDoc.ConsensusParams.Evidence.MaxAgeDuration)
	require.Equal(t, simapp.DefaultConsensusParams.Evidence.MaxAgeNumBlocks, exportedGenDoc.ConsensusParams.Evidence.MaxAgeNumBlocks)

	require.Equal(t, simapp.DefaultConsensusParams.Validator.PubKeyTypes, exportedGenDoc.ConsensusParams.Validator.PubKeyTypes)
}

func TestExportCmd_Height(t *testing.T) {
	testCases := []struct {
		name      string
		flags     []string
		expHeight int64
	}{
		{
			"should export correct height",
			[]string{},
			4,
		},
		{
			"should export height 0 with --for-zero-height",
			[]string{
				fmt.Sprintf("--%s=%s", FlagForZeroHeight, "true"),
			},
			0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempDir, clean := testutil.NewTestCaseDir(t)
			defer clean()

			app, ctx, _, cmd := setupApp(t, tempDir)

			// Fast forward to block 3.
			app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: 2}})
			app.Commit()
			app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: 3}})
			app.Commit()

			output := &bytes.Buffer{}
			cmd.SetOut(output)
			args := append(tc.flags, fmt.Sprintf("--%s=%s", flags.FlagHome, tempDir))
			cmd.SetArgs(args)
			require.NoError(t, cmd.ExecuteContext(ctx))

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
	err := createConfigFolder(tempDir)
	if err != nil {
		t.Fatalf("error creating config folder: %s", err)
	}

	db := dbm.NewMemDB()
	app := simapp.NewSimApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, map[int64]bool{}, tempDir, 0, simapp.MakeEncodingConfig())

	serverCtx := NewDefaultContext()
	serverCtx.Config.RootDir = tempDir

	clientCtx := client.Context{}.WithJSONMarshaler(app.AppCodec())

	genDoc := newDefaultGenesisDoc()
	err = saveGenesisFile(genDoc, serverCtx.Config.GenesisFile())
	require.NoError(t, err)

	app.InitChain(
		abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: simapp.DefaultConsensusParams,
			AppStateBytes:   genDoc.AppState,
		},
	)

	app.Commit()

	cmd := ExportCmd(
		func(logger log.Logger, db dbm.DB, writer io.Writer, i int64, forZeroHeight bool, jailAllowedAddrs []string) (types.ExportedApp, error) {
			return app.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs)
		}, tempDir)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
	ctx = context.WithValue(ctx, ServerContextKey, serverCtx)

	return app, ctx, genDoc, cmd
}

func createConfigFolder(dir string) error {
	return os.Mkdir(path.Join(dir, "config"), 0700)
}

func newDefaultGenesisDoc() *tmtypes.GenesisDoc {
	genesisState := simapp.NewDefaultGenesisState()

	stateBytes, err := json.MarshalIndent(genesisState, "", "  ")
	if err != nil {
		panic(err)
	}

	genDoc := &tmtypes.GenesisDoc{}
	genDoc.ChainID = "theChainId"
	genDoc.Validators = nil
	genDoc.AppState = stateBytes

	return genDoc
}

func saveGenesisFile(genDoc *tmtypes.GenesisDoc, dir string) error {
	err := genutil.ExportGenesisFile(genDoc, dir)
	if err != nil {
		return errors.Wrap(err, "error creating file")
	}

	return nil
}
