package branch

import (
	"reflect"
	"testing"

	corestore "cosmossdk.io/core/store"
)

func TestMergedIterator_Next(t *testing.T) {
	specs := map[string]struct {
		setup func() corestore.Iterator
		exp   [][2]string
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
			exp: [][2]string{{"k1", "1"}},
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
			exp: [][2]string{{"k1", "1"}},
		},
		"both iterators have same key, cache preferred": {
			setup: func() corestore.Iterator {
				parent := newMemState()
				if err := parent.Set([]byte("k1"), []byte("parent-val")); err != nil {
					t.Fatal(err)
				}
				cache := newMemState()
				if err := cache.Set([]byte("k1"), []byte("cache-val")); err != nil {
					t.Fatal(err)
				}
				return mergeIterators(must(parent.Iterator(nil, nil)), must(cache.Iterator(nil, nil)), true)
			},
			exp: [][2]string{{"k1", "cache-val"}},
		},
		"both iterators have same key, but cache value is nil": {
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
			exp: [][2]string{{"k1", "v1"}, {"k2", "v2"}, {"k3", "v3"}, {"k4", "v4"}},
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
			exp: [][2]string{{"k4", "v4"}, {"k3", "v3"}, {"k2", "v2"}, {"k1", "v1"}},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			var got [][2]string
			for iter := spec.setup(); iter.Valid(); iter.Next() {
				got = append(got, [2]string{string(iter.Key()), string(iter.Value())})
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
