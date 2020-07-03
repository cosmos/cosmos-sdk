package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/genutil"
)

func TestExportCmd_ConsensusParams(t *testing.T) {
	tempDir, clean := tests.NewTestCaseDir(t)
	defer clean()

	err := createConfigFolder(tempDir)
	if err != nil {
		t.Fatalf("error creating config folder: %s", err)
	}

	db := dbm.NewMemDB()
	app := simapp.NewSimApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, map[int64]bool{}, tempDir, 0)

	ctx := NewDefaultContext()
	ctx.Config.RootDir = tempDir

	genDoc := newDefaultGenesisDoc(app.Codec())
	err = saveGenesisFile(genDoc, ctx.Config.GenesisFile())

	app.InitChain(
		abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: simapp.DefaultConsensusParams,
			AppStateBytes:   genDoc.AppState,
		},
	)

	app.Commit()

	cmd := ExportCmd(
		ctx,
		app.Codec(),
		func(logger log.Logger, db dbm.DB, writer io.Writer, i int64, b bool, strings []string) (json.RawMessage, []tmtypes.GenesisValidator, *abci.ConsensusParams, error) {
			return app.ExportAppStateAndValidators(true, []string{})
		})

	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetArgs([]string{fmt.Sprintf("--%s=%s", flags.FlagHome, tempDir)})
	require.NoError(t, cmd.Execute())

	var exportedGenDoc tmtypes.GenesisDoc
	err = app.Codec().UnmarshalJSON(output.Bytes(), &exportedGenDoc)
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

func createConfigFolder(dir string) error {
	return os.Mkdir(path.Join(dir, "config"), 0700)
}

func newDefaultGenesisDoc(cdc *codec.Codec) *tmtypes.GenesisDoc {
	genesisState := simapp.NewDefaultGenesisState()

	stateBytes, err := codec.MarshalJSONIndent(cdc, genesisState)
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
