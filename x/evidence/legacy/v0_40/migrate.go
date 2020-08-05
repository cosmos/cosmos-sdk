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
	// We need to register the Evidence interface.
	clientCtx.InterfaceRegistry.RegisterInterface(
		"cosmos_sdk.evidence.v1.Evidence",
		(*v038evidence.Evidence)(nil),
		&v038evidence.Equivocation{},
	)

	return NewGenesisState(evidenceState.Evidence)
}
