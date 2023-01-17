package textual

import (
	"bytes"
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	"cosmossdk.io/x/tx/signing"
	"cosmossdk.io/x/tx/textual/internal/textualpb"
	cosmos_proto "github.com/cosmos/cosmos-proto"
)

// CoinMetadataQueryFn defines a function that queries state for the coin denom
// metadata. It is meant to be passed as an argument into `NewTextual`.
type CoinMetadataQueryFn func(ctx context.Context, denom string) (*bankv1beta1.Metadata, error)

// ValueRendererCreator is a function returning a textual.
type ValueRendererCreator func(protoreflect.FieldDescriptor) ValueRenderer

// Textual holds the configuration for dispatching
// to specific value renderers for SIGN_MODE_TEXTUAL.
type Textual struct {
	// coinMetadataQuerier defines a function to query the coin metadata from
	// state. It should use bank module's `DenomsMetadata` gRPC query to fetch
	// each denom's associated metadata, either using the bank keeper (for
	// server-side code) or a gRPC query client (for client-side code).
	coinMetadataQuerier CoinMetadataQueryFn
	// scalars defines a registry for Cosmos scalars.
	scalars map[string]ValueRendererCreator
	// messages defines a registry for custom message renderers.
	// Note that we also use this same registry for the
	// following messages, as they can be thought of custom message rendering:
	// - SDK coin and coins
	// - Protobuf timestamp
	// - Protobuf duration
	messages map[protoreflect.FullName]ValueRenderer
}

// NewTextual returns a new Textual which provides
// value renderers.
func NewTextual(q CoinMetadataQueryFn) Textual {
	t := Textual{coinMetadataQuerier: q}
	t.init()
	return t
}

// GetFieldValueRenderer returns the value renderer for the given FieldDescriptor.
func (r *Textual) GetFieldValueRenderer(fd protoreflect.FieldDescriptor) (ValueRenderer, error) {
	switch {
	// Scalars, such as sdk.Int and sdk.Dec encoded as strings.
	case fd.Kind() == protoreflect.StringKind:
		if proto.GetExtension(fd.Options(), cosmos_proto.E_Scalar) != "" {
			scalar, ok := proto.GetExtension(fd.Options(), cosmos_proto.E_Scalar).(string)
			if !ok || scalar == "" {
				return nil, fmt.Errorf("got extension option %s of type %T", scalar, scalar)
			}

			vr := r.scalars[scalar]
			if vr != nil {
				return vr(fd), nil
			}
		}

		return NewStringValueRenderer(), nil

	case fd.Kind() == protoreflect.BytesKind:
		return NewBytesValueRenderer(), nil

	// Integers
	case fd.Kind() == protoreflect.Uint32Kind ||
		fd.Kind() == protoreflect.Uint64Kind ||
		fd.Kind() == protoreflect.Int32Kind ||
		fd.Kind() == protoreflect.Int64Kind:
		return NewIntValueRenderer(fd), nil

	case fd.Kind() == protoreflect.EnumKind:
		return NewEnumValueRenderer(fd), nil

	case fd.Kind() == protoreflect.MessageKind:
		md := fd.Message()
		fullName := md.FullName()

		vr, found := r.messages[fullName]
		if found {
			return vr, nil
		}

		if fd.IsMap() {
			return nil, fmt.Errorf("value renderers cannot format value of type map")
		}
		return NewMessageValueRenderer(r, md), nil

	default:
		return nil, fmt.Errorf("value renderers cannot format value of type %s", fd.Kind())
	}
}

// GetMessageValueRenderer is a specialization of GetValueRenderer for messages.
// It is useful when the message type is discovered outside the context of a field,
// e.g. when handling a google.protobuf.Any.
func (r *Textual) GetMessageValueRenderer(md protoreflect.MessageDescriptor) (ValueRenderer, error) {
	fullName := md.FullName()
	vr, found := r.messages[fullName]
	if found {
		return vr, nil
	}
	return NewMessageValueRenderer(r, md), nil
}

// init initializes Textual's internal `scalars` and `messages` registry for
// custom scalar and message renderers.
//
// It is an idempotent method.
func (r *Textual) init() {
	if r.scalars == nil {
		r.scalars = map[string]ValueRendererCreator{}
		r.scalars["cosmos.Int"] = func(fd protoreflect.FieldDescriptor) ValueRenderer { return NewIntValueRenderer(fd) }
		r.scalars["cosmos.Dec"] = func(_ protoreflect.FieldDescriptor) ValueRenderer { return NewDecValueRenderer() }
	}
	if r.messages == nil {
		r.messages = map[protoreflect.FullName]ValueRenderer{}
		r.messages[(&basev1beta1.Coin{}).ProtoReflect().Descriptor().FullName()] = NewCoinsValueRenderer(r.coinMetadataQuerier)
		r.messages[(&durationpb.Duration{}).ProtoReflect().Descriptor().FullName()] = NewDurationValueRenderer()
		r.messages[(&timestamppb.Timestamp{}).ProtoReflect().Descriptor().FullName()] = NewTimestampValueRenderer()
		r.messages[(&anypb.Any{}).ProtoReflect().Descriptor().FullName()] = NewAnyValueRenderer(r)
		r.messages[(&textualpb.TextualData{}).ProtoReflect().Descriptor().FullName()] = NewTxValueRenderer(r)
	}
}

// DefineScalar adds a value renderer to the given Cosmos scalar.
func (r *Textual) DefineScalar(scalar string, vr ValueRendererCreator) {
	r.init()
	r.scalars[scalar] = vr
}

// DefineMessageRenderer adds a new custom message renderer.
func (r *Textual) DefineMessageRenderer(name protoreflect.FullName, vr ValueRenderer) {
	r.init()
	r.messages[name] = vr
}

// GetSignBytes returns the transaction sign bytes.
func (r *Textual) GetSignBytes(ctx context.Context, bodyBz, authInfoBz []byte, signerData signing.SignerData) ([]byte, error) {
	data := &textualpb.TextualData{
		BodyBytes:     bodyBz,
		AuthInfoBytes: authInfoBz,
		SignerData: &textualpb.SignerData{
			Address:       signerData.Address,
			ChainId:       signerData.ChainId,
			AccountNumber: signerData.AccountNumber,
			Sequence:      signerData.Sequence,
			PubKey:        signerData.PubKey,
		},
	}

	vr, err := r.GetMessageValueRenderer(data.ProtoReflect().Descriptor())
	if err != nil {
		return nil, err
	}

	screens, err := vr.Format(ctx, protoreflect.ValueOf(data.ProtoReflect()))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = encode(screens, &buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
