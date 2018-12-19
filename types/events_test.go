package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEvents(t *testing.T) {
	e := NewEventManager()

	resTags := e.Collect()
	require.Empty(t, resTags)

	tags := NewTags(
		"action", []byte("bond"),
		"action", []byte("unbond"),
		"delegator", []byte("cosmos00000"),
		"delegator", []byte("cosmos11111"),
		"recipient", []byte("cosmos00000"),
	)
	e.Event(tags)

	resTags = e.ToTags()
	require.Len(t, resTags, len(tags))
}
