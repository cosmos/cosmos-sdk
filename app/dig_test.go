package app

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/dig"
)

type Inputs struct {
	dig.In

	Ints   []int     `group:"abc"`
	Floats []float64 `group:"abc"`
}

func TestDig(t *testing.T) {
	ctr := dig.New()
	require.NoError(t, ctr.Provide(func() int {
		return 4
	}, dig.Group("abc")))
	require.NoError(t, ctr.Provide(func() int {
		return 5
	}, dig.Group("abc")))
	require.NoError(t, ctr.Provide(func() float64 {
		return 7
	}, dig.Group("abc")))
	require.NoError(t, ctr.Provide(func() float64 {
		return 10
	}, dig.Group("abc")))
	require.NoError(t, ctr.Invoke(func(in Inputs) {
		t.Logf("%+v\n", in)
	}))
}
