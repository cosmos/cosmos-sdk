package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/protocolpool/types"
)

const (
	Params = "params"
)

func GenParams(r *rand.Rand) types.Params {
	params := types.DefaultParams()

	windowLen := r.Intn(20) + 1
	params.DistributionFrequency = uint64(windowLen)

	return params
}

// RandomizedGenState generates a random GenesisState for gov
func RandomizedGenState(simState *module.SimulationState) {
	params := types.DefaultParams()
	simState.AppParams.GetOrGenerate(Params, &params, simState.Rand, func(r *rand.Rand) { params = GenParams(r) })

	genesis := types.DefaultGenesisState()
	genesis.Params = params

	bz, err := json.MarshalIndent(genesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated protocolpool parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(genesis)
}
