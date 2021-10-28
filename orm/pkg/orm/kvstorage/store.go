package kvstorage

import (
	"context"
	"fmt"
	"github.com/cosmos/cosmos-sdk/orm/apis/orm/v1alpha1"
	"github.com/cosmos/cosmos-sdk/orm/pkg/kv"
	"github.com/cosmos/cosmos-sdk/orm/pkg/orm"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func NewStorage(kv kv.KV) Storage {
	return Storage{
		kv: kv,
	}
}

type Storage struct {
	kv           kv.KV
	codecs       map[protoreflect.FullName]*Codec
	typePrefixes map[protoreflect.FullName][]byte
}

func (s Storage) Create(ctx context.Context, object proto.Message) ([]byte, error) {
	fullName := object.ProtoReflect().Descriptor().FullName()

	cdc, exists := s.codecs[fullName]
	if !exists {
		return nil, fmt.Errorf("unregistered state object: %s", object.ProtoReflect().Descriptor().FullName())
	}
	primaryKey, err := cdc.EncodePrimaryKey(object)
	if err != nil {
		return nil, err
	}

	typePrefixedPrimaryKey := append(s.typePrefixes[fullName], primaryKey...)

	if s.kv.Has(ctx, typePrefixedPrimaryKey) {
		return nil, fmt.Errorf("%w: in type prefix %s, %s", orm.ErrAlreadyExists, fullName, typePrefixedPrimaryKey)
	}

	bytes, err := cdc.EncodeObject(object)
	if err != nil {
		return nil, err
	}

	s.kv.Set(ctx, typePrefixedPrimaryKey, bytes)

	return primaryKey, nil
}

func (s Storage) Get(ctx context.Context, id []byte, target proto.Message) error {
	fullName := target.ProtoReflect().Descriptor().FullName()

	cdc, exists := s.codecs[fullName]
	if !exists {
		return fmt.Errorf("unregistered state object: %s", fullName)
	}

	typePrefix := s.typePrefixes[fullName]

	typePrefixedPrimaryKey := append(typePrefix, id...)

	bytes := s.kv.Get(ctx, typePrefixedPrimaryKey)

	return cdc.DecodeObject(id, bytes, target)
}

func (s Storage) Update(ctx context.Context, object proto.Message) ([]byte, error) {
	fullName := object.ProtoReflect().Descriptor().FullName()

	cdc, exists := s.codecs[fullName]
	if !exists {
		return nil, fmt.Errorf("unregistered state object: %s", fullName)
	}

	primaryKey, err := cdc.EncodePrimaryKey(object)
	if err != nil {
		return nil, err
	}

	typePrefix := s.typePrefixes[fullName]

	typePrefixedPrimaryKey := append(typePrefix, primaryKey...)

	if !s.kv.Has(ctx, typePrefixedPrimaryKey) {
		return nil, fmt.Errorf("%w: cannot update in type %s, %s", orm.ErrNotFound, fullName, primaryKey)
	}

	bytes, err := cdc.EncodeObject(object)
	if err != nil {
		return nil, err
	}

	s.kv.Set(ctx, typePrefixedPrimaryKey, bytes)

	return primaryKey, nil
}

func (s Storage) Delete(ctx context.Context, object proto.Message) ([]byte, error) {
	fullName := object.ProtoReflect().Descriptor().FullName()

	cdc, exists := s.codecs[fullName]
	if !exists {
		return nil, fmt.Errorf("unregistered state object: %s", fullName)
	}

	primaryKey, err := cdc.EncodePrimaryKey(object)
	if err != nil {
		return nil, err
	}

	typePrefix := s.typePrefixes[fullName]

	typePrefixedPrimaryKey := append(typePrefix, primaryKey...)

	if !s.kv.Has(ctx, typePrefixedPrimaryKey) {
		return nil, fmt.Errorf("%w: cannot delete in type %s, %s", orm.ErrNotFound, fullName, primaryKey)
	}

	s.kv.Delete(ctx, typePrefixedPrimaryKey)

	return primaryKey, nil
}

func (s Storage) RegisterObject(ctx context.Context, descriptor *v1alpha1.StateObjectDescriptor, messageType protoreflect.MessageType) error {
	md := messageType.Descriptor()
	if _, exists := s.codecs[md.FullName()]; exists {
		return fmt.Errorf("already registered: %s", md.FullName())
	}

	cdc, err := NewCodec(descriptor.TableDescriptor, messageType)
	if err != nil {
		return err
	}

	s.typePrefixes[md.FullName()] = descriptor.TypePrefix
	s.codecs[md.FullName()] = cdc

	return nil
}
