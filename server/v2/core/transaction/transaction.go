package transaction

import (
	"context"

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

// Validator is a transaction validator that validates transactions based off an existing set of handlers
// Validators can be designed to be asynchronous or synchronous
type Validator[T Tx] interface {
	// Validate validates the transactions
	// it returns the context used and a map of which txs failed.
	// It does not take into account what information is needed to be returned to the consensus engine, this must be extracted from the context
	Validate(ctx context.Context, txs []T) (context.Context, map[[32]byte]error)
}
