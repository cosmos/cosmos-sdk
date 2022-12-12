package collections

import (
	"fmt"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"regexp"
)

// NewSchema creates a new schema from the provided store key.
func NewSchema(storeKey storetypes.StoreKey) Schema {
	return Schema{
		storeKey:            storeKey,
		collectionsByName:   map[string]Collection{},
		collectionsByPrefix: map[string]Collection{},
	}
}

// Schema specifies a group of collections stored within the storage specified
// by a single store key. All the collections within the schema must have a
// unique binary prefix and human-readable name. Schema will eventually include
// methods for importing/exporting genesis data and for schema reflection for
// clients.
type Schema struct {
	storeKey            storetypes.StoreKey
	collectionsByPrefix map[string]Collection
	collectionsByName   map[string]Collection
}

func (s Schema) addCollection(collection Collection) {
	prefix := collection.Prefix()
	name := collection.Name()

	if _, ok := s.collectionsByPrefix[string(prefix)]; ok {
		panic(fmt.Errorf("prefix %v already taken within schema", prefix))
	}

	if _, ok := s.collectionsByName[name]; ok {
		panic(fmt.Errorf("name %s already taken within schema", name))
	}

	if !nameRegex.MatchString(name) {
		panic(fmt.Errorf("name must match regex %s, got %s", nameRegex.String(), name))
	}

	s.collectionsByPrefix[string(prefix)] = collection
	s.collectionsByName[name] = collection
}

// NameRegex is the regular expression that all valid collection names must match.
const NameRegex = "^[A-Za-z][A-Za-z0-9_]*$"

var nameRegex = regexp.MustCompile(NameRegex)
