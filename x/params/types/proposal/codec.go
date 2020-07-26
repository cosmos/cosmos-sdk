package proposal

import (
	"github.com/KiraCore/cosmos-sdk/codec"
	"github.com/KiraCore/cosmos-sdk/codec/types"
	govtypes "github.com/KiraCore/cosmos-sdk/x/gov/types"
)

type Codec struct {
	codec.Marshaler

	// Keep reference to the amino codec to allow backwards compatibility along
	// with type, and interface registration.
	amino *codec.Codec
}

func NewCodec(amino *codec.Codec) *Codec {
	return &Codec{Marshaler: codec.NewHybridCodec(amino, types.NewInterfaceRegistry()), amino: amino}
}

// ModuleCdc is the module codec.
var ModuleCdc *Codec

func init() {
	ModuleCdc = NewCodec(codec.New())

	RegisterCodec(ModuleCdc.amino)
	ModuleCdc.amino.Seal()
}

// RegisterCodec registers all necessary param module types with a given codec.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(&ParameterChangeProposal{}, "cosmos-sdk/ParameterChangeProposal", nil)
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*govtypes.Content)(nil),
		&ParameterChangeProposal{},
	)
}
