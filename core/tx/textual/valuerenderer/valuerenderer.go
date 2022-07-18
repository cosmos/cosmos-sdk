package valuerenderer

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	cosmos_proto "github.com/cosmos/cosmos-proto"
)

func GetADR050ValueRenderer(fd protoreflect.FieldDescriptor) (ValueRenderer, error) {
	switch {
	// Integers
	case fd.Kind() == protoreflect.Uint32Kind ||
		fd.Kind() == protoreflect.Uint64Kind ||
		fd.Kind() == protoreflect.Int32Kind ||
		fd.Kind() == protoreflect.Int64Kind ||
		(fd.Kind() == protoreflect.StringKind && isCosmosScalar(fd, "cosmos.Int")):
		{
			return intValueRenderer{}, nil
		}
	// Decimals
	case fd.Kind() == protoreflect.StringKind && isCosmosScalar(fd, "cosmos.Dec"):
		{
			return decValueRenderer{}, nil
		}

	default:
		return nil, fmt.Errorf("value renderers cannot format value of type %s", fd.Kind())
	}
}

// isCosmosScalar returns true if a field has the `cosmos_proto.scalar` field
// option.
func isCosmosScalar(fd protoreflect.FieldDescriptor, scalar string) bool {
	opts := fd.Options().(*descriptorpb.FieldOptions)
	if proto.GetExtension(opts, cosmos_proto.E_Scalar).(string) == scalar {
		return true
	}

	return false
}
