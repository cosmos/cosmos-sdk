package pruning

import (
	"cosmossdk.io/store/v2"
)

// Manager is an abstraction to handle the pruning logic.
type Manager struct {
	stateStore      store.VersionedDatabase
	stateCommitment store.Committer
}
