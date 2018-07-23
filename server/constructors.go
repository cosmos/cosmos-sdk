package server

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
)

type (
	// AppCreator reflects a function that allows us to lazily initialize an
	// application using various configurations.
	AppCreator func(home string, logger log.Logger, traceStore string) (abci.Application, error)

	// AppExporter reflects a function that dumps all app state to
	// JSON-serializable structure and returns the current validator set.
	AppExporter func(home string, logger log.Logger, traceStore string) (json.RawMessage, []tmtypes.GenesisValidator, error)

	// AppCreatorInit reflects a function that performs initialization of an
	// AppCreator.
	AppCreatorInit func(log.Logger, dbm.DB, io.Writer) abci.Application

	// AppExporterInit reflects a function that performs initialization of an
	// AppExporter.
	AppExporterInit func(log.Logger, dbm.DB, io.Writer) (json.RawMessage, []tmtypes.GenesisValidator, error)
)

// ConstructAppCreator returns an application generation function.
func ConstructAppCreator(appFn AppCreatorInit, name string) AppCreator {
	return func(rootDir string, logger log.Logger, traceStore string) (abci.Application, error) {
		dataDir := filepath.Join(rootDir, "data")

		db, err := dbm.NewGoLevelDB(name, dataDir)
		if err != nil {
			return nil, err
		}

		var traceStoreWriter io.Writer
		if traceStore != "" {
			traceStoreWriter, err = os.OpenFile(
				traceStore,
				os.O_WRONLY|os.O_APPEND|os.O_CREATE,
				0666,
			)
			if err != nil {
				return nil, err
			}
		}

		app := appFn(logger, db, traceStoreWriter)
		return app, nil
	}
}

// ConstructAppExporter returns an application export function.
func ConstructAppExporter(appFn AppExporterInit, name string) AppExporter {
	return func(rootDir string, logger log.Logger, traceStore string) (json.RawMessage, []tmtypes.GenesisValidator, error) {
		dataDir := filepath.Join(rootDir, "data")

		db, err := dbm.NewGoLevelDB(name, dataDir)
		if err != nil {
			return nil, nil, err
		}

		var traceStoreWriter io.Writer
		if traceStore != "" {
			traceStoreWriter, err = os.OpenFile(
				traceStore,
				os.O_WRONLY|os.O_APPEND|os.O_CREATE,
				0666,
			)
			if err != nil {
				return nil, nil, err
			}
		}

		return appFn(logger, db, traceStoreWriter)
	}
}
