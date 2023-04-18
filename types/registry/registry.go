package registry

import (
	"sync"

	"github.com/cosmos/gogoproto/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"cosmossdk.io/x/tx/signing"
)

var (
	once           sync.Once
	mergedRegistry *protoregistry.Files
	_              signing.ProtoFileResolver = lazyProtoRegistry{}
)

// lazyProtoRegistry is a lazy loading wrapper around the global protobuf registry.
type lazyProtoRegistry struct{}

func (l lazyProtoRegistry) getRegistry() (*protoregistry.Files, error) {
	var err error
	once.Do(func() {
		mergedRegistry, err = proto.MergedRegistry()
	})
	return mergedRegistry, err
}

func (l lazyProtoRegistry) FindFileByPath(s string) (protoreflect.FileDescriptor, error) {
	reg, err := l.getRegistry()
	if err != nil {
		return nil, err
	}
	return reg.FindFileByPath(s)
}

func (l lazyProtoRegistry) FindDescriptorByName(name protoreflect.FullName) (protoreflect.Descriptor, error) {
	reg, err := l.getRegistry()
	if err != nil {
		return nil, err
	}
	return reg.FindDescriptorByName(name)
}

func (l lazyProtoRegistry) RangeFiles(f func(protoreflect.FileDescriptor) bool) {
	reg, err := l.getRegistry()
	if err != nil {
		panic(err)
	}
	reg.RangeFiles(f)
}

// MergedProtoRegistry returns a lazy loading wrapper around the global protobuf registry, a merged registry
// containing both gogo proto and pulsar types.
func MergedProtoRegistry() signing.ProtoFileResolver {
	return lazyProtoRegistry{}
}
