package orm

import (
	stdErrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
)

func TestIndexerOnCreate(t *testing.T) {
	var myRowID RowID = EncodeSequence(1)

	specs := map[string]struct {
		srcFunc          IndexerFunc
		expIndexKeys     []RowID
		expRowIDs        []RowID
		expAddFuncCalled bool
		expErr           error
	}{
		"single key": {
			srcFunc: func(value interface{}) ([]RowID, error) {
				return []RowID{{0, 0, 0, 0, 0, 0, 0, 1}}, nil
			},
			expAddFuncCalled: true,
			expIndexKeys:     []RowID{{0, 0, 0, 0, 0, 0, 0, 1}},
			expRowIDs:        []RowID{myRowID},
		},
		"multi key": {
			srcFunc: func(value interface{}) ([]RowID, error) {
				return []RowID{{0, 0, 0, 0, 0, 0, 0, 1}, {1, 0, 0, 0, 0, 0, 0, 0}}, nil
			},
			expAddFuncCalled: true,
			expIndexKeys:     []RowID{{0, 0, 0, 0, 0, 0, 0, 1}, {1, 0, 0, 0, 0, 0, 0, 0}},
			expRowIDs:        []RowID{myRowID, myRowID},
		},
		"empty key in slice": {
			srcFunc: func(value interface{}) ([]RowID, error) {
				return []RowID{{}}, nil
			},
			expAddFuncCalled: false,
		},
		"nil key in slice": {
			srcFunc: func(value interface{}) ([]RowID, error) {
				return []RowID{nil}, nil
			},
			expAddFuncCalled: false,
		},
		"empty key": {
			srcFunc: func(value interface{}) ([]RowID, error) {
				return []RowID{}, nil
			},
			expAddFuncCalled: false,
		},
		"nil key": {
			srcFunc: func(value interface{}) ([]RowID, error) {
				return nil, nil
			},
			expAddFuncCalled: false,
		},
		"error case": {
			srcFunc: func(value interface{}) ([]RowID, error) {
				return nil, stdErrors.New("test")
			},
			expErr:           stdErrors.New("test"),
			expAddFuncCalled: false,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			mockPolicy := &addFuncRecorder{}
			idx := NewIndexer(spec.srcFunc, Max255DynamicLengthIndexKeyCodec{})
			idx.addFunc = mockPolicy.add

			err := idx.OnCreate(nil, myRowID, nil)
			if spec.expErr != nil {
				require.Equal(t, spec.expErr, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, spec.expIndexKeys, mockPolicy.secondaryIndexKeys)
			assert.Equal(t, spec.expRowIDs, mockPolicy.rowIDs)
			assert.Equal(t, spec.expAddFuncCalled, mockPolicy.called)
		})
	}
}

