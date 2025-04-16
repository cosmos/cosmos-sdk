package orm

import (
	stdErrors "errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/x/group/errors"
)

func TestNewIndexer(t *testing.T) {
	testCases := []struct {
		name        string
		indexerFunc IndexerFunc
		expectErr   bool
		expectedErr string
	}{
		{
			name:        "nil indexer func",
			indexerFunc: nil,
			expectErr:   true,
			expectedErr: "Indexer func must not be nil",
		},
		{
			name:        "all not nil",
			indexerFunc: func(any) ([]any, error) { return nil, nil },
			expectErr:   false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			indexer, err := NewIndexer(tc.indexerFunc)
			if tc.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, indexer)
			}
		})
	}
}

func TestNewUniqueIndexer(t *testing.T) {
	testCases := []struct {
		name        string
		indexerFunc UniqueIndexerFunc
		expectErr   bool
		expectedErr string
	}{
		{
			name:        "nil indexer func",
			indexerFunc: nil,
			expectErr:   true,
			expectedErr: "Indexer func must not be nil",
		},
		{
			name:        "all not nil",
			indexerFunc: func(any) (any, error) { return nil, nil },
			expectErr:   false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			indexer, err := NewUniqueIndexer(tc.indexerFunc)
			if tc.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, indexer)
			}
		})
	}
}

