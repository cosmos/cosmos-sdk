package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(cdc *codec.Codec, r *rand.Rand) []simulation.ParamChange {
	return []simulation.ParamChange{
		simulation.NewSimParamChange("mint", "InflationRateChange", "",
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenInflationRateChange(cdc, r))
			},
		),
		simulation.NewSimParamChange("mint", "Inflation", "",
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenInflation(cdc, r))
			},
		),
		simulation.NewSimParamChange("mint", "InflationMax", "",
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenInflationMax(cdc, r))
			},
		),
		simulation.NewSimParamChange("mint", "InflationMin", "",
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenInflationMin(cdc, r))
			},
		),
		simulation.NewSimParamChange("mint", "GoalBonded", "",
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenGoalBonded(cdc, r))
			},
		),
	}
}
