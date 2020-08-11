package types

import (
	"fmt"
	"reflect"

	"github.com/gogo/protobuf/proto"
)

// AnyUnpacker is an interface which allows safely unpacking types packed
// in Any's against a whitelist of registered types
type AnyUnpacker interface {
	// UnpackAny unpacks the value in any to the interface pointer passed in as
	// iface. Note that the type in any must have been registered in the
	// underlying whitelist registry as a concrete type for that interface
	// Ex:
	//    var msg sdk.Msg
	//    err := cdc.UnpackAny(any, &msg)
	//    ...
	UnpackAny(any *Any, iface interface{}) error
}

// InterfaceRegistry provides a mechanism for registering interfaces and
// implementations that can be safely unpacked from Any
type InterfaceRegistry interface {
	AnyUnpacker

	// RegisterInterface associates protoName as the public name for the
	// interface passed in as iface. This is to be used primarily to create
	// a public facing registry of interface implementations for clients.
	// protoName should be a well-chosen public facing name that remains stable.
	// RegisterInterface takes an optional list of impls to be registered
	// as implementations of iface.
	//
	// Ex:
	//   registry.RegisterInterface("cosmos.v1beta1.Msg", (*sdk.Msg)(nil))
	RegisterInterface(protoName string, iface interface{}, impls ...proto.Message)

	// RegisterImplementations registers impls as concrete implementations of
	// the interface iface.
	//
	// Ex:
	//  registry.RegisterImplementations((*sdk.Msg)(nil), &MsgSend{}, &MsgMultiSend{})
	RegisterImplementations(iface interface{}, impls ...proto.Message)
}

// UnpackInterfacesMessage is meant to extend protobuf types (which implement
// proto.Message) to support a post-deserialization phase which unpacks
// types packed within Any's using the whitelist provided by AnyUnpacker
type UnpackInterfacesMessage interface {
	// UnpackInterfaces is implemented in order to unpack values packed within
	// Any's using the AnyUnpacker. It should generally be implemented as
	// follows:
	//   func (s *MyStruct) UnpackInterfaces(unpacker AnyUnpacker) error {
	//		var x AnInterface
	//		// where X is an Any field on MyStruct
	//		err := unpacker.UnpackAny(s.X, &x)
	//		if err != nil {
	//			return nil
	//		}
	//		// where Y is a field on MyStruct that implements UnpackInterfacesMessage itself
	//		err = s.Y.UnpackInterfaces(unpacker)
	//		if err != nil {
	//			return nil
	//		}
	//		return nil
	//	 }
	UnpackInterfaces(unpacker AnyUnpacker) error
}

type interfaceRegistry struct {
	interfaceNames map[string]reflect.Type
	interfaceImpls map[reflect.Type]interfaceMap
}

type interfaceMap = map[string]reflect.Type

// NewInterfaceRegistry returns a new InterfaceRegistry
func NewInterfaceRegistry() InterfaceRegistry {
	return &interfaceRegistry{
		interfaceNames: map[string]reflect.Type{},
		interfaceImpls: map[reflect.Type]interfaceMap{},
	}
}

func (registry *interfaceRegistry) RegisterInterface(protoName string, iface interface{}, impls ...proto.Message) {
	typ := reflect.TypeOf(iface)
	if typ.Elem().Kind() != reflect.Interface {
		panic(fmt.Errorf("%T is not an interface type", iface))
	}
	registry.interfaceNames[protoName] = typ
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
			panic(fmt.Errorf("type %T doesn't actually implement interface %+v", impl, ityp))
		}

		imap["/"+proto.MessageName(impl)] = implType
	}

	registry.interfaceImpls[ityp] = imap
}

func (registry *interfaceRegistry) UnpackAny(any *Any, iface interface{}) error {
	if any.TypeUrl == "" {
		// if TypeUrl is empty return nil because without it we can't actually unpack anything
		return nil
	}

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
		return fmt.Errorf("no registered implementations of type %+v", rt)
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

// UnpackInterfaces is a convenience function that calls UnpackInterfaces
// on x if x implements UnpackInterfacesMessage
func UnpackInterfaces(x interface{}, unpacker AnyUnpacker) error {
	if msg, ok := x.(UnpackInterfacesMessage); ok {
		return msg.UnpackInterfaces(unpacker)
	}
	return nil
}
