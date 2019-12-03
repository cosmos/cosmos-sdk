package subspace

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testparams struct {
	i int64
	b bool
}

type testParam struct {
	key []byte
}

func (tp testParam) Key() []byte     { return tp.key }
func (tp testParam) Validate() error { return nil }

func (tp *testparams) ParamSetPairs() ParamSetPairs {
	return ParamSetPairs{
		{testParam{[]byte("i")}, &tp.i},
		{testParam{[]byte("b")}, &tp.b},
	}
}

func TestKeyTable(t *testing.T) {
	table := NewKeyTable()

	require.Panics(t, func() { table.RegisterType(ParamSetPair{testParam{[]byte("")}, nil}) })
	require.Panics(t, func() { table.RegisterType(ParamSetPair{testParam{[]byte("!@#$%")}, nil}) })
	require.Panics(t, func() { table.RegisterType(ParamSetPair{testParam{[]byte("hello,")}, nil}) })
	require.Panics(t, func() { table.RegisterType(ParamSetPair{testParam{[]byte("hello")}, nil}) })

	require.NotPanics(t, func() { table.RegisterType(ParamSetPair{testParam{[]byte("hello")}, bool(false)}) })
	require.NotPanics(t, func() { table.RegisterType(ParamSetPair{testParam{[]byte("world")}, int64(0)}) })
	require.Panics(t, func() { table.RegisterType(ParamSetPair{testParam{[]byte("hello")}, bool(false)}) })

	require.NotPanics(t, func() { table.RegisterParamSet(&testparams{}) })
	require.Panics(t, func() { table.RegisterParamSet(&testparams{}) })
}
