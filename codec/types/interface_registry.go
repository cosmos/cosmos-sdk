package types

import (
	"fmt"
	"reflect"

	"github.com/gogo/protobuf/proto"
)

type AnyUnpacker interface {
	UnpackAny(any *Any, iface interface{}) error
}

type InterfaceRegistry interface {
	AnyUnpacker

	RegisterInterface(protoName string, iface interface{}, impls ...proto.Message)
	RegisterImplementations(iface interface{}, impls ...proto.Message)
}

type UnpackInterfacesMsg interface {
	UnpackInterfaces(ctx AnyUnpacker) error
}

type interfaceRegistry struct {
	interfaceNames map[string]reflect.Type
	interfaceImpls map[reflect.Type]interfaceMap
}

type interfaceMap = map[string]reflect.Type

func NewInterfaceRegistry() InterfaceRegistry {
	return &interfaceRegistry{
		interfaceNames: map[string]reflect.Type{},
		interfaceImpls: map[reflect.Type]interfaceMap{},
	}
}

func (registry *interfaceRegistry) RegisterInterface(protoName string, iface interface{}, impls ...proto.Message) {
	registry.interfaceNames[protoName] = reflect.TypeOf(iface)
	registry.RegisterImplementations(iface, impls...)
}

func (registry *interfaceRegistry) RegisterImplementations(iface interface{}, impls ...proto.Message) {
	ityp := reflect.TypeOf(iface).Elem()
	imap, found := registry.interfaceImpls[ityp]
	if !found {
		imap = map[string]reflect.Type{}
	}
	for _, impl := range impls {
		implType := reflect.TypeOf(impl)
		if !implType.AssignableTo(ityp) {
			panic(fmt.Errorf("type %T doesn't actually implement interface %T", implType, ityp))
		}
		imap["/"+proto.MessageName(impl)] = implType
	}
	registry.interfaceImpls[ityp] = imap
}

func (registry *interfaceRegistry) UnpackAny(any *Any, iface interface{}) error {
	rv := reflect.ValueOf(iface)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("UnpackAny expects a pointer")
	}
	rt := rv.Elem().Type()
	cachedValue := any.cachedValue
	if cachedValue != nil {
		if reflect.TypeOf(cachedValue).AssignableTo(rt) {
			rv.Elem().Set(reflect.ValueOf(cachedValue))
			return nil
		}
	}
	imap, found := registry.interfaceImpls[rt]
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
	err = UnpackInterfaces(msg, registry)
	if err != nil {
		return err
	}
	rv.Elem().Set(reflect.ValueOf(msg))
	any.cachedValue = msg
	return nil
}

func UnpackInterfaces(x interface{}, ctx AnyUnpacker) error {
	if msg, ok := x.(UnpackInterfacesMsg); ok {
		return msg.UnpackInterfaces(ctx)
	}
	return nil
}
