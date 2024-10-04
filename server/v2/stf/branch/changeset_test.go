package branch

import (
	"testing"
)

func TestMemIteratorWithWriteToRebalance(t *testing.T) {
	t.Run("iter is invalid after close", func(t *testing.T) {
		cs := newChangeSet()
		for i := byte(0); i < 32; i++ {
			cs.set([]byte{0, i}, []byte{i})
		}

		it, err := cs.iterator(nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		err = it.Close()
		if err != nil {
			t.Fatal(err)
		}

		if it.Valid() {
			t.Fatal("iterator must be invalid")
		}
	})
}

func TestKeyInRange(t *testing.T) {
	specs := map[string]struct {
		mi  *memIterator
		src []byte
		exp bool
	}{
		"equal start": {
			mi:  &memIterator{ascending: true, start: []byte{0}, end: []byte{2}},
			src: []byte{0},
			exp: true,
		},
		"equal end": {
			mi:  &memIterator{ascending: true, start: []byte{0}, end: []byte{2}},
			src: []byte{2},
			exp: false,
		},
		"between": {
			mi:  &memIterator{ascending: true, start: []byte{0}, end: []byte{2}},
			src: []byte{1},
			exp: true,
		},
		"equal start - open end": {
			mi:  &memIterator{ascending: true, start: []byte{0}},
			src: []byte{0},
			exp: true,
		},
		"greater start - open end": {
			mi:  &memIterator{ascending: true, start: []byte{0}},
			src: []byte{2},
			exp: true,
		},
		"equal end - open start": {
			mi:  &memIterator{ascending: true, end: []byte{2}},
			src: []byte{2},
			exp: false,
		},
		"smaller end - open start": {
			mi:  &memIterator{ascending: true, end: []byte{2}},
			src: []byte{1},
			exp: true,
		},
	}
	for name, spec := range specs {
		for _, asc := range []bool{true, false} {
			order := "asc_"
			if !asc {
				order = "desc_"
			}
			t.Run(order+name, func(t *testing.T) {
				spec.mi.ascending = asc
				got := spec.mi.keyInRange(spec.src)
				if spec.exp != got {
					t.Errorf("expected %v, got %v", spec.exp, got)
				}
			})
		}
	}
}
