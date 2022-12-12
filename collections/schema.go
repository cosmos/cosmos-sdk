package collections

import (
	"fmt"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"regexp"
)

func NewSchema(storeKey storetypes.StoreKey) Schema {
	return Schema{
		storeKey:            storeKey,
		collectionsByName:   map[string]Collection{},
		collectionsByPrefix: map[string]Collection{},
	}
}

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

var nameRegex = regexp.MustCompile("^[A-Za-z][A-Za-z0-9_]*$")
