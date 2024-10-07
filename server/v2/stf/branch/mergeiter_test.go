package branch

import (
	corestore "cosmossdk.io/core/store"
	"reflect"
	"testing"
)

func TestMergedIterator_Next(t *testing.T) {
	specs := map[string]struct {
		setup func() corestore.Iterator
		exp   []string
	}{
		"both iterators are empty": {
			setup: func() corestore.Iterator {
				parent := newMemState()
				cache := newMemState()
				return mergeIterators(must(parent.Iterator(nil, nil)), must(cache.Iterator(nil, nil)), true)
			},
		},
		"parent iterator has one item, cache is empty": {
			setup: func() corestore.Iterator {
				parent := newMemState()
				if err := parent.Set([]byte("k1"), []byte("1")); err != nil {
					t.Fatal(err)
				}
				cache := newMemState()
				return mergeIterators(must(parent.Iterator(nil, nil)), must(cache.Iterator(nil, nil)), true)
			},
			exp: []string{"k1"},
		},
		"cache has one item, parent is empty": {
			setup: func() corestore.Iterator {
				parent := newMemState()
				cache := newMemState()
				if err := cache.Set([]byte("k1"), []byte("1")); err != nil {
					t.Fatal(err)
				}
				return mergeIterators(must(parent.Iterator(nil, nil)), must(cache.Iterator(nil, nil)), true)
			},
			exp: []string{"k1"},
		},
		"both iterators have items, but cache value is nil": {
			setup: func() corestore.Iterator {
				parent := newMemState()
				if err := parent.Set([]byte("k1"), []byte("1")); err != nil {
					t.Fatal(err)
				}
				cache := newMemState()
				if err := cache.Set([]byte("k1"), nil); err != nil {
					t.Fatal(err)
				}
				return mergeIterators(must(parent.Iterator(nil, nil)), must(cache.Iterator(nil, nil)), true)
			},
		},
		"parent and cache are ascending": {
			setup: func() corestore.Iterator {
				parent := newMemState()
				if err := parent.Set([]byte("k2"), []byte("v2")); err != nil {
					t.Fatal(err)
				}
				if err := parent.Set([]byte("k3"), []byte("v3")); err != nil {
					t.Fatal(err)
				}
				cache := newMemState()
				if err := cache.Set([]byte("k1"), []byte("v1")); err != nil {
					t.Fatal(err)
				}
				if err := cache.Set([]byte("k4"), []byte("v4")); err != nil {
					t.Fatal(err)
				}
				return mergeIterators(must(parent.Iterator(nil, nil)), must(cache.Iterator(nil, nil)), true)
			},
			exp: []string{"k1", "k2", "k3", "k4"},
		},
		"parent and cache are descending": {
			setup: func() corestore.Iterator {
				parent := newMemState()
				if err := parent.Set([]byte("k3"), []byte("v3")); err != nil {
					t.Fatal(err)
				}
				if err := parent.Set([]byte("k2"), []byte("v2")); err != nil {
					t.Fatal(err)
				}
				cache := newMemState()
				if err := cache.Set([]byte("k4"), []byte("v4")); err != nil {
					t.Fatal(err)
				}
				if err := cache.Set([]byte("k1"), []byte("v1")); err != nil {
					t.Fatal(err)
				}
				return mergeIterators(must(parent.ReverseIterator(nil, nil)), must(cache.ReverseIterator(nil, nil)), false)
			},
			exp: []string{"k4", "k3", "k2", "k1"},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			var got []string
			for iter := spec.setup(); iter.Valid(); iter.Next() {
				got = append(got, string(iter.Key()))
			}
			if !reflect.DeepEqual(spec.exp, got) {
				t.Errorf("expected: %#v, got: %#v", spec.exp, got)
			}
		})
	}
}

func must[T any](r T, err error) T {
	if err != nil {
		panic(err)
	}
	return r
}
