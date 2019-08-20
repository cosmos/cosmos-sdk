package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(r *rand.Rand) []simulation.ParamChange {
	return []simulation.ParamChange{
		simulation.NewSimParamChange("staking", "MaxValidators", "",
			func(r *rand.Rand) string {
				return fmt.Sprintf("%d", GenMaxValidators(r))
			},
		),
		simulation.NewSimParamChange("staking", "UnbondingTime", "",
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%d\"", GenUnbondingTime(r))
			},
		),
	}
}
