package simapp

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"

	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

//---------------------------------------------------------------------
// Flags

// List of available flags for the simulator
var (
	flagGenesisFileValue        string
	flagParamsFileValue         string
	flagExportParamsPathValue   string
	flagExportParamsHeightValue int
	flagExportStatePathValue    string
	flagExportStatsPathValue    string
	flagSeedValue               int64
	flagInitialBlockHeightValue int
	flagNumBlocksValue          int
	flagBlockSizeValue          int
	flagLeanValue               bool
	flagCommitValue             bool
	flagOnOperationValue        bool // TODO: Remove in favor of binary search for invariant violation
	flagAllInvariantsValue      bool

	flagEnabledValue     bool
	flagVerboseValue     bool
	flagPeriodValue      uint
	flagGenesisTimeValue int64
)

// GetSimulatorFlags gets the values of all the available simulation flags
func GetSimulatorFlags() {

	// Config fields
	flag.StringVar(&flagGenesisFileValue, "Genesis", "", "custom simulation genesis file; cannot be used with params file")
	flag.StringVar(&flagParamsFileValue, "Params", "", "custom simulation params file which overrides any random params; cannot be used with genesis")
	flag.StringVar(&flagExportParamsPathValue, "ExportParamsPath", "", "custom file path to save the exported params JSON")
	flag.IntVar(&flagExportParamsHeightValue, "ExportParamsHeight", 0, "height to which export the randomly generated params")
	flag.StringVar(&flagExportStatePathValue, "ExportStatePath", "", "custom file path to save the exported app state JSON")
	flag.StringVar(&flagExportStatsPathValue, "ExportStatsPath", "", "custom file path to save the exported simulation statistics JSON")
	flag.Int64Var(&flagSeedValue, "Seed", 42, "simulation random seed")
	flag.IntVar(&flagInitialBlockHeightValue, "InitialBlockHeight", 1, "initial block to start the simulation")
	flag.IntVar(&flagNumBlocksValue, "NumBlocks", 500, "number of new blocks to simulate from the initial block height")
	flag.IntVar(&flagBlockSizeValue, "BlockSize", 200, "operations per block")
	flag.BoolVar(&flagLeanValue, "Lean", false, "lean simulation log output")
	flag.BoolVar(&flagCommitValue, "Commit", false, "have the simulation commit")
	flag.BoolVar(&flagOnOperationValue, "SimulateEveryOperation", false, "run slow invariants every operation")
	flag.BoolVar(&flagAllInvariantsValue, "PrintAllInvariants", false, "print all invariants if a broken invariant is found")

	// SimApp flags
	flag.BoolVar(&flagEnabledValue, "Enabled", false, "enable the simulation")
	flag.BoolVar(&flagVerboseValue, "Verbose", false, "verbose log output")
	flag.UintVar(&flagPeriodValue, "Period", 0, "run slow invariants only once every period assertions")
	flag.Int64Var(&flagGenesisTimeValue, "GenesisTime", 0, "override genesis UNIX time instead of using a random UNIX time")
}

// NewConfigFromFlags creates a simulation from the retrieved values of the flags
func NewConfigFromFlags() simulation.Config {
	return simulation.Config{
		GenesisFile:        flagGenesisFileValue,
		ParamsFile:         flagParamsFileValue,
		ExportParamsPath:   flagExportParamsPathValue,
		ExportParamsHeight: flagExportParamsHeightValue,
		ExportStatePath:    flagExportStatePathValue,
		ExportStatsPath:    flagExportStatsPathValue,
		Seed:               flagSeedValue,
		InitialBlockHeight: flagInitialBlockHeightValue,
		NumBlocks:          flagNumBlocksValue,
		BlockSize:          flagBlockSizeValue,
		Lean:               flagLeanValue,
		Commit:             flagCommitValue,
		OnOperation:        flagOnOperationValue,
		AllInvariants:      flagAllInvariantsValue,
	}
}

//---------------------------------------------------------------------
// Simulation Utils

// ExportStateToJSON util function to export the app state to JSON
func ExportStateToJSON(app *SimApp, path string) error {
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
