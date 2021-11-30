package ormkv

import (
	"bytes"
	"io"

	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"

	"google.golang.org/protobuf/proto"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type PrimaryKeyCodec struct {
	*KeyCodec
	msgType          protoreflect.MessageType
	unmarshalOptions proto.UnmarshalOptions
}

func NewPrimaryKeyCodec(keyCodec *KeyCodec, msgType protoreflect.MessageType, unmarshalOptions proto.UnmarshalOptions) *PrimaryKeyCodec {
	return &PrimaryKeyCodec{
		KeyCodec:         keyCodec,
		msgType:          msgType,
		unmarshalOptions: unmarshalOptions,
	}
}

var _ IndexCodec = PrimaryKeyCodec{}

func (p PrimaryKeyCodec) DecodeIndexKey(k, _ []byte) (indexFields, primaryKey []protoreflect.Value, err error) {
	indexFields, err = p.Decode(bytes.NewReader(k))

	// got prefix key
	if err == io.EOF {
		return indexFields, nil, nil
	} else if err != nil {
		return nil, nil, err
	}

	if len(indexFields) == len(p.fieldCodecs) {
		// for primary keys the index fields are the primary key
		// but only if we don't have a prefix key
		primaryKey = indexFields
	}
	return indexFields, primaryKey, nil

}

func (p PrimaryKeyCodec) DecodeEntry(k, v []byte) (Entry, error) {
	values, err := p.Decode(bytes.NewReader(k))
	if err != nil {
		return nil, err
	}

	msg := p.msgType.New().Interface()
	err = p.unmarshalOptions.Unmarshal(v, msg)
	if err != nil {
		return nil, err
	}

	return PrimaryKeyEntry{
		Key:   values,
		Value: msg,
	}, nil
}

func (p PrimaryKeyCodec) EncodeEntry(entry Entry) (k, v []byte, err error) {
	pkEntry, ok := entry.(PrimaryKeyEntry)
	if !ok {
		return nil, nil, ormerrors.BadDecodeEntry
	}

	if pkEntry.Value.ProtoReflect().Descriptor().FullName() != p.msgType.Descriptor().FullName() {
		return nil, nil, ormerrors.BadDecodeEntry
	}

	k, err = p.KeyCodec.Encode(pkEntry.Key)
	if err != nil {
		return nil, nil, err
	}

	v, err = p.marshal(pkEntry.Value)
	return k, v, err
}

func (p PrimaryKeyCodec) marshal(message proto.Message) (v []byte, err error) {
	v, err = proto.MarshalOptions{Deterministic: true}.Marshal(message)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (p *PrimaryKeyCodec) ClearValues(message protoreflect.Message) {
	for _, f := range p.fieldDescriptors {
		message.Clear(f)
	}
}

func (p *PrimaryKeyCodec) Unmarshal(key []protoreflect.Value, value []byte, message proto.Message) error {
	err := p.unmarshalOptions.Unmarshal(value, message)
	if err != nil {
		return err
	}

	// rehydrate primary key
	p.SetValues(message.ProtoReflect(), key)
	return nil
}

func (p PrimaryKeyCodec) EncodeKVFromMessage(message protoreflect.Message) (k, v []byte, err error) {
	_, k, err = p.KeyCodec.EncodeFromMessage(message)
	if err != nil {
		return nil, nil, err
	}

	v, err = p.marshal(message.Interface())
	return k, v, err
}
