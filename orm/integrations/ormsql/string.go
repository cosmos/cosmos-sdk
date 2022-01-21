package ormsql

import (
	"reflect"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type stringCodec struct{}

func (s stringCodec) goType() reflect.Type {
	return reflect.TypeOf("")
}

func (s stringCodec) encode(protoValue protoreflect.Value, goValue reflect.Value) error {
	// here we use any string generated from the proto value
	goValue.SetString(protoValue.String())
	return nil
}

func (s stringCodec) decode(goValue reflect.Value) (protoreflect.Value, error) {
	return protoreflect.ValueOfString(goValue.String()), nil
}
