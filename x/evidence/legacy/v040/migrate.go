package v040

import (
	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	v038evidence "github.com/cosmos/cosmos-sdk/x/evidence/legacy/v038"
	v040evidence "github.com/cosmos/cosmos-sdk/x/evidence/types"
)

// Migrate accepts exported v0.38 x/evidence genesis state and migrates it to
// v0.40 x/evidence genesis state. The migration includes:
//
// - Removing the `Params` field.
// - Converting Equivocations into Anys.
func Migrate(evidenceState v038evidence.GenesisState, _ client.Context) *v040evidence.GenesisState {
	var newEquivocations = make([]v040evidence.Equivocation, len(evidenceState.Evidence))
	for i, evidence := range evidenceState.Evidence {
		equivocation, ok := evidence.(v038evidence.Equivocation)
		if !ok {
			// There's only equivocation in 0.38.
			continue
		}

		newEquivocations[i] = v040evidence.Equivocation{
			Height:           equivocation.Height,
			Time:             equivocation.Time,
			Power:            equivocation.Power,
			ConsensusAddress: equivocation.ConsensusAddress.String(),
		}
	}

	// Then convert the equivocations into Any.
	newEvidence := make([]*codectypes.Any, len(newEquivocations))
	for i := range newEquivocations {
		any, err := codectypes.NewAnyWithValue(&newEquivocations[i])
		if err != nil {
			panic(err)
		}

		newEvidence[i] = any
	}

	return &v040evidence.GenesisState{
		Evidence: newEvidence,
	}
}
