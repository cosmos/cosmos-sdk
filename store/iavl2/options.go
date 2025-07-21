package iavl2

import (
	"cosmossdk.io/log"

	"cosmossdk.io/store/metrics"
	"cosmossdk.io/store/types"
)

// Config defines the configuration for the IAVL2 store.
// This is intended to be user-level configuration,
type Config struct {
	// Path is the path to the IAVL2 store directory.
	Path string
}

// Options defines the options for creating an IAVL2 store.
// The difference between Config and Options is that Config are user-level options
// where as Options are used internally by the store framework.
type Options struct {
	Logger         log.Logger
	Metrics        metrics.StoreMetrics
	Key            types.StoreKey
	CommitID       types.CommitID
	InitialVersion uint64
}
