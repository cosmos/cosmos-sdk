package types

import (
	"sync"

	"github.com/cosmos/gogoproto/proto"
	"google.golang.org/protobuf/reflect/protoregistry"
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
	mu             sync.Mutex
	mergedRegistry *protoregistry.Files
)

func MergedProtoRegistry() *protoregistry.Files {
	if mergedRegistry != nil {
		return mergedRegistry
	}

	mu.Lock()
	defer mu.Unlock()

	var err error
	mergedRegistry, err = proto.MergedRegistry()
	if err != nil {
		panic(err)
	}

	return mergedRegistry
}
