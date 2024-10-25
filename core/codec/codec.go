package codec

import "cosmossdk.io/core/transaction"

// Codec defines a Binary Codec and JSON Codec for modules to encode and decode data.
type Codec interface {
	BinaryCodec
	JSONCodec
}

// BinaryCodec defines a binary encoding and decoding interface for modules to encode and decode data.
type BinaryCodec interface {
	Marshal(transaction.Msg) ([]byte, error)
	Unmarshal([]byte, transaction.Msg) error
}

// JSONCodec defines a JSON encoding and decoding interface for modules to encode and decode data.
type JSONCodec interface {
	MarshalJSON(transaction.Msg) ([]byte, error)
	UnmarshalJSON([]byte, transaction.Msg) error
}
