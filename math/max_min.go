package math

import "golang.org/x/exp/constraints"

func Max[T constraints.Ordered](a, b T, rest ...T) T {
	maximum := a
	if b > a {
		maximum = b
	}
	for _, val := range rest {
		if val > maximum {
			maximum = val
		}
	}
	return maximum
}

func Min[T constraints.Ordered](a, b T, rest ...T) T {
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
