package collections

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNameRegex(t *testing.T) {
	require.Regexp(t, nameRegex, "a")
	require.Regexp(t, nameRegex, "ABC")
	require.Regexp(t, nameRegex, "foo1_xyz")
	require.NotRegexp(t, nameRegex, "1foo")
	require.NotRegexp(t, nameRegex, "_bar")
	require.NotRegexp(t, nameRegex, "abc-xyz")
}

func TestAddCollection(t *testing.T) {
	require.NotPanics(t, func() {
		schema := NewSchema(storetypes.NewKVStoreKey("test"))
		NewMap(schema, NewPrefix(1), "abc", Uint64Key, Uint64Value)
		NewMap(schema, NewPrefix(2), "def", Uint64Key, Uint64Value)
	})

	require.PanicsWithError(t, "name must match regex [A-Za-z][A-Za-z0-9_]*, got 123", func() {
		schema := NewSchema(storetypes.NewKVStoreKey("test"))
		NewMap(schema, NewPrefix(1), "123", Uint64Key, Uint64Value)
	})

	require.PanicsWithError(t, "prefix [1] already taken within schema", func() {
		schema := NewSchema(storetypes.NewKVStoreKey("test"))
		NewMap(schema, NewPrefix(1), "abc", Uint64Key, Uint64Value)
		NewMap(schema, NewPrefix(1), "def", Uint64Key, Uint64Value)
	})

	require.PanicsWithError(t, "name abc already taken within schema", func() {
		schema := NewSchema(storetypes.NewKVStoreKey("test"))
		NewMap(schema, NewPrefix(1), "abc", Uint64Key, Uint64Value)
		NewMap(schema, NewPrefix(2), "abc", Uint64Key, Uint64Value)
	})
}
