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
	// Hash returns the unique identifier for the Tx.
	Hash() [32]byte // TODO evaluate if 32 bytes is the right size & benchmark overhead of hashing instead of using identifier
	// GetMessages returns the list of state transitions of the Tx.
	GetMessages() []Type
	// GetSenders returns the tx state transition sender.
	GetSenders() []Identity // TODO reduce this to a single identity if accepted
	// GetGasLimit returns the gas limit of the tx. Must return math.MaxUint64 for infinite gas
	// txs.
	GetGasLimit() uint64
	// Bytes returns the encoded version of this tx. Note: this is ideally cached
	// from the first instance of the decoding of the tx.
	Bytes() []byte
}
