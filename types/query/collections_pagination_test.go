package query

import (
	"context"
	"testing"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	db "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"
)

func TestCollectionPagination(t *testing.T) {
	sk, ctx := deps()
	sb := collections.NewSchemaBuilder(sk)
	m := collections.NewMap(sb, collections.NewPrefix(0), "_", collections.Uint64Key, collections.Uint64Value)

	for i := uint64(0); i < 300; i++ {
		require.NoError(t, m.Set(ctx, i, i))
	}

	createResults := func(from, to uint64) []collections.KeyValue[uint64, uint64] {
		var res []collections.KeyValue[uint64, uint64]
		if from <= to {
			for i := from; i <= to; i++ {
				res = append(res, collections.KeyValue[uint64, uint64]{
					Key:   i,
					Value: i,
				})
			}
		} else {
			for i := from; i >= to; i-- {
				res = append(res, collections.KeyValue[uint64, uint64]{
					Key:   i,
					Value: i,
				})
			}
		}
		return res
	}

	encodeKey := func(key uint64) []byte {
		b, err := encodeCollKey[uint64, uint64](m, key)
		require.NoError(t, err)
		return b
	}

	type test struct {
		req        *PageRequest
		expResp    *PageResponse
		filter     func(key uint64, value uint64) bool
		expResults []collections.KeyValue[uint64, uint64]
		wantErr    error
	}

	tcs := map[string]test{
		"nil pagination": {
			req: nil,
			expResp: &PageResponse{
				NextKey: encodeKey(100),
				Total:   300,
			},
			expResults: createResults(0, 99),
		},
		"with key and limit": {
			req: &PageRequest{
				Key:   encodeKey(100),
				Limit: 149,
			},
			expResp: &PageResponse{
				NextKey: encodeKey(249),
			},
			expResults: createResults(100, 248),
		},
		"with reverse": {
			req: &PageRequest{
				Reverse: true,
			},
			expResp: &PageResponse{
				NextKey: encodeKey(199),
				Total:   300,
			},
			expResults: createResults(299, 200),
		},
		"with offset and count total": {
			req: &PageRequest{
				Offset:     50,
				Limit:      100,
				CountTotal: true,
			},
			expResp: &PageResponse{
				NextKey: encodeKey(150),
				Total:   300,
			},
			expResults: createResults(50, 149),
		},
		"filtered no key": {
			req: &PageRequest{
				Limit: 3,
			},
			expResp: &PageResponse{
				NextKey: encodeKey(5),
			},
			filter: func(key uint64, value uint64) bool {
				return key%2 == 0
			},
			expResults: []collections.KeyValue[uint64, uint64]{
				{Key: 0, Value: 0},
				{Key: 2, Value: 2},
				{Key: 4, Value: 4},
			},
		},
		"filtered with key": {
			req: &PageRequest{
				Key:   encodeKey(2),
				Limit: 3,
			},
			expResp: &PageResponse{
				NextKey: encodeKey(7),
			},
			filter: func(key uint64, value uint64) bool {
				return key%2 == 0
			},
			expResults: []collections.KeyValue[uint64, uint64]{
				{Key: 2, Value: 2},
				{Key: 4, Value: 4},
				{Key: 6, Value: 6},
			},
		},
		"error offset > total items": {
			req:     &PageRequest{Offset: 500},
			wantErr: collections.ErrInvalidIterator,
		},
	}

	for name, tc := range tcs {
		tc := tc
		t.Run(name, func(t *testing.T) {
			gotResults, gotResponse, err := CollectionFilteredPaginate(ctx, m, tc.req, tc.filter)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expResults, gotResults)
			require.Equal(t, tc.expResp, gotResponse)
		})
	}
}

type testStore struct {
	db db.DB
}

func (t testStore) OpenKVStore(ctx context.Context) store.KVStore {
	return t
}

func (t testStore) Get(key []byte) []byte {
	res, err := t.db.Get(key)
	if err != nil {
		panic(err)
	}
	return res
}

func (t testStore) Has(key []byte) bool {
	res, err := t.db.Has(key)
	if err != nil {
		panic(err)
	}
	return res
}

func (t testStore) Set(key, value []byte) {
	err := t.db.Set(key, value)
	if err != nil {
		panic(err)
	}
}

func (t testStore) Delete(key []byte) {
	err := t.db.Delete(key)
	if err != nil {
		panic(err)
	}
}

func (t testStore) Iterator(start, end []byte) store.Iterator {
	res, err := t.db.Iterator(start, end)
	if err != nil {
		panic(err)
	}
	return res
}

func (t testStore) ReverseIterator(start, end []byte) store.Iterator {
	res, err := t.db.ReverseIterator(start, end)
	if err != nil {
		panic(err)
	}
	return res
}

var _ store.KVStore = testStore{}

func deps() (store.KVStoreService, context.Context) {
	kv := db.NewMemDB()
	return &testStore{kv}, context.Background()
}
