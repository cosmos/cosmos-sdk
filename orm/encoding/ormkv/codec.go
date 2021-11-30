package ormkv

import "google.golang.org/protobuf/reflect/protoreflect"

type EntryCodec interface {
	DecodeEntry(k, v []byte) (Entry, error)
	EncodeEntry(entry Entry) (k, v []byte, err error)
}

type IndexCodec interface {
	EntryCodec
	DecodeIndexKey(k, v []byte) (indexFields, primaryKey []protoreflect.Value, err error)
	EncodeKVFromMessage(message protoreflect.Message) (k, v []byte, err error)
}
