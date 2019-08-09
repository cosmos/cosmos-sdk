//nolint
package simapp

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

//---------------------------------------------------------------------
// Flags

// List of available flags for the simulator
var (
	genesisFile        string
	paramsFile         string
	exportParamsPath   string
	exportParamsHeight int
	exportStatePath    string
	exportStatsPath    string
	seed               int64
	initialBlockHeight int
	numBlocks          int
	blockSize          int
	enabled            bool
	verbose            bool
	lean               bool
	commit             bool
	period             int
	onOperation        bool // TODO Remove in favor of binary search for invariant violation
	allInvariants      bool
	genesisTime        int64
)

// getSimulatorFlags gets the values of all the available simulation flags
func getSimulatorFlags() {
	flag.StringVar(&genesisFile, "Genesis", "", "custom simulation genesis file; cannot be used with params file")
	flag.StringVar(&paramsFile, "Params", "", "custom simulation params file which overrides any random params; cannot be used with genesis")
	flag.StringVar(&exportParamsPath, "ExportParamsPath", "", "custom file path to save the exported params JSON")
	flag.IntVar(&exportParamsHeight, "ExportParamsHeight", 0, "height to which export the randomly generated params")
	flag.StringVar(&exportStatePath, "ExportStatePath", "", "custom file path to save the exported app state JSON")
	flag.StringVar(&exportStatsPath, "ExportStatsPath", "", "custom file path to save the exported simulation statistics JSON")
	flag.Int64Var(&seed, "Seed", 42, "simulation random seed")
	flag.IntVar(&initialBlockHeight, "InitialBlockHeight", 1, "initial block to start the simulation")
	flag.IntVar(&numBlocks, "NumBlocks", 500, "number of new blocks to simulate from the initial block height")
	flag.IntVar(&blockSize, "BlockSize", 200, "operations per block")
	flag.BoolVar(&enabled, "Enabled", false, "enable the simulation")
	flag.BoolVar(&verbose, "Verbose", false, "verbose log output")
	flag.BoolVar(&lean, "Lean", false, "lean simulation log output")
	flag.BoolVar(&commit, "Commit", false, "have the simulation commit")
	flag.IntVar(&period, "Period", 1, "run slow invariants only once every period assertions")
	flag.BoolVar(&onOperation, "SimulateEveryOperation", false, "run slow invariants every operation")
	flag.BoolVar(&allInvariants, "PrintAllInvariants", false, "print all invariants if a broken invariant is found")
	flag.Int64Var(&genesisTime, "GenesisTime", 0, "override genesis UNIX time instead of using a random UNIX time")
}

//---------------------------------------------------------------------
// Simulation Utils

// GetSimulationLog unmarshals the KVPair's Value to the corresponding type based on the
// each's module store key and the prefix bytes of the KVPair's key.
func GetSimulationLog(storeName string, sdr sdk.StoreDecoderRegistry, cdc *codec.Codec, kvs []cmn.KVPair) (log string) {
	if len(kvs)%2 != 0 {
		panic("KVPairs are not multiple of 2. There should be one for each app store")
	}

	var kvA, kvB cmn.KVPair
	for i := 0; i < len(kvs); i += 2 {
		kvA = kvs[i]
		kvB = kvs[i+1]

		if len(kvA.Value) == 0 && len(kvB.Value) == 0 {
			// skip if the value doesn't have any bytes
			continue
		}

		decoder, ok := sdr[storeName]
		if ok {
			log += decoder(cdc, kvA, kvB)
		} else {
			log += fmt.Sprintf("store A %X => %X\nstore B %X => %X\n", kvA.Key, kvA.Value, kvB.Key, kvB.Value)
		}
	}

	return
}
