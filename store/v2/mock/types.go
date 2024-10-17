package mock

import "cosmossdk.io/store/v2"

// StateCommitter is a mock of store.Committer
type StateCommitter interface {
	store.Committer
	store.Pruner
	store.PausablePruner
	store.UpgradeableStore
}

// StateStorage is a mock of store.VersionedWriter
type StateStorage interface {
	store.VersionedWriter
	store.UpgradableDatabase
	store.Pruner
	store.PausablePruner
}
