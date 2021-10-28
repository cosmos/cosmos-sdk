package kvindexer

import (
	"context"
	"fmt"
	"github.com/cosmos/cosmos-sdk/orm/apis/orm/v1alpha1"
	"github.com/cosmos/cosmos-sdk/orm/internal/kvindexer"
	"github.com/cosmos/cosmos-sdk/orm/pkg/kv"
	"github.com/cosmos/cosmos-sdk/orm/pkg/orm"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	objectIndexesPrefix   = []byte{0x0} // maps secondary keys to primary keys
	objectToIndexesPrefix = []byte{0x1} // maps primary key to all their known indexed secondary keys
)

const (
	TypePrefixLength = 2
)

func NewIndexer(kv kv.KV) *Indexer {
	return &Indexer{
		kv:              kv,
		fieldIndexers:   map[protoreflect.FullName]*FieldKeyEncoder{},
		objectsIndexers: map[protoreflect.FullName][]*FieldKeyEncoder{},
		typePrefix:      map[protoreflect.FullName][]byte{},
	}
}

type Indexer struct {
	kv              kv.KV
	fieldIndexers   map[protoreflect.FullName]*FieldKeyEncoder   // maps many objects fields to its indexer
	objectsIndexers map[protoreflect.FullName][]*FieldKeyEncoder // maps one object to all its indexers
	typePrefix      map[protoreflect.FullName][]byte             // typePrefix must be of constant size
}

func (i *Indexer) Index(ctx context.Context, primaryKey []byte, object proto.Message) error {
	// TODO(fdymylja): should check not singleton, and that it has indexes
	fullName := object.ProtoReflect().Descriptor().FullName()
	indexers, exists := i.objectsIndexers[fullName]
	if !exists {
		return fmt.Errorf("unregistered object: %s", object.ProtoReflect().Descriptor().FullName())
	}
	typePrefix := i.typePrefix[fullName]

	mappingKeys := &kvindexer.MappingKeysList{Keys: make([][]byte, len(indexers))}
	for i, indexer := range indexers {
		mappingKey, err := indexer.EncodePrimaryKey(primaryKey, object)
		if err != nil {
			return err
		}

		mappingKeys.Keys[i] = mappingKey
	}

	for _, typePrefixedMappingKey := range mappingKeys.Keys {
		if i.kv.Has(ctx, typePrefixedMappingKey) {
			return fmt.Errorf("type %s object %s is already indexed", fullName, primaryKey)
		}
		i.kv.Set(ctx, typePrefixedMappingKey, []byte{})
	}

	// save mappings list
	typePrefixedMappingKey := NewTypePrefixedIndexListKey(typePrefix, primaryKey)
	if i.kv.Has(ctx, typePrefixedMappingKey) {
		return fmt.Errorf("type %s object %s has already an index list", fullName, primaryKey) // TODO(Fdymylja): this is data corruption
	}

	x, err := proto.MarshalOptions{Deterministic: true}.Marshal(mappingKeys)
	if err != nil {
		panic(err)
	}
	i.kv.Set(ctx, typePrefixedMappingKey, x)

	return nil
}

func (i *Indexer) ClearIndexes(ctx context.Context, primaryKey []byte, object proto.Message) error {
	// TODO(fdymylja): should check not singleton, and that it has indexes
	typePrefix, exist := i.typePrefix[object.ProtoReflect().Descriptor().FullName()]
	if !exist {
		return fmt.Errorf("unknown object: %s", object.ProtoReflect().Descriptor().FullName())
	}
	key := NewTypePrefixedIndexListKey(typePrefix, primaryKey)

	if !i.kv.Has(ctx, key) {
		return fmt.Errorf("unknown in type %s key: %s", object.ProtoReflect().Descriptor().FullName(), key)
	}

	// clear indexes
	indexesList := &kvindexer.MappingKeysList{}
	v := i.kv.Get(ctx, key)
	if err := (proto.UnmarshalOptions{}.Unmarshal(v, indexesList)); err != nil {
		panic(err)
	}

	for _, index := range indexesList.Keys {
		if !i.kv.Has(ctx, key) {
			panic("data corruption")
		}
		i.kv.Delete(ctx, index)
	}

	i.kv.Delete(ctx, key)
	return nil
}

func (i *Indexer) List(ctx context.Context, object proto.Message, options orm.ListOptions) (kv.Iterator, error) {
	// TODO(fdymylja): implement joint iterator logic
	name := object.ProtoReflect().Descriptor().FullName().Append(options.FieldsToMatch[0].FieldName)
	indexer, exists := i.fieldIndexers[name]
	if !exists {
		return nil, fmt.Errorf("unrecognized index %s in object %s", name.Name(), name.Parent())
	}
	// TODO(fdymylja): value validity check
	key := indexer.IndexPrefix(options.FieldsToMatch[0].Value)
	iter := i.kv.IteratePrefix(ctx, key)
	if !iter.Valid() {
		return nil, fmt.Errorf("%w: no results in type %s for query %s", orm.ErrNotFound, object.ProtoReflect().Descriptor().FullName(), options)
	}

	return iter, nil
}

func (i *Indexer) RegisterObject(ctx context.Context, descriptor *v1alpha1.StateObjectDescriptor, messageType protoreflect.MessageType) error {
	md := messageType.Descriptor()

	_, exists := i.objectsIndexers[md.FullName()]
	if exists {
		return fmt.Errorf("object already registered: %s", md.FullName())
	}

	if len(descriptor.TypePrefix) != TypePrefixLength {
		return fmt.Errorf("invalid type prefix length")
	}

	i.typePrefix[md.FullName()] = descriptor.TypePrefix
	i.objectsIndexers[md.FullName()] = []*FieldKeyEncoder{}

	// singleton registers no op encodings
	if descriptor.TableDescriptor.Singleton {
		return nil
	}

	for _, sk := range descriptor.TableDescriptor.SecondaryKeys {
		fieldEncoder, err := NewFieldIndexer(descriptor.TypePrefix, sk, messageType)
		if err != nil {
			return err
		}

		if _, exists := i.fieldIndexers[fieldEncoder.fd.FullName()]; exists {
			return fmt.Errorf("duplicate index: %s", sk.FieldName)
		}

		i.objectsIndexers[md.FullName()] = append(i.objectsIndexers[md.FullName()], fieldEncoder)
		i.fieldIndexers[fieldEncoder.fd.FullName()] = fieldEncoder
	}

	return nil
}

// NewTypePrefixedIndexListKey returns the type prefixed key used to save
// primary key to full indexes list.
func NewTypePrefixedIndexListKey(prefix []byte, key []byte) []byte {
	return append([]byte{objectToIndexesPrefix[0], prefix[0], prefix[1]}, key...)
}
