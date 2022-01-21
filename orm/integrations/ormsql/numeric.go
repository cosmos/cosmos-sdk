package ormsql

import (
	"reflect"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type int32Codec struct{}

func (i int32Codec) goType() reflect.Type { return reflect.TypeOf(int32(0)) }

func (i int32Codec) encode(protoValue protoreflect.Value, goValue reflect.Value) error {
	goValue.SetInt(protoValue.Int())
	return nil
}

func (i int32Codec) decode(goValue reflect.Value) (protoreflect.Value, error) {
	return protoreflect.ValueOfInt32(int32(goValue.Int())), nil
}

type uint32Codec struct{}

func (u uint32Codec) goType() reflect.Type { return reflect.TypeOf(uint32(0)) }

func (u uint32Codec) encode(protoValue protoreflect.Value, goValue reflect.Value) error {
	goValue.SetUint(protoValue.Uint())
	return nil
}

func (u uint32Codec) decode(goValue reflect.Value) (protoreflect.Value, error) {
	return protoreflect.ValueOfUint32(uint32(goValue.Uint())), nil
}

type int64Codec struct{}

func (i int64Codec) goType() reflect.Type { return reflect.TypeOf(int64(0)) }

func (i int64Codec) encode(protoValue protoreflect.Value, goValue reflect.Value) error {
	goValue.SetInt(protoValue.Int())
	return nil
}

func (i int64Codec) decode(goValue reflect.Value) (protoreflect.Value, error) {
	return protoreflect.ValueOfInt64(goValue.Int()), nil
}
