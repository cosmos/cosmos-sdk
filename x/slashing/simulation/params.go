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
func ParamChanges(cdc *codec.Codec, r *rand.Rand) []simulation.SimParamChange {
	return []simulation.SimParamChange{
		simulation.NewSimParamChange("slashing", "SignedBlocksWindow", "",
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%d\"", GenSignedBlocksWindow(cdc, r))
			},
		),
		simulation.NewSimParamChange("slashing", "MinSignedPerWindow", "",
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenMinSignedPerWindow(cdc, r))
			},
		),
		simulation.NewSimParamChange("slashing", "SlashFractionDowntime", "",
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenSlashFractionDowntime(cdc, r))
			},
		),
	}
}
