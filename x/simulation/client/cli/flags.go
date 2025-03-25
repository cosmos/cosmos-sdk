package cli

import (
	"flag"
	"time"

	"github.com/cosmos/cosmos-sdk/types/simulation"
)

const DefaultSeedValue = 42

// List of available flags for the simulator
var (
	FlagGenesisFileValue        string
	FlagParamsFileValue         string
	FlagExportParamsPathValue   string
	FlagExportParamsHeightValue int
	FlagExportStatePathValue    string
	FlagExportStatsPathValue    string
	FlagSeedValue               int64
	FlagInitialBlockHeightValue int
	FlagNumBlocksValue          int
	FlagBlockSizeValue          int
	FlagLeanValue               bool
	FlagCommitValue             bool
	FlagDBBackendValue          string

	FlagVerboseValue     bool
	FlagGenesisTimeValue int64
	FlagSigverifyTxValue bool
	FlagFauxMerkle       bool

	// Deprecated: This flag is unused and will be removed in a future release.
	FlagPeriodValue uint
	// Deprecated: This flag is unused and will be removed in a future release.
	FlagEnabledValue bool
	// Deprecated: This flag is unused and will be removed in a future release.
	FlagOnOperationValue bool
	// Deprecated: This flag is unused and will be removed in a future release.
	FlagAllInvariantsValue bool
)

// GetSimulatorFlags gets the values of all the available simulation flags
func GetSimulatorFlags() {
	// config fields
	flag.StringVar(&FlagGenesisFileValue, "Genesis", "", "custom simulation genesis file; cannot be used with params file")
	flag.StringVar(&FlagParamsFileValue, "Params", "", "custom simulation params file which overrides any random params; cannot be used with genesis")
	flag.StringVar(&FlagExportParamsPathValue, "ExportParamsPath", "", "custom file path to save the exported params JSON")
	flag.IntVar(&FlagExportParamsHeightValue, "ExportParamsHeight", 0, "height to which export the randomly generated params")
	flag.StringVar(&FlagExportStatePathValue, "ExportStatePath", "", "custom file path to save the exported app state JSON")
	flag.StringVar(&FlagExportStatsPathValue, "ExportStatsPath", "", "custom file path to save the exported simulation statistics JSON")
	flag.Int64Var(&FlagSeedValue, "Seed", DefaultSeedValue, "simulation random seed")
	flag.IntVar(&FlagInitialBlockHeightValue, "InitialBlockHeight", 1, "initial block to start the simulation")
	flag.IntVar(&FlagNumBlocksValue, "NumBlocks", 500, "number of new blocks to simulate from the initial block height")
	flag.IntVar(&FlagBlockSizeValue, "BlockSize", 200, "operations per block")
	flag.BoolVar(&FlagLeanValue, "Lean", false, "lean simulation log output")
	flag.BoolVar(&FlagCommitValue, "Commit", true, "have the simulation commit")
	flag.StringVar(&FlagDBBackendValue, "DBBackend", "memdb", "custom db backend type: goleveldb, pebbledb, memdb")

	// simulation flags
	flag.BoolVar(&FlagVerboseValue, "Verbose", false, "verbose log output")
	flag.Int64Var(&FlagGenesisTimeValue, "GenesisTime", time.Now().Unix(), "use current time as genesis UNIX time for default")
	flag.BoolVar(&FlagSigverifyTxValue, "SigverifyTx", true, "whether to sigverify check for transaction ")
	flag.BoolVar(&FlagFauxMerkle, "FauxMerkle", false, "use faux merkle instead of iavl")

	flag.UintVar(&FlagPeriodValue, "Period", 0, "This parameter is unused and will be removed")
	flag.BoolVar(&FlagEnabledValue, "Enabled", false, "This parameter is unused and will be removed")
	flag.BoolVar(&FlagOnOperationValue, "SimulateEveryOperation", false, "This parameter is unused and will be removed")
	flag.BoolVar(&FlagAllInvariantsValue, "PrintAllInvariants", false, "This parameter is unused and will be removed")
}

// NewConfigFromFlags creates a simulation from the retrieved values of the flags.
func NewConfigFromFlags() simulation.Config {
	return simulation.Config{
		GenesisFile:        FlagGenesisFileValue,
		ParamsFile:         FlagParamsFileValue,
		ExportParamsPath:   FlagExportParamsPathValue,
		ExportParamsHeight: FlagExportParamsHeightValue,
		ExportStatePath:    FlagExportStatePathValue,
		ExportStatsPath:    FlagExportStatsPathValue,
		Seed:               FlagSeedValue,
		InitialBlockHeight: FlagInitialBlockHeightValue,
		GenesisTime:        FlagGenesisTimeValue,
		NumBlocks:          FlagNumBlocksValue,
		BlockSize:          FlagBlockSizeValue,
		Lean:               FlagLeanValue,
		Commit:             FlagCommitValue,
		DBBackend:          FlagDBBackendValue,
	}
}

// GetDeprecatedFlagUsed return list of deprecated flag names that are being used.
// This function is for internal usage only and may be removed with the deprecated fields.
func GetDeprecatedFlagUsed() []string {
	var usedFlags []string
	for _, flagName := range []string{
		"Enabled",
		"SimulateEveryOperation",
		"PrintAllInvariants",
		"Period",
	} {
		if flag.Lookup(flagName) != nil {
			usedFlags = append(usedFlags, flagName)
		}
	}
	return usedFlags
}
