package v040

import (
	"github.com/cosmos/cosmos-sdk/client"
	v038evidence "github.com/cosmos/cosmos-sdk/x/evidence/legacy/v0_38"
)

// Migrate accepts exported v0.38 x/evidence genesis state and migrates it to
// v0.40 x/evidence genesis state. The migration includes:
//
// - Removing the `Params` field.
func Migrate(evidenceState v038evidence.GenesisState, clientCtx client.Context) GenesisState {
	return NewGenesisState(evidenceState.Evidence)
}