func TestIndexerOnCreate(t *testing.T) {
	var myRowID RowID = EncodeSequence(1)

	specs := map[string]struct {
		srcFunc          IndexerFunc
		expIndexKeys     []any
		expRowIDs        []RowID
		expAddFuncCalled bool
		expErr           error
	}{
		"single key": {
			srcFunc: func(value any) ([]any, error) {
				return []any{uint64(1)}, nil
			},
			expAddFuncCalled: true,
			expIndexKeys:     []any{uint64(1)},
			expRowIDs:        []RowID{myRowID},
		},
		"multi key": {
			srcFunc: func(value any) ([]any, error) {
				return []any{uint64(1), uint64(128)}, nil
			},
			expAddFuncCalled: true,
			expIndexKeys:     []any{uint64(1), uint64(128)},
			expRowIDs:        []RowID{myRowID, myRowID},
		},
		"empty key in slice": {
			srcFunc: func(value any) ([]any, error) {
				return []any{[]byte{}}, nil
			},
			expAddFuncCalled: false,
		},
		"nil key in slice": {
			srcFunc: func(value any) ([]any, error) {
				return []any{nil}, nil
			},
			expErr:           fmt.Errorf("type %T not allowed as key part", nil),
			expAddFuncCalled: false,
		},
		"empty key": {
			srcFunc: func(value any) ([]any, error) {
				return []any{}, nil
			},
			expAddFuncCalled: false,
		},
		"nil key": {
			srcFunc: func(value any) ([]any, error) {
				return nil, nil
			},
			expAddFuncCalled: false,
		},
		"error case": {
			srcFunc: func(value any) ([]any, error) {
				return nil, stdErrors.New("test")
			},
			expErr:           stdErrors.New("test"),
			expAddFuncCalled: false,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			mockPolicy := &addFuncRecorder{}
			idx, err := NewIndexer(spec.srcFunc)
			require.NoError(t, err)
			idx.addFunc = mockPolicy.add

			err = idx.OnCreate(nil, myRowID, nil)
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
	storeKey := storetypes.NewKVStoreKey("test")
	store := prefix.NewStore(ctx.KVStore(storeKey), []byte{multiKeyIndex.prefix})

	specs := map[string]struct {
		srcFunc        IndexerFunc
		expDeletedKeys []RowID
		expErr         error
	}{
		"single key": {
			srcFunc: func(value any) ([]any, error) {
				return []any{uint64(1)}, nil
			},
			expDeletedKeys: []RowID{append(EncodeSequence(1), myRowID...)},
		},
		"multi key": {
			srcFunc: func(value any) ([]any, error) {
				return []any{uint64(1), uint64(128)}, nil
			},
			expDeletedKeys: []RowID{
				append(EncodeSequence(1), myRowID...),
				append(EncodeSequence(128), myRowID...),
			},
		},
		"empty key": {
			srcFunc: func(value any) ([]any, error) {
				return []any{}, nil
			},
		},
		"nil key": {
			srcFunc: func(value any) ([]any, error) {
				return nil, nil
			},
		},
		"empty key in slice": {
			srcFunc: func(value any) ([]any, error) {
				return []any{[]byte{}}, nil
			},
		},
		"nil key in slice": {
			srcFunc: func(value any) ([]any, error) {
				return []any{nil}, nil
			},
			expErr: fmt.Errorf("type %T not allowed as key part", nil),
		},
		"error case": {
			srcFunc: func(value any) ([]any, error) {
				return nil, stdErrors.New("test")
			},
			expErr: stdErrors.New("test"),
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			idx, err := NewIndexer(spec.srcFunc)
			require.NoError(t, err)

			if spec.expErr == nil {
				err = idx.OnCreate(store, myRowID, nil)
				require.NoError(t, err)
				for _, key := range spec.expDeletedKeys {
					require.Equal(t, true, store.Has(key))
				}
			}

			err = idx.OnDelete(store, myRowID, nil)
			if spec.expErr != nil {
				require.Equal(t, spec.expErr, err)
				return
			}
			require.NoError(t, err)
			for _, key := range spec.expDeletedKeys {
				require.Equal(t, false, store.Has(key))
			}
		})
	}
}

func TestIndexerOnUpdate(t *testing.T) {
	myRowID := EncodeSequence(1)

	var multiKeyIndex MultiKeyIndex
	ctx := NewMockContext()
	storeKey := storetypes.NewKVStoreKey("test")
	store := prefix.NewStore(ctx.KVStore(storeKey), []byte{multiKeyIndex.prefix})

	specs := map[string]struct {
		srcFunc        IndexerFunc
		expAddedKeys   []RowID
		expDeletedKeys []RowID
		expErr         error
		addFunc        func(storetypes.KVStore, any, RowID) error
	}{
		"single key - same key, no update": {
			srcFunc: func(value any) ([]any, error) {
				return []any{uint64(1)}, nil
			},
		},
		"single key - different key, replaced": {
			srcFunc: func(value any) ([]any, error) {
				keys := []uint64{1, 2}
				return []any{keys[value.(int)]}, nil
			},
			expAddedKeys: []RowID{
				append(EncodeSequence(2), myRowID...),
			},
			expDeletedKeys: []RowID{
				append(EncodeSequence(1), myRowID...),
			},
		},
		"multi key - same key, no update": {
			srcFunc: func(value any) ([]any, error) {
				return []any{uint64(1), uint64(2)}, nil
			},
		},
		"multi key - replaced": {
			srcFunc: func(value any) ([]any, error) {
				keys := []uint64{1, 2, 3, 4}
				return []any{keys[value.(int)], keys[value.(int)+2]}, nil
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
			srcFunc: func(value any) ([]any, error) {
				return []any{}, nil
			},
		},
		"nil key": {
			srcFunc: func(value any) ([]any, error) {
				return nil, nil
			},
		},
		"empty key in slice": {
			srcFunc: func(value any) ([]any, error) {
				return []any{[]byte{}}, nil
			},
		},
		"nil key in slice": {
			srcFunc: func(value any) ([]any, error) {
				return []any{nil}, nil
			},
			expErr: fmt.Errorf("type %T not allowed as key part", nil),
		},
		"error case with new value": {
			srcFunc: func(value any) ([]any, error) {
				return nil, stdErrors.New("test")
			},
			expErr: stdErrors.New("test"),
		},
		"error case with old value": {
			srcFunc: func(value any) ([]any, error) {
				var err error
				if value.(int)%2 == 1 {
					err = stdErrors.New("test")
				}
				return []any{uint64(1)}, err
			},
			expErr: stdErrors.New("test"),
		},
		"error case on persisting new keys": {
			srcFunc: func(value any) ([]any, error) {
				keys := []uint64{1, 2}
				return []any{keys[value.(int)]}, nil
			},
			addFunc: func(_ storetypes.KVStore, _ any, _ RowID) error {
				return stdErrors.New("test")
			},
			expErr: stdErrors.New("test"),
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			idx, err := NewIndexer(spec.srcFunc)
			require.NoError(t, err)

			if spec.expErr == nil {
				err = idx.OnCreate(store, myRowID, 0)
				require.NoError(t, err)
			}

			if spec.addFunc != nil {
				idx.addFunc = spec.addFunc
			}
			err = idx.OnUpdate(store, myRowID, 1, 0)
			if spec.expErr != nil {
				require.Equal(t, spec.expErr, err)
				return
			}
			require.NoError(t, err)
			for _, key := range spec.expAddedKeys {
				require.Equal(t, true, store.Has(key))
			}
			for _, key := range spec.expDeletedKeys {
				require.Equal(t, false, store.Has(key))
			}
		})
	}
}

func TestUniqueKeyAddFunc(t *testing.T) {
	myRowID := EncodeSequence(1)
	presetKeyPart := []byte("my-preset-key")
	presetKey := append(AddLengthPrefix(presetKeyPart), myRowID...)

	specs := map[string]struct {
		srcKey           []byte
		expErr           *errorsmod.Error
		expExistingEntry []byte
	}{
		"create when not exists": {
			srcKey:           []byte("my-index-key"),
			expExistingEntry: append(AddLengthPrefix([]byte("my-index-key")), myRowID...),
		},
		"error when exists already": {
			srcKey: presetKeyPart,
			expErr: errors.ErrORMUniqueConstraint,
		},
		"nil key not allowed": {
			srcKey: nil,
			expErr: errors.ErrORMInvalidArgument,
		},
		"empty key not allowed": {
			srcKey: []byte{},
			expErr: errors.ErrORMInvalidArgument,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			storeKey := storetypes.NewKVStoreKey("test")
			store := NewMockContext().KVStore(storeKey)
			store.Set(presetKey, []byte{})

			err := uniqueKeysAddFunc(store, spec.srcKey, myRowID)
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
	presetKeyPart := []byte("my-preset-key")
	presetKey := append(AddLengthPrefix(presetKeyPart), myRowID...)

	specs := map[string]struct {
		srcKey           []byte
		expErr           *errorsmod.Error
		expExistingEntry []byte
	}{
		"create when not exists": {
			srcKey:           []byte("my-index-key"),
			expExistingEntry: append(AddLengthPrefix([]byte("my-index-key")), myRowID...),
		},
		"noop when exists already": {
			srcKey:           presetKeyPart,
			expExistingEntry: presetKey,
		},
		"nil key not allowed": {
			srcKey: nil,
			expErr: errors.ErrORMInvalidArgument,
		},
		"empty key not allowed": {
			srcKey: []byte{},
			expErr: errors.ErrORMInvalidArgument,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			storeKey := storetypes.NewKVStoreKey("test")
			store := NewMockContext().KVStore(storeKey)
			store.Set(presetKey, []byte{})

			err := multiKeyAddFunc(store, spec.srcKey, myRowID)
			require.True(t, spec.expErr.Is(err))
			if spec.expErr != nil {
				return
			}
			assert.True(t, store.Has(spec.expExistingEntry))
		})
	}
}

func TestDifference(t *testing.T) {
	specs := map[string]struct {
		srcA      []any
		srcB      []any
		expResult []any
		expErr    bool
	}{
		"all of A": {
			srcA:      []any{"a", "b"},
			srcB:      []any{"c"},
			expResult: []any{"a", "b"},
		},
		"A - B": {
			srcA:      []any{"a", "b"},
			srcB:      []any{"b", "c", "d"},
			expResult: []any{"a"},
		},
		"type in A not allowed": {
			srcA:   []any{1},
			srcB:   []any{"b", "c", "d"},
			expErr: true,
		},
		"type in B not allowed": {
			srcA:   []any{"b", "c", "d"},
			srcB:   []any{1},
			expErr: true,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			got, err := difference(spec.srcA, spec.srcB)
			if spec.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, spec.expResult, got)
			}
		})
	}
}

