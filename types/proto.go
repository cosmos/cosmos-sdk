package types

import (
	"sync"

	"github.com/cosmos/gogoproto/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	txsigning "cosmossdk.io/x/tx/signing"
)

// CustomProtobufType defines the interface custom gogo proto types must implement
// in order to be used as a "customtype" extension.
//
// ref: https://github.com/cosmos/gogoproto/blob/master/custom_types.md
type CustomProtobufType interface {
	Marshal() ([]byte, error)
	MarshalTo(data []byte) (n int, err error)
	Unmarshal(data []byte) error
	Size() int

	MarshalJSON() ([]byte, error)
	UnmarshalJSON(data []byte) error
}

var (
	mu             sync.RWMutex
	mergedRegistry *protoregistry.Files
	_              txsigning.ProtoFileResolver = lazyProtoRegistry{}
)

type lazyProtoRegistry struct{}

func (l lazyProtoRegistry) init() error {
	mu.Lock()
	defer mu.Unlock()

	if mergedRegistry != nil {
		return nil
	}

	var err error
	mergedRegistry, err = proto.MergedRegistry()
	if err != nil {
		return err
	}

	return nil
}

func (l lazyProtoRegistry) FindFileByPath(s string) (protoreflect.FileDescriptor, error) {
	mu.RLock()
	defer mu.RUnlock()
	if mergedRegistry == nil {
		if err := l.init(); err != nil {
			return nil, err
		}
	}
	return mergedRegistry.FindFileByPath(s)
}

func (l lazyProtoRegistry) FindDescriptorByName(name protoreflect.FullName) (protoreflect.Descriptor, error) {
	mu.RLock()
	defer mu.RUnlock()
	if mergedRegistry == nil {
		if err := l.init(); err != nil {
			return nil, err
		}
	}
	return mergedRegistry.FindDescriptorByName(name)
}

func (l lazyProtoRegistry) RangeFiles(f func(protoreflect.FileDescriptor) bool) {
	mu.RLock()
	defer mu.RUnlock()
	if mergedRegistry == nil {
		if err := l.init(); err != nil {
			panic(err)
		}
	}
	mergedRegistry.RangeFiles(f)
}

func MergedProtoRegistry() txsigning.ProtoFileResolver {
	return lazyProtoRegistry{}
}
