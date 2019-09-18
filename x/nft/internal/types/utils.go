package types

import "strings"

// Findable is an interface for iterable types that allows the FindUtil function to work
type Findable interface {
	ElAtIndex(index int) string
	Len() int
}

// FindUtil is a find funcion for types that support the Findable interface
func FindUtil(group Findable, el string) int {
	if group.Len() == 0 {
		return -1
	}
	low := 0
	high := group.Len() - 1
	median := 0
	for low <= high {
		median = (low + high) / 2
		compare := strings.Compare(group.ElAtIndex(median), el)
		if compare == 0 {
			// if sa[median] == el
			return median
		} else if compare == -1 {
			// if sa[median] < el
			low = median + 1
		} else {
			// if sa[median] > el
			high = median - 1
		}
	}
	return -1
}
