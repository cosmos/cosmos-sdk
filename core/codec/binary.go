package codec

import "github.com/cosmos/gogoproto/proto"

type BinaryCodec interface {
	// Marshal returns binary encoding of v.
	Marshal(o proto.Message) ([]byte, error)

	// Unmarshal parses the data encoded with Marshal method and stores the result
	// in the value pointed to by v.
	Unmarshal(bz []byte, ptr proto.Message) error
}
