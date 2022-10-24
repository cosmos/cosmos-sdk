package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/pointnetwork/cosmos-point-sdk/x/simulation"

	simtypes "github.com/pointnetwork/cosmos-point-sdk/types/simulation"
	"github.com/pointnetwork/cosmos-point-sdk/x/bank/types"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, string(types.KeySendEnabled),
			func(r *rand.Rand) string {
				paramsBytes, err := json.Marshal(RandomGenesisSendParams(r))
				if err != nil {
					panic(err)
				}
				return string(paramsBytes)
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyDefaultSendEnabled),
			func(r *rand.Rand) string {
				return fmt.Sprintf("%v", RandomGenesisDefaultSendParam(r))
			},
		),
	}
}
