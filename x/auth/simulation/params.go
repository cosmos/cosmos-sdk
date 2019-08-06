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
		simulation.NewSimParamChange("auth", "MaxMemoCharacters", "",
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%d\"", GenMaxMemoChars(cdc, r, ap))
			},
		),
		simulation.NewSimParamChange("auth", "TxSigLimit", "",
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%d\"", GenTxSigLimit(cdc, r, ap))
			},
		),
		simulation.NewSimParamChange("auth", "TxSizeCostPerByte", "",
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%d\"", GenTxSizeCostPerByte(cdc, r, ap))
			},
		),
	}
}
