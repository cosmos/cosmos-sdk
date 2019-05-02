package simulation

import (
	"fmt"
	"math/rand"
	"time"

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
// NOTE: All parameter value ranges are adapted from appStateRandomizedFn.
//
// TODO: governance parameters (blocked on an upgrade to go-amino)
var paramChangePool = []simParamChange{
	// staking parameters
	{
		"staking",
		"MaxValidators",
		"",
		func(r *rand.Rand) string {
			return fmt.Sprintf("%d", r.Intn(250)+1)
		},
	},
	{
		"staking",
		"UnbondingTime",
		"",
		func(r *rand.Rand) string {
			return fmt.Sprintf("\"%d\"", time.Duration(simulation.RandIntBetween(r, 60, 60*60*24*3*2))*time.Second)
		},
	},
	// slashing parameters
	{
		"slashing",
		"SignedBlocksWindow",
		"",
		func(r *rand.Rand) string {
			return fmt.Sprintf("\"%d\"", simulation.RandIntBetween(r, 10, 1000))
		},
	},
	{
		"slashing",
		"MinSignedPerWindow",
		"",
		func(r *rand.Rand) string {
			return fmt.Sprintf("\"%d\"", sdk.NewDecWithPrec(int64(r.Intn(10)), 1))
		},
	},
	{
		"slashing",
		"SlashFractionDowntime",
		"",
		func(r *rand.Rand) string {
			return fmt.Sprintf("\"%d\"", sdk.NewDec(1).Quo(sdk.NewDec(int64(r.Intn(200)+1))))
		},
	},
	// minting parameters
	{
		"mint",
		"InflationRateChange",
		"",
		func(r *rand.Rand) string {
			return fmt.Sprintf("\"%d\"", sdk.NewDecWithPrec(int64(r.Intn(99)), 2))
		},
	},
	// auth parameters
	{
		"auth",
		"MaxMemoCharacters",
		"",
		func(r *rand.Rand) string {
			return fmt.Sprintf("\"%d\"", simulation.RandIntBetween(r, 100, 200))
		},
	},
	{
		"auth",
		"TxSigLimit",
		"",
		func(r *rand.Rand) string {
			return fmt.Sprintf("\"%d\"", r.Intn(7)+1)
		},
	},
	{
		"auth",
		"TxSizeCostPerByte",
		"",
		func(r *rand.Rand) string {
			return fmt.Sprintf("\"%d\"", simulation.RandIntBetween(r, 5, 15))
		},
	},
}

// SimulateParamChangeProposalContent returns random parameter change content.
// It will generate a ParameterChangeProposal object with anywhere between 1 and
// 3 parameter changes all of which have random, but valid values.
func SimulateParamChangeProposalContent(r *rand.Rand) gov.Content {
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
		simulation.RandStringOfLength(r, 200),
		simulation.RandStringOfLength(r, 6000),
		paramChanges,
	)
}
