package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	"github.com/KiraCore/cosmos-sdk/x/simulation"

	simtypes "github.com/KiraCore/cosmos-sdk/types/simulation"
	"github.com/KiraCore/cosmos-sdk/x/bank/types"
)

const keySendEnabled = "sendenabled"

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, keySendEnabled,
			func(r *rand.Rand) string {
				return fmt.Sprintf("%v", GenSendEnabled(r))
			},
		),
	}
}
