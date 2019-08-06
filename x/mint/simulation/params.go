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
func ParamChanges(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams) []simulation.SimParamChange {
	return []simulation.SimParamChange{
		simulation.NewSimParamChange("mint", "InflationRateChange", "",
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenInflationRateChange(cdc, r, ap))
			},
		),
	}
}
