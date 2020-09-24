package v040

import (
	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	v038evidence "github.com/cosmos/cosmos-sdk/x/evidence/legacy/v0_38"
)

// Migrate accepts exported v0.38 x/evidence genesis state and migrates it to
// v0.40 x/evidence genesis state. The migration includes:
//
// - Removing the `Params` field.
func Migrate(evidenceState v038evidence.GenesisState, _ client.Context) *GenesisState {
	var newEquivocations = make([]Equivocation, 0, len(evidenceState.Evidence))
	for i, evidence := range evidenceState.Evidence {
		equivocation, ok := evidence.(v038evidence.Equivocation)
		if !ok {
			// There's only equivocation in 0.38.
			continue
		}

		newEquivocations[i] = Equivocation{
			Height:           equivocation.Height,
			Time:             equivocation.Time,
			Power:            equivocation.Power,
			ConsensusAddress: equivocation.ConsensusAddress,
		}
	}

	// Then convert the equivocations into Any.
	newEvidence := make([]*codectypes.Any, 0, len(newEquivocations))
	for i, equi := range newEquivocations {
		newEvidence[i] = codectypes.UnsafePackAny(&equi)
	}

	return &GenesisState{
		Evidence: newEvidence,
	}
}
