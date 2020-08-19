package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"

	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/ibc-transfer/types"
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
		PortId: portID,
	}

	bz, err := json.MarshalIndent(&transferGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&transferGenesis)
}