func TestPruneEmptyKeys(t *testing.T) {
	specs := map[string]struct {
		srcFunc   IndexerFunc
		expResult []any
		expError  error
	}{
		"non empty": {
			srcFunc: func(v any) ([]any, error) {
				return []any{uint64(0), uint64(1)}, nil
			},
			expResult: []any{uint64(0), uint64(1)},
		},
		"empty": {
			srcFunc: func(v any) ([]any, error) {
				return []any{}, nil
			},
			expResult: []any{},
		},
		"nil": {
			srcFunc: func(v any) ([]any, error) {
				return nil, nil
			},
		},
		"empty in the beginning": {
			srcFunc: func(v any) ([]any, error) {
				return []any{[]byte{}, uint64(0), uint64(1)}, nil
			},
			expResult: []any{uint64(0), uint64(1)},
		},
		"empty in the middle": {
			srcFunc: func(v any) ([]any, error) {
				return []any{uint64(0), []byte{}, uint64(1)}, nil
			},
			expResult: []any{uint64(0), uint64(1)},
		},
		"empty at the end": {
			srcFunc: func(v any) ([]any, error) {
				return []any{uint64(0), uint64(1), []byte{}}, nil
			},
			expResult: []any{uint64(0), uint64(1)},
		},
		"error passed": {
			srcFunc: func(v any) ([]any, error) {
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
	secondaryIndexKeys []any
	rowIDs             []RowID
	called             bool
}

func (c *addFuncRecorder) add(_ storetypes.KVStore, key any, rowID RowID) error {
	c.secondaryIndexKeys = append(c.secondaryIndexKeys, key)
	c.rowIDs = append(c.rowIDs, rowID)
	c.called = true
	return nil
}
