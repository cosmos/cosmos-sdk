package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// GenGovGenesisState generates a random GenesisState for gov
func GenGovGenesisState(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams, genesisState map[string]json.RawMessage) {
	var vp time.Duration
	ap.GetOrGenerate(cdc, simulation.VotingParamsVotingPeriod, &vp, r,
		func(r *rand.Rand) {
			vp = simulation.ModuleParamSimulator[simulation.VotingParamsVotingPeriod](r).(time.Duration)
		})

	govGenesis := gov.NewGenesisState(
		uint64(r.Intn(100)),
		gov.NewDepositParams(
			func(r *rand.Rand) sdk.Coins {
				var v sdk.Coins
				ap.GetOrGenerate(cdc, simulation.DepositParamsMinDeposit, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.DepositParamsMinDeposit](r).(sdk.Coins)
					})
				return v
			}(r),
			vp,
		),
		gov.NewVotingParams(vp),
		gov.NewTallyParams(
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(cdc, simulation.TallyParamsQuorum, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.TallyParamsQuorum](r).(sdk.Dec)
					})
				return v
			}(r),
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(cdc, simulation.TallyParamsThreshold, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.TallyParamsThreshold](r).(sdk.Dec)
					})
				return v
			}(r),
			func(r *rand.Rand) sdk.Dec {
				var v sdk.Dec
				ap.GetOrGenerate(cdc, simulation.TallyParamsVeto, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.TallyParamsVeto](r).(sdk.Dec)
					})
				return v
			}(r),
		),
	)

	fmt.Printf("Selected randomly generated governance parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, govGenesis))
	genesisState[gov.ModuleName] = cdc.MustMarshalJSON(govGenesis)
}
