package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAppendEvents(t *testing.T) {
	e1 := NewEvent("transfer", NewAttribute("sender", "foo"))
	e2 := NewEvent("transfer", NewAttribute("sender", "bar"))
	a := Events{e1}
	b := Events{e2}
	c := a.AppendEvents(b)
	require.Equal(t, c, Events{e1, e2})
	require.Equal(t, c, Events{e1}.AppendEvent("transfer", NewAttribute("sender", "bar")))
	require.Equal(t, c, Events{e1}.AppendEvents(Events{e2}))
}

func TestAppendAttributes(t *testing.T) {
	e := NewEvent("transfer", NewAttribute("sender", "foo"))
	e = e.AppendAttributes(NewAttribute("recipient", "bar"))
	require.Len(t, e.Attributes, 2)
	require.Equal(t, e, NewEvent("transfer", NewAttribute("sender", "foo"), NewAttribute("recipient", "bar")))
}

func TestEmptyEvents(t *testing.T) {
	require.Equal(t, EmptyEvents(), Events{})
}

func TestAttributeString(t *testing.T) {
	require.Equal(t, "foo: bar", NewAttribute("foo", "bar").String())
}
