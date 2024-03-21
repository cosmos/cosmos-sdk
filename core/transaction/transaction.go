package transaction

import gogoproto "github.com/cosmos/gogoproto/proto"

type (
	Type     = gogoproto.Message
	Identity = []byte
)

// Codec defines the TX codec, which converts a TX from bytes to its concrete representation.
type Codec[T Tx] interface {
	// Decode decodes the tx bytes into a DecodedTx, containing
	// both concrete and bytes representation of the tx.
	Decode([]byte) (T, error)
}

// Message represents a transaction message.
type Message struct {
	Type   Type     // The type of the message.
	Signer Identity // The identity of the signer.
}

type Tx interface {
	// Hash returns the unique identifier for the Tx.
	Hash() [32]byte
	// GetMessages returns the list of state transitions of the Tx.
	GetMessages() []Message
	// GetGasLimit returns the gas limit of the tx. Must return math.MaxUint64 for infinite gas
	// txs.
	GetGasLimit() uint64
	// Bytes returns the encoded version of this tx. Note: this is ideally cached
	// from the first instance of the decoding of the tx.
	Bytes() []byte
}
