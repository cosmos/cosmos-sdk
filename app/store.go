package app

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-wire"
	"github.com/tendermint/iavl"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/state"
)

// Store contains the merkle tree, and all info to handle abci requests
type Store struct {
	state.State
	height    uint64
	hash      []byte
	persisted bool

	logger log.Logger
}

var stateKey = []byte("merkle:state") // Database key for merkle tree save value db values

// ChainState contains the latest Merkle root hash and the number of times `Commit` has been called
type ChainState struct {
	Hash   []byte
	Height uint64
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
	// start at 1 so the height returned by query is for the
	// next block, ie. the one that includes the AppHash for our current state
	initialHeight := uint64(1)

	// Non-persistent case
	if dbName == "" {
		tree := iavl.NewVersionedTree(
			0,
			dbm.NewMemDB(),
		)
		store := &Store{
			State:  state.NewState(tree, true),
			height: initialHeight,
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

	var chainState ChainState
	if empty {
		logger.Info("no existing db, creating new db")
		chainState = ChainState{
			Hash:   nil,
			Height: initialHeight,
		}
		db.Set(stateKey, wire.BinaryBytes(chainState))
	} else {
		logger.Info("loading existing db")
		eyesStateBytes := db.Get(stateKey)

		if err = wire.ReadBinaryBytes(eyesStateBytes, &chainState); err != nil {
			return nil, errors.Wrap(err, "Reading MerkleEyesState")
		}
		if err = tree.Load(); err != nil {
			return nil, errors.Wrap(err, "Loading tree")
		}
	}

	res := &Store{
		State:     state.NewState(tree, true),
		height:    chainState.Height,
		hash:      chainState.Hash,
		persisted: true,
		logger:    logger,
	}
	return res, nil
}

// Info implements abci.Application. It returns the height, hash and size (in the data).
// The height is the block that holds the transactions, not the apphash itself.
func (s *Store) Info() abci.ResponseInfo {
	s.logger.Info("Info synced",
		"height", s.height,
		"hash", fmt.Sprintf("%X", s.hash))
	return abci.ResponseInfo{
		Data:             cmn.Fmt("size:%v", s.State.Size()),
		LastBlockHeight:  s.height - 1,
		LastBlockAppHash: s.hash,
	}
}

// Commit implements abci.Application
func (s *Store) Commit() abci.Result {
	var err error
	s.height++
	s.hash, err = s.State.Hash()
	if err != nil {
		return abci.NewError(abci.CodeType_InternalError, err.Error())
	}

	s.logger.Debug("Commit synced",
		"height", s.height,
		"hash", fmt.Sprintf("%X", s.hash))

	s.State.BatchSet(stateKey, wire.BinaryBytes(ChainState{
		Hash:   s.hash,
		Height: s.height,
	}))

	hash, err := s.State.Commit(s.height)
	if err != nil {
		return abci.NewError(abci.CodeType_InternalError, err.Error())
	}
	if !bytes.Equal(hash, s.hash) {
		return abci.NewError(abci.CodeType_InternalError, "AppHash is incorrect")
	}

	if s.State.Size() == 0 {
		return abci.NewResultOK(nil, "Empty hash for empty tree")
	}
	return abci.NewResultOK(s.hash, "")
}

// Query implements abci.Application
func (s *Store) Query(reqQuery abci.RequestQuery) (resQuery abci.ResponseQuery) {
	// set the query response height to current
	tree := s.State.Committed()

	height := reqQuery.Height
	if height == 0 {
		if tree.Tree.VersionExists(s.height - 1) {
			height = s.height - 1
		} else {
			height = s.height
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
