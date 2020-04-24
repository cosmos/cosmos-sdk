package types

import (
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"reflect"
)

type interfaceMap = map[string]reflect.Type

type anyContext struct {
	interfaceNames map[string]reflect.Type
	interfaceImpls map[reflect.Type]interfaceMap
}

func NewAnyContext() AnyContext {
	return &anyContext{
		interfaceNames: map[string]reflect.Type{},
		interfaceImpls: map[reflect.Type]interfaceMap{},
	}
}

type AnyContext interface {
	RegisterInterface(protoName string, ptr interface{})
	RegisterImplementation(iface interface{}, impl proto.Message)
	UnpackAny(any *Any, iface interface{}) (interface{}, error)
}

func (a *anyContext) RegisterInterface(protoName string, ptr interface{}) {
	a.interfaceNames[protoName] = reflect.TypeOf(ptr)
}

func (a *anyContext) RegisterImplementation(iface interface{}, impl proto.Message) {
	ityp := reflect.TypeOf(iface)
	imap, found := a.interfaceImpls[ityp]
	if !found {
		imap = map[string]reflect.Type{}
	}
	imap["/"+proto.MessageName(impl)] = reflect.TypeOf(impl)
	a.interfaceImpls[ityp] = imap
}

func (any *Any) Pack(x proto.Message) error {
	any.TypeUrl = "/" + proto.MessageName(x)
	bz, err := proto.Marshal(x)
	if err != nil {
		return err
	}
	any.Value = bz
	return nil
}

func (ctx *anyContext) UnpackAny(any *Any, iface interface{}) (interface{}, error) {
	imap, found := ctx.interfaceImpls[reflect.TypeOf(iface)]
	if !found {
		return nil, fmt.Errorf("no registered implementations of interface type %T", iface)
	}
	typ, found := imap[any.TypeUrl]
	if !found {
		return nil, fmt.Errorf("no concrete type registered for type URL %s against interface %T", any.TypeUrl, iface)
	}
	msg, ok := reflect.New(typ).Interface().(proto.Message)
	if !ok {
		return nil, fmt.Errorf("can't proto unmarshal %T", msg)
	}
	err := proto.Unmarshal(any.Value, msg)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (any Any) UnsafeUnpack() (interface{}, error) {
	gogoAny := &types.Any{TypeUrl: any.TypeUrl}
	msg, err := types.EmptyAny(gogoAny)
	if err != nil {
		return nil, err
	}
	err = proto.Unmarshal(any.Value, msg)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func MarshalAny(x interface{}) ([]byte, error) {
	msg, ok := x.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("can't proto marshal %T", x)
	}
	any := &Any{}
	err := any.Pack(msg)
	if err != nil {
		return nil, err
	}
	return any.Marshal()
}

func UnmarshalAny(ctx AnyContext, iface interface{}, bz []byte) (interface{}, error) {
	any := &Any{}
	err := any.Unmarshal(bz)
	if err != nil {
		return nil, err
	}
	return ctx.UnpackAny(any, iface)
}
