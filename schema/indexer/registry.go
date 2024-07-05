package indexer

import "fmt"

// Register registers an indexer type with the given initialization function.
func Register(indexerType string, initFunc InitFunc) {
	if _, ok := indexerRegistry[indexerType]; ok {
		panic(fmt.Sprintf("indexer %s already registered", indexerType))
	}

	indexerRegistry[indexerType] = initFunc
}

// Lookup returns the initialization function for the given indexer type or nil.
func Lookup(indexerType string) InitFunc {
	return indexerRegistry[indexerType]
}

var indexerRegistry = map[string]InitFunc{}
