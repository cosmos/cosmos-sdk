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

func UnpackInterfaces(x interface{}, ctx AnyUnpacker) error {
	if msg, ok := x.(UnpackInterfacesMsg); ok {
		return msg.UnpackInterfaces(ctx)
	}
	return nil
}
