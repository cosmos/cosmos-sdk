package appmanager

type Store interface {
	NewBlockWithVersion(version uint64) ReadonlyStore
	ReadonlyWithVersion(version uint64) ReadonlyStore
	CommitChanges(changes []ChangeSet) (Hash, error)
}

type ChangeSet struct {
	Key, Value []byte
	Remove     bool
}

type BranchStore interface {
	ReadonlyStore
	Set(key []byte, value []byte) error
	Delete(key []byte) error
	ApplyChangeSets(changes []ChangeSet) error
	ChangeSets() ([]ChangeSet, error)
}

type Iterator interface {
	Next() error
	Valid() bool
	Key() []byte
	Value() []byte
	Close() error
}

type ReadonlyStore interface {
	Get([]byte) ([]byte, error)
	Iterate(start, end []byte) Iterator // consider removing iterate?
}
