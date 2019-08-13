package operations

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govsimops "github.com/cosmos/cosmos-sdk/x/gov/simulation/operations"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// SimulateParamChangeProposalContent returns random parameter change content.
// It will generate a ParameterChangeProposal object with anywhere between 1 and
// 3 parameter changes all of which have random, but valid values.
func SimulateParamChangeProposalContent(paramChangePool []simulation.ParamChange) govsimops.ContentSimulator {
	return func(r *rand.Rand, _ sdk.Context, _ []simulation.Account) govtypes.Content {

		// TODO: add comment of why is the number of changes capped
		numChanges := simulation.RandIntBetween(r, 1, len(paramChangePool)/2)
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
}
