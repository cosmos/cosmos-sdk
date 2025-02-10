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

	// MessageType returns the message type this index codec applies to.
	MessageType() protoreflect.MessageType

	// GetFieldNames returns the field names in the key of this index.
	GetFieldNames() []protoreflect.Name

	// DecodeIndexKey decodes a kv-pair into index-fields and primary-key field
	// values. These fields may or may not overlap depending on the index.
	DecodeIndexKey(k, v []byte) (indexFields, primaryKey []protoreflect.Value, err error)

	// EncodeKVFromMessage encodes a kv-pair for the index from a message.
	EncodeKVFromMessage(message protoreflect.Message) (k, v []byte, err error)

	// CompareKeys compares the provided values which must correspond to the
	// fields in this key. Prefix keys of different lengths are supported but the
	// function will panic if either array is too long. A negative value is returned
	// if values1 is less than values2, 0 is returned if the two arrays are equal,
	// and a positive value is returned if values2 is greater.
	CompareKeys(key1, key2 []protoreflect.Value) int

	// EncodeKeyFromMessage encodes the key part of this index and returns both
	// index values and encoded key.
	EncodeKeyFromMessage(message protoreflect.Message) (keyValues []protoreflect.Value, key []byte, err error)

	// IsFullyOrdered returns true if all fields in the key are also ordered.
	IsFullyOrdered() bool
}
