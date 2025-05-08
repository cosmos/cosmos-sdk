package corecompat

// Codec defines a Binary Codec and JSON Codec for modules to encode and decode data.
type Codec interface {
	BinaryCodec
	JSONCodec
}

// BinaryCodec defines a binary encoding and decoding interface for modules to encode and decode data.
type BinaryCodec interface {
	Marshal(ProtoMsg) ([]byte, error)
	Unmarshal([]byte, ProtoMsg) error
}

// JSONCodec defines a JSON encoding and decoding interface for modules to encode and decode data.
type JSONCodec interface {
	MarshalJSON(ProtoMsg) ([]byte, error)
	UnmarshalJSON([]byte, ProtoMsg) error
}

// ProtoMsg defines the legacy golang proto message interface.
type ProtoMsg = interface {
	Reset()
	String() string
	ProtoMessage()
}
