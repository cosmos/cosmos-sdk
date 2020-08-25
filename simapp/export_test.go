package simapp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/genutil"
)

func TestSimappInitialHeight(t *testing.T) {
	// Create new simapp with a new genesis doc.
	tempDir1, clean := testutil.NewTestCaseDir(t)
	defer clean()
	app, clientCtx, serverCtx := setupApp(t, tempDir1)
	genDoc := newDefaultGenesisDoc()
	err := saveGenesisFile(genDoc, serverCtx.Config.GenesisFile())
	require.NoError(t, err)

	// Fast forward to block 3.
	app.InitChain(
		abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: simapp.DefaultConsensusParams,
			AppStateBytes:   genDoc.AppState,
		},
	)
	app.Commit()
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: 2}})
	app.Commit()
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: 3}})
	app.Commit()

	// Create the export command.
	cmd := server.ExportCmd(
		func(logger log.Logger, db dbm.DB, writer io.Writer, i int64, b bool, strings []string) (json.RawMessage, []tmtypes.GenesisValidator, int64, *abci.ConsensusParams, error) {
			return app.ExportAppStateAndValidators(true, []string{})
		}, tempDir1)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
	ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

	// Run the export command, save the genesis file to another folder.
	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetArgs([]string{fmt.Sprintf("--%s=%s", flags.FlagHome, tempDir1)})
	require.NoError(t, cmd.ExecuteContext(ctx))
	tempDir2, clean := testutil.NewTestCaseDir(t)
	defer clean()
	var exportedGenDoc tmtypes.GenesisDoc
	err = tmjson.Unmarshal(output.Bytes(), &exportedGenDoc)
	if err != nil {
		t.Fatalf("error unmarshaling exported genesis doc: %s", err)
	}
	err = saveGenesisFile(genDoc, tempDir2)

	// Run a new app, with exported genesis.
	app, clientCtx, serverCtx = setupApp(t, tempDir2)
	app.InitChain(
		abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: simapp.DefaultConsensusParams,
			AppStateBytes:   exportedGenDoc.AppState,
			InitialHeight:   exportedGenDoc.InitialHeight,
		},
	)
	app.Commit()

	// Check that initial height is taken into account.
	require.Equal(t, 4, app.LastBlockHeight())
}

func setupApp(t *testing.T, tempDir string) (*simapp.SimApp, client.Context, *server.Context) {
	err := os.Mkdir(path.Join(tempDir, "config"), 0700)
	if err != nil {
		t.Fatalf("error creating config folder: %s", err)
	}

	db := dbm.NewMemDB()
	app := simapp.NewSimApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, map[int64]bool{}, tempDir, 0, simapp.MakeEncodingConfig())

	serverCtx := server.NewDefaultContext()
	serverCtx.Config.RootDir = tempDir

	clientCtx := client.Context{}.WithJSONMarshaler(app.AppCodec())

	return app, clientCtx, serverCtx
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
