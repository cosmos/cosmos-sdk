package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

const (
	keyVotingParams  = "votingparams"
	keyDepositParams = "depositparams"
	keyTallyParams   = "tallyparams"
	subkeyQuorum     = "quorum"
	subkeyThreshold  = "threshold"
	subkeyVeto       = "veto"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, keyVotingParams,
			func(r *rand.Rand) string {
				return fmt.Sprintf(`{"voting_period": "%d"}`, GenVotingParamsVotingPeriod(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, keyDepositParams,
			func(r *rand.Rand) string {
				return fmt.Sprintf(`{"max_deposit_period": "%d"}`, GenDepositParamsDepositPeriod(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, keyTallyParams,
			func(r *rand.Rand) string {
				changes := []struct {
					key   string
					value sdk.Dec
				}{
					{subkeyQuorum, GenTallyParamsQuorum(r)},
					{subkeyThreshold, GenTallyParamsThreshold(r)},
					{subkeyVeto, GenTallyParamsVeto(r)},
				}

				pc := make(map[string]string)
				numChanges := simtypes.RandIntBetween(r, 1, len(changes))
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
