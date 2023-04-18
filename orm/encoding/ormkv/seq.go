package ormkv

import (
	"bytes"
	"encoding/binary"

	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/orm/types/ormerrors"
)

// SeqCodec is the codec for auto-incrementing uint64 primary key sequences.
type SeqCodec struct {
	messageType protoreflect.FullName
	prefix      []byte
}

// NewSeqCodec creates a new SeqCodec.
func NewSeqCodec(messageType protoreflect.MessageType, prefix []byte) *SeqCodec {
	return &SeqCodec{messageType: messageType.Descriptor().FullName(), prefix: prefix}
}

var _ EntryCodec = &SeqCodec{}

func (s SeqCodec) DecodeEntry(k, v []byte) (Entry, error) {
	if !bytes.Equal(k, s.prefix) {
		return nil, ormerrors.UnexpectedDecodePrefix
	}

	x, err := s.DecodeValue(v)
	if err != nil {
		return nil, err
	}

	return &SeqEntry{
		TableName: s.messageType,
		Value:     x,
	}, nil
}

func (s SeqCodec) EncodeEntry(entry Entry) (k, v []byte, err error) {
	seqEntry, ok := entry.(*SeqEntry)
	if !ok {
		return nil, nil, ormerrors.BadDecodeEntry
	}

	if seqEntry.TableName != s.messageType {
		return nil, nil, ormerrors.BadDecodeEntry
	}

	return s.prefix, s.EncodeValue(seqEntry.Value), nil
}

func (s SeqCodec) Prefix() []byte {
	return s.prefix
}

func (s SeqCodec) EncodeValue(seq uint64) (v []byte) {
	bz := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(bz, seq)
	return bz[:n]
}

func (s SeqCodec) DecodeValue(v []byte) (uint64, error) {
	if len(v) == 0 {
		return 0, nil
	}
	return binary.ReadUvarint(bytes.NewReader(v))
}
