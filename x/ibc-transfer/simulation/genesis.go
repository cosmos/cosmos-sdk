package simulation

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/KiraCore/cosmos-sdk/codec"
	"github.com/KiraCore/cosmos-sdk/types/module"
	simtypes "github.com/KiraCore/cosmos-sdk/types/simulation"
	"github.com/KiraCore/cosmos-sdk/x/ibc-transfer/types"
)

// Simulation parameter constants
const port = "port_id"

// RandomizedGenState generates a random GenesisState for transfer.
func RandomizedGenState(simState *module.SimulationState) {
	var portID string
	simState.AppParams.GetOrGenerate(
		simState.Cdc, port, &portID, simState.Rand,
		func(r *rand.Rand) { portID = strings.ToLower(simtypes.RandStringOfLength(r, 20)) },
	)

	transferGenesis := types.GenesisState{
		PortID: portID,
	}

	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, codec.MustMarshalJSONIndent(simState.Cdc, transferGenesis))
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(transferGenesis)
}
