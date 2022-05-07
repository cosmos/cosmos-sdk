package internal

import (
	"reflect"

	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/container"
)

var ModuleRegistry = map[protoreflect.FullName]*ModuleInitializer{}

type ModuleInitializer struct {
	ConfigGoType    reflect.Type
	ConfigProtoType protoreflect.MessageType
	Error           error
	Providers       []container.ProviderDescriptor
}
