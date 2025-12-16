package math

import "cmp"

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
