package proposal

import (
	"cosmossdk.io/core/registry"
	govtypes "cosmossdk.io/x/gov/types/v1beta1"

	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterLegacyAminoCodec registers all necessary param module types with a given LegacyAmino codec.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&ParameterChangeProposal{}, "cosmos-sdk/ParameterChangeProposal", nil)
}

func RegisterInterfaces(registrar registry.InterfaceRegistrar) {
	registrar.RegisterImplementations(
		(*govtypes.Content)(nil),
		&ParameterChangeProposal{},
	)
}
