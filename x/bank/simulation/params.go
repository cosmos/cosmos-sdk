package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/x/bank/internal/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

const keySendEnabled = "sendenabled"

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(r *rand.Rand) []simulation.ParamChange {
	return []simulation.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, keySendEnabled,
			func(r *rand.Rand) string {
				return fmt.Sprintf("%v", GenSendEnabled(r))
			},
		),
	}
}
