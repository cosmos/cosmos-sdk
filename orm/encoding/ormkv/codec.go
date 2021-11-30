package ormkv

import "google.golang.org/protobuf/reflect/protoreflect"

type Codec interface {
	DecodeKV(k, v []byte) (Entry, error)
	EncodeKV(entry Entry) (k, v []byte, err error)
}

type IndexCodec interface {
	Codec
	DecodeIndexKey(k, v []byte) (indexFields, primaryKey []protoreflect.Value, err error)
}
