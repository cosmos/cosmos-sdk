package simapp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	"github.com/cosmos/cosmos-sdk/simapp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// SetupSimulation creates the config, db (levelDB), temporary directory and logger for
// the simulation tests. If `FlagEnabledValue` is false it skips the current test.
// Returns error on an invalid db intantiation or temp dir creation.
func SetupSimulation(dirPrefix, dbName string) (simulation.Config, dbm.DB, string, log.Logger, bool, error) {
	if !FlagEnabledValue {
		return simulation.Config{}, nil, "", nil, true, nil
	}

	config := NewConfigFromFlags()
	config.ChainID = helpers.SimAppChainID

	var logger log.Logger
	if FlagVerboseValue {
		logger = log.TestingLogger()
	} else {
		logger = log.NewNopLogger()
	}

	dir, err := ioutil.TempDir("", dirPrefix)
	if err != nil {
		return simulation.Config{}, nil, "", nil, false, err
	}

	db, err := sdk.NewLevelDB(dbName, dir)
	if err != nil {
		return simulation.Config{}, nil, "", nil, false, err
	}

	return config, db, dir, logger, false, nil
}

// RunSimulation runs a randomized simulation from the operations defined in an
// app. It also exports the state and params and prints the DB stats.
func RunSimulation(
	tb testing.TB, w io.Writer, db dbm.DB, app types.App, bapp *baseapp.BaseApp,
	config simulation.Config,
) error {
	// run randomized simulation
	stopEarly, simParams, simErr := simulation.SimulateFromSeed(
		tb, w, bapp, AppStateFn(app.Codec(), app.SimulationManager()),
		SimulationOperations(app, app.Codec(), config),
		app.ModuleAccountAddrs(), config,
	)

	// export state and params before the simulation error is checked
	if config.ExportStatePath != "" {
		if err := ExportStateToJSON(app, config.ExportStatePath); err != nil {
			return err
		}
	}

	if config.ExportParamsPath != "" {
		if err := ExportParamsToJSON(simParams, config.ExportParamsPath); err != nil {
			return err
		}
	}

	if simErr != nil {
		return simErr
	}

	if config.Commit {
		fmt.Println("\nLevelDB Stats")
		fmt.Println(db.Stats()["leveldb.stats"])
		fmt.Println("LevelDB cached block size", db.Stats()["leveldb.cachedblock"])
	}

	if stopEarly {
		return errors.New("can't export or import a zero-validator genesis")
	}

	return nil
}

// SimulationOperations retrieves the simulation params from the provided file path
// and returns all the modules weighted operations
func SimulationOperations(app types.App, cdc *codec.Codec, config simulation.Config) []simulation.WeightedOperation {
	simState := module.SimulationState{
		AppParams: make(simulation.AppParams),
		Cdc:       cdc,
	}

	if config.ParamsFile != "" {
		bz, err := ioutil.ReadFile(config.ParamsFile)
		if err != nil {
			panic(err)
		}

		app.Codec().MustUnmarshalJSON(bz, &simState.AppParams)
	}

	simState.ParamChanges = app.SimulationManager().GenerateParamChanges(config.Seed)
	simState.Contents = app.SimulationManager().GetProposalContents(simState)
	return app.SimulationManager().WeightedOperations(simState)
}

// ExportStateToJSON util function to export the app state to JSON
func ExportStateToJSON(app types.App, path string) error {
	fmt.Println("exporting app state...")
	appState, _, err := app.ExportAppStateAndValidators(false, nil)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, []byte(appState), 0644)
}

// ExportParamsToJSON util function to export the simulation parameters to JSON
func ExportParamsToJSON(params simulation.Params, path string) error {
	fmt.Println("exporting simulation params...")
	paramsBz, err := json.MarshalIndent(params, "", " ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, paramsBz, 0644)
}

// GetSimulationLog unmarshals the KVPair's Value to the corresponding type based on the
// each's module store key and the prefix bytes of the KVPair's key.
func GetSimulationLog(storeName string, sdr sdk.StoreDecoderRegistry, cdc *codec.Codec, kvAs, kvBs []cmn.KVPair) (log string) {
	for i := 0; i < len(kvAs); i++ {

		if len(kvAs[i].Value) == 0 && len(kvBs[i].Value) == 0 {
			// skip if the value doesn't have any bytes
			continue
		}

		decoder, ok := sdr[storeName]
		if ok {
			log += decoder(cdc, kvAs[i], kvBs[i])
		} else {
			log += fmt.Sprintf("store A %X => %X\nstore B %X => %X\n", kvAs[i].Key, kvAs[i].Value, kvBs[i].Key, kvBs[i].Value)
		}
	}

	return
}
