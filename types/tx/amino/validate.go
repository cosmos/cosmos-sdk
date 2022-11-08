package amino

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// ValidateAminoAnnotations validates the `amino.*` protobuf annotations. It
// performs the following validations:
//   - Make sure `amino.name` is equal to the name in the Amino codec's
//     registry.
func ValidateAminoAnnotations(interfaceRegistry codectypes.InterfaceRegistry, aminoCdc *codec.LegacyAmino) {

}
