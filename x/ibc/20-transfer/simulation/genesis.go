package simulation

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
)

// Simulation parameter constants
const port = "port_id"

// RandomizedGenState generates a random GenesisState for transfer.
func RandomizedGenState(simState *module.SimulationState) {
	var portID string

	simState.AppParams.GetOrGenerate(
		simState.Cdc, port, &portID, simState.Rand,
		func(r *rand.Rand) { portID = simtypes.RandStringOfLength(simState.Rand, 20) },
	)

	transferGenesis := types.GenesisState{
		PortID: portID,
	}

	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, codec.MustMarshalJSONIndent(simState.Cdc, transferGenesis))
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(transferGenesis)
}
