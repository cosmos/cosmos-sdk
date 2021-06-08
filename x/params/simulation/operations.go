package simulation

import (
	"fmt"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
)

// SimulateParamChangeProposalContent returns random parameter change content.
// It will generate a ParameterChangeProposal object with anywhere between 1 and
// the total amount of defined parameters changes, all of which have random valid values.
func SimulateParamChangeProposalContent(paramChangePool []simulation.ParamChange) simulation.ContentSimulatorFn {
	numProposals := 0
	return func(r *rand.Rand, _ sdk.Context, _ []simulation.Account) simulation.Content {

		lenParamChange := len(paramChangePool)
		if lenParamChange == 0 {
			panic("param changes array is empty")
		}

		numChanges := simulation.RandIntBetween(r, 1, lenParamChange)
		paramChanges := make([]proposal.ParamChange, numChanges)

		// map from key to empty struct; used only for look-up of the keys of the
		// parameters that are already in the random set of changes.
		paramChangesKeys := make(map[string]struct{})

		for i := 0; i < numChanges; i++ {
			spc := paramChangePool[r.Intn(len(paramChangePool))]
			composedkey := spc.ComposedKey()

			// do not include duplicate parameter changes for a given subspace/key
			_, ok := paramChangesKeys[composedkey]
			for ok {
				spc = paramChangePool[r.Intn(len(paramChangePool))]
				_, ok = paramChangesKeys[composedkey]
			}

			// add a new distinct parameter to the set of changes and register the key
			// to avoid further duplicates
			paramChangesKeys[composedkey] = struct{}{}
			paramChanges[i] = proposal.NewParamChange(spc.Subspace(), spc.Key(), spc.SimValue()(r))
		}

		title := fmt.Sprintf("title from SimulateParamChangeProposalContent-%d", numProposals)
		desc := fmt.Sprintf("desc from SimulateParamChangeProposalContent-%d. Random short desc: %s",
			numProposals, simulation.RandStringOfLength(r, 20))
		numProposals += 1
		return proposal.NewParameterChangeProposal(
			title,        // title
			desc,         // description
			paramChanges, // set of changes
		)
	}
}
