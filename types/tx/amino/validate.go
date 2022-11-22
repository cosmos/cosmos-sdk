package amino

import (
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// ValidateAminoAnnotations validates the `amino.*` protobuf annotations. It
// performs the following validations:
//   - Make sure `amino.name` is equal to the name in the Amino codec's registry.
//
// If `fileResolver` is nil, then protoregistry.GlobalFile will be used.
func ValidateAminoAnnotations(fileResolver protodesc.Resolver, interfaceRegistry codectypes.InterfaceRegistry, aminoCdc *codec.LegacyAmino) {
	if fileResolver == nil {
		fileResolver = protoregistry.GlobalFiles
	}
}
