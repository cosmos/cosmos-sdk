package collections

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"cosmossdk.io/core/appmodule"

	"cosmossdk.io/core/store"

	"github.com/hashicorp/go-multierror"
)

// SchemaBuilder is used for building schemas. The Build method should always
// be called after all collections have been initialized. Initializing new
// collections with the builder after initialization will result in panics.
type SchemaBuilder struct {
	schema *Schema
	err    *multierror.Error
}

// NewSchemaBuilder creates a new schema builder from the provided store key.
// Callers should always call the SchemaBuilder.Build method when they are
// done adding collections to the schema.
func NewSchemaBuilder(service store.KVStoreService) *SchemaBuilder {
	return &SchemaBuilder{
		schema: &Schema{
			storeAccessor:       service.OpenKVStore,
			collectionsByName:   map[string]collection{},
			collectionsByPrefix: map[string]collection{},
		},
	}
}

// Build should be called after all collections that are part of the schema
// have been initialized in order to get a reference to the Schema. It is
// important to check the returned error for any initialization errors.
// The SchemaBuilder CANNOT be used after Build is called - doing so will
// result in panics.
func (s *SchemaBuilder) Build() (Schema, error) {
	if s.err != nil {
		return Schema{}, s.err
	}

	// check for any overlapping prefixes
	for prefix := range s.schema.collectionsByPrefix {
		for prefix2 := range s.schema.collectionsByPrefix {
			// don't compare the prefix to itself
			if prefix == prefix2 {
				continue
			}

			// if one prefix is the prefix of the other we have an overlap and
			// this schema is corrupt
			if strings.HasPrefix(prefix, prefix2) {
				return Schema{}, fmt.Errorf("schema has overlapping prefixes 0x%x and 0x%x", prefix, prefix2)
			}
		}
	}

	if s.schema == nil {
		// explicit panic to avoid nil pointer dereference
		panic("schema is nil")
	}

	schema := *s.schema

	s.schema = nil // this makes the builder unusable

	return schema, nil
}

func (s *SchemaBuilder) addCollection(collection collection) {
	prefix := collection.getPrefix()
	name := collection.getName()

	if _, ok := s.schema.collectionsByPrefix[string(prefix)]; ok {
		s.err = multierror.Append(s.err, fmt.Errorf("prefix %v already taken within schema", prefix))
		return
	}

	if _, ok := s.schema.collectionsByName[name]; ok {
		s.err = multierror.Append(s.err, fmt.Errorf("name %s already taken within schema", name))
		return
	}

	if !nameRegex.MatchString(name) {
		s.err = multierror.Append(s.err, fmt.Errorf("name must match regex %s, got %s", NameRegex, name))
		return
	}

	s.schema.collectionsByPrefix[string(prefix)] = collection
	s.schema.collectionsByName[name] = collection
}

// NameRegex is the regular expression that all valid collection names must match.
const NameRegex = "[A-Za-z][A-Za-z0-9_]*"

var nameRegex = regexp.MustCompile("^" + NameRegex + "$")

// Schema specifies a group of collections stored within the storage specified
// by a single store key. All the collections within the schema must have a
// unique binary prefix and human-readable name. Schema will eventually include
// methods for importing/exporting genesis data and for schema reflection for
// clients.
type Schema struct {
	storeAccessor       func(context.Context) store.KVStore
	collectionsOrdered  []string
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
