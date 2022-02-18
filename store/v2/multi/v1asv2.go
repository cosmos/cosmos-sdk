package multi

import (
	"io"

	"github.com/cosmos/cosmos-sdk/db"
	dbutil "github.com/cosmos/cosmos-sdk/internal/db"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	v1 "github.com/cosmos/cosmos-sdk/store/types"
	v2 "github.com/cosmos/cosmos-sdk/store/v2"
)

var (
	_ v2.CommitMultiStore = (*store1as2)(nil)
	_ v2.Queryable        = (*store1as2)(nil)
	_ v2.CacheMultiStore  = cacheStore1as2{}
)

type store1as2 struct {
	*rootmulti.Store
	database *dbutil.TmdbAdapter
}

type cacheStore1as2 struct {
	v1.CacheMultiStore
}

// NewV1MultiStoreAsV2 constructs a v1 CommitMultiStore from v2.StoreParams
func NewV1MultiStoreAsV2(database db.DBConnection, opts StoreParams) (v2.CommitMultiStore, error) {
	adapter := dbutil.ConnectionAsTmdb(database)
	cms := rootmulti.NewStore(adapter)
	for name, typ := range opts.StoreSchema {
		switch typ {
		case v2.StoreTypePersistent:
			typ = v1.StoreTypeIAVL
		}
		skey, err := opts.storeKey(name)
		if err != nil {
			return nil, err
		}
		cms.MountStoreWithDB(skey, typ, nil)
	}
	cms.SetPruning(opts.Pruning)
	err := cms.SetInitialVersion(int64(opts.InitialVersion))
	if err != nil {
		return nil, err
	}
	err = cms.LoadLatestVersionAndUpgrade(opts.Upgrades)
	if err != nil {
		return nil, err
	}
	for skey, ls := range opts.listeners {
		cms.AddListeners(skey, ls)
	}
	cms.SetTracer(opts.TraceWriter)
	cms.SetTracingContext(opts.TraceContext)
	return &store1as2{cms, adapter}, nil
}

// MultiStore

func (s *store1as2) CacheWrap() v2.CacheMultiStore {
	return cacheStore1as2{s.CacheMultiStore()}
}
func (s *store1as2) GetVersion(ver int64) (v2.MultiStore, error) {
	ret, err := s.CacheMultiStoreWithVersion(ver)
	return cacheStore1as2{ret}, err
}

// CommitMultiStore

func (s *store1as2) Close() error {
	return s.database.CloseTx()
}

func (s *store1as2) Commit() v2.CommitID {
	ret := s.Store.Commit()
	_, err := s.database.Commit()
	if err != nil {
		panic(err)
	}
	return ret
}

func (s *store1as2) SetInitialVersion(ver uint64) error {
	return s.Store.SetInitialVersion(int64(ver))
}

func (s *store1as2) SetTracer(w io.Writer)                { s.SetTracer(w) }
func (s *store1as2) SetTracingContext(tc v2.TraceContext) { s.SetTracingContext(tc) }

func (s *store1as2) GetAllVersions() []int { panic("unsupported: GetAllVersions") }

// CacheMultiStore

func (s cacheStore1as2) CacheWrap() v2.CacheMultiStore {
	return cacheStore1as2{s.CacheMultiStore.CacheMultiStore()}
}

func (s cacheStore1as2) SetTracer(w io.Writer)                { s.SetTracer(w) }
func (s cacheStore1as2) SetTracingContext(tc v2.TraceContext) { s.SetTracingContext(tc) }
