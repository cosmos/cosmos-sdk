package collections

import (
	"context"
	"fmt"
	"regexp"

	"cosmossdk.io/core/appmodule"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
)

// Schema specifies a group of collections stored within the storage specified
// by a single store key. All the collections within the schema must have a
// unique binary prefix and human-readable name. Schema will eventually include
// methods for importing/exporting genesis data and for schema reflection for
// clients.
type Schema struct {
	storeKey            storetypes.StoreKey
	collectionsByPrefix map[string]collection
	collectionsByName   map[string]collection
}

// NewSchema creates a new schema from the provided store key.
func NewSchema(storeKey storetypes.StoreKey) Schema {
	return Schema{
		storeKey:            storeKey,
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
