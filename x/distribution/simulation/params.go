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
		simulation.NewSimParamChange("distribution", "communitytax", "",
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenCommunityTax(r))
			},
		),
		simulation.NewSimParamChange("distribution", "baseproposerreward", "",
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenBaseProposerReward(r))
			},
		),
		simulation.NewSimParamChange("distribution", "bonusproposerreward", "",
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenBonusProposerReward(r))
			},
		),
	}
}
