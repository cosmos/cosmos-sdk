package flat

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"math"
	"sync"

	dbm "github.com/cosmos/cosmos-sdk/db"
	"github.com/cosmos/cosmos-sdk/db/prefix"
	abci "github.com/tendermint/tendermint/abci/types"

	util "github.com/cosmos/cosmos-sdk/internal"
	"github.com/cosmos/cosmos-sdk/store/cachekv"
	"github.com/cosmos/cosmos-sdk/store/listenkv"
	"github.com/cosmos/cosmos-sdk/store/tracekv"
	"github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/store/v2/smt"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

var (
	_ types.KVStore       = (*Store)(nil)
	_ types.CommitKVStore = (*Store)(nil)
	_ types.Queryable     = (*Store)(nil)
)

var (
	merkleRootKey     = []byte{0} // Key for root hash of Merkle tree
	dataPrefix        = []byte{1} // Prefix for state mappings
	indexPrefix       = []byte{2} // Prefix for Store reverse index
	merkleNodePrefix  = []byte{3} // Prefix for Merkle tree nodes
	merkleValuePrefix = []byte{4} // Prefix for Merkle value mappings
)

var (
	ErrVersionDoesNotExist = errors.New("version does not exist")
	ErrMaximumHeight       = errors.New("maximum block height reached")
)

type StoreConfig struct {
	// Version pruning options for backing DBs.
	Pruning types.PruningOptions
	// The backing DB to use for the state commitment Merkle tree data.
	// If nil, Merkle data is stored in the state storage DB under a separate prefix.
	MerkleDB       dbm.DBConnection
	InitialVersion uint64
}

// Store is a CommitKVStore which handles state storage and commitments as separate concerns,
// optionally using separate backing key-value DBs for each.
// Allows synchronized R/W access by locking.
type Store struct {
	stateDB   dbm.DBConnection
	stateTxn  dbm.DBReadWriter
	dataTxn   dbm.DBReadWriter
	merkleTxn dbm.DBReadWriter
	indexTxn  dbm.DBReadWriter
	// State commitment (SC) KV store for current version
	merkleStore *smt.Store

	opts StoreConfig
	mtx  sync.RWMutex
}

var DefaultStoreConfig = StoreConfig{Pruning: types.PruneDefault, MerkleDB: nil}

// NewStore creates a new Store, or loads one if db contains existing data.
func NewStore(db dbm.DBConnection, opts StoreConfig) (ret *Store, err error) {
	versions, err := db.Versions()
	if err != nil {
		return
	}
	loadExisting := false
	// If the DB is not empty, attempt to load existing data
	if saved := versions.Count(); saved != 0 {
		if opts.InitialVersion != 0 && versions.Last() < opts.InitialVersion {
			return nil, fmt.Errorf("latest saved version is less than initial version: %v < %v",
				versions.Last(), opts.InitialVersion)
		}
		loadExisting = true
	}
	err = db.Revert()
	if err != nil {
		return
	}
	stateTxn := db.ReadWriter()
	defer func() {
		if err != nil {
			err = util.CombineErrors(err, stateTxn.Discard(), "stateTxn.Discard also failed")
		}
	}()
	merkleTxn := stateTxn
	if opts.MerkleDB != nil {
		var mversions dbm.VersionSet
		mversions, err = opts.MerkleDB.Versions()
		if err != nil {
			return
		}
		// Version sets of each DB must match
		if !versions.Equal(mversions) {
			err = fmt.Errorf("Storage and Merkle DB have different version history")
			return
		}
		err = opts.MerkleDB.Revert()
		if err != nil {
			return
		}
		merkleTxn = opts.MerkleDB.ReadWriter()
	}

	var merkleStore *smt.Store
	if loadExisting {
		var root []byte
		root, err = stateTxn.Get(merkleRootKey)
		if err != nil {
			return
		}
		if root == nil {
			err = fmt.Errorf("could not get root of SMT")
			return
		}
		merkleStore = loadSMT(merkleTxn, root)
	} else {
		merkleNodes := prefix.NewPrefixReadWriter(merkleTxn, merkleNodePrefix)
		merkleValues := prefix.NewPrefixReadWriter(merkleTxn, merkleValuePrefix)
		merkleStore = smt.NewStore(merkleNodes, merkleValues)
	}
	return &Store{
		stateDB:     db,
		stateTxn:    stateTxn,
		dataTxn:     prefix.NewPrefixReadWriter(stateTxn, dataPrefix),
		indexTxn:    prefix.NewPrefixReadWriter(stateTxn, indexPrefix),
		merkleTxn:   merkleTxn,
		merkleStore: merkleStore,
		opts:        opts,
	}, nil
}

func (s *Store) Close() error {
	err := s.stateTxn.Discard()
	if s.opts.MerkleDB != nil {
		err = util.CombineErrors(err, s.merkleTxn.Discard(), "merkleTxn.Discard also failed")
	}
	return err
}

// Get implements KVStore.
func (s *Store) Get(key []byte) []byte {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	val, err := s.dataTxn.Get(key)
	if err != nil {
		panic(err)
	}
	return val
}

