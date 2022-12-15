package collections

import (
	"context"
	"cosmossdk.io/core/store"
	"fmt"
	"regexp"

	"cosmossdk.io/core/appmodule"
)

// Schema specifies a group of collections stored within the storage specified
// by a single store key. All the collections within the schema must have a
// unique binary prefix and human-readable name. Schema will eventually include
// methods for importing/exporting genesis data and for schema reflection for
// clients.
type Schema struct {
	storeAccessor       func(context.Context) store.KVStore
	collectionsByPrefix map[string]collection
	collectionsByName   map[string]collection
}

// NewSchema creates a new schema for the provided KVStoreService.
func NewSchema(service store.KVStoreService) Schema {
	return NewSchemaFromAccessor(func(ctx context.Context) store.KVStore {
		return service.OpenKVStore(ctx)
	})
}

// NewMemoryStoreSchema creates a new schema for the provided MemoryStoreService.
func NewMemoryStoreSchema(service store.MemoryStoreService) Schema {
	return NewSchemaFromAccessor(func(ctx context.Context) store.KVStore {
		return service.OpenMemoryStore(ctx)
	})
}

// NewSchemaFromAccessor creates a new schema for the provided store accessor
// function. Modules built against versions of the SDK which do not support
// the cosmossdk.io/core/appmodule APIs should use this method.
// Ex:

//	NewSchemaFromAccessor(func(ctx context.Context) store.KVStore {
//			return sdk.UnwrapSDKContext(ctx).KVStore(kvStoreKey)
//	}
func NewSchemaFromAccessor(accessor func(context.Context) store.KVStore) Schema {
	return Schema{
		storeAccessor:       accessor,
		collectionsByName:   map[string]collection{},
		collectionsByPrefix: map[string]collection{},
	}
}

func (s Schema) addCollection(collection collection) {
	prefix := collection.getPrefix()
	name := collection.getName()

	if _, ok := s.collectionsByPrefix[string(prefix)]; ok {
		panic(fmt.Errorf("prefix %v already taken within schema", prefix))
	}

	if _, ok := s.collectionsByName[name]; ok {
		panic(fmt.Errorf("name %s already taken within schema", name))
	}

	if !nameRegex.MatchString(name) {
		panic(fmt.Errorf("name must match regex %s, got %s", NameRegex, name))
	}

	s.collectionsByPrefix[string(prefix)] = collection
	s.collectionsByName[name] = collection
}

// NameRegex is the regular expression that all valid collection names must match.
const NameRegex = "[A-Za-z][A-Za-z0-9_]*"

var nameRegex = regexp.MustCompile("^" + NameRegex + "$")

// DefaultGenesis implements the appmodule.HasGenesis.DefaultGenesis method.
func (s Schema) DefaultGenesis(target appmodule.GenesisTarget) error {
	for name, coll := range s.collectionsByName {
		writer, err := target(name)
		if err != nil {
			return err
		}

		err = coll.defaultGenesis(writer)
		if err != nil {
			return err
		}

		err = writer.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// ValidateGenesis implements the appmodule.HasGenesis.ValidateGenesis method.
func (s Schema) ValidateGenesis(source appmodule.GenesisSource) error {
	for name, coll := range s.collectionsByName {
		reader, err := source(name)
		if err != nil {
			return err
		}

		err = coll.validateGenesis(reader)
		if err != nil {
			return err
		}

		err = reader.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// InitGenesis implements the appmodule.HasGenesis.InitGenesis method.
func (s Schema) InitGenesis(ctx context.Context, source appmodule.GenesisSource) error {
	for name, coll := range s.collectionsByName {
		reader, err := source(name)
		if err != nil {
			return err
		}

		err = coll.importGenesis(ctx, reader)
		if err != nil {
			return err
		}

		err = reader.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// ExportGenesis implements the appmodule.HasGenesis.ExportGenesis method.
func (s Schema) ExportGenesis(ctx context.Context, target appmodule.GenesisTarget) error {
	for name, coll := range s.collectionsByName {
		writer, err := target(name)
		if err != nil {
			return err
		}

		err = coll.exportGenesis(ctx, writer)
		if err != nil {
			return err
		}

		err = writer.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
