package sims

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

// SetupSimulation creates the config, db (levelDB), temporary directory and logger for the simulation tests.
// If `skip` is false it skips the current test. `skip` should be set using the `FlagEnabledValue` flag.
// Returns error on an invalid db intantiation or temp dir creation.
func SetupSimulation(config simtypes.Config, dirPrefix, dbName string, verbose, skip bool) (dbm.DB, string, log.Logger, bool, error) {
	if !skip {
		return nil, "", nil, true, nil
	}

	var logger log.Logger
	if verbose {
		logger = log.NewLogger(os.Stdout) // TODO(mr): enable selection of log destination.
	} else {
		logger = log.NewNopLogger()
	}

	dir, err := os.MkdirTemp("", dirPrefix)
	if err != nil {
		return nil, "", nil, false, err
	}

	db, err := dbm.NewDB(dbName, dbm.BackendType(config.DBBackend), dir)
	if err != nil {
		return nil, "", nil, false, err
	}

	return db, dir, logger, false, nil
}

// SimulationOperations retrieves the simulation params from the provided file path
// and returns all the modules weighted operations
func SimulationOperations(app runtime.AppI, cdc codec.JSONCodec, config simtypes.Config) []simtypes.WeightedOperation {
	simState := module.SimulationState{
		AppParams: make(simtypes.AppParams),
		Cdc:       cdc,
		BondDenom: sdk.DefaultBondDenom,
	}

	if config.ParamsFile != "" {
		bz, err := os.ReadFile(config.ParamsFile)
		if err != nil {
			panic(err)
		}

		err = json.Unmarshal(bz, &simState.AppParams)
		if err != nil {
			panic(err)
		}
	}

	simState.LegacyProposalContents = app.SimulationManager().GetProposalContents(simState) //nolint:staticcheck
	simState.ProposalMsgs = app.SimulationManager().GetProposalMsgs(simState)
	return app.SimulationManager().WeightedOperations(simState)
}

// CheckExportSimulation exports the app state and simulation parameters to JSON
// if the export paths are defined.
func CheckExportSimulation(app runtime.AppI, config simtypes.Config, params simtypes.Params) error {
	if config.ExportStatePath != "" {
		fmt.Println("exporting app state...")
		exported, err := app.ExportAppStateAndValidators(false, nil, nil)
		if err != nil {
			return err
		}

		if err := os.WriteFile(config.ExportStatePath, []byte(exported.AppState), 0o600); err != nil {
			return err
		}
	}

	if config.ExportParamsPath != "" {
		fmt.Println("exporting simulation params...")
		paramsBz, err := json.MarshalIndent(params, "", " ")
		if err != nil {
			return err
		}

		if err := os.WriteFile(config.ExportParamsPath, paramsBz, 0o600); err != nil {
			return err
		}
	}
	return nil
}

// PrintStats prints the corresponding statistics from the app DB.
func PrintStats(db dbm.DB) {
	fmt.Println("\nLevelDB Stats")
	fmt.Println(db.Stats()["leveldb.stats"])
	fmt.Println("LevelDB cached block size", db.Stats()["leveldb.cachedblock"])
}

// GetSimulationLog unmarshals the KVPair's Value to the corresponding type based on the
// each's module store key and the prefix bytes of the KVPair's key.
func GetSimulationLog(storeName string, sdr simtypes.StoreDecoderRegistry, kvAs, kvBs []kv.Pair) (log string) {
	for i := 0; i < len(kvAs); i++ {
		if len(kvAs[i].Value) == 0 && len(kvBs[i].Value) == 0 {
			// skip if the value doesn't have any bytes
			continue
		}

		decoder, ok := sdr[storeName]
		if ok {
			log += decoder(kvAs[i], kvBs[i])
		} else {
			log += fmt.Sprintf("store A %X => %X\nstore B %X => %X\n", kvAs[i].Key, kvAs[i].Value, kvBs[i].Key, kvBs[i].Value)
		}
	}

	return log
}

// DiffKVStores compares two KVstores and returns all the key/value pairs
// that differ from one another. It also skips value comparison for a set of provided prefixes.
func DiffKVStores(a, b storetypes.KVStore, prefixesToSkip [][]byte) (kvAs, kvBs []kv.Pair) {
	iterA := a.Iterator(nil, nil)

	defer iterA.Close()

	iterB := b.Iterator(nil, nil)

	defer iterB.Close()

	for {
		if !iterA.Valid() && !iterB.Valid() {
			return kvAs, kvBs
		}

		var kvA, kvB kv.Pair
		if iterA.Valid() {
			kvA = kv.Pair{Key: iterA.Key(), Value: iterA.Value()}

			iterA.Next()
		}

		if iterB.Valid() {
			kvB = kv.Pair{Key: iterB.Key(), Value: iterB.Value()}
		}

		compareValue := true

		for _, prefix := range prefixesToSkip {
			// Skip value comparison if we matched a prefix
			if bytes.HasPrefix(kvA.Key, prefix) {
				compareValue = false
				break
			}
		}

		if !compareValue {
			// We're skipping this key due to an exclusion prefix.  If it's present in B, iterate past it.  If it's
			// absent don't iterate.
			if bytes.Equal(kvA.Key, kvB.Key) {
				iterB.Next()
			}
			continue
		}

		// always iterate B when comparing
		iterB.Next()

		if !bytes.Equal(kvA.Key, kvB.Key) || !bytes.Equal(kvA.Value, kvB.Value) {
			kvAs = append(kvAs, kvA)
			kvBs = append(kvBs, kvB)
		}
	}
}
