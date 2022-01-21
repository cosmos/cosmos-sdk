package module

import (
	msgv1 "github.com/cosmos/cosmos-sdk/api/cosmos/msg/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"testing"
)

func Test_assertSigners(t *testing.T) {

	type test struct {
		messages          []*descriptorpb.DescriptorProto
		signersDefinition [][]string // what is set in the extension, messages & signers definition must have same length
		wantErr           bool       // expected error, nil means none
	}

	newFD := func() *descriptorpb.FileDescriptorProto {
		return &descriptorpb.FileDescriptorProto{
			Name: proto.String("path/path.proto"),
		}
	}

	protoType := func(typ descriptorpb.FieldDescriptorProto_Type) *descriptorpb.FieldDescriptorProto_Type {
		return &typ
	}

	tests := map[string]test{
		"ok string field": {
			messages: []*descriptorpb.DescriptorProto{
				{
					Name: proto.String("OkStringField"),
					Field: []*descriptorpb.FieldDescriptorProto{
						{
							Name:   proto.String("string_signer"),
							Number: proto.Int32(1),
							Label:  nil,
							Type:   protoType(descriptorpb.FieldDescriptorProto_TYPE_STRING),
						},
					},
				},
			},
			signersDefinition: [][]string{{"string_signer"}},
		},
		"ok message field": {
			messages: []*descriptorpb.DescriptorProto{
				{
					Name: proto.String("OkMessageField"),
					Field: []*descriptorpb.FieldDescriptorProto{
						{
							Name:     proto.String("message_signer"),
							Number:   proto.Int32(1),
							Type:     protoType(descriptorpb.FieldDescriptorProto_TYPE_MESSAGE),
							TypeName: proto.String("OkStringField"),
						},
					},
				},
				{
					Name: proto.String("OkStringField"),
					Field: []*descriptorpb.FieldDescriptorProto{
						{
							Name:   proto.String("string_signer"),
							Number: proto.Int32(1),
							Label:  nil,
							Type:   protoType(descriptorpb.FieldDescriptorProto_TYPE_STRING),
						},
					},
				},
			},
			signersDefinition: [][]string{{"message_signer"}, {"string_signer"}},
		},
		"no signers": {
			messages: []*descriptorpb.DescriptorProto{
				{
					Name: proto.String("OkStringField"),
					Field: []*descriptorpb.FieldDescriptorProto{
						{
							Name:   proto.String("string_signer"),
							Number: proto.Int32(1),
							Label:  nil,
							Type:   protoType(descriptorpb.FieldDescriptorProto_TYPE_STRING),
						},
					},
				},
			},
			wantErr: true,
		},
		"bad kind": {
			messages: []*descriptorpb.DescriptorProto{
				{
					Name: proto.String("OkStringField"),
					Field: []*descriptorpb.FieldDescriptorProto{
						{
							Name:   proto.String("bad_kind"),
							Number: proto.Int32(1),
							Label:  nil,
							Type:   protoType(descriptorpb.FieldDescriptorProto_TYPE_BOOL),
						},
					},
				},
			},
			signersDefinition: [][]string{{"bad_kind"}},
			wantErr:           true,
		},

		"signer not found": {
			messages: []*descriptorpb.DescriptorProto{
				{
					Name: proto.String("OkStringField"),
					Field: []*descriptorpb.FieldDescriptorProto{
						{
							Name:   proto.String("string_signer"),
							Number: proto.Int32(1),
							Label:  nil,
							Type:   protoType(descriptorpb.FieldDescriptorProto_TYPE_STRING),
						},
					},
				},
			},
			signersDefinition: [][]string{{"not_found"}},
			wantErr:           true,
		},
		"recursive": {
			messages: []*descriptorpb.DescriptorProto{
				{
					Name: proto.String("RecursiveA"),
					Field: []*descriptorpb.FieldDescriptorProto{
						{
							Name:     proto.String("recursive"),
							Number:   proto.Int32(1),
							Type:     protoType(descriptorpb.FieldDescriptorProto_TYPE_MESSAGE),
							TypeName: proto.String("RecursiveB"),
						},
					},
				},
				{
					Name: proto.String("RecursiveB"),
					Field: []*descriptorpb.FieldDescriptorProto{
						{
							Name:     proto.String("recursive"),
							Number:   proto.Int32(1),
							Type:     protoType(descriptorpb.FieldDescriptorProto_TYPE_MESSAGE),
							TypeName: proto.String("RecursiveA"),
						},
					},
				},
			},
			signersDefinition: [][]string{{"recursive"}, {"recursive"}},
			wantErr:           true,
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			fdRaw := newFD()
			for _, m := range tc.messages {
				m.Options = &descriptorpb.MessageOptions{} // so it is not nil
				fdRaw.MessageType = append(fdRaw.MessageType, m)
			}

			fd, err := protodesc.NewFile(fdRaw, nil)
			require.NoError(t, err)

			for i, signersXT := range tc.signersDefinition {
				proto.SetExtension(
					fd.Messages().ByName(protoreflect.Name(*tc.messages[i].Name)).Options().(*descriptorpb.MessageOptions),
					msgv1.E_Signer,
					signersXT,
				)
			}

			err = assertSigners(fd.Messages().ByName(protoreflect.Name(*tc.messages[0].Name)), map[protoreflect.FullName]struct{}{})
			require.Equal(t, tc.wantErr, err != nil, "unmatched error criteria %s <->", tc.wantErr, err)
		})
	}
}
