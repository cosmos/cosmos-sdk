package sims

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/types/module"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
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
		TxConfig:  moduletestutil.MakeTestTxConfig(),
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

	//nolint:staticcheck // used for legacy testing
	simState.LegacyProposalContents = app.SimulationManager().GetProposalContents(simState)
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
func DiffKVStores(a, b storetypes.KVStore, prefixesToSkip [][]byte) (diffA, diffB []kv.Pair) {
	iterA := a.Iterator(nil, nil)
	defer iterA.Close()

	iterB := b.Iterator(nil, nil)
	defer iterB.Close()

	var wg sync.WaitGroup

	wg.Add(1)
	kvAs := make([]kv.Pair, 0)
	go func() {
		defer wg.Done()
		kvAs = getKVPairs(iterA, prefixesToSkip)
	}()

	wg.Add(1)
	kvBs := make([]kv.Pair, 0)
	go func() {
		defer wg.Done()
		kvBs = getKVPairs(iterB, prefixesToSkip)
	}()

	wg.Wait()

	if len(kvAs) != len(kvBs) {
		fmt.Printf("KV stores are different: %d key/value pairs in store A and %d key/value pairs in store B\n", len(kvAs), len(kvBs))
	}

	return getDiffFromKVPair(kvAs, kvBs)
}

// getDiffFromKVPair compares two KVstores and returns all the key/value pairs
func getDiffFromKVPair(kvAs, kvBs []kv.Pair) (diffA, diffB []kv.Pair) {
	// we assume that kvBs is equal or larger than kvAs
	// if not, we swap the two
	if len(kvAs) > len(kvBs) {
		kvAs, kvBs = kvBs, kvAs
		// we need to swap the diffA and diffB as well
		defer func() {
			diffA, diffB = diffB, diffA
		}()
	}

	// in case kvAs is empty we can return early
	// since there is nothing to compare
	// if kvAs == kvBs, then diffA and diffB will be empty
	if len(kvAs) == 0 {
		return []kv.Pair{}, kvBs
	}

	index := make(map[string][]byte, len(kvBs))
	for _, kv := range kvBs {
		index[string(kv.Key)] = kv.Value
	}

	for _, kvA := range kvAs {
		if kvBValue, ok := index[string(kvA.Key)]; !ok {
			diffA = append(diffA, kvA)
			diffB = append(diffB, kv.Pair{Key: kvA.Key}) // the key is missing from kvB so we append a pair with an empty value
		} else if !bytes.Equal(kvA.Value, kvBValue) {
			diffA = append(diffA, kvA)
			diffB = append(diffB, kv.Pair{Key: kvA.Key, Value: kvBValue})
		} else {
			// values are equal, so we remove the key from the index
			delete(index, string(kvA.Key))
		}
	}

	// add the remaining keys from kvBs
	for key, value := range index {
		diffA = append(diffA, kv.Pair{Key: []byte(key)}) // the key is missing from kvA so we append a pair with an empty value
		diffB = append(diffB, kv.Pair{Key: []byte(key), Value: value})
	}

	return diffA, diffB
}

func getKVPairs(iter dbm.Iterator, prefixesToSkip [][]byte) (kvs []kv.Pair) {
	for iter.Valid() {
		key, value := iter.Key(), iter.Value()

		// do not add the KV pair if the key is prefixed to be skipped.
		skip := false
		for _, prefix := range prefixesToSkip {
			if bytes.HasPrefix(key, prefix) {
				skip = true
				break
			}
		}

		if !skip {
			kvs = append(kvs, kv.Pair{Key: key, Value: value})
		}

		iter.Next()
	}

	return kvs
}
