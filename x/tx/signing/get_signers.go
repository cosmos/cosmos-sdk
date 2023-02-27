package signing

import (
	"fmt"
	"strings"

	msgv1 "cosmossdk.io/api/cosmos/msg/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/anypb"
)

type getSignersFunc func(proto.Message) []string

func getSignersFieldNames(descriptor protoreflect.MessageDescriptor) ([]string, error) {
	signersFields := proto.GetExtension(descriptor.Options(), msgv1.E_Signer).([]string)
	if signersFields == nil || len(signersFields) == 0 {
		return nil, fmt.Errorf("no cosmos.msg.v1.signersFields option found for message %s", descriptor.FullName())
	}

	return signersFields, nil
}

func (*MsgContext) makeGetSignersFunc(descriptor protoreflect.MessageDescriptor) (getSignersFunc, error) {
	signersFields, err := getSignersFieldNames(descriptor)
	if err != nil {
		return nil, err
	}

	fieldGetters := make([]func(proto.Message, []string) []string, len(signersFields))
	for i, fieldName := range signersFields {
		field := descriptor.Fields().ByName(protoreflect.Name(fieldName))
		if field == nil {
			return nil, fmt.Errorf("field %s not found in message %s", fieldName, descriptor.FullName())
		}

		if field.IsMap() || field.HasOptionalKeyword() {
			return nil, fmt.Errorf("cosmos.msg.v1.signer field %s in message %s must not be a map or optional", fieldName, descriptor.FullName())
		}

		switch field.Kind() {
		case protoreflect.StringKind:
			if field.IsList() {
				fieldGetters[i] = func(msg proto.Message, arr []string) []string {
					signers := msg.ProtoReflect().Get(field).List()
					n := signers.Len()
					for i := 0; i < n; i++ {
						arr = append(arr, signers.Get(i).String())
					}
					return arr
				}
			} else {
				fieldGetters[i] = func(msg proto.Message, arr []string) []string {
					return append(arr, msg.ProtoReflect().Get(field).String())
				}
			}
		case protoreflect.MessageKind:
			isList := field.IsList()
			nestedMessage := field.Message()
			nestedSignersFields, err := getSignersFieldNames(nestedMessage)
			if err != nil {
				return nil, err
			}

			if len(nestedSignersFields) != 1 {
				return nil, fmt.Errorf("nested cosmos.msg.v1.signer option in message %s must contain only one value", nestedMessage.FullName())
			}

			nestedFieldName := nestedSignersFields[0]
			nestedField := nestedMessage.Fields().ByName(protoreflect.Name(nestedFieldName))
			nestedIsList := nestedField.IsList()
			if nestedField == nil {
				return nil, fmt.Errorf("field %s not found in message %s", nestedFieldName, nestedMessage.FullName())
			}

			if nestedField.Kind() != protoreflect.StringKind || nestedField.IsMap() || nestedField.HasOptionalKeyword() {
				return nil, fmt.Errorf("nested signer field %s in message %s must be a simple string", nestedFieldName, nestedMessage.FullName())
			}

			if isList {
				if nestedIsList {
					fieldGetters[i] = func(msg proto.Message, arr []string) []string {
						msgs := msg.ProtoReflect().Get(field).List()
						m := msgs.Len()
						for i := 0; i < m; i++ {
							signers := msgs.Get(i).Message().Get(nestedField).List()
							n := signers.Len()
							for j := 0; j < n; j++ {
								arr = append(arr, signers.Get(j).String())
							}
						}
						return arr
					}
				} else {
					fieldGetters[i] = func(msg proto.Message, arr []string) []string {
						msgs := msg.ProtoReflect().Get(field).List()
						m := msgs.Len()
						for i := 0; i < m; i++ {
							arr = append(arr, msgs.Get(i).Message().Get(nestedField).String())
						}
						return arr
					}
				}
			} else {
				if nestedIsList {
					fieldGetters[i] = func(msg proto.Message, arr []string) []string {
						nestedMsg := msg.ProtoReflect().Get(field).Message()
						signers := nestedMsg.Get(nestedField).List()
						n := signers.Len()
						for j := 0; j < n; j++ {
							arr = append(arr, signers.Get(j).String())
						}
						return arr
					}
				} else {
					fieldGetters[i] = func(msg proto.Message, arr []string) []string {
						return append(arr, msg.ProtoReflect().Get(field).Message().Get(nestedField).String())
					}
				}
			}

		default:
			return nil, fmt.Errorf("unexpected field type %s for field %s in message %s", field.Kind(), fieldName, descriptor.FullName())
		}
	}

	return func(message proto.Message) []string {
		var signers []string
		for _, getter := range fieldGetters {
			signers = getter(message, signers)
		}
		return signers
	}, nil
}

func (c *MsgContext) GetSignersForAny(msg *anypb.Any) ([]string, error) {
	messageType, err := c.protoTypes.FindMessageByURL(msg.TypeUrl)
	if err == protoregistry.NotFound {
		messageName := protoreflect.FullName(msg.TypeUrl)
		if i := strings.LastIndexByte(string(messageName), '/'); i >= 0 {
			messageName = messageName[i+1:]
		}

		messageDesc, err := c.protoFiles.FindDescriptorByName(messageName)
		if err != nil {
			return nil, err
		}
		messageType = dynamicpb.NewMessageType(messageDesc.(protoreflect.MessageDescriptor))
	} else if err != nil {
		return nil, err
	}

	message := messageType.New().Interface()
	if err := msg.UnmarshalTo(message); err != nil {
		return nil, err
	}

	return c.GetSignersForMessage(message)
}

func (c *MsgContext) GetSignersForMessage(msg proto.Message) ([]string, error) {
	messageDescriptor := msg.ProtoReflect().Descriptor()
	f, ok := c.getSignersFuncs[messageDescriptor.FullName()]
	if !ok {
		var err error
		f, err = c.makeGetSignersFunc(messageDescriptor)
		if err != nil {
			return nil, err
		}
		c.getSignersFuncs[messageDescriptor.FullName()] = f
	}

	return f(msg), nil
}
