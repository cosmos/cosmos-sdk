package v040

import (
	v038evidence "github.com/cosmos/cosmos-sdk/x/evidence/legacy/v0_38"
	"github.com/cosmos/cosmos-sdk/x/evidence/types"
)

// Migrate accepts exported v0.38 x/evidence genesis state and migrates it to
// v0.40 x/evidence genesis state. The migration includes:
//
// - Removing the `Params` field.
func Migrate(oldState v038evidence.GenesisState) types.GenesisState {
	return types.NewGenesisState(oldState.Evidence)
}
