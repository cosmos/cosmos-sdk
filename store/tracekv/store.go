package tracekv

import (
	"io"

	"cosmossdk.io/store/v2"
)

type (
	Store struct {
		parent  store.KVStore
		context store.TraceContext
		writer  io.Writer
	}

	// traceOperation implements a traced KVStore operation, such as a read or write
	traceOperation struct {
		Operation string                 `json:"operation"`
		Key       string                 `json:"key"`
		Value     string                 `json:"value"`
		Metadata  map[string]interface{} `json:"metadata"`
	}
)
