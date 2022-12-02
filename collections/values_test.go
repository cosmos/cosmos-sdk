package collections

import "testing"

func TestUint64Value(t *testing.T) {
	t.Run("bijective", func(t *testing.T) {
		assertValueBijective(t, Uint64Value, 555)
	})
}