func TestIndexerOnDelete(t *testing.T) {
	myRowID := EncodeSequence(1)

	var multiKeyIndex MultiKeyIndex
	ctx := NewMockContext()
	storeKey := sdk.NewKVStoreKey("test")
	store := prefix.NewStore(ctx.KVStore(storeKey), []byte{multiKeyIndex.prefix})

	specs := map[string]struct {
		srcFunc      IndexerFunc
		expIndexKeys []RowID
		expErr       error
	}{
		"single key": {
			srcFunc: func(value interface{}) ([]RowID, error) {
				return []RowID{{0, 0, 0, 0, 0, 0, 0, 1}}, nil
			},
			expIndexKeys: []RowID{append([]byte{0, 0, 0, 0, 0, 0, 0, 1}, myRowID...)},
		},
		"multi key": {
			srcFunc: func(value interface{}) ([]RowID, error) {
				return []RowID{{0, 0, 0, 0, 0, 0, 0, 1}, {1, 0, 0, 0, 0, 0, 0, 0}}, nil
			},
			expIndexKeys: []RowID{
				append([]byte{0, 0, 0, 0, 0, 0, 0, 1}, myRowID...),
				append([]byte{1, 0, 0, 0, 0, 0, 0, 0}, myRowID...),
			},
		},
		"empty key": {
			srcFunc: func(value interface{}) ([]RowID, error) {
				return []RowID{}, nil
			},
		},
		"nil key": {
			srcFunc: func(value interface{}) ([]RowID, error) {
				return nil, nil
			},
		},
		"empty key in slice": {
			srcFunc: func(value interface{}) ([]RowID, error) {
				return []RowID{{}}, nil
			},
		},
		"nil key in slice": {
			srcFunc: func(value interface{}) ([]RowID, error) {
				return []RowID{nil}, nil
			},
		},
		"error case": {
			srcFunc: func(value interface{}) ([]RowID, error) {
				return nil, stdErrors.New("test")
			},
			expErr: stdErrors.New("test"),
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			codec := FixLengthIndexKeys(EncodedSeqLength)
			idx := NewIndexer(spec.srcFunc, codec)
			err := idx.OnDelete(store, myRowID, nil)
			if spec.expErr != nil {
				require.Equal(t, spec.expErr, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestIndexerOnUpdate(t *testing.T) {
	myRowID := EncodeSequence(1)
	codec := FixLengthIndexKeys(EncodedSeqLength)

	var multiKeyIndex MultiKeyIndex
	ctx := NewMockContext()
	storeKey := sdk.NewKVStoreKey("test")
	store := prefix.NewStore(ctx.KVStore(storeKey), []byte{multiKeyIndex.prefix})

	specs := map[string]struct {
		srcFunc        IndexerFunc
		mockStore      *updateKVStoreRecorder
		expAddedKeys   []RowID
		expDeletedKeys []RowID
		expErr         error
		addFunc        func(sdk.KVStore, IndexKeyCodec, []byte, RowID) error
	}{
		"single key - same key, no update": {
			srcFunc: func(value interface{}) ([]RowID, error) {
				return []RowID{EncodeSequence(1)}, nil
			},
		},
		"single key - different key, replaced": {
			srcFunc: func(value interface{}) ([]RowID, error) {
				keys := []RowID{EncodeSequence(1), EncodeSequence(2)}
				return []RowID{keys[value.(int)]}, nil
			},
			expAddedKeys: []RowID{
				append(EncodeSequence(2), myRowID...),
			},
			expDeletedKeys: []RowID{
				append(EncodeSequence(1), myRowID...),
			},
		},
		"multi key - same key, no update": {
			srcFunc: func(value interface{}) ([]RowID, error) {
				return []RowID{EncodeSequence(1), EncodeSequence(2)}, nil
			},
		},
		"multi key - replaced": {
			srcFunc: func(value interface{}) ([]RowID, error) {
				keys := []RowID{EncodeSequence(1), EncodeSequence(2), EncodeSequence(3), EncodeSequence(4)}
				return []RowID{keys[value.(int)], keys[value.(int)+2]}, nil
			},
			expAddedKeys: []RowID{
				append(EncodeSequence(2), myRowID...),
				append(EncodeSequence(4), myRowID...),
			},
			expDeletedKeys: []RowID{
				append(EncodeSequence(1), myRowID...),
				append(EncodeSequence(3), myRowID...),
			},
		},
		"empty key": {
			srcFunc: func(value interface{}) ([]RowID, error) {
				return []RowID{}, nil
			},
		},
		"nil key": {
			srcFunc: func(value interface{}) ([]RowID, error) {
				return nil, nil
			},
		},
		"empty key in slice": {
			srcFunc: func(value interface{}) ([]RowID, error) {
				return []RowID{{}}, nil
			},
		},
		"nil key in slice": {
			srcFunc: func(value interface{}) ([]RowID, error) {
				return []RowID{nil}, nil
			},
		},
		"error case with new value": {
			srcFunc: func(value interface{}) ([]RowID, error) {
				return nil, stdErrors.New("test")
			},
			expErr: stdErrors.New("test"),
		},
		"error case with old value": {
			srcFunc: func(value interface{}) ([]RowID, error) {
				var err error
				if value.(int)%2 == 1 {
					err = stdErrors.New("test")
				}
				return []RowID{myRowID}, err
			},
			expErr: stdErrors.New("test"),
		},
		"error case on persisting new keys": {
			srcFunc: func(value interface{}) ([]RowID, error) {
				keys := []RowID{EncodeSequence(1), EncodeSequence(2)}
				return []RowID{keys[value.(int)]}, nil
			},
			addFunc: func(_ sdk.KVStore, _ IndexKeyCodec, _ []byte, _ RowID) error {
				return stdErrors.New("test")
			},
			expErr: stdErrors.New("test"),
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			idx := NewIndexer(spec.srcFunc, codec)
			if spec.addFunc != nil {
				idx.addFunc = spec.addFunc
			}
			err := idx.OnUpdate(store, myRowID, 1, 0)
			if spec.expErr != nil {
				require.Equal(t, spec.expErr, err)
				return
			}
			require.NoError(t, err)

		})
	}
}

func TestUniqueKeyAddFunc(t *testing.T) {
	myRowID := EncodeSequence(1)
	myPresetKey := append([]byte("my-preset-key"), myRowID...)

	specs := map[string]struct {
		srcKey           []byte
		expErr           *errors.Error
		expExistingEntry []byte
	}{

		"create when not exists": {
			srcKey:           []byte("my-index-key"),
			expExistingEntry: append([]byte("my-index-key"), myRowID...),
		},
		"error when exists already": {
			srcKey: []byte("my-preset-key"),
			expErr: ErrUniqueConstraint,
		},
		"nil key not allowed": {
			srcKey: nil,
			expErr: ErrArgument,
		},
		"empty key not allowed": {
			srcKey: []byte{},
			expErr: ErrArgument,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			storeKey := sdk.NewKVStoreKey("test")
			store := NewMockContext().KVStore(storeKey)
			store.Set(myPresetKey, []byte{})

			codec := FixLengthIndexKeys(EncodedSeqLength)
			err := uniqueKeysAddFunc(store, codec, spec.srcKey, myRowID)
			require.True(t, spec.expErr.Is(err))
			if spec.expErr != nil {
				return
			}
			assert.True(t, store.Has(spec.expExistingEntry), "not found")
		})
	}
}

func TestMultiKeyAddFunc(t *testing.T) {
	myRowID := EncodeSequence(1)
	myPresetKey := append([]byte("my-preset-key"), myRowID...)

	specs := map[string]struct {
		srcKey           []byte
		expErr           *errors.Error
		expExistingEntry []byte
	}{

		"create when not exists": {
			srcKey:           []byte("my-index-key"),
			expExistingEntry: append([]byte("my-index-key"), myRowID...),
		},
		"noop when exists already": {
			srcKey:           []byte("my-preset-key"),
			expExistingEntry: myPresetKey,
		},
		"nil key not allowed": {
			srcKey: nil,
			expErr: ErrArgument,
		},
		"empty key not allowed": {
			srcKey: []byte{},
			expErr: ErrArgument,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			storeKey := sdk.NewKVStoreKey("test")
			store := NewMockContext().KVStore(storeKey)
			store.Set(myPresetKey, []byte{})

			codec := FixLengthIndexKeys(EncodedSeqLength)
			err := multiKeyAddFunc(store, codec, spec.srcKey, myRowID)
			require.True(t, spec.expErr.Is(err))
			if spec.expErr != nil {
				return
			}
			assert.True(t, store.Has(spec.expExistingEntry))
		})
	}
}

func TestDifference(t *testing.T) {
	asByte := func(s []string) []RowID {
		r := make([]RowID, len(s))
		for i := 0; i < len(s); i++ {
			r[i] = []byte(s[i])
		}
		return r
	}

	specs := map[string]struct {
		srcA      []string
		srcB      []string
		expResult []RowID
	}{
		"all of A": {
			srcA:      []string{"a", "b"},
			srcB:      []string{"c"},
			expResult: []RowID{[]byte("a"), []byte("b")},
		},
		"A - B": {
			srcA:      []string{"a", "b"},
			srcB:      []string{"b", "c", "d"},
			expResult: []RowID{[]byte("a")},
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			got := difference(asByte(spec.srcA), asByte(spec.srcB))
			assert.Equal(t, spec.expResult, got)
		})
	}
}

func TestPruneEmptyKeys(t *testing.T) {
	specs := map[string]struct {
		srcFunc   IndexerFunc
		expResult []RowID
		expError  error
	}{
		"non empty": {
			srcFunc: func(v interface{}) ([]RowID, error) {
				return []RowID{{0}, {1}}, nil
			},
			expResult: []RowID{{0}, {1}},
		},
		"empty": {
			srcFunc: func(v interface{}) ([]RowID, error) {
				return []RowID{}, nil
			},
			expResult: []RowID{},
		},
		"nil": {
			srcFunc: func(v interface{}) ([]RowID, error) {
				return nil, nil
			},
		},
		"nil in the beginning": {
			srcFunc: func(v interface{}) ([]RowID, error) {
				return []RowID{nil, {0}, {1}}, nil
			},
			expResult: []RowID{{0}, {1}},
		},
		"nil in the middle": {
			srcFunc: func(v interface{}) ([]RowID, error) {
				return []RowID{{0}, nil, {1}}, nil
			},
			expResult: []RowID{{0}, {1}},
		},
		"nil at the end": {
			srcFunc: func(v interface{}) ([]RowID, error) {
				return []RowID{{0}, nil, {1}}, nil
			},
			expResult: []RowID{{0}, {1}},
		},
		"empty in the beginning": {
			srcFunc: func(v interface{}) ([]RowID, error) {
				return []RowID{{}, {0}, {1}}, nil
			},
			expResult: []RowID{{0}, {1}},
		},
		"empty in the middle": {
			srcFunc: func(v interface{}) ([]RowID, error) {
				return []RowID{{0}, {}, {1}}, nil
			},
			expResult: []RowID{{0}, {1}},
		},
		"empty at the end": {
			srcFunc: func(v interface{}) ([]RowID, error) {
				return []RowID{{0}, {}, {1}}, nil
			},
			expResult: []RowID{{0}, {1}},
		},
		"error passed": {
			srcFunc: func(v interface{}) ([]RowID, error) {
				return nil, stdErrors.New("test")
			},
			expError: stdErrors.New("test"),
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			r, err := pruneEmptyKeys(spec.srcFunc)(nil)
			require.Equal(t, spec.expError, err)
			if spec.expError != nil {
				return
			}
			assert.Equal(t, spec.expResult, r)
		})
	}
}

type addFuncRecorder struct {
	secondaryIndexKeys []RowID
	rowIDs             []RowID
	called             bool
}

func (c *addFuncRecorder) add(_ sdk.KVStore, _ IndexKeyCodec, key []byte, rowID RowID) error {
	c.secondaryIndexKeys = append(c.secondaryIndexKeys, key)
	c.rowIDs = append(c.rowIDs, rowID)
	c.called = true
	return nil
}

type deleteKVStoreRecorder struct {
	AlwaysPanicKVStore
	deletes []RowID
}

func (m *deleteKVStoreRecorder) Delete(key []byte) {
	m.deletes = append(m.deletes, key)
}

type updateKVStoreRecorder struct {
	deleteKVStoreRecorder
	stored    tuples
	hasResult bool
}

func (u *updateKVStoreRecorder) Set(key, value []byte) {
	u.stored = append(u.stored, tuple{key, value})
}

func (u updateKVStoreRecorder) Has(key []byte) bool {
	return u.hasResult
}

type tuple struct {
	key, val []byte
}

type tuples []tuple

func (t tuples) Keys() []RowID {
	if t == nil {
		return nil
	}
	r := make([]RowID, len(t))
	for i, v := range t {
		r[i] = v.key
	}
	return r
}
