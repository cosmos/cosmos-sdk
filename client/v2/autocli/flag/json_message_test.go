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

	ctx := context.Background()
	value := jsonMessageFlagType{messageDesc: descriptor}.NewValue(
		&ctx,
		&Builder{TypeResolver: types},
	)
	require.NoError(t, value.Set(`{"inflationMax":"0.020000000000000000"}`))

	got, err := value.Get(protoreflect.Value{})
	require.NoError(t, err)
	require.Equal(
		t,
		"20000000000000000",
		got.Message().Get(descriptor.Fields().ByName("inflation_max")).String(),
	)
}