// Has implements KVStore.
func (s *Store) Has(key []byte) bool {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	has, err := s.dataTxn.Has(key)
	if err != nil {
		panic(err)
	}
	return has
}

// Set implements KVStore.
func (s *Store) Set(key, value []byte) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	err := s.dataTxn.Set(key, value)
	if err != nil {
		panic(err)
	}
	s.merkleStore.Set(key, value)
	khash := sha256.Sum256(key)
	err = s.indexTxn.Set(khash[:], key)
	if err != nil {
		panic(err)
	}
}

// Delete implements KVStore.
func (s *Store) Delete(key []byte) {
	khash := sha256.Sum256(key)
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.merkleStore.Delete(key)
	_ = s.indexTxn.Delete(khash[:])
	_ = s.dataTxn.Delete(key)
}

type contentsIterator struct {
	dbm.Iterator
	valid bool
}

func newIterator(source dbm.Iterator) *contentsIterator {
	ret := &contentsIterator{Iterator: source}
	ret.Next()
	return ret
}

func (it *contentsIterator) Next()       { it.valid = it.Iterator.Next() }
func (it *contentsIterator) Valid() bool { return it.valid }

// Iterator implements KVStore.
func (s *Store) Iterator(start, end []byte) types.Iterator {
	iter, err := s.dataTxn.Iterator(start, end)
	if err != nil {
		panic(err)
	}
	return newIterator(iter)
}

// ReverseIterator implements KVStore.
func (s *Store) ReverseIterator(start, end []byte) types.Iterator {
	iter, err := s.dataTxn.ReverseIterator(start, end)
	if err != nil {
		panic(err)
	}
	return newIterator(iter)
}

// GetStoreType implements Store.
func (s *Store) GetStoreType() types.StoreType {
	return types.StoreTypeDecoupled
}

// Commit implements Committer.
func (s *Store) Commit() types.CommitID {
	versions, err := s.stateDB.Versions()
	if err != nil {
		panic(err)
	}
	target := versions.Last() + 1
	if target > math.MaxInt64 {
		panic(ErrMaximumHeight)
	}
	// Fast forward to initialversion if needed
	if s.opts.InitialVersion != 0 && target < s.opts.InitialVersion {
		target = s.opts.InitialVersion
	}
	cid, err := s.commit(target)
	if err != nil {
		panic(err)
	}

	previous := cid.Version - 1
	if s.opts.Pruning.KeepEvery != 1 && s.opts.Pruning.Interval != 0 && cid.Version%int64(s.opts.Pruning.Interval) == 0 {
		// The range of newly prunable versions
		lastPrunable := previous - int64(s.opts.Pruning.KeepRecent)
		firstPrunable := lastPrunable - int64(s.opts.Pruning.Interval)
		for version := firstPrunable; version <= lastPrunable; version++ {
			if s.opts.Pruning.KeepEvery == 0 || version%int64(s.opts.Pruning.KeepEvery) != 0 {
				s.stateDB.DeleteVersion(uint64(version))
				if s.opts.MerkleDB != nil {
					s.opts.MerkleDB.DeleteVersion(uint64(version))
				}
			}
		}
	}
	return *cid
}

func (s *Store) commit(target uint64) (id *types.CommitID, err error) {
	root := s.merkleStore.Root()
	err = s.stateTxn.Set(merkleRootKey, root)
	if err != nil {
		return
	}
	err = s.stateTxn.Commit()
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			err = util.CombineErrors(err, s.stateDB.Revert(), "stateDB.Revert also failed")
		}
	}()
	err = s.stateDB.SaveVersion(target)
	if err != nil {
		return
	}

	stateTxn := s.stateDB.ReadWriter()
	defer func() {
		if err != nil {
			err = util.CombineErrors(err, stateTxn.Discard(), "stateTxn.Discard also failed")
		}
	}()
	merkleTxn := stateTxn

	// If DBs are not separate, Merkle state has been commmitted & snapshotted
	if s.opts.MerkleDB != nil {
		defer func() {
			if err != nil {
				if delerr := s.stateDB.DeleteVersion(target); delerr != nil {
					err = fmt.Errorf("%w: commit rollback failed: %v", err, delerr)
				}
			}
		}()

		err = s.merkleTxn.Commit()
		if err != nil {
			return
		}
		defer func() {
			if err != nil {
				err = util.CombineErrors(err, s.opts.MerkleDB.Revert(), "merkleDB.Revert also failed")
			}
		}()

		err = s.opts.MerkleDB.SaveVersion(target)
		if err != nil {
			return
		}
		merkleTxn = s.opts.MerkleDB.ReadWriter()
	}

	s.stateTxn = stateTxn
	s.dataTxn = prefix.NewPrefixReadWriter(stateTxn, dataPrefix)
	s.indexTxn = prefix.NewPrefixReadWriter(stateTxn, indexPrefix)
	s.merkleTxn = merkleTxn
	s.merkleStore = loadSMT(merkleTxn, root)

	return &types.CommitID{Version: int64(target), Hash: root}, nil
}

