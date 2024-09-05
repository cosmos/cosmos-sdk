package branch

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_memIterator(t *testing.T) {
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

		require.False(t, it.Valid())
	})
}
