package types

import (
	"fmt"
	"reflect"

	"github.com/gogo/protobuf/jsonpb"

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
	jsonpb.AnyResolver

	// RegisterInterface associates protoName as the public name for the
	// interface passed in as iface. This is to be used primarily to create
	// a public facing registry of interface implementations for clients.
	// protoName should be a well-chosen public facing name that remains stable.
	// RegisterInterface takes an optional list of impls to be registered
	// as implementations of iface.
	//
	// Ex:
	//   registry.RegisterInterface("cosmos.base.v1beta1.Msg", (*sdk.Msg)(nil))
	RegisterInterface(protoName string, iface interface{}, impls ...proto.Message)

	// RegisterImplementations registers impls as concrete implementations of
	// the interface iface.
	//
	// Ex:
	//  registry.RegisterImplementations((*sdk.Msg)(nil), &MsgSend{}, &MsgMultiSend{})
	RegisterImplementations(iface interface{}, impls ...proto.Message)

	// ListAllInterfaces list the type URLs of all registered interfaces.
	ListAllInterfaces() []string

	// ListImplementations lists the valid type URLs for the given interface name that can be used
	// for the provided interface type URL.
	ListImplementations(ifaceTypeURL string) []string
}

// UnpackInterfacesMessage is meant to extend protobuf types (which implement
// proto.Message) to support a post-deserialization phase which unpacks
// types packed within Any's using the whitelist provided by AnyUnpacker
type UnpackInterfacesMessage interface {
	// UnpackInterfaces is implemented in order to unpack values packed within
	// Any's using the AnyUnpacker. It should generally be implemented as
	// follows:
	//   func (s *MyStruct) UnpackInterfaces(unpacker AnyUnpacker) error {
	//		var x AnyInterface
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
	typeURLMap     map[string]reflect.Type
}

type interfaceMap = map[string]reflect.Type

// NewInterfaceRegistry returns a new InterfaceRegistry
func NewInterfaceRegistry() InterfaceRegistry {
	return &interfaceRegistry{
		interfaceNames: map[string]reflect.Type{},
		interfaceImpls: map[reflect.Type]interfaceMap{},
		typeURLMap:     map[string]reflect.Type{},
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

// RegisterImplementations registers a concrete proto Message which implements
// the given interface.
//
// This function PANICs if different concrete types are registered under the
// same typeURL.
func (registry *interfaceRegistry) RegisterImplementations(iface interface{}, impls ...proto.Message) {
	for _, impl := range impls {
		typeURL := "/" + proto.MessageName(impl)
		registry.registerImpl(iface, typeURL, impl)
	}
}

// RegisterCustomTypeURL registers a concrete type which implements the given
// interface under `typeURL`.
//
// This function PANICs if different concrete types are registered under the
// same typeURL.
func (registry *interfaceRegistry) RegisterCustomTypeURL(iface interface{}, typeURL string, impl proto.Message) {
	registry.registerImpl(iface, typeURL, impl)
}

// registerImpl registers a concrete type which implements the given
// interface under `typeURL`.
//
// This function PANICs if different concrete types are registered under the
// same typeURL.
func (registry *interfaceRegistry) registerImpl(iface interface{}, typeURL string, impl proto.Message) {
	ityp := reflect.TypeOf(iface).Elem()
	imap, found := registry.interfaceImpls[ityp]
	if !found {
		imap = map[string]reflect.Type{}
	}

	implType := reflect.TypeOf(impl)
	if !implType.AssignableTo(ityp) {
		panic(fmt.Errorf("type %T doesn't actually implement interface %+v", impl, ityp))
	}

	// Check if we already registered something under the given typeURL. It's
	// okay to register the same concrete type again, but if we are registering
	// a new concrete type under the same typeURL, then we throw an error (here,
	// we panic).
	foundImplType, found := imap[typeURL]
	if found && foundImplType != implType {
		panic(
			fmt.Errorf(
				"concrete type %s has already been registered under typeURL %s, cannot register %s under same typeURL. "+
					"This usually means that there are conflicting modules registering different concrete types "+
					"for a same interface implementation",
				foundImplType,
				typeURL,
				implType,
			),
		)
	}

	imap[typeURL] = implType
	registry.typeURLMap[typeURL] = implType

	registry.interfaceImpls[ityp] = imap
}

func (registry *interfaceRegistry) ListAllInterfaces() []string {
	interfaceNames := registry.interfaceNames
	keys := make([]string, 0, len(interfaceNames))
	for key := range interfaceNames {
		keys = append(keys, key)
	}
	return keys
}

func (registry *interfaceRegistry) ListImplementations(ifaceName string) []string {
	typ, ok := registry.interfaceNames[ifaceName]
	if !ok {
		return []string{}
	}

	impls, ok := registry.interfaceImpls[typ.Elem()]
	if !ok {
		return []string{}
	}

	keys := make([]string, 0, len(impls))
	for key := range impls {
		keys = append(keys, key)
	}
	return keys
}

func (registry *interfaceRegistry) UnpackAny(any *Any, iface interface{}) error {
	// here we gracefully handle the case in which `any` itself is `nil`, which may occur in message decoding
	if any == nil {
		return nil
	}

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

// Resolve returns the proto message given its typeURL. It works with types
// registered with RegisterInterface/RegisterImplementations, as well as those
// registered with RegisterWithCustomTypeURL.
func (registry *interfaceRegistry) Resolve(typeURL string) (proto.Message, error) {
	typ, found := registry.typeURLMap[typeURL]
	if !found {
		return nil, fmt.Errorf("unable to resolve type URL %s", typeURL)
	}

	msg, ok := reflect.New(typ.Elem()).Interface().(proto.Message)
	if !ok {
		return nil, fmt.Errorf("can't resolve type URL %s", typeURL)
	}

	return msg, nil
}

// UnpackInterfaces is a convenience function that calls UnpackInterfaces
// on x if x implements UnpackInterfacesMessage
func UnpackInterfaces(x interface{}, unpacker AnyUnpacker) error {
	if msg, ok := x.(UnpackInterfacesMessage); ok {
		return msg.UnpackInterfaces(unpacker)
	}
	return nil
}
