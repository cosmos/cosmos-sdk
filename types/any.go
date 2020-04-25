package types

import (
	"fmt"
	"github.com/gogo/protobuf/proto"
	"reflect"
)

type interfaceMap = map[string]reflect.Type

type interfaceContext struct {
	interfaceNames map[string]reflect.Type
	interfaceImpls map[reflect.Type]interfaceMap
}

func NewInterfaceContext() InterfaceContext {
	return &interfaceContext{
		interfaceNames: map[string]reflect.Type{},
		interfaceImpls: map[reflect.Type]interfaceMap{},
	}
}

type InterfaceContext interface {
	RegisterInterface(protoName string, ptr interface{})
	RegisterImplementation(iface interface{}, impl proto.Message)
	UnpackAny(any *Any, iface interface{}) error
}

func (a *interfaceContext) RegisterInterface(protoName string, ptr interface{}) {
	a.interfaceNames[protoName] = reflect.TypeOf(ptr)
}

func (a *interfaceContext) RegisterImplementation(iface interface{}, impl proto.Message) {
	ityp := reflect.TypeOf(iface).Elem()
	imap, found := a.interfaceImpls[ityp]
	if !found {
		imap = map[string]reflect.Type{}
	}
	implType := reflect.TypeOf(impl)
	if !implType.AssignableTo(ityp) {
		panic(fmt.Errorf("type %T doesn't actually implement interface %T", implType, ityp))
	}
	imap["/"+proto.MessageName(impl)] = implType
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

func (ctx *interfaceContext) UnpackAny(any *Any, iface interface{}) error {
	rv := reflect.ValueOf(iface)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("UnpackAny expects a pointer")
	}
	imap, found := ctx.interfaceImpls[rv.Elem().Type()]
	if !found {
		return fmt.Errorf("no registered implementations of interface type %T", iface)
	}
	typ, found := imap[any.TypeUrl]
	if !found {
		return fmt.Errorf("no concrete type registered for type URL %s against interface %T", any.TypeUrl, iface)
	}
	msg, ok := reflect.New(typ.Elem()).Interface().(proto.Message)
	if !ok {
		return fmt.Errorf("can't proto unmarshal %T", msg)
	}
	err := proto.Unmarshal(any.Value, msg)
	if err != nil {
		return err
	}
	rv.Elem().Set(reflect.ValueOf(msg))
	return nil
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

func UnmarshalAny(ctx InterfaceContext, iface interface{}, bz []byte) error {
	any := &Any{}
	err := any.Unmarshal(bz)
	if err != nil {
		return err
	}
	return ctx.UnpackAny(any, iface)
}
