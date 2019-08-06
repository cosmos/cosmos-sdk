package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// SimulateParamChangeProposalContent returns random parameter change content.
// It will generate a ParameterChangeProposal object with anywhere between 1 and
// 3 parameter changes all of which have random, but valid values.
func SimulateParamChangeProposalContent(r *rand.Rand, paramChangePool []simulation.SimParamChange) gov.Content {
	numChanges := simulation.RandIntBetween(r, 1, len(paramChangePool))
	paramChanges := make([]params.ParamChange, numChanges, numChanges)
	paramChangesKeys := make(map[string]struct{})

	for i := 0; i < numChanges; i++ {
		spc := paramChangePool[r.Intn(len(paramChangePool))]

		// do not include duplicate parameter changes for a given subspace/key
		_, ok := paramChangesKeys[spc.ComposedKey()]
		for ok {
			spc = paramChangePool[r.Intn(len(paramChangePool))]
			_, ok = paramChangesKeys[spc.ComposedKey()]
		}

		paramChangesKeys[spc.ComposedKey()] = struct{}{}
		paramChanges[i] = params.NewParamChangeWithSubkey(spc.Subspace, spc.Key, spc.Subkey, spc.SimValue(r))
	}

	return params.NewParameterChangeProposal(
		simulation.RandStringOfLength(r, 140),  // title
		simulation.RandStringOfLength(r, 5000), // description
		paramChanges,
	)
}
