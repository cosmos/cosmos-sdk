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
	once           sync.Once
	mergedRegistry *protoregistry.Files
	_              txsigning.ProtoFileResolver = lazyProtoRegistry{}
)

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

func MergedProtoRegistry() txsigning.ProtoFileResolver {
	return lazyProtoRegistry{}
}
