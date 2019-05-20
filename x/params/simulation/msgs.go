package simulation

import (
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
//
// TODO: governance parameters (blocked on an upgrade to go-amino)
var paramChangePool = []simParamChange{
	// staking parameters
	{
		"staking",
		"MaxValidators",
		"",
		func(r *rand.Rand) string {
			return fmt.Sprintf("%d", simulation.ModuleParamSimulator["MaxValidators"](r).(uint16))
		},
	},
	{
		"staking",
		"UnbondingTime",
		"",
		func(r *rand.Rand) string {
			return fmt.Sprintf("\"%d\"", simulation.ModuleParamSimulator["UnbondingTime"](r).(time.Duration))
		},
	},
	// slashing parameters
	{
		"slashing",
		"SignedBlocksWindow",
		"",
		func(r *rand.Rand) string {
			return fmt.Sprintf("\"%d\"", simulation.ModuleParamSimulator["SignedBlocksWindow"](r).(int64))
		},
	},
	{
		"slashing",
		"MinSignedPerWindow",
		"",
		func(r *rand.Rand) string {
			return fmt.Sprintf("\"%s\"", simulation.ModuleParamSimulator["MinSignedPerWindow"](r).(sdk.Dec))
		},
	},
	{
		"slashing",
		"SlashFractionDowntime",
		"",
		func(r *rand.Rand) string {
			return fmt.Sprintf("\"%s\"", simulation.ModuleParamSimulator["SlashFractionDowntime"](r).(sdk.Dec))
		},
	},
	// minting parameters
	{
		"mint",
		"InflationRateChange",
		"",
		func(r *rand.Rand) string {
			return fmt.Sprintf("\"%s\"", simulation.ModuleParamSimulator["InflationRateChange"](r).(sdk.Dec))
		},
	},
	// auth parameters
	{
		"auth",
		"MaxMemoCharacters",
		"",
		func(r *rand.Rand) string {
			return fmt.Sprintf("\"%d\"", simulation.ModuleParamSimulator["MaxMemoCharacters"](r).(uint64))
		},
	},
	{
		"auth",
		"TxSigLimit",
		"",
		func(r *rand.Rand) string {
			return fmt.Sprintf("\"%d\"", simulation.ModuleParamSimulator["TxSigLimit"](r).(uint64))
		},
	},
	{
		"auth",
		"TxSizeCostPerByte",
		"",
		func(r *rand.Rand) string {
			return fmt.Sprintf("\"%d\"", simulation.ModuleParamSimulator["TxSizeCostPerByte"](r).(uint64))
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
		paramChanges[i] = params.NewParamChange(spc.subspace, spc.key, spc.subkey, spc.simValue(r))
	}

	return params.NewParameterChangeProposal(
		simulation.RandStringOfLength(r, 140),
		simulation.RandStringOfLength(r, 5000),
		paramChanges,
	)
}
