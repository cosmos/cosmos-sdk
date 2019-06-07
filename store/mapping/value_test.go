package mapping

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

type valuepair struct {
	key   []byte
	value test
}

func TestValue(t *testing.T) {
	key, ctx, cdc := defaultComponents()
	base := NewBase(cdc, key)

	cases := make([]valuepair, testsize)
	for i := range cases {
		cases[i].key = make([]byte, 20)
		rand.Read(cases[i].key)
		cases[i].value = newtest()
		NewValue(base, cases[i].key).Set(ctx, cases[i].value)
	}

	for i := range cases {
		var val test
		NewValue(base, cases[i].key).Get(ctx, &val)
		require.Equal(t, cases[i].value, val)
	}
}
