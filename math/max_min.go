package math

import "cmp"

// Max returns the maximum value among the provided arguments.
// It accepts two required arguments and any number of additional values.
// For empty or single-value cases, returns the first argument.
// Uses Go's generic constraints to work with any ordered type.
func Max[T cmp.Ordered](a, b T, rest ...T) T {
	max := a
	if b > a {
		max = b
	}
	for _, val := range rest {
		if val > max {
			max = val
		}
	}
	return max
}

// Min returns the minimum value among the provided arguments.
// It accepts two required arguments and any number of additional values.
// For empty or single-value cases, returns the first argument.
// Uses Go's generic constraints to work with any ordered type.
func Min[T cmp.Ordered](a, b T, rest ...T) T {
	min := a
	if b < a {
		min = b
	}
	for _, val := range rest {
		if val < min {
			min = val
		}
	}
	return min
}
