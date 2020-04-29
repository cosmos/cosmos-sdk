package types

import (
	"fmt"
	"reflect"

	"github.com/gogo/protobuf/proto"
)

type Any struct {
	// A URL/resource name that uniquely identifies the type of the serialized
	// protocol buffer message. This string must contain at least
	// one "/" character. The last segment of the URL's path must represent
	// the fully qualified name of the type (as in
	// `path/google.protobuf.Duration`). The name should be in a canonical form
	// (e.g., leading "." is not accepted).
	//
	// In practice, teams usually precompile into the binary all types that they
	// expect it to use in the context of Any. However, for URLs which use the
	// scheme `http`, `https`, or no scheme, one can optionally set up a type
	// server that maps type URLs to message definitions as follows:
	//
	// * If no scheme is provided, `https` is assumed.
	// * An HTTP GET on the URL must yield a [google.protobuf.Type][]
	//   value in binary format, or produce an error.
	// * Applications are allowed to cache lookup results based on the
	//   URL, or have them precompiled into a binary to avoid any
	//   lookup. Therefore, binary compatibility needs to be preserved
	//   on changes to types. (Use versioned type names to manage
	//   breaking changes.)
	//
	// Note: this functionality is not currently available in the official
	// protobuf release, and it is not used for type URLs beginning with
	// type.googleapis.com.
	//
	// Schemes other than `http`, `https` (or the empty scheme) might be
	// used with implementation specific semantics.
	//
	TypeUrl string `protobuf:"bytes,1,opt,name=type_url,json=typeUrl,proto3" json:"type_url,omitempty"`
	// Must be a valid serialized protocol buffer of the above specified type.
	Value                []byte   `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
	cachedValue          interface{}
}

func NewAnyWithValue(value interface{}) (*Any, error) {
	any := &Any{}
	msg, ok := value.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("can't pack %T", msg)
	}
	err := any.Pack(msg)
	if err != nil {
		return nil, err
	}
	return any, nil
}

type interfaceMap = map[string]reflect.Type

type interfaceRegistry struct {
	interfaceNames map[string]reflect.Type
	interfaceImpls map[reflect.Type]interfaceMap
}

func NewInterfaceRegistry() InterfaceRegistry {
	return &interfaceRegistry{
		interfaceNames: map[string]reflect.Type{},
		interfaceImpls: map[reflect.Type]interfaceMap{},
	}
}

type InterfaceRegistry interface {
	AnyUnpacker

	RegisterInterface(protoName string, iface interface{}, impls ...proto.Message)
	RegisterImplementations(iface interface{}, impls ...proto.Message)
}

type AnyUnpacker interface {
	UnpackAny(any *Any, iface interface{}) error
}

func (a *interfaceRegistry) RegisterInterface(protoName string, iface interface{}, impls ...proto.Message) {
	a.interfaceNames[protoName] = reflect.TypeOf(iface)
	a.RegisterImplementations(iface, impls...)
}

func (a *interfaceRegistry) RegisterImplementations(iface interface{}, impls ...proto.Message) {
	ityp := reflect.TypeOf(iface).Elem()
	imap, found := a.interfaceImpls[ityp]
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
	a.interfaceImpls[ityp] = imap
}

func (any *Any) Pack(x proto.Message) error {
	any.TypeUrl = "/" + proto.MessageName(x)
	bz, err := proto.Marshal(x)
	if err != nil {
		return err
	}
	any.Value = bz
	any.cachedValue = x
	return nil
}

func (any *Any) CachedValue() interface{} {
	return any.cachedValue
}

func (ctx *interfaceRegistry) UnpackAny(any *Any, iface interface{}) error {
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
	imap, found := ctx.interfaceImpls[rt]
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
	err = UnpackInterfaces(msg, ctx)
	if err != nil {
		return err
	}
	rv.Elem().Set(reflect.ValueOf(msg))
	any.cachedValue = msg
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

func UnmarshalAny(ctx AnyUnpacker, iface interface{}, bz []byte) error {
	any := &Any{}
	err := any.Unmarshal(bz)
	if err != nil {
		return err
	}
	return ctx.UnpackAny(any, iface)
}

type UnpackInterfacesMsg interface {
	UnpackInterfaces(ctx AnyUnpacker) error
}

func UnpackInterfaces(x interface{}, ctx AnyUnpacker) error {
	if msg, ok := x.(UnpackInterfacesMsg); ok {
		return msg.UnpackInterfaces(ctx)
	}
	return nil
}
