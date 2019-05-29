package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

type simParamChange struct {
	subspace string
	key      string
	subkey   string
	simValue func(r *rand.Rand) string
}

func (spc simParamChange) compKey() string {
	return fmt.Sprintf("%s/%s/%s", spc.subkey, spc.key, spc.subkey)
}

// paramChangePool defines a static slice of possible simulated parameter changes
// where each simParamChange corresponds to a ParamChange with a simValue
// function to generate a simulated new value.
var paramChangePool = []simParamChange{
	// staking parameters
	{
		"staking",
		"MaxValidators",
		"",
		func(r *rand.Rand) string {
			return fmt.Sprintf("%d", simulation.ModuleParamSimulator["max_validators"](r).(uint16))
		},
	},
	{
		"staking",
		"UnbondingTime",
		"",
		func(r *rand.Rand) string {
			return fmt.Sprintf("\"%d\"", simulation.ModuleParamSimulator["unbonding_time"](r).(time.Duration))
		},
	},
	// slashing parameters
	{
		"slashing",
		"SignedBlocksWindow",
		"",
		func(r *rand.Rand) string {
			return fmt.Sprintf("\"%d\"", simulation.ModuleParamSimulator["signed_blocks_window"](r).(int64))
		},
	},
	{
		"slashing",
		"MinSignedPerWindow",
		"",
		func(r *rand.Rand) string {
			return fmt.Sprintf("\"%s\"", simulation.ModuleParamSimulator["min_signed_per_window"](r).(sdk.Dec))
		},
	},
	{
		"slashing",
		"SlashFractionDowntime",
		"",
		func(r *rand.Rand) string {
			return fmt.Sprintf("\"%s\"", simulation.ModuleParamSimulator["slash_fraction_downtime"](r).(sdk.Dec))
		},
	},
	// minting parameters
	{
		"mint",
		"InflationRateChange",
		"",
		func(r *rand.Rand) string {
			return fmt.Sprintf("\"%s\"", simulation.ModuleParamSimulator["inflation_rate_change"](r).(sdk.Dec))
		},
	},
	// gov parameters
	{
		"gov",
		"votingparams",
		"",
		func(r *rand.Rand) string {
			return fmt.Sprintf(`{"voting_period": "%d"}`, simulation.ModuleParamSimulator["voting_params_voting_period"](r).(time.Duration))
		},
	},
	{
		"gov",
		"depositparams",
		"",
		func(r *rand.Rand) string {
			return fmt.Sprintf(`{"max_deposit_period": "%d"}`, simulation.ModuleParamSimulator["voting_params_voting_period"](r).(time.Duration))
		},
	},
	{
		"gov",
		"tallyparams",
		"",
		func(r *rand.Rand) string {
			changes := []struct {
				key   string
				value sdk.Dec
			}{
				{"quorum", simulation.ModuleParamSimulator["tally_params_quorum"](r).(sdk.Dec)},
				{"threshold", simulation.ModuleParamSimulator["tally_params_threshold"](r).(sdk.Dec)},
				{"veto", simulation.ModuleParamSimulator["tally_params_veto"](r).(sdk.Dec)},
			}

			pc := make(map[string]string)
			numChanges := simulation.RandIntBetween(r, 1, len(changes))
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
	},
	// auth parameters
	{
		"auth",
		"MaxMemoCharacters",
		"",
		func(r *rand.Rand) string {
			return fmt.Sprintf("\"%d\"", simulation.ModuleParamSimulator["max_memo_characters"](r).(uint64))
		},
	},
	{
		"auth",
		"TxSigLimit",
		"",
		func(r *rand.Rand) string {
			return fmt.Sprintf("\"%d\"", simulation.ModuleParamSimulator["tx_sig_limit"](r).(uint64))
		},
	},
	{
		"auth",
		"TxSizeCostPerByte",
		"",
		func(r *rand.Rand) string {
			return fmt.Sprintf("\"%d\"", simulation.ModuleParamSimulator["tx_size_cost_per_byte"](r).(uint64))
		},
	},
}

// SimulateParamChangeProposalContent returns random parameter change content.
// It will generate a ParameterChangeProposal object with anywhere between 1 and
// 3 parameter changes all of which have random, but valid values.
func SimulateParamChangeProposalContent(r *rand.Rand, _ *baseapp.BaseApp, _ sdk.Context, _ []simulation.Account) gov.Content {
	numChanges := simulation.RandIntBetween(r, 1, len(paramChangePool)/2)
	paramChanges := make([]params.ParamChange, numChanges, numChanges)
	paramChangesKeys := make(map[string]struct{})

	for i := 0; i < numChanges; i++ {
		spc := paramChangePool[r.Intn(len(paramChangePool))]

		// do not include duplicate parameter changes for a given subspace/key
		_, ok := paramChangesKeys[spc.compKey()]
		for ok {
			spc = paramChangePool[r.Intn(len(paramChangePool))]
			_, ok = paramChangesKeys[spc.compKey()]
		}

		paramChangesKeys[spc.compKey()] = struct{}{}
		paramChanges[i] = params.NewParamChangeWithSubkey(spc.subspace, spc.key, spc.subkey, spc.simValue(r))
	}

	return params.NewParameterChangeProposal(
		simulation.RandStringOfLength(r, 140),
		simulation.RandStringOfLength(r, 5000),
		paramChanges,
	)
}
