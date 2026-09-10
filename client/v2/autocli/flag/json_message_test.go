package flag

import (
	"context"
	"testing"

	cosmos_proto "github.com/cosmos/cosmos-proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestJSONMessageFlagLegacyDec(t *testing.T) {
	options := &descriptorpb.FieldOptions{}
	proto.SetExtension(options, cosmos_proto.E_Scalar, DecScalarType)
	file, err := protodesc.NewFile(&descriptorpb.FileDescriptorProto{
		Syntax:  proto.String("proto3"),
		Name:    proto.String("params.proto"),
		Package: proto.String("cosmos.test.v1"),
		MessageType: []*descriptorpb.DescriptorProto{{
			Name: proto.String("Params"),
			Field: []*descriptorpb.FieldDescriptorProto{{
				Name:     proto.String("inflation_max"),
				JsonName: proto.String("inflationMax"),
				Number:   proto.Int32(1),
				Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
				Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
				Options:  options,
			}},
		}},
	}, nil)
	require.NoError(t, err)

	descriptor := file.Messages().ByName("Params")
	messageType := dynamicpb.NewMessageType(descriptor)
	types := new(protoregistry.Types)
	require.NoError(t, types.RegisterMessage(messageType))

	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "human readable decimal is encoded as atomics",
			input: `{"inflationMax":"0.020000000000000000"}`,
			want:  "20000000000000000",
		},
		{
			// Atomics are indistinguishable from a human readable integer, so
			// input that already encodes the field must survive untouched
			// instead of being scaled by LegacyPrecision a second time.
			name:  "atomics are left as supplied",
			input: `{"inflationMax":"20000000000000000"}`,
			want:  "20000000000000000",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			value := jsonMessageFlagType{messageDesc: descriptor}.NewValue(
				&ctx,
				&Builder{TypeResolver: types},
			)
			require.NoError(t, value.Set(tc.input))

			got, err := value.Get(protoreflect.Value{})
			require.NoError(t, err)
			require.Equal(
				t,
				tc.want,
				got.Message().Get(descriptor.Fields().ByName("inflation_max")).String(),
			)
		})
	}
}

func TestJSONMessageFlagLegacyDecEncodesBytesFields(t *testing.T) {
	options := &descriptorpb.FieldOptions{}
	proto.SetExtension(options, cosmos_proto.E_Scalar, DecScalarType)
	file, err := protodesc.NewFile(&descriptorpb.FileDescriptorProto{
		Syntax:  proto.String("proto3"),
		Name:    proto.String("bytes_params.proto"),
		Package: proto.String("cosmos.test.v1"),
		MessageType: []*descriptorpb.DescriptorProto{{
			Name: proto.String("BytesParams"),
			Field: []*descriptorpb.FieldDescriptorProto{{
				Name:     proto.String("inflation_max"),
				JsonName: proto.String("inflationMax"),
				Number:   proto.Int32(1),
				Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
				Type:     descriptorpb.FieldDescriptorProto_TYPE_BYTES.Enum(),
				Options:  options,
			}},
		}},
	}, nil)
	require.NoError(t, err)

	descriptor := file.Messages().ByName("BytesParams")
	messageType := dynamicpb.NewMessageType(descriptor)
	types := new(protoregistry.Types)
	require.NoError(t, types.RegisterMessage(messageType))

	ctx := context.Background()
	value := jsonMessageFlagType{messageDesc: descriptor}.NewValue(
		&ctx,
		&Builder{TypeResolver: types},
	)

	// A bytes-backed scalar requires protobuf JSON's base64 representation.
	// Writing the atomics as a plain JSON string would make protojson decode
	// the digits themselves as base64 and silently corrupt the field bytes.
	require.NoError(t, value.Set(`{"inflationMax":"0.000000000000001234"}`))

	got, err := value.Get(protoreflect.Value{})
	require.NoError(t, err)
	require.Equal(
		t,
		[]byte("1234"),
		got.Message().Get(descriptor.Fields().ByName("inflation_max")).Bytes(),
	)

	// Existing protobuf JSON for bytes is already base64 and must pass through.
	require.NoError(t, value.Set(`{"inflationMax":"MTIzNA=="}`))
	got, err = value.Get(protoreflect.Value{})
	require.NoError(t, err)
	require.Equal(
		t,
		[]byte("1234"),
		got.Message().Get(descriptor.Fields().ByName("inflation_max")).Bytes(),
	)
}

