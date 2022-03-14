package ormsql

import (
	"reflect"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/structpb"
	"gorm.io/datatypes"
)

type mapCodec struct {
	jsonMarshalOptions protojson.MarshalOptions
}

func (m mapCodec) goType() reflect.Type {
	return reflect.TypeOf(datatypes.JSON{})
}

func (m mapCodec) encode(protoValue protoreflect.Value, goValue reflect.Value) error {
	protoMap := protoValue.Map()
	goMap := map[string]interface{}{}
	protoMap.Range(func(key protoreflect.MapKey, value protoreflect.Value) bool {
		goMap[key.String()] = value.Interface()
		return true
	})

	protoStruct, err := structpb.NewStruct(goMap)
	if err != nil {
		return err
	}

	bz, err := m.jsonMarshalOptions.Marshal(protoStruct)
	if err != nil {
		return err
	}

	goValue.Set(reflect.ValueOf(datatypes.JSON(bz)))
	return nil
}

func (m mapCodec) decode(goValue reflect.Value) (protoreflect.Value, error) {
	//TODO implement me
	panic("implement me")
}
