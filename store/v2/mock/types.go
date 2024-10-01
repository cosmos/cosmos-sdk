package mock

import "cosmossdk.io/store/v2"

// StateCommitter is a mock of store.Committer
type StateCommitter interface {
	store.Committer
	store.Pruner
	store.PausablePruner
	store.UpgradeableStore
}

// StateStorage is a mock of store.VersionedDatabase
type StateStorage interface {
	store.VersionedDatabase
	store.UpgradableDatabase
	store.Pruner
	store.PausablePruner
}
