package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"
)

func TestStoreUpgrades(t *testing.T) {
	t.Parallel()
	type toAdd struct {
		key string
	}
	type toDelete struct {
		key    string
		delete bool
	}
	type toRename struct {
		newkey string
		result string
	}

	cases := map[string]struct {
		upgrades     *StoreUpgrades
		expectAdd    []toAdd
		expectDelete []toDelete
		expectRename []toRename
	}{
		"empty upgrade": {
			expectDelete: []toDelete{{"foo", false}},
			expectRename: []toRename{{"foo", ""}},
		},
		"simple matches": {
			upgrades: &StoreUpgrades{
				Deleted: []string{"foo"},
				Renamed: []StoreRename{{"bar", "baz"}},
			},
			expectDelete: []toDelete{{"foo", true}, {"bar", false}, {"baz", false}},
			expectRename: []toRename{{"foo", ""}, {"bar", ""}, {"baz", "bar"}},
		},
		"many data points": {
			upgrades: &StoreUpgrades{
				Added:   []string{"foo", "bar", "baz"},
				Deleted: []string{"one", "two", "three", "four", "five"},
				Renamed: []StoreRename{{"old", "new"}, {"white", "blue"}, {"black", "orange"}, {"fun", "boring"}},
			},
			expectAdd:    []toAdd{{"foo"}, {"bar"}, {"baz"}},
			expectDelete: []toDelete{{"four", true}, {"six", false}, {"baz", false}},
			expectRename: []toRename{{"white", ""}, {"blue", "white"}, {"boring", "fun"}, {"missing", ""}},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			for _, r := range tc.expectAdd {
				assert.Equal(t, tc.upgrades.IsAdded(r.key), true)
			}
			for _, d := range tc.expectDelete {
				assert.Equal(t, tc.upgrades.IsDeleted(d.key), d.delete)
			}
			for _, r := range tc.expectRename {
				assert.Equal(t, tc.upgrades.RenamedFrom(r.newkey), r.result)
			}
		})
	}
}

func TestCommitID(t *testing.T) {
	t.Parallel()
	require.True(t, CommitID{}.IsZero())
	require.False(t, CommitID{Version: int64(1)}.IsZero())
	require.False(t, CommitID{Hash: []byte("x")}.IsZero())
	require.Equal(t, "CommitID{[120 120 120 120]:64}", CommitID{Version: int64(100), Hash: []byte("xxxx")}.String())
}

func TestKVStoreKey(t *testing.T) {
	t.Parallel()
	key := NewKVStoreKey("test")
	require.Equal(t, "test", key.name)
	require.Equal(t, key.name, key.Name())
	require.Equal(t, fmt.Sprintf("KVStoreKey{%p, test}", key), key.String())
}

func TestNilKVStoreKey(t *testing.T) {
	t.Parallel()

	require.Panics(t, func() {
		_ = NewKVStoreKey("")
	}, "setting an empty key should panic")
}

func TestTransientStoreKey(t *testing.T) {
	t.Parallel()
	key := NewTransientStoreKey("test")
	require.Equal(t, "test", key.name)
	require.Equal(t, key.name, key.Name())
	require.Equal(t, fmt.Sprintf("TransientStoreKey{%p, test}", key), key.String())
}

func TestObjectStoreKey(t *testing.T) {
	t.Parallel()
	key := NewObjectStoreKey("test")
	require.Equal(t, "test", key.name)
	require.Equal(t, key.name, key.Name())
	require.Equal(t, fmt.Sprintf("ObjectStoreKey{%p, test}", key), key.String())
}

func TestMemoryStoreKey(t *testing.T) {
	t.Parallel()
	key := NewMemoryStoreKey("test")
	require.Equal(t, "test", key.name)
	require.Equal(t, key.name, key.Name())
	require.Equal(t, fmt.Sprintf("MemoryStoreKey{%p, test}", key), key.String())
}

func TestNewTransientStoreKeys(t *testing.T) {
	assert.DeepEqual(t, map[string]*TransientStoreKey{}, NewTransientStoreKeys())
	assert.DeepEqual(t, 1, len(NewTransientStoreKeys("one")))
}

func TestNewObjectStoreKeys(t *testing.T) {
	assert.DeepEqual(t, map[string]*ObjectStoreKey{}, NewObjectStoreKeys())
	assert.DeepEqual(t, 1, len(NewObjectStoreKeys("one")))
}

func TestNewInfiniteGasMeter(t *testing.T) {
	gm := NewInfiniteGasMeter()
	require.NotNil(t, gm)
}

func TestStoreTypes(t *testing.T) {
	assert.DeepEqual(t, InclusiveEndBytes([]byte("endbytes")), InclusiveEndBytes([]byte("endbytes")))
}
