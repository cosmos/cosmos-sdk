package v040

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	v038evidence "github.com/cosmos/cosmos-sdk/x/evidence/legacy/v0_38"
)

// RegisterCodec registers all the necessary types and interfaces for the
// evidence module.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*v038evidence.Evidence)(nil), nil)
	cdc.RegisterConcrete(&v038evidence.Equivocation{}, "cosmos-sdk/Equivocation", nil)
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterInterface(
		"cosmos_sdk.evidence.v1.Evidence",
		(*v038evidence.Evidence)(nil),
		&v038evidence.Equivocation{},
	)
}

// Migrate accepts exported v0.38 x/evidence genesis state and migrates it to
// v0.40 x/evidence genesis state. The migration includes:
//
// - Removing the `Params` field.
func Migrate(evidenceState v038evidence.GenesisState, clientCtx client.Context) GenesisState {
	RegisterCodec(clientCtx.Codec)
	RegisterInterfaces(clientCtx.InterfaceRegistry)

	return NewGenesisState(evidenceState.Evidence)
}
