package simsx

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCollect tests the Collect function, which applies a transformation function to each element of a slice.
// It checks if the transformation is applied correctly and returns the expected results.
func TestCollect(t *testing.T) {
	src := []int{1, 2, 3}
	got := Collect(src, func(a int) int { return a * 2 })
	assert.Equal(t, []int{2, 4, 6}, got)
	gotStrings := Collect(src, strconv.Itoa)
	assert.Equal(t, []string{"1", "2", "3"}, gotStrings)
}

// TestFirst tests the First function, which finds the first element in a slice that satisfies a given predicate.
// It checks if the correct element is returned and if nil is returned when no element satisfies the predicate.
func TestFirst(t *testing.T) {
	src := []string{"a", "b"}
	assert.Equal(t, &src[1], First(src, func(a string) bool { return a == "b" }))
	assert.Nil(t, First(src, func(a string) bool { return false }))
}

// TestOneOf tests the OneOf function, which selects a random element from a slice using a mock random number generator.
// It checks if the correct element is selected based on the mock's configuration.
func TestOneOf(t *testing.T) {
	src := []string{"a", "b"}
	got := OneOf(randMock{next: 1}, src)
	assert.Equal(t, "b", got)
	// test with other type
	src2 := []int{1, 2, 3}
	got2 := OneOf(randMock{next: 2}, src2)
	assert.Equal(t, 3, got2)
}

type randMock struct {
	next int
}

// Intn returns the fixed value specified in the `next` field of the randMock struct.
func (x randMock) Intn(n int) int {
	return x.next
}
