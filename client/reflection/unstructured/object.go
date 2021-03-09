package unstructured

import (
	"fmt"
	"reflect"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

// Object defines an unstructured object that can be used to fill protobuf objects recursively
type Object map[string]interface{}

func (o Object) MarshalFromDescriptor(desc protoreflect.MessageDescriptor) (proto.Message, error) {
	dyn := dynamicpb.NewMessage(desc)
	for fieldName, interfaceValue := range o {
		fieldDesc := desc.Fields().ByName(protoreflect.Name(fieldName))
		if fieldDesc == nil {
			return nil, fmt.Errorf("descriptor %s does not contain field named %s", desc.FullName(), fieldName)
		}
		v := reflect.ValueOf(interfaceValue)
		var pv protoreflect.Value // pv is what we will use to set the field
		// TODO eventually implement all...
		switch fieldDesc.Kind() {
		// bool
		case protoreflect.BoolKind:
			if v.Kind() != reflect.Bool {
				return nil, errTypeMismatch(desc, fieldDesc, v)
			}
			pv = protoreflect.ValueOfBool(v.Bool())
		// string
		case protoreflect.StringKind:
			if v.Kind() != reflect.String {
				return nil, errTypeMismatch(desc, fieldDesc, v)
			}
			pv = protoreflect.ValueOfString(v.String())
		default:
			return nil, fmt.Errorf("descriptor %s field %s unsupported type: %s", desc.FullName(), fieldDesc.FullName(), fieldDesc.Kind().String())
		}

		// set the field
		dyn.Set(fieldDesc, pv)
	}

	return dyn, nil
}

func errTypeMismatch(desc protoreflect.MessageDescriptor, field protoreflect.FieldDescriptor, v interface{}) error {
	return fmt.Errorf("descriptor %s field %s expects a bool, got: %T", desc.FullName(), field.FullName(), v)
}
