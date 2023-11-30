package iavl_v2

import (
	"fmt"
	io "io"
	"path/filepath"
	"time"

	"github.com/cosmos/iavl/v2"
	abci "github.com/tendermint/tendermint/abci/types"

	pruningtypes "github.com/cosmos/cosmos-sdk/pruning/types"
	"github.com/cosmos/cosmos-sdk/store/cachekv"
	"github.com/cosmos/cosmos-sdk/store/listenkv"
	"github.com/cosmos/cosmos-sdk/store/tracekv"
	"github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
)

var (
	_ types.KVStore                 = (*Store)(nil)
	_ types.CommitStore             = (*Store)(nil)
	_ types.CommitKVStore           = (*Store)(nil)
	_ types.Queryable               = (*Store)(nil)
	_ types.StoreWithInitialVersion = (*Store)(nil)
)

type Store struct {
	iavl.Tree
}

func LoadStoreWithInitialVersion(v2RootPath string, key types.StoreKey, id types.CommitID, _ uint64) (types.CommitKVStore, error) {
	// TODO
	// handle initialVersion (last param). This parameter is non-zero when the storeKey is flagged as added in upgrades.
	// i.e. not the happy path.
	path := filepath.Join(v2RootPath, key.Name())
	pool := iavl.NewNodePool()
	sqlOpts := iavl.SqliteDbOptions{Path: path}
	var err error
	sqlOpts.MmapSize, err = sqlOpts.EstimateMmapSize()
	if err != nil {
		return nil, fmt.Errorf("failed to estimate mmap size for sqlite db path=%s: %w", path, err)
	}
	sql, err := iavl.NewSqliteDb(pool, sqlOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db path=%s: %w", path, err)
	}

	tree := iavl.NewTree(sql, pool, iavl.TreeOptions{
		StateStorage: true,
		MetricsProxy: &telemetry.GlobalMetricProxy{},
		HeightFilter: 1,
	})
	if key.Name() == "ibc" {
		err = tree.LoadVersion(id.Version)
	} else {
		err = tree.LoadVersion(id.Version)
	}
	if err != nil {
		return nil, err
	}
	if err = sql.WarmLeaves(); err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return &Store{*tree}, nil
}

func (s *Store) SetInitialVersion(version int64) {
	//TODO implement me
	panic("implement me")
}

func (s *Store) Query(query abci.RequestQuery) abci.ResponseQuery {
	//TODO implement me
	panic("implement me")
}

func (s *Store) Commit() types.CommitID {
	defer telemetry.MeasureSince(time.Now(), "store", "iavl", "commit")

	hash, version, err := s.Tree.SaveVersion()
	if err != nil {
		panic(err)
	}

	return types.CommitID{
		Version: version,
		Hash:    hash,
	}
}

func (s *Store) LastCommitID() types.CommitID {
	hash := s.Tree.Hash()
	fmt.Printf("IAVLV2 Get LastCommitID: %x\n", hash)

	return types.CommitID{
		Version: s.Tree.Version(),
		Hash:    hash,
	}
}

func (s *Store) SetPruning(options pruningtypes.PruningOptions) {
	panic("cannot set pruning options on an initialized IAVL store")
}

func (s *Store) GetPruning() pruningtypes.PruningOptions {
	panic("cannot get pruning options on an initialized IAVL store")
}

func (s *Store) GetStoreType() types.StoreType {
	return types.StoreTypeIAVL
}

func (s *Store) CacheWrap() types.CacheWrap {
	return cachekv.NewStore(s)
}

func (s *Store) CacheWrapWithTrace(w io.Writer, tc types.TraceContext) types.CacheWrap {
	return cachekv.NewStore(tracekv.NewStore(s, w, tc))
}

func (s *Store) CacheWrapWithListeners(storeKey types.StoreKey, listeners []types.WriteListener) types.CacheWrap {
	return cachekv.NewStore(listenkv.NewStore(s, storeKey, listeners))
}

func (s *Store) Get(key []byte) []byte {
	defer telemetry.MeasureSince(time.Now(), "store", "iavl", "get")
	value, err := s.Tree.Get(key)
	if err != nil {
		panic(err)
	}
	return value
}

func (s *Store) Has(key []byte) bool {
	defer telemetry.MeasureSince(time.Now(), "store", "iavl", "has")
	has, err := s.Tree.Has(key)
	if err != nil {
		panic(err)
	}
	return has
}

func (s *Store) Set(key, value []byte) {
	defer telemetry.MeasureSince(time.Now(), "store", "iavl", "set")
	types.AssertValidKey(key)
	types.AssertValidValue(value)
	_, err := s.Tree.Set(key, value)
	if err != nil {
		panic(err)
	}
}

func (s *Store) Delete(key []byte) {
	defer telemetry.MeasureSince(time.Now(), "store", "iavl", "delete")
	s.Tree.Remove(key)
}

func (s *Store) Iterator(start, end []byte) types.Iterator {
	itr, err := s.Tree.Iterator(start, end, false)
	if err != nil {
		panic(err)
	}
	return itr
}

func (s *Store) ReverseIterator(start, end []byte) types.Iterator {
	itr, err := s.Tree.ReverseIterator(start, end)
	if err != nil {
		panic(err)
	}
	return itr
}
