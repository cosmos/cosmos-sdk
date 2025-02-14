package util

import (
	"cmp"
	"maps"
	"slices"
)

// IterateMapOrdered iterates over the map with keys sorted in ascending order
// calling forEach for each key-value pair as long as forEach does not return an error.
func IterateMapOrdered[K cmp.Ordered, V any](m map[K]V, forEach func(k K, v V) error) error {
	keys := OrderedMapKeys(m)
	for _, k := range keys {
		if err := forEach(k, m[k]); err != nil {
			return err
		}
	}
	return nil
}

// OrderedMapKeys returns the map keys in ascending order.
func OrderedMapKeys[K cmp.Ordered, V any](m map[K]V) []K {
	return slices.Sorted(maps.Keys(m))
}
