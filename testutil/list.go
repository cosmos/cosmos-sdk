package testutil

import (
	"math/rand"
	"sort"

	"golang.org/x/exp/constraints"
)

func RandSliceElem[E any](r *rand.Rand, elems []E) (E, bool) {
	if len(elems) == 0 {
		var e E
		return e, false
	}

	return elems[r.Intn(len(elems))], true
}

// SortSlice sorts a slice of type T elements that implement constraints.Ordered.
// Mutates input slice s
func SortSlice[T constraints.Ordered](s []T) {
	sort.Slice(s, func(i, j int) bool {
		return s[i] < s[j]
	})
}
