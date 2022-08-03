package valuerenderer

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	cosmos_proto "github.com/cosmos/cosmos-proto"
)

type Textual struct {
	scalars map[string]ValueRenderer
}

func NewTextual() Textual {
	return Textual{}
}

// GetValueRenderer returns the value renderer for the given FieldDescriptor.
func (r Textual) GetValueRenderer(fd protoreflect.FieldDescriptor) (ValueRenderer, error) {
	switch {
	// Scalars, such as sdk.Int and sdk.Dec.
	case fd.Kind() == protoreflect.StringKind && proto.GetExtension(fd.Options(), cosmos_proto.E_Scalar) != "":
		{
			scalar, ok := proto.GetExtension(fd.Options(), cosmos_proto.E_Scalar).(string)
			if !ok || scalar == "" {
				return nil, fmt.Errorf("got extension option %s of type %T", scalar, scalar)
			}

			if r.scalars == nil {
				r.init()
			}

			vr := r.scalars[scalar]
			if vr == nil {
				return nil, fmt.Errorf("got empty value renderer for scalar %s", scalar)
			}

			return vr, nil
		}
	case fd.Kind() == protoreflect.BytesKind:
		return bytesValueRenderer{}, nil

	// Integers
	case fd.Kind() == protoreflect.Uint32Kind ||
		fd.Kind() == protoreflect.Uint64Kind ||
		fd.Kind() == protoreflect.Int32Kind ||
		fd.Kind() == protoreflect.Int64Kind:
		{
			return intValueRenderer{}, nil
		}

	default:
		return nil, fmt.Errorf("value renderers cannot format value of type %s", fd.Kind())
	}
}

func (r *Textual) init() {
	if r.scalars == nil {
		r.scalars = map[string]ValueRenderer{}
		r.scalars["cosmos.Int"] = intValueRenderer{}
		r.scalars["cosmos.Dec"] = decValueRenderer{}
	}
}

// DefineScalar adds a value renderer to the given Cosmos scalar.
func (r *Textual) DefineScalar(scalar string, vr ValueRenderer) {
	r.init()
	r.scalars[scalar] = vr
}