func TestJSONMessageFlagPreservesTimestampJSON(t *testing.T) {
	file, err := protodesc.NewFile(&descriptorpb.FileDescriptorProto{
		Syntax:     proto.String("proto3"),
		Name:       proto.String("timestamp_params.proto"),
		Package:    proto.String("cosmos.test.v1"),
		Dependency: []string{"google/protobuf/timestamp.proto"},
		MessageType: []*descriptorpb.DescriptorProto{{
			Name: proto.String("TimestampParams"),
			Field: []*descriptorpb.FieldDescriptorProto{{
				Name:     proto.String("updated_at"),
				JsonName: proto.String("updatedAt"),
				Number:   proto.Int32(1),
				Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
				Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
				TypeName: proto.String(".google.protobuf.Timestamp"),
			}},
		}},
	}, protoregistry.GlobalFiles)
	require.NoError(t, err)

	descriptor := file.Messages().ByName("TimestampParams")
	messageType := dynamicpb.NewMessageType(descriptor)
	types := new(protoregistry.Types)
	require.NoError(t, types.RegisterMessage(messageType))
	require.NoError(t, types.RegisterMessage((&timestamppb.Timestamp{}).ProtoReflect().Type()))

	ctx := context.Background()
	value := jsonMessageFlagType{messageDesc: descriptor}.NewValue(
		&ctx,
		&Builder{TypeResolver: types},
	)
	require.NoError(t, value.Set(`{"updatedAt":"2026-08-16T00:00:00Z"}`))

	got, err := value.Get(protoreflect.Value{})
	require.NoError(t, err)
	timestamp := got.Message().Get(descriptor.Fields().ByName("updated_at")).Message()
	require.Equal(
		t,
		int64(1786838400),
		timestamp.Get(timestamp.Descriptor().Fields().ByName("seconds")).Int(),
	)
}

func TestJSONMessageFlagPreservesStructJSON(t *testing.T) {
	file, err := protodesc.NewFile(&descriptorpb.FileDescriptorProto{
		Syntax:     proto.String("proto3"),
		Name:       proto.String("struct_params.proto"),
		Package:    proto.String("cosmos.test.v1"),
		Dependency: []string{"google/protobuf/struct.proto"},
		MessageType: []*descriptorpb.DescriptorProto{{
			Name: proto.String("StructParams"),
			Field: []*descriptorpb.FieldDescriptorProto{{
				Name:     proto.String("metadata"),
				JsonName: proto.String("metadata"),
				Number:   proto.Int32(1),
				Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
				Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
				TypeName: proto.String(".google.protobuf.Struct"),
			}},
		}},
	}, protoregistry.GlobalFiles)
	require.NoError(t, err)

	descriptor := file.Messages().ByName("StructParams")
	messageType := dynamicpb.NewMessageType(descriptor)
	types := new(protoregistry.Types)
	require.NoError(t, types.RegisterMessage(messageType))
	require.NoError(t, types.RegisterMessage((&structpb.Struct{}).ProtoReflect().Type()))

	ctx := context.Background()
	value := jsonMessageFlagType{messageDesc: descriptor}.NewValue(
		&ctx,
		&Builder{TypeResolver: types},
	)
	require.NoError(t, value.Set(`{"metadata":{"fields":"literal"}}`))
	require.JSONEq(t, `{"metadata":{"fields":"literal"}}`, value.String())
}
