package valuerenderer

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	cosmos_proto "github.com/cosmos/cosmos-proto"
)

type CoinMetadataQueryFn func(ctx context.Context, denom string) (*bankv1beta1.Metadata, error)

type Textual struct {
	// coinMetadataQuerier defines a function to query the coin metadata from
	// state.
	coinMetadataQuerier CoinMetadataQueryFn
	// scalars defines a registry for Cosmos scalars.
	scalars map[string]ValueRenderer
}

// NewTextual creates a new SIGN_MODE_TEXTUAL renderer.
func NewTextual(q CoinMetadataQueryFn) Textual {
	return Textual{coinMetadataQuerier: q}
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

	// Integers
	case fd.Kind() == protoreflect.Uint32Kind ||
		fd.Kind() == protoreflect.Uint64Kind ||
		fd.Kind() == protoreflect.Int32Kind ||
		fd.Kind() == protoreflect.Int64Kind:
		{
			return intValueRenderer{}, nil
		}

	// Coin and Coins
	case fd.Kind() == protoreflect.MessageKind && (&basev1beta1.Coin{}).ProtoReflect().Descriptor() == fd.Message():
		{
			if fd.Cardinality() == protoreflect.Repeated {
				return coinsValueRenderer{r.coinMetadataQuerier}, nil
			} else {
				return coinValueRenderer{r.coinMetadataQuerier}, nil
			}
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
