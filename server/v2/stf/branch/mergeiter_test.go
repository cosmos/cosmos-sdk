package branch

import (
	"reflect"
	"testing"

	corestore "cosmossdk.io/core/store"
)

func TestMergedIterator_Validity(t *testing.T) {
	panics := func(f func()) {
		defer func() {
			r := recover()
			if r == nil {
				t.Error("panic expected")
			}
		}()

		f()
	}

	t.Run("panics when calling key on invalid iter", func(t *testing.T) {
		parent, err := newMemState().Iterator(nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		cache, err := newMemState().Iterator(nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		it := mergeIterators(parent, cache, true)
		panics(func() {
			it.Key()
		})
	})

	t.Run("panics when calling value on invalid iter", func(t *testing.T) {
		parent, err := newMemState().Iterator(nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		cache, err := newMemState().Iterator(nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		it := mergeIterators(parent, cache, true)

		panics(func() {
			it.Value()
		})
	})
}

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
		"parent iterator has one item, child is empty": {
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
		"child has one item, parent is empty": {
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
		"both iterators have same key, child preferred": {
			setup: func() corestore.Iterator {
				parent := newMemState()
				if err := parent.Set([]byte("k1"), []byte("parent-val")); err != nil {
					t.Fatal(err)
				}
				cache := newMemState()
				if err := cache.Set([]byte("k1"), []byte("child-val")); err != nil {
					t.Fatal(err)
				}
				return mergeIterators(must(parent.Iterator(nil, nil)), must(cache.Iterator(nil, nil)), true)
			},
			exp: [][2]string{{"k1", "child-val"}},
		},
		"both iterators have same key, but child value is nil": {
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
		"parent and child are ascending": {
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
		"parent and child are descending": {
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
