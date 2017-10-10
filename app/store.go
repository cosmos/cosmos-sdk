package app

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/iavl"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/state"
)

// Store contains the merkle tree, and all info to handle abci requests
type Store struct {
	state.State
	logger log.Logger
}

// MockStore returns an in-memory store only intended for testing
func MockStore() *Store {
	res, err := NewStore("", 0, log.NewNopLogger())
	if err != nil {
		// should never happen, abort test if it does
		panic(err)
	}
	return res
}

// NewStore initializes an in-memory iavl.VersionedTree, or attempts to load a
// persistant tree from disk
func NewStore(dbName string, cacheSize int, logger log.Logger) (*Store, error) {
	// memory backed case, just for testing
	if dbName == "" {
		tree := iavl.NewVersionedTree(
			0,
			dbm.NewMemDB(),
		)
		store := &Store{
			State:  state.NewState(tree),
			logger: logger,
		}
		return store, nil
	}

	// Expand the path fully
	dbPath, err := filepath.Abs(dbName)
	if err != nil {
		return nil, errors.Wrap(err, "Invalid Database Name")
	}

	// Some external calls accidently add a ".db", which is now removed
	dbPath = strings.TrimSuffix(dbPath, path.Ext(dbPath))

	// Split the database name into it's components (dir, name)
	dir := path.Dir(dbPath)
	name := path.Base(dbPath)

	// Make sure the path exists
	empty, _ := cmn.IsDirEmpty(dbPath + ".db")

	// Open database called "dir/name.db", if it doesn't exist it will be created
	db := dbm.NewDB(name, dbm.LevelDBBackendStr, dir)
	tree := iavl.NewVersionedTree(cacheSize, db)

	if empty {
		logger.Info("no existing db, creating new db")
	} else {
		logger.Info("loading existing db")
		if err = tree.Load(); err != nil {
			return nil, errors.Wrap(err, "Loading tree")
		}
	}

	res := &Store{
		State:  state.NewState(tree),
		logger: logger,
	}
	return res, nil
}

// Height gets the last height stored in the database
func (s *Store) Height() uint64 {
	return s.State.LatestHeight()
}

// Hash gets the last hash stored in the database
func (s *Store) Hash() []byte {
	return s.State.LatestHash()
}

// Info implements abci.Application. It returns the height, hash and size (in the data).
// The height is the block that holds the transactions, not the apphash itself.
func (s *Store) Info() abci.ResponseInfo {
	s.logger.Info("Info synced",
		"height", s.Height(),
		"hash", fmt.Sprintf("%X", s.Hash()))
	return abci.ResponseInfo{
		Data:             cmn.Fmt("size:%v", s.State.Size()),
		LastBlockHeight:  s.Height() - 1,
		LastBlockAppHash: s.Hash(),
	}
}

// Commit implements abci.Application
func (s *Store) Commit() abci.Result {
	height := s.Height() + 1

	hash, err := s.State.Commit(height)
	if err != nil {
		return abci.NewError(abci.CodeType_InternalError, err.Error())
	}
	s.logger.Debug("Commit synced",
		"height", height,
		"hash", fmt.Sprintf("%X", hash),
	)

	if s.State.Size() == 0 {
		return abci.NewResultOK(nil, "Empty hash for empty tree")
	}
	fmt.Printf("ABCI Commit: %d / %X\n", height, hash)
	return abci.NewResultOK(hash, "")
}

// Query implements abci.Application
func (s *Store) Query(reqQuery abci.RequestQuery) (resQuery abci.ResponseQuery) {
	// set the query response height to current
	tree := s.State.Committed()

	height := reqQuery.Height
	if height == 0 {
		if tree.Tree.VersionExists(s.Height() - 1) {
			height = s.Height() - 1
		} else {
			height = s.Height()
		}
	}
	resQuery.Height = height

	switch reqQuery.Path {
	case "/store", "/key": // Get by key
		key := reqQuery.Data // Data holds the key bytes
		resQuery.Key = key
		if reqQuery.Prove {
			value, proof, err := tree.GetVersionedWithProof(key, height)
			if err != nil {
				resQuery.Log = err.Error()
				break
			}
			resQuery.Value = value
			resQuery.Proof = proof.Bytes()
		} else {
			value := tree.Get(key)
			resQuery.Value = value
		}

	default:
		resQuery.Code = abci.CodeType_UnknownRequest
		resQuery.Log = cmn.Fmt("Unexpected Query path: %v", reqQuery.Path)
	}
	return
}