// LastCommitID implements Committer.
func (s *Store) LastCommitID() types.CommitID {
	versions, err := s.stateDB.Versions()
	if err != nil {
		panic(err)
	}
	last := versions.Last()
	if last == 0 {
		return types.CommitID{}
	}
	// Latest Merkle root is the one currently stored
	hash, err := s.stateTxn.Get(merkleRootKey)
	if err != nil {
		panic(err)
	}
	return types.CommitID{Version: int64(last), Hash: hash}
}

func (s *Store) GetPruning() types.PruningOptions   { return s.opts.Pruning }
func (s *Store) SetPruning(po types.PruningOptions) { s.opts.Pruning = po }

// Query implements ABCI interface, allows queries.
//
// by default we will return from (latest height -1),
// as we will have merkle proofs immediately (header height = data height + 1)
// If latest-1 is not present, use latest (which must be present)
// if you care to have the latest data to see a tx results, you must
// explicitly set the height you want to see
func (s *Store) Query(req abci.RequestQuery) (res abci.ResponseQuery) {
	if len(req.Data) == 0 {
		return sdkerrors.QueryResult(sdkerrors.Wrap(sdkerrors.ErrTxDecode, "query cannot be zero length"), false)
	}

	// if height is 0, use the latest height
	height := req.Height
	if height == 0 {
		versions, err := s.stateDB.Versions()
		if err != nil {
			return sdkerrors.QueryResult(errors.New("failed to get version info"), false)
		}
		latest := versions.Last()
		if versions.Exists(latest - 1) {
			height = int64(latest - 1)
		} else {
			height = int64(latest)
		}
	}
	if height < 0 {
		return sdkerrors.QueryResult(fmt.Errorf("height overflow: %v", height), false)
	}
	res.Height = height

	switch req.Path {
	case "/key":
		var err error
		res.Key = req.Data // data holds the key bytes

		dbr, err := s.stateDB.ReaderAt(uint64(height))
		if err != nil {
			if errors.Is(err, dbm.ErrVersionDoesNotExist) {
				err = sdkerrors.ErrInvalidHeight
			}
			return sdkerrors.QueryResult(err, false)
		}
		defer dbr.Discard()
		contents := prefix.NewPrefixReader(dbr, dataPrefix)
		res.Value, err = contents.Get(res.Key)
		if err != nil {
			return sdkerrors.QueryResult(err, false)
		}
		if !req.Prove {
			break
		}
		merkleView := dbr
		if s.opts.MerkleDB != nil {
			merkleView, err = s.opts.MerkleDB.ReaderAt(uint64(height))
			if err != nil {
				return sdkerrors.QueryResult(
					fmt.Errorf("version exists in state DB but not Merkle DB: %v", height), false)
			}
			defer merkleView.Discard()
		}
		root, err := dbr.Get(merkleRootKey)
		if err != nil {
			return sdkerrors.QueryResult(err, false)
		}
		if root == nil {
			return sdkerrors.QueryResult(errors.New("Merkle root hash not found"), false)
		}
		merkleStore := loadSMT(dbm.ReaderAsReadWriter(merkleView), root)
		res.ProofOps, err = merkleStore.GetProof(res.Key)
		if err != nil {
			return sdkerrors.QueryResult(fmt.Errorf("Merkle proof creation failed for key: %v", res.Key), false)
		}

	case "/subspace":
		pairs := kv.Pairs{
			Pairs: make([]kv.Pair, 0),
		}

		subspace := req.Data
		res.Key = subspace

		iterator := s.Iterator(subspace, types.PrefixEndBytes(subspace))
		for ; iterator.Valid(); iterator.Next() {
			pairs.Pairs = append(pairs.Pairs, kv.Pair{Key: iterator.Key(), Value: iterator.Value()})
		}
		iterator.Close()

		bz, err := pairs.Marshal()
		if err != nil {
			panic(fmt.Errorf("failed to marshal KV pairs: %w", err))
		}

		res.Value = bz

	default:
		return sdkerrors.QueryResult(sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unexpected query path: %v", req.Path), false)
	}

	return res
}

func loadSMT(merkleTxn dbm.DBReadWriter, root []byte) *smt.Store {
	merkleNodes := prefix.NewPrefixReadWriter(merkleTxn, merkleNodePrefix)
	merkleValues := prefix.NewPrefixReadWriter(merkleTxn, merkleValuePrefix)
	return smt.LoadStore(merkleNodes, merkleValues, root)
}

func (st *Store) CacheWrap() types.CacheWrap {
	return cachekv.NewStore(st)
}

func (st *Store) CacheWrapWithTrace(w io.Writer, tc types.TraceContext) types.CacheWrap {
	return cachekv.NewStore(tracekv.NewStore(st, w, tc))
}

func (st *Store) CacheWrapWithListeners(storeKey types.StoreKey, listeners []types.WriteListener) types.CacheWrap {
	return cachekv.NewStore(listenkv.NewStore(st, storeKey, listeners))
}
