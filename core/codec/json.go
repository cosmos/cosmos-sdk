package codec

import "github.com/cosmos/gogoproto/proto"

type JSONCodec interface {
	// MarshalJSON returns JSON encoding of v.
	MarshalJSON(o proto.Message) ([]byte, error)

	// UnmarshalJSON parses the data encoded with MarshalJSON method and stores the result
	// in the value pointed to by v.
	UnmarshalJSON(bz []byte, ptr proto.Message) error
}
