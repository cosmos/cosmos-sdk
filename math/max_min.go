package math

import "golang.org/x/exp/constraints"

func Maximum[T constraints.Ordered](a, b T, rest ...T) T {
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

func Minimum[T constraints.Ordered](a, b T, rest ...T) T {
	minimum := a
	if b < a {
		minimum = b
	}
	for _, val := range rest {
		if val < minimum {
			minimum = val
		}
	}
	return minimum
}
