package unstructured

import (
	"fmt"
	"reflect"

	"github.com/spf13/cast"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

// Object represents the behaviour of a raw object that can marshal itself
// to a proto dynamic message given its file descriptor
type Object interface {
	Marshal(desc protoreflect.MessageDescriptor) (*dynamicpb.Message, error)
}

// Map defines an unstructured map object that can be used to fill protobuf objects recursively
// types should be either pointers or golang primitive types, as of now, using type aliases
// is not supported.
type Map map[string]interface{}

func (o Map) Marshal(md protoreflect.MessageDescriptor) (*dynamicpb.Message, error) {
	dyn := dynamicpb.NewMessage(md)
	for fieldName, interfaceValue := range o {
		fd := md.Fields().ByName(protoreflect.Name(fieldName))
		if fd == nil {
			return nil, fmt.Errorf("descriptor %s does not contain field named %s", md.FullName(), fieldName)
		}
		switch {
		case fd.IsList():
			value := dyn.NewField(fd)
			listValue := value.List()
			err := fillList(interfaceValue, listValue, fd, md)
			if err != nil {
				return nil, err
			}
			dyn.Set(fd, value)
		case fd.IsMap():
			rawField := dyn.NewField(fd)
			pMap := rawField.Map()
			err := fillMap(interfaceValue, pMap, fd, md)
			if err != nil {
				return nil, err
			}
			dyn.Set(fd, rawField)
		default:
			pv, err := interfaceToProtoValue(interfaceValue, fd, md)
			if err != nil {
				return nil, err
			}
			// set the field
			dyn.Set(fd, pv)
		}
	}

	return dyn, nil
}

// TODO: indirect
func fillMap(v interface{}, pMap protoreflect.Map, fd protoreflect.FieldDescriptor, md protoreflect.MessageDescriptor) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Map {
		return errTypeMismatch(md, fd, v)
	}

	keyDesc := fd.MapKey()
	valueDesc := fd.MapValue()

	mapIter := rv.MapRange()

	for mapIter.Next() {
		k := mapIter.Key().Interface()
		v := mapIter.Value().Interface()

		// cast k and v to kDesc and vDesc
		kValue, err := interfaceToProtoValue(k, keyDesc, md)
		if err != nil {
			return fmt.Errorf("unable to set map key for map field descriptor %s: %w", fd, err)
		}
		vValue, err := interfaceToProtoValue(v, valueDesc, md)
		if err != nil {
			return fmt.Errorf("unable to set map value for map field descriptor %s: %w", fd, err)
		}
		// set in map
		pMap.Set(kValue.MapKey(), vValue)
	}

	return nil
}

// TODO: indirect
func fillList(v interface{}, list protoreflect.List, fd protoreflect.FieldDescriptor, md protoreflect.MessageDescriptor) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		return errTypeMismatch(md, fd, v)
	}
	for i := 0; i < rv.Len(); i++ {
		iv := rv.Index(i).Interface()
		v, err := interfaceToProtoValue(iv, fd, md)
		if err != nil {
			return err
		}
		list.Append(v)
	}
	return nil
}

func interfaceToProtoValue(interfaceValue interface{}, fieldDesc protoreflect.FieldDescriptor, desc protoreflect.MessageDescriptor) (protoreflect.Value, error) {
	switch fieldDesc.Kind() {
	// bool
	case protoreflect.BoolKind:
		v, err := cast.ToBoolE(interfaceValue)
		if err != nil {
			return protoreflect.Value{}, errTypeMismatch(desc, fieldDesc, interfaceValue)
		}
		return protoreflect.ValueOfBool(v), nil
	// enum
	case protoreflect.EnumKind:
		v, err := cast.ToInt32E(interfaceValue)
		if err != nil {
			return protoreflect.Value{}, errTypeMismatch(desc, fieldDesc, interfaceValue)
		}
		return protoreflect.ValueOfEnum((protoreflect.EnumNumber)(v)), nil
	// int32
	case protoreflect.Int32Kind:
		v, err := cast.ToInt32E(interfaceValue)
		if err != nil {
			return protoreflect.Value{}, errTypeMismatch(desc, fieldDesc, interfaceValue)
		}
		return protoreflect.ValueOfInt32(v), nil
	// uint32
	case protoreflect.Uint32Kind:
		v, err := cast.ToUint32E(interfaceValue)
		if err != nil {
			return protoreflect.Value{}, errTypeMismatch(desc, fieldDesc, interfaceValue)
		}
		return protoreflect.ValueOfUint32(v), nil
	// int64
	case protoreflect.Int64Kind:
		v, err := cast.ToInt64E(interfaceValue)
		if err != nil {
			return protoreflect.Value{}, errTypeMismatch(desc, fieldDesc, interfaceValue)
		}
		return protoreflect.ValueOfInt64(v), nil
	// uint64
	case protoreflect.Uint64Kind:
		v, err := cast.ToUint64E(interfaceValue)
		if err != nil {
			return protoreflect.Value{}, errTypeMismatch(desc, fieldDesc, interfaceValue)
		}
		return protoreflect.ValueOfUint64(v), nil
	// float
	case protoreflect.FloatKind:
		v, err := cast.ToFloat32E(interfaceValue)
		if err != nil {
			return protoreflect.Value{}, errTypeMismatch(desc, fieldDesc, interfaceValue)
		}
		return protoreflect.ValueOfFloat32(v), nil
	// handle double
	case protoreflect.DoubleKind:
		v, err := cast.ToFloat64E(interfaceValue)
		if err != nil {
			return protoreflect.Value{}, errTypeMismatch(desc, fieldDesc, interfaceValue)
		}
		return protoreflect.ValueOfFloat64(v), nil
	// string
	case protoreflect.StringKind:
		v, err := cast.ToStringE(interfaceValue)
		if err != nil {
			return protoreflect.Value{}, errTypeMismatch(desc, fieldDesc, interfaceValue)
		}
		return protoreflect.ValueOfString(v), nil
	// bytes
	case protoreflect.BytesKind:
		v, err := castToBytes(interfaceValue)
		if err != nil {
			return protoreflect.Value{}, errTypeMismatch(desc, fieldDesc, interfaceValue)
		}
		return protoreflect.ValueOfBytes(v), nil
	// handle message kind
	case protoreflect.MessageKind:
		ob, ok := interfaceValue.(Object)
		if !ok {
			return protoreflect.Value{}, fmt.Errorf("descriptor %s expected %s kind at %s which should be castable to unstructured.Object, got: %T", desc, protoreflect.MessageKind, fieldDesc, interfaceValue)
		}
		dpb, err := ob.Marshal(fieldDesc.Message())
		if err != nil {
			return protoreflect.Value{}, fmt.Errorf("descriptor %s failed to marshal Message for field descriptor %s: %w", desc, fieldDesc, err)
		}
		return protoreflect.ValueOfMessage(dpb), nil
	default:
		return protoreflect.Value{}, fmt.Errorf("descriptor %s field %s unsupported type: %s", desc.FullName(), fieldDesc.FullName(), fieldDesc.Kind().String())
	}
}

func castToBytes(value interface{}) ([]byte, error) {
	switch casted := value.(type) {
	case []byte:
		return casted, nil
	}
	return nil, fmt.Errorf("unable to cast %#v of type %T to string", value, value)
}

func errTypeMismatch(desc protoreflect.MessageDescriptor, field protoreflect.FieldDescriptor, v interface{}) error {
	return fmt.Errorf("descriptor %s field %s expects %s, got: %T", desc.FullName(), field.FullName(), field.Kind(), v)
}
