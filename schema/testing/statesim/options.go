package statesim

// Options are options for object, module and app state simulators.
type Options struct {
	// CanRetainDeletions indicates that the simulator can retain deletions when that flag is enabled
	// on object types. This should be set to match the indexers ability to retain deletions or not
	// for accurately testing the expected state in the simulator with the indexer's actual state.
	CanRetainDeletions bool
}
