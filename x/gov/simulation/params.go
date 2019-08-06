package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	simutil "github.com/cosmos/cosmos-sdk/x/simulation/util"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams) []simulation.SimParamChange {
	return []simulation.SimParamChange{
		simulation.NewSimParamChange("gov", "votingparams", "",
			func(r *rand.Rand) string {
				return fmt.Sprintf(`{"voting_period": "%d"}`, GenVotingParamsVotingPeriod(cdc, r, ap))
			},
		),
		simulation.NewSimParamChange("gov", "depositparams", "",
			func(r *rand.Rand) string {
				return fmt.Sprintf(`{"max_deposit_period": "%d"}`, GenDepositParamsDepositPeriod(cdc, r, ap))
			},
		),
		simulation.NewSimParamChange("gov", "tallyparams", "",
			func(r *rand.Rand) string {
				changes := []struct {
					key   string
					value sdk.Dec
				}{
					{"quorum", GenTallyParamsQuorum(cdc, r, ap)},
					{"threshold", GenTallyParamsThreshold(cdc, r, ap)},
					{"veto", GenTallyParamsVeto(cdc, r, ap)},
				}

				pc := make(map[string]string)
				numChanges := simutil.RandIntBetween(r, 1, len(changes))
				for i := 0; i < numChanges; i++ {
					c := changes[r.Intn(len(changes))]

					_, ok := pc[c.key]
					for ok {
						c := changes[r.Intn(len(changes))]
						_, ok = pc[c.key]
					}

					pc[c.key] = c.value.String()
				}

				bz, _ := json.Marshal(pc)
				return string(bz)
			},
		),
	}
}
