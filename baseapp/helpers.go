package baseapp

import (
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/tendermint/abci/server"
	abci "github.com/tendermint/abci/types"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"
)

// RunForever - BasecoinApp execution and cleanup
func RunForever(app abci.Application) {

	// Start the ABCI server
	srv, err := server.NewServer("0.0.0.0:46658", "socket", app)
	if err != nil {
		cmn.Exit(err.Error())
	}
	srv.Start()

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		srv.Stop()
	})
}

// AppCreator lets us lazily initialize app, using home dir
// and other flags (?) to start
type AppCreator func(string, log.Logger) (abci.Application, error)

// AppExporter dumps all app state to JSON-serializable structure
type AppExporter func(home string, log log.Logger) (interface{}, *wire.Codec, error)

// GenerateFn returns an application generation function
func GenerateFn(appFn func(log.Logger, dbm.DB) abci.Application, name string) AppCreator {
	return func(rootDir string, logger log.Logger) (abci.Application, error) {
		dataDir := filepath.Join(rootDir, "data")
		db, err := dbm.NewGoLevelDB(name, dataDir)
		if err != nil {
			return nil, err
		}
		app := appFn(logger, db)
		return app, nil
	}
}

// ExportFn returns an application export function
func ExportFn(appFn func(log.Logger, dbm.DB) (interface{}, *wire.Codec), name string) AppExporter {
	return func(rootDir string, logger log.Logger) (interface{}, *wire.Codec, error) {
		dataDir := filepath.Join(rootDir, "data")
		db, err := dbm.NewGoLevelDB(name, dataDir)
		if err != nil {
			return nil, nil, err
		}
		genesis, codec := appFn(logger, db)
		return genesis, codec, nil
	}
}
