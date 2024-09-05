package ormkv

import (
	"bytes"
	"errors"
	"io"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/orm/types/ormerrors"
)

// PrimaryKeyCodec is the codec for primary keys.
type PrimaryKeyCodec struct {
	*KeyCodec
	unmarshalOptions proto.UnmarshalOptions
}

var _ IndexCodec = &PrimaryKeyCodec{}

// NewPrimaryKeyCodec creates a new PrimaryKeyCodec for the provided msg and
// fields, with an optional prefix and unmarshal options.
func NewPrimaryKeyCodec(prefix []byte, msgType protoreflect.MessageType, fieldNames []protoreflect.Name, unmarshalOptions proto.UnmarshalOptions) (*PrimaryKeyCodec, error) {
	keyCodec, err := NewKeyCodec(prefix, msgType, fieldNames)
	if err != nil {
		return nil, err
	}

	return &PrimaryKeyCodec{
		KeyCodec:         keyCodec,
		unmarshalOptions: unmarshalOptions,
	}, nil
}

var _ IndexCodec = PrimaryKeyCodec{}

func (p PrimaryKeyCodec) DecodeIndexKey(k, _ []byte) (indexFields, primaryKey []protoreflect.Value, err error) {
	indexFields, err = p.DecodeKey(bytes.NewReader(k))

	// got prefix key
	if errors.Is(err, io.EOF) {
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
	values, err := p.DecodeKey(bytes.NewReader(k))
	if errors.Is(err, io.EOF) {
		return &PrimaryKeyEntry{
			TableName: p.messageType.Descriptor().FullName(),
			Key:       values,
		}, nil
	} else if err != nil {
		return nil, err
	}

	msg := p.messageType.New().Interface()
	err = p.Unmarshal(values, v, msg)

	return &PrimaryKeyEntry{
		TableName: p.messageType.Descriptor().FullName(),
		Key:       values,
		Value:     msg,
	}, err
}

func (p PrimaryKeyCodec) EncodeEntry(entry Entry) (k, v []byte, err error) {
	pkEntry, ok := entry.(*PrimaryKeyEntry)
	if !ok {
		return nil, nil, ormerrors.BadDecodeEntry.Wrapf("expected %T, got %T", &PrimaryKeyEntry{}, entry)
	}

	if pkEntry.TableName != p.messageType.Descriptor().FullName() {
		return nil, nil, ormerrors.BadDecodeEntry.Wrapf(
			"wrong table name, got %s, expected %s",
			pkEntry.TableName,
			p.messageType.Descriptor().FullName(),
		)
	}

	k, err = p.KeyCodec.EncodeKey(pkEntry.Key)
	if err != nil {
		return nil, nil, err
	}

	v, err = p.marshal(pkEntry.Key, pkEntry.Value)
	return k, v, err
}

func (p PrimaryKeyCodec) marshal(key []protoreflect.Value, message proto.Message) (v []byte, err error) {
	// first clear the priamry key values because these are already stored in
	// the key so we don't need to store them again in the value
	p.ClearValues(message.ProtoReflect())

	v, err = proto.MarshalOptions{Deterministic: true}.Marshal(message)
	if err != nil {
		return nil, err
	}

	// set the primary key values again returning the message to its original state
	p.SetKeyValues(message.ProtoReflect(), key)

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
	p.SetKeyValues(message.ProtoReflect(), key)
	return nil
}

func (p PrimaryKeyCodec) EncodeKVFromMessage(message protoreflect.Message) (k, v []byte, err error) {
	ks, k, err := p.KeyCodec.EncodeKeyFromMessage(message)
	if err != nil {
		return nil, nil, err
	}

	v, err = p.marshal(ks, message.Interface())
	return k, v, err
}
