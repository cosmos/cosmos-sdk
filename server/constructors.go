package server

import (
	"encoding/json"
	"path/filepath"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	tmtypes "github.com/tendermint/tendermint/types"
)

// AppCreator lets us lazily initialize app, using home dir
// and other flags (?) to start
type AppCreator func(string, *sdk.ServerContext) (abci.Application, error)

// AppExporter dumps all app state to JSON-serializable structure and returns the current validator set
type AppExporter func(home string, ctx *sdk.ServerContext) (json.RawMessage, []tmtypes.GenesisValidator, error)

// ConstructAppCreator returns an application generation function
func ConstructAppCreator(appFn func(*sdk.ServerContext, dbm.DB) abci.Application, name string) AppCreator {
	return func(rootDir string, ctx *sdk.ServerContext) (abci.Application, error) {
		dataDir := filepath.Join(rootDir, "data")
		db, err := dbm.NewGoLevelDB(name, dataDir)
		if err != nil {
			return nil, err
		}
		app := appFn(ctx, db)
		return app, nil
	}
}

// ConstructAppExporter returns an application export function
func ConstructAppExporter(appFn func(*sdk.ServerContext, dbm.DB) (json.RawMessage, []tmtypes.GenesisValidator, error), name string) AppExporter {
	return func(rootDir string, ctx *sdk.ServerContext) (json.RawMessage, []tmtypes.GenesisValidator, error) {
		dataDir := filepath.Join(rootDir, "data")
		db, err := dbm.NewGoLevelDB(name, dataDir)
		if err != nil {
			return nil, nil, err
		}
		return appFn(ctx, db)
	}
}
