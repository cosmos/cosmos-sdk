package simulation

// Config contains the necessary configuration flags for the simulator
type Config struct {
	GenesisFile string // custom simulation genesis file; cannot be used with params file
	ParamsFile  string // custom simulation params file which overrides any random params; cannot be used with genesis

	ExportParamsPath   string // custom file path to save the exported params JSON
	ExportParamsHeight int    //height to which export the randomly generated params
	ExportStatePath    string //custom file path to save the exported app state JSON
	ExportStatsPath    string // custom file path to save the exported simulation statistics JSON

	Seed               int64 // simulation random seed
	InitialBlockHeight int   // initial block to start the simulation
	NumBlocks          int   // number of new blocks to simulate from the initial block height
	BlockSize          int   // operations per block

	Lean   bool // lean simulation log output
	Commit bool // have the simulation commit

	OnOperation   bool // run slow invariants every operation
	AllInvariants bool // print all failed invariants if a broken invariant is found
}

// NewConfig creates a new simulation config instance
func NewConfig(genesisFile, paramsFile, exportParamsPath string, exportParamsHeight int,
	exportStatePath, exportStatsPath string, seed int64,
	initialBlockHeight, numBlocks, blockSize int,
	lean, commit, onOperation,
	allInvariants bool) Config {

	return Config{
		GenesisFile:        genesisFile,
		ParamsFile:         paramsFile,
		ExportParamsPath:   exportParamsPath,
		ExportParamsHeight: exportParamsHeight,
		ExportStatePath:    exportStatePath,
		ExportStatsPath:    exportStatsPath,
		Seed:               seed,
		InitialBlockHeight: initialBlockHeight,
		NumBlocks:          numBlocks,
		BlockSize:          blockSize,
		Lean:               lean,
		Commit:             commit,
		OnOperation:        onOperation,
		AllInvariants:      allInvariants,
	}
}
