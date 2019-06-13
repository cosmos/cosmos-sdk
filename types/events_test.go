package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAppendEvents(t *testing.T) {
	e1 := NewEvent("transfer", Attribute{"sender", "foo"})
	e2 := NewEvent("transfer", Attribute{"sender", "bar"})
	a := Events{e1}
	b := Events{e2}
	c := a.AppendEvents(b)
	require.Equal(t, c, Events{e1, e2})
	require.Equal(t, c, Events{e1}.AppendEvent("transfer", Attribute{"sender", "bar"}))
	require.Equal(t, c, Events{e1}.AppendEvents(Events{e2}))
}

func TestEmptyEvents(t *testing.T) {
	require.Equal(t, EmptyEvents(), Events{})
}
