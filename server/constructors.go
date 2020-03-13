package server

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/store/snapshot"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	// AppCreator is a function that allows us to lazily initialize an
	// application using various configurations.
	AppCreator func(log.Logger, dbm.DB, *snapshot.Store, io.Writer) abci.Application

	// AppExporter is a function that dumps all app state to
	// JSON-serializable structure and returns the current validator set.
	AppExporter func(log.Logger, dbm.DB, io.Writer, int64, bool, []string) (json.RawMessage, []tmtypes.GenesisValidator, error)
)

func openDB(rootDir string) (dbm.DB, error) {
	dataDir := filepath.Join(rootDir, "data")
	db, err := sdk.NewLevelDB("application", dataDir)
	return db, err
}

func openSnapshotStore(dataDir string) (*snapshot.Store, error) {
	db, err := sdk.NewLevelDB("snapshot", dataDir)
	if err != nil {
		return nil, err
	}
	return snapshot.New(db, filepath.Join(dataDir, "snapshots"))
}

func openTraceWriter(traceWriterFile string) (w io.Writer, err error) {
	if traceWriterFile != "" {
		w, err = os.OpenFile(
			traceWriterFile,
			os.O_WRONLY|os.O_APPEND|os.O_CREATE,
			0666,
		)
		return
	}
	return
}
