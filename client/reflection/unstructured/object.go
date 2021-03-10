package unstructured

import (
	"fmt"

	"github.com/spf13/cast"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

// Object defines an unstructured object that can be used to fill protobuf objects recursively
// types should be either pointers or golang primitive types, as of now, using type aliases
// is not supported.
type Object map[string]interface{}

func (o Object) Marshal(desc protoreflect.MessageDescriptor) (*dynamicpb.Message, error) {
	dyn := dynamicpb.NewMessage(desc)
	for fieldName, interfaceValue := range o {
		fieldDesc := desc.Fields().ByName(protoreflect.Name(fieldName))
		if fieldDesc == nil {
			return nil, fmt.Errorf("descriptor %s does not contain field named %s", desc.FullName(), fieldName)
		}
		var pv protoreflect.Value // pv is what we will use to set the field
		// TODO eventually implement all...
		// TODO map list
		switch fieldDesc.Kind() {
		// bool
		case protoreflect.BoolKind:
			v, err := cast.ToBoolE(interfaceValue)
			if err != nil {
				return nil, errTypeMismatch(desc, fieldDesc, interfaceValue)
			}
			pv = protoreflect.ValueOfBool(v)
		// enum
		case protoreflect.EnumKind:
			v, err := cast.ToInt32E(interfaceValue)
			if err != nil {
				return nil, errTypeMismatch(desc, fieldDesc, interfaceValue)
			}
			pv = protoreflect.ValueOfEnum((protoreflect.EnumNumber)(v))
		// int32
		case protoreflect.Int32Kind:
			v, err := cast.ToInt32E(interfaceValue)
			if err != nil {
				return nil, errTypeMismatch(desc, fieldDesc, interfaceValue)
			}
			pv = protoreflect.ValueOfInt32(v)
		// int64
		case protoreflect.Int64Kind:
			v, err := cast.ToInt64E(interfaceValue)
			if err != nil {
				return nil, errTypeMismatch(desc, fieldDesc, interfaceValue)
			}
			pv = protoreflect.ValueOfInt64(v)
		// uint64
		case protoreflect.Uint64Kind:
			v, err := cast.ToUint64E(interfaceValue)
			if err != nil {
				return nil, errTypeMismatch(desc, fieldDesc, interfaceValue)
			}
			pv = protoreflect.ValueOfUint64(v)
		// float
		case protoreflect.FloatKind:
			v, err := cast.ToFloat32E(interfaceValue)
			if err != nil {
				return nil, errTypeMismatch(desc, fieldDesc, interfaceValue)
			}
			pv = protoreflect.ValueOfFloat32(v)
		// handle double
		case protoreflect.DoubleKind:
			v, err := cast.ToFloat64E(interfaceValue)
			if err != nil {
				return nil, errTypeMismatch(desc, fieldDesc, interfaceValue)
			}
			pv = protoreflect.ValueOfFloat64(v)
		// string
		case protoreflect.StringKind:
			v, err := cast.ToStringE(interfaceValue)
			if err != nil {
				return nil, errTypeMismatch(desc, fieldDesc, interfaceValue)
			}
			pv = protoreflect.ValueOfString(v)
		// bytes
		case protoreflect.BytesKind:
			v, err := castToBytes(interfaceValue)
			if err != nil {
				return nil, errTypeMismatch(desc, fieldDesc, interfaceValue)
			}
			pv = protoreflect.ValueOfBytes(v)
		// handle messages
		case protoreflect.MessageKind:
			recObj, ok := interfaceValue.(Object)
			if !ok {
				return nil, fmt.Errorf("descriptor %s expects a message type at field %s which should be expressed as unstructured.Object, got: %T", desc, fieldDesc, interfaceValue)
			}
			msg, err := recObj.Marshal(fieldDesc.Message())
			if err != nil {
				return nil, fmt.Errorf("descriptor %s: unable to unmarshal recursive type at field %s: %w", desc, fieldDesc, err)
			}
			pv = protoreflect.ValueOfMessage(msg.ProtoReflect())
		default:
			return nil, fmt.Errorf("descriptor %s field %s unsupported type: %s", desc.FullName(), fieldDesc.FullName(), fieldDesc.Kind().String())
		}

		// set the field
		dyn.Set(fieldDesc, pv)
	}

	return dyn, nil
}

func castToBytes(value interface{}) ([]byte, error) {
	switch casted := value.(type) {
	case []byte:
		return casted, nil
	}
	return nil, fmt.Errorf("unable to cast %#v of type %T to string", value, value)
}

func errTypeMismatch(desc protoreflect.MessageDescriptor, field protoreflect.FieldDescriptor, v interface{}) error {
	return fmt.Errorf("descriptor %s field %s expects a bool, got: %T", desc.FullName(), field.FullName(), v)
}
