package store

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/tendermint/iavl"
	dbm "github.com/tendermint/tmlibs/db"
)

// NewIAVLLoader returns a CommitterLoader that returns
// an IAVLStore
func NewIAVLLoader(dbName string, cacheSize int, history uint64) CommitterLoader {
	l := iavlLoader{
		dbName:    dbName,
		cacheSize: cacheSize,
		history:   history,
	}
	return CommitterLoader(l.Load)
}

// IAVLStore Implements IterKVStore and Committer
type IAVLStore struct {
	// we must store the last height here, as it is needed
	// for saving the versioned tree
	lastHeight uint64

	// history is how many old versions we hold onto,
	// uses a naive "hold last X versions" algorithm
	history uint64

	tree *iavl.VersionedTree
}

// Commit writes another version to the
func (i *IAVLStore) Commit() CommitID {
	// save a new version
	i.lastHeight++
	hash, err := i.tree.SaveVersion(i.lastHeight)
	if err != nil {
		// TODO: do we want to extend Commit to
		// allow returning errors?
		panic(err)
	}

	// release an old version
	if i.history <= i.lastHeight {
		release := i.lastHeight - i.history
		i.tree.DeleteVersion(release)
	}

	return CommitID{
		Version: i.lastHeight,
		Hash:    hash,
	}
}

// var _ IterKVStore = (*IAVLStore)(nil)
var _ KVStore = (*IAVLStore)(nil)
var _ Committer = (*IAVLStore)(nil)

// iavlLoader contains info on what store we want to load from
type iavlLoader struct {
	dbName    string
	cacheSize int
	history   uint64
}

// Load implements CommitLoader type
func (l iavlLoader) Load(id CommitID) (Committer, error) {
	// memory backed case, just for testing
	if l.dbName == "" {
		tree := iavl.NewVersionedTree(0, dbm.NewMemDB())
		store := &IAVLStore{
			tree:    tree,
			history: l.history,
		}
		return store, nil
	}

	// Expand the path fully
	dbPath, err := filepath.Abs(l.dbName)
	if err != nil {
		return nil, errors.New("Invalid Database Name")
	}

	// Some external calls accidently add a ".db", which is now removed
	dbPath = strings.TrimSuffix(dbPath, path.Ext(dbPath))

	// Split the database name into it's components (dir, name)
	dir := filepath.Dir(dbPath)
	name := filepath.Base(dbPath)

	// Open database called "dir/name.db", if it doesn't exist it will be created
	db := dbm.NewDB(name, dbm.LevelDBBackendStr, dir)
	tree := iavl.NewVersionedTree(l.cacheSize, db)
	if err = tree.Load(); err != nil {
		return nil, errors.New("Loading tree: " + err.Error())
	}

	// TODO: load the version stored in id
	store := &IAVLStore{
		tree:       tree,
		lastHeight: tree.LatestVersion(),
		history:    l.history,
	}

	return store, nil
}
