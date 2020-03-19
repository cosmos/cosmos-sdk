package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	keyMaxValidators = "MaxValidators"
	keyUnbondingTime = "UnbondingTime"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(r *rand.Rand) []module.ParamChange {
	return []module.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, keyMaxValidators,
			func(r *rand.Rand) string {
				return fmt.Sprintf("%d", GenMaxValidators(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, keyUnbondingTime,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%d\"", GenUnbondingTime(r))
			},
		),
	}
}
