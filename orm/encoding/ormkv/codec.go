package ormkv

import "google.golang.org/protobuf/reflect/protoreflect"

// EntryCodec defines an interfaces for decoding and encoding entries in the
// kv-store backing an ORM instance. EntryCodec's enable full logical decoding
// of ORM data.
type EntryCodec interface {

	// DecodeEntry decodes a kv-pair into an Entry.
	DecodeEntry(k, v []byte) (Entry, error)

	// EncodeEntry encodes an entry into a kv-pair.
	EncodeEntry(entry Entry) (k, v []byte, err error)
}

// IndexCodec defines an interfaces for encoding and decoding index-keys in the
// kv-store.
type IndexCodec interface {
	EntryCodec

	// DecodeIndexKey decodes a kv-pair into index-fields and primary-key field
	// values. These fields may or may not overlap depending on the index.
	DecodeIndexKey(k, v []byte) (indexFields, primaryKey []protoreflect.Value, err error)

	// EncodeKVFromMessage encodes a kv-pair for the index from a message.
	EncodeKVFromMessage(message protoreflect.Message) (k, v []byte, err error)
}
