package subspace

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testparams struct {
	i int64
	b bool
}

func (tp *testparams) KeyValuePairs() KeyValuePairs {
	return KeyValuePairs{
		{[]byte("i"), &tp.i},
		{[]byte("b"), &tp.b},
	}
}

func TestTypeTable(t *testing.T) {
	table := NewTypeTable()

	require.Panics(t, func() { table.RegisterType([]byte(""), nil) })
	require.Panics(t, func() { table.RegisterType([]byte("!@#$%"), nil) })
	require.Panics(t, func() { table.RegisterType([]byte("hello,"), nil) })
	require.Panics(t, func() { table.RegisterType([]byte("hello"), nil) })

	require.NotPanics(t, func() { table.RegisterType([]byte("hello"), bool(false)) })
	require.NotPanics(t, func() { table.RegisterType([]byte("world"), int64(0)) })
	require.Panics(t, func() { table.RegisterType([]byte("hello"), bool(false)) })

	require.NotPanics(t, func() { table.RegisterParamSet(&testparams{}) })
	require.Panics(t, func() { table.RegisterParamSet(&testparams{}) })
}
