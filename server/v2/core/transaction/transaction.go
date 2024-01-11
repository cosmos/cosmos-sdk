package transaction

import (
	"google.golang.org/protobuf/proto"
)

type (
	Type     = proto.Message
	Identity = []byte
)

// Codec defines the TX codec, which converts a TX from bytes to its concrete representation.
type Codec[T Tx] interface {
	// Decode decodes the tx bytes into a DecodedTx, containing
	// both concrete and bytes representation of the tx.
	Decode([]byte) (T, error)
}

type Tx interface {
	Hash() [32]byte // TODO evaluate if 32 bytes is the right size & benchmark overhead of hashing instead of using identifier
	GetMessages() []Type
	GetSenders() []Identity // TODO reduce this to a single identity if accepted
	GetGasLimit() uint64
}
