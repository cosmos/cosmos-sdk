package indexing

import (
	"context"
	"fmt"

	"cosmossdk.io/schema/appdata"
	"cosmossdk.io/schema/logutil"
)

type Indexer interface {
	Initialize(context.Context, InitializationData) (InitializationResult, error)
}

type IndexerResources struct {
	Logger logutil.Logger
}

type IndexerFactory = func(options map[string]interface{}, resources IndexerResources) (Indexer, error)

type InitializationData struct{}

type InitializationResult struct {
	Listener           appdata.Listener
	LastBlockPersisted int64
}

func RegisterIndexer(name string, factory IndexerFactory) {
	if _, ok := indexerRegistry[name]; ok {
		panic(fmt.Sprintf("indexer %s already registered", name))
	}

	indexerRegistry[name] = factory
}

var indexerRegistry = map[string]IndexerFactory{}
