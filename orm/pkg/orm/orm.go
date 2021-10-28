package orm

import (
	"context"
	"errors"
	"github.com/cosmos/cosmos-sdk/orm/apis/orm/v1alpha1"
	"github.com/cosmos/cosmos-sdk/orm/pkg/kv"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// errors
var (
	ErrAlreadyExists = errors.New("orm: object already exists")
	ErrNotFound      = errors.New("orm: object not found")
)

type ListOptions struct {
	FieldsToMatch []FieldMatch
}

type FieldMatch struct {
	FieldName protoreflect.Name
	Value     protoreflect.Value
}

type ObjectIterator struct {
	kv.Iterator
	store *ORM
}

func (i ObjectIterator) Get(o proto.Message) error {
	return i.store.Get(i.Context(), i.Key(), o)
}

type Indexer interface {
	Index(ctx context.Context, primaryKey []byte, object proto.Message) error
	ClearIndexes(ctx context.Context, primaryKey []byte, object proto.Message) error
	List(ctx context.Context, object proto.Message, options ListOptions) (kv.Iterator, error)
	RegisterObject(ctx context.Context, descriptor *v1alpha1.StateObjectDescriptor, messageType protoreflect.MessageType) error
}

type Storage interface {
	Create(ctx context.Context, object proto.Message) (primaryKey []byte, err error)
	Get(ctx context.Context, primaryKey []byte, target proto.Message) error
	Update(ctx context.Context, object proto.Message) (primaryKey []byte, err error)
	Delete(ctx context.Context, object proto.Message) (primaryKey []byte, err error)
	RegisterObject(ctx context.Context, descriptor *v1alpha1.StateObjectDescriptor, messageType protoreflect.MessageType) error
}

type ORM struct {
	indexer Indexer
	storage Storage
}

func NewORM(indexer Indexer, storage Storage) *ORM {
	return &ORM{
		indexer: indexer,
		storage: storage,
	}
}

func (s *ORM) Create(ctx context.Context, object proto.Message) error {
	// create object in storage
	pk, err := s.storage.Create(ctx, object)
	if err != nil {
		return err
	}
	// index object
	err = s.indexer.Index(ctx, pk, object)
	if err != nil {
		return err
	}
	return nil
}

func (s *ORM) Update(ctx context.Context, object proto.Message) error {
	// update object in storage
	pk, err := s.storage.Update(ctx, object)
	if err != nil {
		return err
	}
	// clear indexes
	err = s.indexer.ClearIndexes(ctx, pk, object)
	if err != nil {
		return err
	}

	// update indexes
	err = s.indexer.Index(ctx, pk, object)
	return nil
}

func (s *ORM) Get(ctx context.Context, id []byte, target proto.Message) error {
	return s.storage.Get(ctx, id, target)
}

func (s *ORM) Delete(ctx context.Context, target proto.Message) error {
	pk, err := s.storage.Delete(ctx, target)
	if err != nil {
		return err
	}
	return s.indexer.ClearIndexes(ctx, pk, target)
}

func (s *ORM) List(ctx context.Context, object proto.Message, options ListOptions) (ObjectIterator, error) {
	keyIter, err := s.indexer.List(ctx, object, options)
	if err != nil {
		return ObjectIterator{}, err
	}
	return ObjectIterator{Iterator: keyIter, store: s}, nil
}

func (s *ORM) RegisterObject(ctx context.Context, descriptor *v1alpha1.StateObjectDescriptor, messageType protoreflect.MessageType) error {
	panic("impl")
}
