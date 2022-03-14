package ormsql

import (
	"reflect"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/structpb"
	"gorm.io/datatypes"
)

type listCodec struct {
	jsonMarshalOptions   protojson.MarshalOptions
	jsonUnmarshalOptions protojson.UnmarshalOptions
}

func (l listCodec) goType() reflect.Type {
	return reflect.TypeOf(datatypes.JSON{})
}

func (l listCodec) encode(protoValue protoreflect.Value, goValue reflect.Value) error {
	list := protoValue.List()
	n := list.Len()
	values := make([]interface{}, n)
	for i := 0; i < n; i++ {
		values[i] = list.Get(i).Interface()
	}

	structList, err := structpb.NewList(values)
	if err != nil {
		return err
	}

	bz, err := l.jsonMarshalOptions.Marshal(structList)
	if err != nil {
		return err
	}

	goValue.Set(reflect.ValueOf(datatypes.JSON(bz)))
	return nil
}

func (l listCodec) decode(goValue reflect.Value) (protoreflect.Value, error) {
	var structList structpb.ListValue
	err := l.jsonUnmarshalOptions.Unmarshal(goValue.Interface().(datatypes.JSON), &structList)
	if err != nil {
		return protoreflect.Value{}, err
	}

	//n := len(structList.Values)
	//values := make([]protoreflect.Value, n)
	//for i := 0; i < n; i++ {
	//	values[n] = structList.Values[i].AsInterface()
	//}
	panic("TODO: structpb is lossy with int64's!")
}
