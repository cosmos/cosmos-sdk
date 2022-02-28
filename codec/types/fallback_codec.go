package types

import (
	"errors"
	"github.com/gogo/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var _ InterfaceRegistry = FallbackInterfaceRegistry{}

// FallbackInterfaceRegistry is used by the fallback codec
// in case Context's Codec is not set.
type FallbackInterfaceRegistry struct{}

func (f FallbackInterfaceRegistry) isInterfaceRegistry() {
	panic("implement me")
}

func (f FallbackInterfaceRegistry) FindExtensionByName(field protoreflect.FullName) (protoreflect.ExtensionType, error) {
	panic("implement me")
}

func (f FallbackInterfaceRegistry) FindExtensionByNumber(message protoreflect.FullName, field protoreflect.FieldNumber) (protoreflect.ExtensionType, error) {
	panic("implement me")
}

func (f FallbackInterfaceRegistry) FindMessageByName(message protoreflect.FullName) (protoreflect.MessageType, error) {
	panic("implement me")
}

func (f FallbackInterfaceRegistry) FindMessageByURL(url string) (protoreflect.MessageType, error) {
	panic("implement me")
}

// errCodecNotSet is return by FallbackInterfaceRegistry in case there are attempt to decode
// or encode a type which contains an interface field.
var errCodecNotSet = errors.New("client: cannot encode or decode type which requires the application specific codec")

func (f FallbackInterfaceRegistry) UnpackAny(any *Any, iface interface{}) error {
	return errCodecNotSet
}

func (f FallbackInterfaceRegistry) Resolve(typeUrl string) (proto.Message, error) {
	return nil, errCodecNotSet
}

func (f FallbackInterfaceRegistry) RegisterInterface(protoName string, iface interface{}, impls ...interface{}) {
	panic("cannot be called")
}

func (f FallbackInterfaceRegistry) RegisterImplementations(iface interface{}, impls ...interface{}) {
	panic("cannot be called")
}

func (f FallbackInterfaceRegistry) ListAllInterfaces() []string {
	panic("cannot be called")
}

func (f FallbackInterfaceRegistry) ListImplementations(ifaceTypeURL string) []string {
	panic("cannot be called")
}
