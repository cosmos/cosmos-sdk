package collections

import (
	"fmt"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"regexp"
	"strings"

	"github.com/hashicorp/go-multierror"
)

type SchemaBuilder struct {
	schema *Schema
	err    *multierror.Error
}

// NewSchemaBuilder creates a new schema builder from the provided store key.
// Callers should always call the SchemaBuilder.Build method when they are
// done adding collections to the schema.
func NewSchemaBuilder(storeKey storetypes.StoreKey) *SchemaBuilder {
	return &SchemaBuilder{
		schema: &Schema{
			storeKey:            storeKey,
			collectionsByName:   map[string]collection{},
			collectionsByPrefix: map[string]collection{},
		},
	}
}

// Build should be called after all collections that are part of the schema
// have been initialized in order to get a reference to the Schema and to
// check for any initialization errors.
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

	return *s.schema, nil
}

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
